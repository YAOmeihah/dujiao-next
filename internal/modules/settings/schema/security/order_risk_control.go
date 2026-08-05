package settingssecurity

import (
	"encoding/json"
	"net"
	"regexp"
	"strings"
	"unicode"

	"github.com/dujiao-next/internal/shared/jsonmap"
)

var guestPhonePattern = regexp.MustCompile(`^\+?[0-9]{6,20}$`)

// OrderRateLimitConfig 下单频率限制配置
type OrderRateLimitConfig struct {
	Enabled       bool `json:"enabled"`
	WindowSeconds int  `json:"window_seconds"`
	MaxRequests   int  `json:"max_requests"`
	BlockSeconds  int  `json:"block_seconds"`
}

// OrderRiskControlConfig 订单风控配置
type OrderRiskControlConfig struct {
	Enabled                       bool                 `json:"enabled"`
	MaxPendingOrdersPerUser       int                  `json:"max_pending_orders_per_user"`
	MaxPendingOrdersPerIP         int                  `json:"max_pending_orders_per_ip"`
	MaxPendingOrdersPerGuestPhone int                  `json:"max_pending_orders_per_guest_phone"`
	OrderRateLimit                OrderRateLimitConfig `json:"order_rate_limit"`
	IPBlacklist                   []string             `json:"ip_blacklist"`
	PhoneBlacklist                []string             `json:"phone_blacklist"`
	LegacyEmailBlacklist          []string             `json:"email_blacklist,omitempty"`
}

// DefaultOrderRiskControlConfig 默认风控配置
func DefaultOrderRiskControlConfig() OrderRiskControlConfig {
	return OrderRiskControlConfig{
		Enabled:                       false,
		MaxPendingOrdersPerUser:       3,
		MaxPendingOrdersPerIP:         5,
		MaxPendingOrdersPerGuestPhone: 2,
		OrderRateLimit: OrderRateLimitConfig{
			Enabled:       false,
			WindowSeconds: 60,
			MaxRequests:   5,
			BlockSeconds:  120,
		},
		IPBlacklist:    []string{},
		PhoneBlacklist: []string{},
	}
}

// NormalizeOrderRiskControlConfig 归一化风控配置
func NormalizeOrderRiskControlConfig(cfg OrderRiskControlConfig) OrderRiskControlConfig {
	if cfg.MaxPendingOrdersPerUser < 0 || cfg.MaxPendingOrdersPerUser > 100 {
		cfg.MaxPendingOrdersPerUser = 3
	}
	if cfg.MaxPendingOrdersPerIP < 0 || cfg.MaxPendingOrdersPerIP > 100 {
		cfg.MaxPendingOrdersPerIP = 5
	}
	if cfg.MaxPendingOrdersPerGuestPhone < 0 || cfg.MaxPendingOrdersPerGuestPhone > 100 {
		cfg.MaxPendingOrdersPerGuestPhone = 2
	}

	if cfg.OrderRateLimit.WindowSeconds < 10 || cfg.OrderRateLimit.WindowSeconds > 3600 {
		cfg.OrderRateLimit.WindowSeconds = 60
	}
	if cfg.OrderRateLimit.MaxRequests < 1 || cfg.OrderRateLimit.MaxRequests > 100 {
		cfg.OrderRateLimit.MaxRequests = 5
	}
	if cfg.OrderRateLimit.BlockSeconds < 0 || cfg.OrderRateLimit.BlockSeconds > 86400 {
		cfg.OrderRateLimit.BlockSeconds = 120
	}

	// 归一化 IP 黑名单：去空行、去首尾空格、校验格式
	cleanIPs := make([]string, 0, len(cfg.IPBlacklist))
	for _, entry := range cfg.IPBlacklist {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if isValidIPOrCIDR(entry) {
			cleanIPs = append(cleanIPs, entry)
		}
	}
	cfg.IPBlacklist = cleanIPs

	cleanPhones := make([]string, 0, len(cfg.PhoneBlacklist))
	for _, phone := range cfg.PhoneBlacklist {
		phone = canonicalizeGuestPhone(phone)
		if phone != "" && guestPhonePattern.MatchString(phone) {
			cleanPhones = append(cleanPhones, phone)
		}
	}
	cfg.PhoneBlacklist = cleanPhones

	cleanEmails := make([]string, 0, len(cfg.LegacyEmailBlacklist))
	for _, email := range cfg.LegacyEmailBlacklist {
		email = strings.ToLower(strings.TrimSpace(email))
		if email != "" {
			cleanEmails = append(cleanEmails, email)
		}
	}
	cfg.LegacyEmailBlacklist = cleanEmails

	return cfg
}

func canonicalizeGuestPhone(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(trimmed))
	for _, r := range trimmed {
		switch {
		case unicode.IsSpace(r):
			continue
		case r == '-' || r == '(' || r == ')':
			continue
		default:
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// isValidIPOrCIDR 校验字符串是否为有效的 IP 地址或 CIDR 表示
func isValidIPOrCIDR(s string) bool {
	if net.ParseIP(s) != nil {
		return true
	}
	_, _, err := net.ParseCIDR(s)
	return err == nil
}

// DecodeOrderRiskControlConfig 从 JSON map 解析风控配置
func DecodeOrderRiskControlConfig(raw jsonmap.JSON, fallback OrderRiskControlConfig) OrderRiskControlConfig {
	result := fallback
	if raw == nil {
		return result
	}
	normalizedRaw := make(jsonmap.JSON, len(raw))
	for key, value := range raw {
		normalizedRaw[key] = value
	}
	if _, exists := normalizedRaw["max_pending_orders_per_guest_phone"]; !exists {
		if legacyValue, ok := normalizedRaw["max_pending_orders_per_guest_email"]; ok {
			normalizedRaw["max_pending_orders_per_guest_phone"] = legacyValue
		}
	}
	data, err := json.Marshal(normalizedRaw)
	if err != nil {
		return result
	}
	_ = json.Unmarshal(data, &result)
	return NormalizeOrderRiskControlConfig(result)
}

// EncodeOrderRiskControlConfig 将风控配置转为 map 用于存储
func EncodeOrderRiskControlConfig(cfg OrderRiskControlConfig) jsonmap.JSON {
	normalized := NormalizeOrderRiskControlConfig(cfg)
	data, err := json.Marshal(normalized)
	if err != nil {
		return jsonmap.JSON{}
	}
	var result jsonmap.JSON
	_ = json.Unmarshal(data, &result)
	return result
}

// NormalizeOrderRiskControlConfigJSON 是 Registry 使用的原始 JSON 写入策略。
func NormalizeOrderRiskControlConfigJSON(value jsonmap.JSON) jsonmap.JSON {
	return EncodeOrderRiskControlConfig(DecodeOrderRiskControlConfig(value, DefaultOrderRiskControlConfig()))
}
