package service

import (
	"regexp"
	"strings"

	"github.com/dujiao-next/internal/models"
)

var shippingPhonePattern = regexp.MustCompile(`^[0-9+\-\s]{6,20}$`)

func ValidateAndNormalizeShippingAddress(input models.JSON, addressService ...*AddressService) (models.JSON, error) {
	var svc *AddressService
	if len(addressService) > 0 {
		svc = addressService[0]
	}
	return ValidateAndNormalizeShippingAddressWithService(input, svc)
}

func ValidateAndNormalizeShippingAddressWithService(input models.JSON, addressService *AddressService) (models.JSON, error) {
	if len(input) == 0 {
		return nil, ErrShippingAddressRequired
	}

	normalized := models.JSON{
		"receiver_name":  strings.TrimSpace(toShippingText(input["receiver_name"])),
		"receiver_phone": strings.TrimSpace(toShippingText(input["receiver_phone"])),
		"province":       strings.TrimSpace(toShippingText(input["province"])),
		"province_code":  strings.TrimSpace(toShippingText(input["province_code"])),
		"city":           strings.TrimSpace(toShippingText(input["city"])),
		"city_code":      strings.TrimSpace(toShippingText(input["city_code"])),
		"district":       strings.TrimSpace(toShippingText(input["district"])),
		"district_code":  strings.TrimSpace(toShippingText(input["district_code"])),
		"township":       strings.TrimSpace(toShippingText(input["township"])),
		"township_code":  strings.TrimSpace(toShippingText(input["township_code"])),
		"village":        strings.TrimSpace(toShippingText(input["village"])),
		"village_code":   strings.TrimSpace(toShippingText(input["village_code"])),
		"detail_address": strings.TrimSpace(toShippingText(input["detail_address"])),
	}

	requiredKeys := []string{
		"receiver_name",
		"receiver_phone",
		"province_code",
		"city_code",
		"district_code",
		"township_code",
		"village_code",
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
	if addressService == nil {
		return nil, ErrShippingAddressInvalid
	}

	province, ok := addressService.GetProvince(toShippingText(normalized["province_code"]))
	if !ok {
		return nil, ErrShippingAddressInvalid
	}
	city, ok := addressService.GetCity(toShippingText(normalized["city_code"]))
	if !ok || city.ProvinceCode != province.Code {
		return nil, ErrShippingAddressInvalid
	}
	district, ok := addressService.GetDistrict(toShippingText(normalized["district_code"]))
	if !ok || district.ProvinceCode != province.Code || district.CityCode != city.Code {
		return nil, ErrShippingAddressInvalid
	}
	township, ok := addressService.GetTownship(toShippingText(normalized["township_code"]))
	if !ok || township.ProvinceCode != province.Code || township.CityCode != city.Code || township.DistrictCode != district.Code {
		return nil, ErrShippingAddressInvalid
	}
	village, ok := addressService.GetVillage(toShippingText(normalized["village_code"]))
	if !ok || village.ProvinceCode != province.Code || village.CityCode != city.Code || village.DistrictCode != district.Code || village.TownshipCode != township.Code {
		return nil, ErrShippingAddressInvalid
	}

	normalized["province"] = province.Name
	normalized["city"] = city.Name
	normalized["district"] = district.Name
	normalized["township"] = township.Name
	normalized["village"] = village.Name

	return normalized, nil
}

func toShippingText(value interface{}) string {
	if value == nil {
		return ""
	}
	text, _ := value.(string)
	return text
}
