package vpay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/payment/common"
)

const (
	defaultCreateOrderPath = "/createOrder"
	defaultPayPagePath     = "/payPage/pay.html"
	SignTypeMD5            = "MD5"
	SignTypeHMACSHA256     = "HMAC_SHA256"
)

var (
	ErrConfigInvalid    = errors.New("vpay config invalid")
	ErrRequestFailed    = errors.New("vpay request failed")
	ErrResponseInvalid  = errors.New("vpay response invalid")
	ErrChannelType      = errors.New("vpay channel type invalid")
	ErrSignatureInvalid = errors.New("vpay signature invalid")
)

// Config VPay 商户侧配置。
type Config struct {
	GatewayURL      string `json:"gateway_url"`
	SignKey         string `json:"sign_key"`
	MerchantKey     string `json:"merchant_key"`
	NotifyURL       string `json:"notify_url"`
	ReturnURL       string `json:"return_url"`
	SignType        string `json:"sign_type"`
	CreateOrderPath string `json:"create_order_path"`
	PayPagePath     string `json:"pay_page_path"`
}

// CreateInput VPay 创建订单输入。
type CreateInput struct {
	PayID       string
	Param       string
	ChannelType string
	Price       string
	NotifyURL   string
	ReturnURL   string
}

// CreateResult VPay 创建订单结果。
type CreateResult struct {
	PayID       string
	OrderID     string
	PayType     string
	Price       string
	ReallyPrice string
	PayURL      string
	RedirectURL string
	Raw         map[string]interface{}
}

// ParseConfig 解析 VPay 配置。
func ParseConfig(raw map[string]interface{}) (*Config, error) {
	return common.ParseConfig[Config](raw, ErrConfigInvalid)
}

func (c *Config) Normalize() {
	c.GatewayURL = strings.TrimRight(strings.TrimSpace(c.GatewayURL), "/")
	c.SignKey = strings.TrimSpace(c.SignKey)
	c.MerchantKey = strings.TrimSpace(c.MerchantKey)
	if c.SignKey == "" {
		c.SignKey = c.MerchantKey
	}
	if c.MerchantKey == "" {
		c.MerchantKey = c.SignKey
	}
	c.NotifyURL = strings.TrimSpace(c.NotifyURL)
	c.ReturnURL = strings.TrimSpace(c.ReturnURL)
	c.SignType = normalizeSignType(c.SignType)
	c.CreateOrderPath = normalizePath(c.CreateOrderPath, defaultCreateOrderPath)
	c.PayPagePath = normalizePath(c.PayPagePath, defaultPayPagePath)
}

// ValidateConfig 校验 VPay 配置完整性。
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("%w: config is nil", ErrConfigInvalid)
	}
	if strings.TrimSpace(cfg.GatewayURL) == "" {
		return fmt.Errorf("%w: gateway_url is required", ErrConfigInvalid)
	}
	if strings.TrimSpace(cfg.SignKey) == "" {
		return fmt.Errorf("%w: sign_key is required", ErrConfigInvalid)
	}
	if strings.TrimSpace(cfg.NotifyURL) == "" {
		return fmt.Errorf("%w: notify_url is required", ErrConfigInvalid)
	}
	if strings.TrimSpace(cfg.ReturnURL) == "" {
		return fmt.Errorf("%w: return_url is required", ErrConfigInvalid)
	}
	if !isSupportedSignType(cfg.SignType) {
		return fmt.Errorf("%w: sign_type is unsupported", ErrConfigInvalid)
	}
	return nil
}

// CreatePayment 调用 VPay /createOrder 并返回内置支付页跳转地址。
func CreatePayment(ctx context.Context, cfg *Config, input CreateInput) (*CreateResult, error) {
	if cfg == nil {
		return nil, ErrConfigInvalid
	}
	if ctx == nil {
		ctx = context.Background()
	}
	payID := strings.TrimSpace(input.PayID)
	price := strings.TrimSpace(input.Price)
	if payID == "" || price == "" {
		return nil, ErrConfigInvalid
	}
	payType := resolvePayType(input.ChannelType)
	if payType == "" {
		return nil, ErrChannelType
	}
	notifyURL := strings.TrimSpace(input.NotifyURL)
	if notifyURL == "" {
		notifyURL = strings.TrimSpace(cfg.NotifyURL)
	}
	returnURL := strings.TrimSpace(input.ReturnURL)
	if returnURL == "" {
		returnURL = strings.TrimSpace(cfg.ReturnURL)
	}
	if notifyURL == "" || returnURL == "" || strings.TrimSpace(cfg.GatewayURL) == "" || strings.TrimSpace(cfg.SignKey) == "" {
		return nil, ErrConfigInvalid
	}

	params := map[string]string{
		"payId":     payID,
		"type":      payType,
		"price":     price,
		"param":     strings.TrimSpace(input.Param),
		"notifyUrl": notifyURL,
		"returnUrl": returnURL,
	}
	signType := normalizeSignType(cfg.SignType)
	sign, err := signCreateOrderWithType(params["payId"], params["param"], params["type"], params["price"], cfg.SignKey, signType)
	if err != nil {
		return nil, fmt.Errorf("%w: sign_type is unsupported", ErrConfigInvalid)
	}
	params["signType"] = signType
	params["sign"] = sign

	body, err := postForm(ctx, buildEndpoint(cfg.GatewayURL, cfg.CreateOrderPath), params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	body, err = normalizeResponseBody(body)
	if err != nil {
		return nil, fmt.Errorf("%w: normalize response failed", ErrResponseInvalid)
	}
	result, err := parseCreateResponse(body)
	if err != nil {
		return nil, err
	}
	if result.PayID == "" {
		result.PayID = payID
	}
	if result.OrderID == "" {
		return nil, fmt.Errorf("%w: orderId is empty", ErrResponseInvalid)
	}
	result.RedirectURL = buildPayPageURL(cfg.GatewayURL, cfg.PayPagePath, result.OrderID)
	return result, nil
}

// VerifyCallback 验证 VPay 异步通知或同步跳转签名。
func VerifyCallback(cfg *Config, form map[string][]string) error {
	if cfg == nil {
		return ErrConfigInvalid
	}
	sign := strings.TrimSpace(firstValue(form, "sign"))
	if sign == "" {
		return ErrSignatureInvalid
	}
	expected, err := signCallbackWithType(
		firstValue(form, "payId"),
		firstValue(form, "param"),
		firstValue(form, "type"),
		firstValue(form, "price"),
		firstValue(form, "reallyPrice"),
		cfg.SignKey,
		firstValue(form, "signType"),
	)
	if err != nil {
		return ErrSignatureInvalid
	}
	if !strings.EqualFold(expected, sign) {
		return ErrSignatureInvalid
	}
	return nil
}

// IsSupportedChannelType 判断 VPay 支持的渠道类型。
func IsSupportedChannelType(channelType string) bool {
	return resolvePayType(channelType) != ""
}

// ResolvePayType 将系统渠道类型转换为 VPay 支付方式。
func ResolvePayType(channelType string) string {
	return resolvePayType(channelType)
}

func resolvePayType(channelType string) string {
	switch strings.ToLower(strings.TrimSpace(channelType)) {
	case constants.PaymentChannelTypeWechat, constants.PaymentChannelTypeWxpay:
		return "1"
	case constants.PaymentChannelTypeAlipay:
		return "2"
	default:
		return ""
	}
}

func signCreateOrder(payID, param, payType, price, key string) string {
	return md5Hex(payID + param + payType + price + key)
}

func signCreateOrderWithType(payID, param, payType, price, key, signType string) (string, error) {
	return makeSignature(payID+param+payType+price, key, signType)
}

func signCallback(payID, param, payType, price, reallyPrice, key string) string {
	return md5Hex(payID + param + payType + price + reallyPrice + key)
}

func signCallbackWithType(payID, param, payType, price, reallyPrice, key, signType string) (string, error) {
	return makeSignature(payID+param+payType+price+reallyPrice, key, signType)
}

func makeSignature(payload, key, signType string) (string, error) {
	switch normalizeSignType(signType) {
	case SignTypeMD5:
		return md5Hex(payload + key), nil
	case SignTypeHMACSHA256:
		return hmacSHA256Hex(payload, key), nil
	default:
		return "", ErrSignatureInvalid
	}
}

func md5Hex(raw string) string {
	sum := md5.Sum([]byte(raw))
	return strings.ToLower(hex.EncodeToString(sum[:]))
}

func hmacSHA256Hex(payload, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(payload))
	return strings.ToLower(hex.EncodeToString(mac.Sum(nil)))
}

func normalizeSignType(signType string) string {
	signType = strings.ToUpper(strings.TrimSpace(signType))
	if signType == "" {
		return SignTypeMD5
	}
	return signType
}

func isSupportedSignType(signType string) bool {
	switch normalizeSignType(signType) {
	case SignTypeMD5, SignTypeHMACSHA256:
		return true
	default:
		return false
	}
}

func postForm(ctx context.Context, endpoint string, params map[string]string) ([]byte, error) {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func parseCreateResponse(body []byte) (*CreateResult, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: decode response failed", ErrResponseInvalid)
	}
	if pickInt(raw, "code") != 1 {
		return nil, fmt.Errorf("%w: %s", ErrResponseInvalid, strings.TrimSpace(pickString(raw, "msg")))
	}
	data, _ := raw["data"].(map[string]interface{})
	if data == nil {
		return nil, fmt.Errorf("%w: data is empty", ErrResponseInvalid)
	}
	return &CreateResult{
		PayID:       strings.TrimSpace(pickString(data, "payId")),
		OrderID:     strings.TrimSpace(pickString(data, "orderId")),
		PayType:     strings.TrimSpace(pickString(data, "payType")),
		Price:       strings.TrimSpace(pickString(data, "price")),
		ReallyPrice: strings.TrimSpace(pickString(data, "reallyPrice")),
		PayURL:      strings.TrimSpace(pickString(data, "payUrl")),
		Raw:         raw,
	}, nil
}

func normalizeResponseBody(body []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '"' {
		return trimmed, nil
	}
	var inner string
	if err := json.Unmarshal(trimmed, &inner); err != nil {
		return nil, err
	}
	return bytes.TrimSpace([]byte(inner)), nil
}

func buildEndpoint(gatewayURL, path string) string {
	base := strings.TrimRight(strings.TrimSpace(gatewayURL), "/")
	return base + normalizePath(path, "")
}

func buildPayPageURL(gatewayURL, path, orderID string) string {
	endpoint := buildEndpoint(gatewayURL, path)
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	query := parsed.Query()
	query.Set("orderId", strings.TrimSpace(orderID))
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func normalizePath(path, fallback string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		path = fallback
	}
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func firstValue(form map[string][]string, key string) string {
	if values, ok := form[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

func pickString(data map[string]interface{}, key string) string {
	val, ok := data[key]
	if !ok || val == nil {
		return ""
	}
	switch typed := val.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func pickInt(data map[string]interface{}, key string) int {
	val, ok := data[key]
	if !ok || val == nil {
		return 0
	}
	switch typed := val.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return int(parsed)
		}
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
	}
	return 0
}
