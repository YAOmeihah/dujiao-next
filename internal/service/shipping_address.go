package service

import (
	"regexp"
	"strings"

	"github.com/dujiao-next/internal/models"
)

var shippingPhonePattern = regexp.MustCompile(`^[0-9+\-\s]{6,20}$`)

func ValidateAndNormalizeShippingAddress(input models.JSON) (models.JSON, error) {
	if len(input) == 0 {
		return nil, ErrShippingAddressRequired
	}

	normalized := models.JSON{
		"receiver_name":  strings.TrimSpace(toShippingText(input["receiver_name"])),
		"receiver_phone": strings.TrimSpace(toShippingText(input["receiver_phone"])),
		"province":       strings.TrimSpace(toShippingText(input["province"])),
		"city":           strings.TrimSpace(toShippingText(input["city"])),
		"district":       strings.TrimSpace(toShippingText(input["district"])),
		"detail_address": strings.TrimSpace(toShippingText(input["detail_address"])),
		"postal_code":    strings.TrimSpace(toShippingText(input["postal_code"])),
	}

	requiredKeys := []string{
		"receiver_name",
		"receiver_phone",
		"province",
		"city",
		"district",
		"detail_address",
	}
	for _, key := range requiredKeys {
		if strings.TrimSpace(toShippingText(normalized[key])) == "" {
			return nil, ErrShippingAddressRequired
		}
	}
	if !shippingPhonePattern.MatchString(toShippingText(normalized["receiver_phone"])) {
		return nil, ErrShippingAddressInvalid
	}

	return normalized, nil
}

func toShippingText(value interface{}) string {
	if value == nil {
		return ""
	}
	text, _ := value.(string)
	return text
}
