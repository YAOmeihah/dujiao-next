package vpay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/constants"
)

func TestCreatePaymentPostsSignedFormAndBuildsRedirectURL(t *testing.T) {
	var gotForm url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/createOrder" {
			t.Fatalf("path = %s, want /createOrder", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		gotForm = r.PostForm
		if got := gotForm.Get("payId"); got != "DJP1001" {
			t.Fatalf("payId = %s", got)
		}
		if got := gotForm.Get("type"); got != "1" {
			t.Fatalf("type = %s", got)
		}
		if got := gotForm.Get("price"); got != "99.00" {
			t.Fatalf("price = %s", got)
		}
		if got := gotForm.Get("param"); got != "DJORDER1001" {
			t.Fatalf("param = %s", got)
		}
		if got := gotForm.Get("notifyUrl"); got != "https://api.example.com/api/v1/payments/callback" {
			t.Fatalf("notifyUrl = %s", got)
		}
		if got := gotForm.Get("returnUrl"); !strings.HasPrefix(got, "https://shop.example.com/pay?order_no=DJORDER1001") {
			t.Fatalf("returnUrl = %s", got)
		}
		if got := gotForm.Get("sign"); got != signCreateOrder("DJP1001", "DJORDER1001", "1", "99.00", "secret-key") {
			t.Fatalf("sign = %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"msg":"成功","data":{"payId":"DJP1001","orderId":"VP202604250001","payType":1,"price":"99.00","reallyPrice":"99.01","payUrl":"weixin://pay","state":0}}`))
	}))
	defer server.Close()

	cfg := &Config{
		GatewayURL: server.URL,
		SignKey:    "secret-key",
		NotifyURL:  "https://api.example.com/api/v1/payments/callback",
		ReturnURL:  "https://shop.example.com/pay?order_no=DJORDER1001",
	}
	cfg.Normalize()

	result, err := CreatePayment(context.Background(), cfg, CreateInput{
		PayID:       "DJP1001",
		Param:       "DJORDER1001",
		ChannelType: constants.PaymentChannelTypeWechat,
		Price:       "99.00",
		NotifyURL:   cfg.NotifyURL,
		ReturnURL:   cfg.ReturnURL,
	})
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	if result.OrderID != "VP202604250001" {
		t.Fatalf("order id = %s", result.OrderID)
	}
	if result.ReallyPrice != "99.01" {
		t.Fatalf("really price = %s", result.ReallyPrice)
	}
	redirect, err := url.Parse(result.RedirectURL)
	if err != nil {
		t.Fatalf("parse redirect url failed: %v", err)
	}
	if redirect.Scheme+"://"+redirect.Host != server.URL {
		t.Fatalf("redirect host = %s://%s, want %s", redirect.Scheme, redirect.Host, server.URL)
	}
	if redirect.Path != "/payPage/pay.html" {
		t.Fatalf("redirect path = %s", redirect.Path)
	}
	if got := redirect.Query().Get("orderId"); got != "VP202604250001" {
		t.Fatalf("redirect orderId = %s", got)
	}
	if result.Raw == nil || result.Raw["code"] != float64(1) {
		t.Fatalf("raw response should be recorded, got %#v", result.Raw)
	}
}

func TestVerifyCallbackUsesReceivedFieldText(t *testing.T) {
	cfg := &Config{SignKey: "secret-key"}
	form := map[string][]string{
		"payId":       {"DJP1001"},
		"param":       {"DJORDER1001"},
		"type":        {"1"},
		"price":       {"99.00"},
		"reallyPrice": {"99.01"},
	}
	form["sign"] = []string{signCallback("DJP1001", "DJORDER1001", "1", "99.00", "99.01", "secret-key")}

	if err := VerifyCallback(cfg, form); err != nil {
		t.Fatalf("VerifyCallback failed: %v", err)
	}

	form["price"] = []string{"99.0"}
	if err := VerifyCallback(cfg, form); err == nil {
		t.Fatalf("VerifyCallback should fail when the received field text changes")
	}
}

func TestResolvePayType(t *testing.T) {
	tests := []struct {
		channelType string
		want        string
	}{
		{constants.PaymentChannelTypeWechat, "1"},
		{constants.PaymentChannelTypeWxpay, "1"},
		{constants.PaymentChannelTypeAlipay, "2"},
		{constants.PaymentChannelTypeQqpay, ""},
	}
	for _, tc := range tests {
		if got := resolvePayType(tc.channelType); got != tc.want {
			t.Fatalf("resolvePayType(%s) = %s, want %s", tc.channelType, got, tc.want)
		}
	}
}
