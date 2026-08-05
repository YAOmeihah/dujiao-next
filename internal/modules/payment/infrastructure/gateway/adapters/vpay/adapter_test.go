package vpayadapter

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/constants"
	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/shopspring/decimal"
)

func TestVpayAdapterCreatesRedirectPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.PostForm.Get("payId") != "PAY-1001" || r.PostForm.Get("param") != "ORDER-1001" {
			t.Fatalf("unexpected vpay identifiers: %v", r.PostForm)
		}
		_, _ = w.Write([]byte(`{"code":1,"msg":"ok","data":{"payId":"PAY-1001","orderId":"VPAY-1001","price":"99.00","reallyPrice":"99.01"}}`))
	}))
	defer server.Close()

	adapter := NewVpayAdapter()
	result, err := adapter.CreatePayment(context.Background(), jsonmap.JSON{
		"gateway_url": server.URL,
		"sign_key":    "secret-key",
		"notify_url":  "https://api.example.com/api/v1/payments/callback",
		"return_url":  "https://shop.example.com/pay",
	}, paymentcontract.GatewayCreateInput{
		OrderNo:     "PAY-1001",
		ChannelType: constants.PaymentChannelTypeAlipay,
		Amount:      money.FromDecimal(decimal.RequireFromString("99.00")),
		ReturnURLQuery: map[string]string{
			"order_no":    "ORDER-1001",
			"vpay_return": "1",
		},
		Extra: map[string]interface{}{"interaction_mode": constants.PaymentInteractionRedirect},
	})
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	if result.ProviderRef != "VPAY-1001" || !strings.Contains(result.RedirectURL, "orderId=VPAY-1001") {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.QRCodeURL != "" || result.Payload["really_price"] != "99.01" {
		t.Fatalf("unexpected VPay payload: %+v", result)
	}
}

func TestVpayAdapterVerifiesCallback(t *testing.T) {
	form := map[string][]string{
		"payId": {"PAY-1001"}, "param": {"ORDER-1001"}, "type": {"2"},
		"price": {"99.00"}, "reallyPrice": {"99.01"},
	}
	sum := md5.Sum([]byte("PAY-1001" + "ORDER-1001" + "2" + "99.00" + "99.01" + "secret-key"))
	form["sign"] = []string{strings.ToLower(hex.EncodeToString(sum[:]))}

	result, err := NewVpayAdapter().(paymentcontract.GatewayCallbackVerifier).VerifyCallback(jsonmap.JSON{
		"gateway_url": "https://pay.example.com",
		"sign_key":    "secret-key",
		"notify_url":  "https://api.example.com/callback",
		"return_url":  "https://shop.example.com/pay",
	}, form, nil)
	if err != nil {
		t.Fatalf("VerifyCallback failed: %v", err)
	}
	if result.OrderNo != "PAY-1001" || result.Status != constants.PaymentStatusSuccess || result.Amount.String() != "99.00" {
		t.Fatalf("unexpected callback result: %+v", result)
	}
}
