package captchahttp

import "testing"

func TestCaptchaPayloadRequestIncludesCapToken(t *testing.T) {
	payload := CaptchaPayloadRequest{CapToken: " cap-token "}.ToCaptchaPayload()
	if payload.CapToken != "cap-token" {
		t.Fatalf("cap token = %q", payload.CapToken)
	}
}
