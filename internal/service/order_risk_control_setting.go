package service

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
)

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
		entry = trimString(entry)
		if entry == "" {
			continue
		}
		if isValidIPOrCIDR(entry) {
			cleanIPs = append(cleanIPs, entry)
		}
	}
	cfg.IPBlacklist = cleanIPs

	// 归一化手机号黑名单：去空行、去首尾空格
	cleanPhones := make([]string, 0, len(cfg.PhoneBlacklist))
	for _, phone := range cfg.PhoneBlacklist {
		phone = canonicalizeGuestPhone(phone)
		if phone != "" && guestPhonePattern.MatchString(phone) {
			cleanPhones = append(cleanPhones, phone)
		}
	}
	cfg.PhoneBlacklist = cleanPhones

	cleanLegacyEmails := make([]string, 0, len(cfg.LegacyEmailBlacklist))
	for _, email := range cfg.LegacyEmailBlacklist {
		email = strings.ToLower(strings.TrimSpace(email))
		if email != "" {
			cleanLegacyEmails = append(cleanLegacyEmails, email)
		}
	}
	cfg.LegacyEmailBlacklist = cleanLegacyEmails

	return cfg
}

func trimString(s string) string {
	return strings.TrimSpace(s)
}

// isValidIPOrCIDR 校验字符串是否为有效的 IP 地址或 CIDR 表示
func isValidIPOrCIDR(s string) bool {
	if net.ParseIP(s) != nil {
		return true
	}
	_, _, err := net.ParseCIDR(s)
	return err == nil
}

// orderRiskControlConfigFromJSON 从 JSON map 解析风控配置
func orderRiskControlConfigFromJSON(raw models.JSON, fallback OrderRiskControlConfig) OrderRiskControlConfig {
	result := fallback
	if raw == nil {
		return result
	}
	normalizedRaw := make(models.JSON, len(raw))
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

// OrderRiskControlConfigToMap 将风控配置转为 map 用于存储
func OrderRiskControlConfigToMap(cfg OrderRiskControlConfig) models.JSON {
	normalized := NormalizeOrderRiskControlConfig(cfg)
	data, err := json.Marshal(normalized)
	if err != nil {
		return models.JSON{}
	}
	var result models.JSON
	_ = json.Unmarshal(data, &result)
	return result
}

// GetOrderRiskControlConfig 获取风控配置
func (s *SettingService) GetOrderRiskControlConfig() (OrderRiskControlConfig, error) {
	fallback := DefaultOrderRiskControlConfig()
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyOrderRiskControlConfig)
	if err != nil {
		return fallback, err
	}
	return orderRiskControlConfigFromJSON(value, fallback), nil
}
