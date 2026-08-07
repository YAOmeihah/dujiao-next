package vpayadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dujiao-next/internal/constants"
	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"
	gatewaycommon "github.com/dujiao-next/internal/modules/payment/infrastructure/gateway/common"
	"github.com/dujiao-next/internal/modules/payment/infrastructure/gateway/vpay"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/shopspring/decimal"
)

// vpayAdapter keeps the local VPay integration on the upstream paymentcontract.GatewayProvider registry path.
type vpayAdapter struct{}

func NewVpayAdapter() paymentcontract.GatewayProvider { return &vpayAdapter{} }

var (
	_ paymentcontract.GatewayProvider         = (*vpayAdapter)(nil)
	_ paymentcontract.GatewayCallbackVerifier = (*vpayAdapter)(nil)
)

func (a *vpayAdapter) Type() string {
	return constants.PaymentProviderVpay + ":"
}

func (a *vpayAdapter) parseConfig(raw jsonmap.JSON) (*vpay.Config, error) {
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

func (a *vpayAdapter) ValidateConfig(raw jsonmap.JSON, channelType string) error {
	if channelType != "" && !vpay.IsSupportedChannelType(channelType) {
		return fmt.Errorf("%w: vpay channel_type %s", paymentcontract.ErrGatewayUnsupportedChannel, channelType)
	}
	_, err := a.parseConfig(raw)
	return err
}

func (a *vpayAdapter) CreatePayment(ctx context.Context, raw jsonmap.JSON, input paymentcontract.GatewayCreateInput) (*paymentcontract.GatewayCreateResult, error) {
	if strings.ToUpper(strings.TrimSpace(input.Currency)) != "CNY" {
		return nil, fmt.Errorf("%w: vpay only supports CNY payments", paymentcontract.ErrGatewayConfigInvalid)
	}
	if !vpay.IsSupportedChannelType(input.ChannelType) {
		return nil, fmt.Errorf("%w: vpay channel_type %s", paymentcontract.ErrGatewayUnsupportedChannel, input.ChannelType)
	}

	mode, _ := input.Extra["interaction_mode"].(string)
	if strings.ToLower(strings.TrimSpace(mode)) != constants.PaymentInteractionRedirect {
		return nil, fmt.Errorf("%w: vpay only supports redirect interaction_mode", paymentcontract.ErrGatewayConfigInvalid)
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
	returnURL = gatewaycommon.AppendQueryParams(returnURL, input.ReturnURLQuery)
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

	payload := jsonmap.JSON{}
	if result.Raw != nil {
		payload = jsonmap.JSON(result.Raw)
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

	return &paymentcontract.GatewayCreateResult{
		ProviderRef: result.OrderID,
		RedirectURL: result.RedirectURL,
		QRCodeURL:   "",
		Payload:     payload,
	}, nil
}

func (a *vpayAdapter) VerifyCallback(raw jsonmap.JSON, form map[string][]string, _ []byte) (*paymentcontract.GatewayCallbackResult, error) {
	cfg, err := a.parseConfig(raw)
	if err != nil {
		return nil, err
	}
	if err := vpay.VerifyCallback(cfg, form); err != nil {
		return nil, mapVpayError(err)
	}

	amount := money.Amount{}
	if price := strings.TrimSpace(gatewaycommon.FormValue(form, "price")); price != "" {
		if d, parseErr := decimal.NewFromString(price); parseErr == nil {
			amount = money.FromDecimal(d)
		}
	}

	payload := jsonmap.JSON{}
	if raw, marshalErr := json.Marshal(gatewaycommon.FormToJSON(form)); marshalErr == nil {
		var m map[string]interface{}
		if jsonErr := json.Unmarshal(raw, &m); jsonErr == nil {
			payload = jsonmap.JSON(m)
		}
	}

	return &paymentcontract.GatewayCallbackResult{
		OrderNo:     strings.TrimSpace(gatewaycommon.FormValue(form, "payId")),
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
		return fmt.Errorf("%w: %v", paymentcontract.ErrGatewayConfigInvalid, err)
	case errors.Is(err, vpay.ErrChannelType):
		return fmt.Errorf("%w: %v", paymentcontract.ErrGatewayUnsupportedChannel, err)
	case errors.Is(err, vpay.ErrRequestFailed):
		return fmt.Errorf("%w: %v", paymentcontract.ErrGatewayRequestFailed, err)
	case errors.Is(err, vpay.ErrResponseInvalid):
		return fmt.Errorf("%w: %v", paymentcontract.ErrGatewayResponseInvalid, err)
	case errors.Is(err, vpay.ErrSignatureInvalid):
		return fmt.Errorf("%w: %v", paymentcontract.ErrGatewaySignatureInvalid, err)
	default:
		return err
	}
}
