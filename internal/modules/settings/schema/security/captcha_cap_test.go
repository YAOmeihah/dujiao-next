package settingssecurity

import (
	"testing"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
)

func TestDefaultCaptchaSettingIncludesCap(t *testing.T) {
	setting := DefaultCaptchaSetting(config.CaptchaConfig{
		Provider: constants.CaptchaProviderCap,
		Cap: config.CaptchaCapConfig{
			Endpoint:  " http://127.0.0.1:3001/ ",
			SiteKey:   " site ",
			SecretKey: " secret ",
			TimeoutMS: 0,
		},
	})

	if setting.Provider != constants.CaptchaProviderCap {
		t.Fatalf("provider = %q, want cap", setting.Provider)
	}
	if setting.Cap.Endpoint != "http://127.0.0.1:3001" || setting.Cap.SiteKey != "site" || setting.Cap.SecretKey != "secret" {
		t.Fatalf("cap setting = %#v", setting.Cap)
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
			SiteKey:   "site",
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
			SiteKey:   "site",
			SecretKey: "secret",
			TimeoutMS: 2000,
		},
	})

	adminCap, ok := MaskCaptchaSettingForAdmin(setting)["cap"].(map[string]interface{})
	if !ok || adminCap["secret_key"] != "" || adminCap["has_secret"] != true {
		t.Fatalf("admin cap config = %#v", adminCap)
	}

	publicCap, ok := PublicCaptchaSetting(setting)["cap"].(map[string]interface{})
	if !ok || publicCap["endpoint"] != "http://127.0.0.1:3001" || publicCap["site_key"] != "site" {
		t.Fatalf("public cap config = %#v", publicCap)
	}
	if _, exists := publicCap["secret_key"]; exists {
		t.Fatalf("public config leaked secret: %#v", publicCap)
	}
}
