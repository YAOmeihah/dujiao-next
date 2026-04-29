package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
)

func newCapTestService(endpoint string) *CaptchaService {
	svc := NewCaptchaService(nil, config.CaptchaConfig{
		Provider: constants.CaptchaProviderCap,
		Scenes:   config.CaptchaSceneConfig{Login: true},
		Cap: config.CaptchaCapConfig{
			Endpoint:  endpoint,
			SiteKey:   "site",
			SecretKey: "secret",
			TimeoutMS: 1000,
		},
	})
	svc.cacheTTL = time.Millisecond
	return svc
}

func TestCaptchaServiceVerifyCapSuccess(t *testing.T) {
	var gotSecret string
	var gotResponse string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site/siteverify" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var body struct {
			Secret   string `json:"secret"`
			Response string `json:"response"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		gotSecret = body.Secret
		gotResponse = body.Response
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer server.Close()

	err := newCapTestService(server.URL).Verify(constants.CaptchaSceneLogin, CaptchaVerifyPayload{CapToken: "cap-token"}, "127.0.0.1")
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if gotSecret != "secret" || gotResponse != "cap-token" {
		t.Fatalf("secret/response = %q/%q", gotSecret, gotResponse)
	}
}

func TestCaptchaServiceVerifyCapRequired(t *testing.T) {
	err := newCapTestService("http://127.0.0.1:1").Verify(constants.CaptchaSceneLogin, CaptchaVerifyPayload{}, "")
	if err != ErrCaptchaRequired {
		t.Fatalf("err = %v, want ErrCaptchaRequired", err)
	}
}

func TestCaptchaServiceVerifyCapInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": false})
	}))
	defer server.Close()

	err := newCapTestService(server.URL).Verify(constants.CaptchaSceneLogin, CaptchaVerifyPayload{CapToken: "bad"}, "")
	if err != ErrCaptchaInvalid {
		t.Fatalf("err = %v, want ErrCaptchaInvalid", err)
	}
}

func TestCaptchaServiceVerifyCapDecodeFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{"))
	}))
	defer server.Close()

	err := newCapTestService(server.URL).Verify(constants.CaptchaSceneLogin, CaptchaVerifyPayload{CapToken: "token"}, "")
	if !errors.Is(err, ErrCaptchaVerifyFailed) {
		t.Fatalf("err = %v, want ErrCaptchaVerifyFailed", err)
	}
}
