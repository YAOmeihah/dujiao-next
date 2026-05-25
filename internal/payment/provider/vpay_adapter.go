package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/vpay"

	"github.com/shopspring/decimal"
)

// vpayAdapter keeps the local VPay integration on the upstream Provider registry path.
type vpayAdapter struct{}

func NewVpayAdapter() Provider { return &vpayAdapter{} }

var (
	_ Provider         = (*vpayAdapter)(nil)
	_ CallbackVerifier = (*vpayAdapter)(nil)
)

func (a *vpayAdapter) Type() string {
	return constants.PaymentProviderVpay + ":"
}

func (a *vpayAdapter) parseConfig(raw models.JSON) (*vpay.Config, error) {
	cfg, err := vpay.ParseConfig(raw)
	if err != nil {
		return nil, mapVpayError(err)
	}
	cfg.Normalize()
	if err := vpay.ValidateConfig(cfg); err != nil {
		return nil, mapVpayError(err)
	}
	return cfg, nil
}

func (a *vpayAdapter) ValidateConfig(raw models.JSON, channelType string) error {
	if channelType != "" && !vpay.IsSupportedChannelType(channelType) {
		return fmt.Errorf("%w: vpay channel_type %s", ErrUnsupportedChannel, channelType)
	}
	_, err := a.parseConfig(raw)
	return err
}

func (a *vpayAdapter) CreatePayment(ctx context.Context, raw models.JSON, input CreateInput) (*CreateResult, error) {
	if !vpay.IsSupportedChannelType(input.ChannelType) {
		return nil, fmt.Errorf("%w: vpay channel_type %s", ErrUnsupportedChannel, input.ChannelType)
	}

	mode, _ := input.Extra["interaction_mode"].(string)
	if strings.ToLower(strings.TrimSpace(mode)) != constants.PaymentInteractionRedirect {
		return nil, fmt.Errorf("%w: vpay only supports redirect interaction_mode", ErrConfigInvalid)
	}

	cfg, err := a.parseConfig(raw)
	if err != nil {
		return nil, err
	}

	notifyURL := strings.TrimSpace(input.NotifyURL)
	if notifyURL == "" {
		notifyURL = strings.TrimSpace(cfg.NotifyURL)
	}
	returnURL := strings.TrimSpace(input.ReturnURL)
	if returnURL == "" {
		returnURL = strings.TrimSpace(cfg.ReturnURL)
	}
	returnURL = appendQueryParams(returnURL, input.ReturnURLQuery)
	param := strings.TrimSpace(input.ReturnURLQuery["order_no"])
	if param == "" {
		param = input.OrderNo
	}

	result, err := vpay.CreatePayment(ctx, cfg, vpay.CreateInput{
		PayID:       input.OrderNo,
		Param:       param,
		ChannelType: input.ChannelType,
		Price:       input.Amount.String(),
		NotifyURL:   notifyURL,
		ReturnURL:   returnURL,
	})
	if err != nil {
		return nil, mapVpayError(err)
	}

	payload := models.JSON{}
	if result.Raw != nil {
		payload = models.JSON(result.Raw)
	}
	if strings.TrimSpace(result.Price) != "" {
		payload["price"] = strings.TrimSpace(result.Price)
	}
	if strings.TrimSpace(result.ReallyPrice) != "" {
		payload["really_price"] = strings.TrimSpace(result.ReallyPrice)
	}
	if strings.TrimSpace(result.PayURL) != "" {
		payload["vpay_pay_url"] = strings.TrimSpace(result.PayURL)
	}

	return &CreateResult{
		ProviderRef: result.OrderID,
		RedirectURL: result.RedirectURL,
		QRCodeURL:   "",
		Payload:     payload,
	}, nil
}

func (a *vpayAdapter) VerifyCallback(raw models.JSON, form map[string][]string, _ []byte) (*CallbackResult, error) {
	cfg, err := a.parseConfig(raw)
	if err != nil {
		return nil, err
	}
	if err := vpay.VerifyCallback(cfg, form); err != nil {
		return nil, mapVpayError(err)
	}

	amount := models.Money{}
	if price := strings.TrimSpace(pickFormValue(form, "price")); price != "" {
		if d, parseErr := decimal.NewFromString(price); parseErr == nil {
			amount = models.NewMoneyFromDecimal(d)
		}
	}

	payload := models.JSON{}
	if raw, marshalErr := json.Marshal(formToJSON(form)); marshalErr == nil {
		var m map[string]interface{}
		if jsonErr := json.Unmarshal(raw, &m); jsonErr == nil {
			payload = models.JSON(m)
		}
	}

	return &CallbackResult{
		OrderNo:     strings.TrimSpace(pickFormValue(form, "payId")),
		ProviderRef: "",
		Status:      constants.PaymentStatusSuccess,
		Amount:      amount,
		Currency:    "CNY",
		Payload:     payload,
	}, nil
}

func mapVpayError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, vpay.ErrConfigInvalid):
		return fmt.Errorf("%w: %v", ErrConfigInvalid, err)
	case errors.Is(err, vpay.ErrChannelType):
		return fmt.Errorf("%w: %v", ErrUnsupportedChannel, err)
	case errors.Is(err, vpay.ErrRequestFailed):
		return fmt.Errorf("%w: %v", ErrRequestFailed, err)
	case errors.Is(err, vpay.ErrResponseInvalid):
		return fmt.Errorf("%w: %v", ErrResponseInvalid, err)
	case errors.Is(err, vpay.ErrSignatureInvalid):
		return fmt.Errorf("%w: %v", ErrSignatureInvalid, err)
	default:
		return err
	}
}
