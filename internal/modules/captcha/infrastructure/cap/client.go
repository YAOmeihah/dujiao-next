package cap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dujiao-next/internal/modules/captcha/contract"
	settingssecurity "github.com/dujiao-next/internal/modules/settings/schema/security"
)

// Client 通过 CAP Standalone Siteverify API 校验令牌。
type Client struct{}

var _ contract.CapVerifier = Client{}

// New 创建 CAP 校验客户端。
func New() Client { return Client{} }

type verifyResponse struct {
	Success bool `json:"success"`
}

// Verify 校验令牌。
func (Client) Verify(cfg settingssecurity.CaptchaCapSetting, token string) error {
	endpoint := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	siteKey := strings.Trim(strings.TrimSpace(cfg.SiteKey), "/")
	secret := strings.TrimSpace(cfg.SecretKey)
	if endpoint == "" || siteKey == "" || secret == "" {
		return contract.ErrConfigInvalid
	}

	timeout := cfg.TimeoutMS
	if timeout < 500 || timeout > 10000 {
		timeout = 2000
	}

	body, err := json.Marshal(map[string]string{
		"secret":   secret,
		"response": strings.TrimSpace(token),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", contract.ErrVerifyFailed, err)
	}

	verifyURL := endpoint + "/" + url.PathEscape(siteKey) + "/siteverify"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, verifyURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: %v", contract.ErrVerifyFailed, err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(timeout) * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", contract.ErrVerifyFailed, err)
	}
	defer resp.Body.Close()

	var result verifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("%w: %v", contract.ErrVerifyFailed, err)
	}
	if !result.Success {
		return contract.ErrInvalid
	}
	return nil
}
