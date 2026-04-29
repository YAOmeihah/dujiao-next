package service

import (
	"testing"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
)

func TestCaptchaDefaultSettingIncludesCap(t *testing.T) {
	setting := CaptchaDefaultSetting(config.CaptchaConfig{
		Provider: constants.CaptchaProviderCap,
		Cap: config.CaptchaCapConfig{
			Endpoint:  " http://127.0.0.1:3001/ ",
			SiteKey:   " 3bed8d0d85 ",
			SecretKey: " secret ",
			TimeoutMS: 0,
		},
	})

	if setting.Provider != constants.CaptchaProviderCap {
		t.Fatalf("provider = %q, want cap", setting.Provider)
	}
	if setting.Cap.Endpoint != "http://127.0.0.1:3001" {
		t.Fatalf("endpoint = %q", setting.Cap.Endpoint)
	}
	if setting.Cap.SiteKey != "3bed8d0d85" {
		t.Fatalf("site key = %q", setting.Cap.SiteKey)
	}
	if setting.Cap.SecretKey != "secret" {
		t.Fatalf("secret key = %q", setting.Cap.SecretKey)
	}
	if setting.Cap.TimeoutMS != 2000 {
		t.Fatalf("timeout = %d, want 2000", setting.Cap.TimeoutMS)
	}
}

func TestValidateCaptchaSettingRequiresCapFields(t *testing.T) {
	setting := NormalizeCaptchaSetting(CaptchaSetting{
		Provider: constants.CaptchaProviderCap,
		Scenes:   CaptchaSceneSetting{Login: true},
		Cap: CaptchaCapSetting{
			Endpoint:  "http://127.0.0.1:3001",
			SiteKey:   "3bed8d0d85",
			SecretKey: "secret",
			TimeoutMS: 2000,
		},
	})
	if err := ValidateCaptchaSetting(setting); err != nil {
		t.Fatalf("valid cap setting returned error: %v", err)
	}

	setting.Cap.SecretKey = ""
	if err := ValidateCaptchaSetting(setting); err == nil {
		t.Fatal("missing cap secret key should fail validation")
	}
}

func TestCaptchaSettingPublicAndAdminCapOutput(t *testing.T) {
	setting := NormalizeCaptchaSetting(CaptchaSetting{
		Provider: constants.CaptchaProviderCap,
		Cap: CaptchaCapSetting{
			Endpoint:  "http://127.0.0.1:3001",
			SiteKey:   "3bed8d0d85",
			SecretKey: "secret",
			TimeoutMS: 2000,
		},
	})

	admin := MaskCaptchaSettingForAdmin(setting)
	adminCap, ok := admin["cap"].(map[string]interface{})
	if !ok {
		t.Fatalf("admin cap config missing: %#v", admin)
	}
	if adminCap["secret_key"] != "" {
		t.Fatalf("admin secret should be masked: %#v", adminCap)
	}
	if adminCap["has_secret"] != true {
		t.Fatalf("admin has_secret = %#v", adminCap["has_secret"])
	}

	public := PublicCaptchaSetting(setting)
	publicCap, ok := public["cap"].(map[string]interface{})
	if !ok {
		t.Fatalf("public cap config missing: %#v", public)
	}
	if _, exists := publicCap["secret_key"]; exists {
		t.Fatalf("public config leaked secret: %#v", publicCap)
	}
}
