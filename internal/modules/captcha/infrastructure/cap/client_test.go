package cap

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dujiao-next/internal/modules/captcha/contract"
	settingssecurity "github.com/dujiao-next/internal/modules/settings/schema/security"
)

func TestVerifySuccess(t *testing.T) {
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

	err := New().Verify(settingssecurity.CaptchaCapSetting{
		Endpoint: server.URL, SiteKey: "site", SecretKey: "secret", TimeoutMS: 1000,
	}, "cap-token")
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if gotSecret != "secret" || gotResponse != "cap-token" {
		t.Fatalf("secret/response = %q/%q", gotSecret, gotResponse)
	}
}

func TestVerifyInvalidAndDecodeFailure(t *testing.T) {
	invalidServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": false})
	}))
	defer invalidServer.Close()

	cfg := settingssecurity.CaptchaCapSetting{Endpoint: invalidServer.URL, SiteKey: "site", SecretKey: "secret", TimeoutMS: 1000}
	if err := New().Verify(cfg, "bad"); err != contract.ErrInvalid {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}

	decodeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{"))
	}))
	defer decodeServer.Close()
	cfg.Endpoint = decodeServer.URL
	if err := New().Verify(cfg, "token"); !errors.Is(err, contract.ErrVerifyFailed) {
		t.Fatalf("err = %v, want ErrVerifyFailed", err)
	}
}
