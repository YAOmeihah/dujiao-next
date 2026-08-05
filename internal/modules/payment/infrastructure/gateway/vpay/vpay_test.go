package vpay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
		if got := gotForm.Get("sign"); got != signCreateOrder("DJP1001", "DJORDER1001", "1", "99.00", "secret-key") {
			t.Fatalf("sign = %s", got)
		}
		_, _ = w.Write([]byte(`{"code":1,"msg":"ok","data":{"payId":"DJP1001","orderId":"VP202604250001","payType":1,"price":"99.00","reallyPrice":"99.01","payUrl":"weixin://pay","state":0}}`))
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
	})
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	if result.OrderID != "VP202604250001" || result.ReallyPrice != "99.01" {
		t.Fatalf("unexpected result: %+v", result)
	}
	redirect, err := url.Parse(result.RedirectURL)
	if err != nil {
		t.Fatalf("parse redirect url failed: %v", err)
	}
	if redirect.Path != "/payPage/pay.html" || redirect.Query().Get("orderId") != "VP202604250001" {
		t.Fatalf("unexpected redirect url: %s", result.RedirectURL)
	}
}

func TestCreatePaymentPostsHMACSignType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if got := r.PostForm.Get("signType"); got != "HMAC_SHA256" {
			t.Fatalf("signType = %s", got)
		}
		wantSign := expectedHMACSHA256Hex("DJP1002"+"DJORDER1002"+"1"+"88.00", "secret-key")
		if got := r.PostForm.Get("sign"); got != wantSign {
			t.Fatalf("sign = %s, want %s", got, wantSign)
		}
		_, _ = w.Write([]byte(`{"code":1,"msg":"ok","data":{"payId":"DJP1002","orderId":"VP202604250002"}}`))
	}))
	defer server.Close()

	cfg := &Config{GatewayURL: server.URL, SignKey: "secret-key", NotifyURL: "https://api.example.com/callback", ReturnURL: "https://shop.example.com/pay", SignType: "HMAC_SHA256"}
	cfg.Normalize()
	if _, err := CreatePayment(context.Background(), cfg, CreateInput{PayID: "DJP1002", Param: "DJORDER1002", ChannelType: constants.PaymentChannelTypeWechat, Price: "88.00"}); err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
}

func TestVerifyCallbackUsesReceivedFieldText(t *testing.T) {
	cfg := &Config{SignKey: "secret-key"}
	form := map[string][]string{
		"payId": {"DJP1001"}, "param": {"DJORDER1001"}, "type": {"1"},
		"price": {"99.00"}, "reallyPrice": {"99.01"},
	}
	form["sign"] = []string{signCallback("DJP1001", "DJORDER1001", "1", "99.00", "99.01", "secret-key")}
	if err := VerifyCallback(cfg, form); err != nil {
		t.Fatalf("VerifyCallback failed: %v", err)
	}
	form["price"] = []string{"99.0"}
	if err := VerifyCallback(cfg, form); err == nil {
		t.Fatal("VerifyCallback should fail when signed text changes")
	}
}

func TestResolvePayType(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{constants.PaymentChannelTypeWechat, "1"},
		{constants.PaymentChannelTypeWxpay, "1"},
		{constants.PaymentChannelTypeAlipay, "2"},
		{constants.PaymentChannelTypeQqpay, ""},
	} {
		if got := ResolvePayType(tc.input); got != tc.want {
			t.Fatalf("ResolvePayType(%s) = %s, want %s", tc.input, got, tc.want)
		}
	}
}

func expectedHMACSHA256Hex(payload, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(payload))
	return strings.ToLower(hex.EncodeToString(mac.Sum(nil)))
}
