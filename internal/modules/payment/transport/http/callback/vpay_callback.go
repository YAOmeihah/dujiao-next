package paymentcallbackhttp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dujiao-next/internal/constants"
	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func (h *Handler) handleVpayCallback(c *gin.Context) bool {
	log := ginutil.RequestLog(c)
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

	log.Infow("vpay_callback_received", "client_ip", c.ClientIP(), "pay_id", payID, "param", param,
		"type", payType, "price", price, "really_price", reallyPrice)
	payment, err := h.payments.GetByGatewayOrderNo(payID)
	if err != nil || payment == nil {
		log.Warnw("vpay_callback_payment_not_found", "pay_id", payID, "error", err)
		c.String(http.StatusOK, constants.VpayCallbackFail)
		return true
	}
	channel, err := h.channels.GetByID(payment.ChannelID)
	if err != nil || channel == nil {
		log.Warnw("vpay_callback_channel_not_found", "payment_id", payment.ID, "channel_id", payment.ChannelID, "error", err)
		c.String(http.StatusOK, constants.VpayCallbackFail)
		return true
	}
	if !strings.EqualFold(strings.TrimSpace(channel.ProviderType), constants.PaymentProviderVpay) {
		log.Warnw("vpay_callback_provider_invalid", "payment_id", payment.ID, "channel_id", channel.ID, "provider_type", channel.ProviderType)
		c.String(http.StatusOK, constants.VpayCallbackFail)
		return true
	}
	if err := verifyVpayCallbackPayment(payment, price, reallyPrice); err != nil {
		log.Warnw("vpay_callback_payment_verify_failed", "payment_id", payment.ID, "channel_id", channel.ID, "error", err)
		h.enqueuePaymentExceptionAlert(c, jsonmap.JSON{
			"alert_type": "vpay_callback_verify_failed", "alert_level": "error",
			"payment_id": fmt.Sprintf("%d", payment.ID), "message": strings.TrimSpace(err.Error()),
			"provider": constants.PaymentProviderVpay,
		})
		c.String(http.StatusOK, constants.VpayCallbackFail)
		return true
	}

	updated, err := h.service.HandleSyncCallback(channel, form, nil)
	if err != nil {
		log.Warnw("vpay_callback_handle_failed", "payment_id", payment.ID, "channel_id", channel.ID, "pay_id", payID, "error", err)
		h.enqueuePaymentExceptionAlert(c, jsonmap.JSON{
			"alert_type": "vpay_callback_handle_failed", "alert_level": "error",
			"payment_id": fmt.Sprintf("%d", payment.ID), "order_no": payID,
			"message": strings.TrimSpace(err.Error()), "provider": constants.PaymentProviderVpay,
		})
		c.String(http.StatusOK, constants.VpayCallbackFail)
		return true
	}
	log.Infow("vpay_callback_processed", "payment_id", payment.ID, "channel_id", channel.ID, "pay_id", payID, "status", updated.Status)
	c.String(http.StatusOK, constants.VpayCallbackSuccess)
	return true
}

func verifyVpayCallbackPayment(payment *paymentdomain.Payment, price, reallyPrice string) error {
	if payment == nil {
		return fmt.Errorf("payment not found")
	}
	priceAmount, err := decimal.NewFromString(strings.TrimSpace(price))
	if err != nil || priceAmount.Round(2).Cmp(payment.Amount.Decimal.Round(2)) != 0 {
		return fmt.Errorf("payment amount mismatch")
	}
	expectedReallyPrice := vpayProviderPayloadValue(payment.ProviderPayload, "really_price", "reallyPrice")
	if expectedReallyPrice != "" && expectedReallyPrice != strings.TrimSpace(reallyPrice) {
		return fmt.Errorf("payment amount mismatch")
	}
	return nil
}

func vpayProviderPayloadValue(payload jsonmap.JSON, keys ...string) string {
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
