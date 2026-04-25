package public

import (
	"fmt"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/vpay"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func (h *Handler) HandleVpayCallback(c *gin.Context) bool {
	log := shared.RequestLog(c)
	form, err := parseCallbackForm(c)
	if err != nil {
		log.Warnw("vpay_callback_form_parse_failed", "error", err)
		return false
	}
	payID := strings.TrimSpace(getFirstValue(form, "payId"))
	param := strings.TrimSpace(getFirstValue(form, "param"))
	payType := strings.TrimSpace(getFirstValue(form, "type"))
	price := strings.TrimSpace(getFirstValue(form, "price"))
	reallyPrice := strings.TrimSpace(getFirstValue(form, "reallyPrice"))
	sign := strings.TrimSpace(getFirstValue(form, "sign"))
	if payID == "" || payType == "" || price == "" || reallyPrice == "" || sign == "" {
		log.Debugw("vpay_callback_not_matched", "reason", "missing_required_fields")
		return false
	}

	log.Infow("vpay_callback_received",
		"client_ip", c.ClientIP(),
		"pay_id", payID,
		"param", param,
		"type", payType,
		"price", price,
		"really_price", reallyPrice,
		"raw_form", callbackRawFormForLog(form),
	)

	payment, err := h.PaymentRepo.GetByGatewayOrderNo(payID)
	if err != nil || payment == nil {
		log.Warnw("vpay_callback_payment_not_found", "pay_id", payID, "error", err)
		c.String(200, constants.VpayCallbackFail)
		return true
	}
	channel, err := h.PaymentChannelRepo.GetByID(payment.ChannelID)
	if err != nil || channel == nil {
		log.Warnw("vpay_callback_channel_not_found",
			"payment_id", payment.ID,
			"channel_id", payment.ChannelID,
			"error", err,
		)
		c.String(200, constants.VpayCallbackFail)
		return true
	}
	if strings.ToLower(strings.TrimSpace(channel.ProviderType)) != constants.PaymentProviderVpay {
		log.Warnw("vpay_callback_provider_invalid",
			"payment_id", payment.ID,
			"channel_id", channel.ID,
			"provider_type", channel.ProviderType,
		)
		c.String(200, constants.VpayCallbackFail)
		return true
	}
	cfg, err := vpay.ParseConfig(channel.ConfigJSON)
	if err != nil {
		log.Warnw("vpay_callback_config_parse_failed", "payment_id", payment.ID, "channel_id", channel.ID, "error", err)
		c.String(200, constants.VpayCallbackFail)
		return true
	}
	if err := vpay.ValidateConfig(cfg); err != nil {
		log.Warnw("vpay_callback_config_invalid", "payment_id", payment.ID, "channel_id", channel.ID, "error", err)
		c.String(200, constants.VpayCallbackFail)
		return true
	}
	if err := vpay.VerifyCallback(cfg, form); err != nil {
		log.Warnw("vpay_callback_signature_invalid", "payment_id", payment.ID, "channel_id", channel.ID, "error", err)
		h.enqueuePaymentExceptionAlert(c, models.JSON{
			"alert_type":  "vpay_signature_invalid",
			"alert_level": "error",
			"payment_id":  fmt.Sprintf("%d", payment.ID),
			"message":     strings.TrimSpace(err.Error()),
			"provider":    constants.PaymentProviderVpay,
		})
		c.String(200, constants.VpayCallbackFail)
		return true
	}
	if err := verifyVpayCallbackPayment(payment, price, reallyPrice); err != nil {
		log.Warnw("vpay_callback_payment_verify_failed", "payment_id", payment.ID, "channel_id", channel.ID, "error", err)
		h.enqueuePaymentExceptionAlert(c, models.JSON{
			"alert_type":  "vpay_callback_verify_failed",
			"alert_level": "error",
			"payment_id":  fmt.Sprintf("%d", payment.ID),
			"message":     strings.TrimSpace(err.Error()),
			"provider":    constants.PaymentProviderVpay,
		})
		c.String(200, constants.VpayCallbackFail)
		return true
	}

	amount := models.Money{}
	parsedAmount, parseErr := decimal.NewFromString(price)
	if parseErr != nil {
		log.Warnw("vpay_callback_amount_parse_failed", "payment_id", payment.ID, "price", price, "error", parseErr)
		c.String(200, constants.VpayCallbackFail)
		return true
	}
	amount = models.NewMoneyFromDecimal(parsedAmount)
	payload := make(map[string]interface{}, len(form))
	for key, values := range form {
		if len(values) > 0 {
			payload[key] = values[0]
		}
	}
	callbackInput := service.PaymentCallbackInput{
		PaymentID:   payment.ID,
		OrderNo:     payID,
		ChannelID:   channel.ID,
		Status:      constants.PaymentStatusSuccess,
		ProviderRef: strings.TrimSpace(payment.ProviderRef),
		Amount:      amount,
		Currency:    strings.TrimSpace(payment.Currency),
		Payload:     models.JSON(payload),
	}

	updated, err := h.PaymentService.HandleCallback(callbackInput)
	if err != nil {
		log.Warnw("vpay_callback_handle_failed",
			"payment_id", payment.ID,
			"channel_id", channel.ID,
			"order_no", callbackInput.OrderNo,
			"error", err,
		)
		h.enqueuePaymentExceptionAlert(c, models.JSON{
			"alert_type":  "vpay_callback_handle_failed",
			"alert_level": "error",
			"payment_id":  fmt.Sprintf("%d", payment.ID),
			"order_no":    strings.TrimSpace(callbackInput.OrderNo),
			"message":     strings.TrimSpace(err.Error()),
			"provider":    constants.PaymentProviderVpay,
		})
		c.String(200, constants.VpayCallbackFail)
		return true
	}
	log.Infow("vpay_callback_processed",
		"payment_id", payment.ID,
		"channel_id", channel.ID,
		"order_no", callbackInput.OrderNo,
		"status", updated.Status,
	)
	c.String(200, constants.VpayCallbackSuccess)
	return true
}

func verifyVpayCallbackPayment(payment *models.Payment, price string, reallyPrice string) error {
	if payment == nil {
		return service.ErrPaymentNotFound
	}
	priceAmount, err := decimal.NewFromString(strings.TrimSpace(price))
	if err != nil {
		return service.ErrPaymentInvalid
	}
	if priceAmount.Round(2).Cmp(payment.Amount.Decimal.Round(2)) != 0 {
		return service.ErrPaymentAmountMismatch
	}
	expectedReallyPrice := strings.TrimSpace(vpayProviderPayloadValue(payment.ProviderPayload, "really_price", "reallyPrice"))
	if expectedReallyPrice != "" && expectedReallyPrice != strings.TrimSpace(reallyPrice) {
		return service.ErrPaymentAmountMismatch
	}
	return nil
}

func vpayProviderPayloadValue(payload models.JSON, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok && value != nil {
			return strings.TrimSpace(fmt.Sprintf("%v", value))
		}
	}
	if data, ok := payload["data"].(map[string]interface{}); ok {
		for _, key := range keys {
			if value, exists := data[key]; exists && value != nil {
				return strings.TrimSpace(fmt.Sprintf("%v", value))
			}
		}
	}
	return ""
}
