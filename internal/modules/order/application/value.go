package application

import (
	"regexp"
	"strings"

	addresscontract "github.com/dujiao-next/internal/modules/addressdivision/contract"
	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

var shippingPhonePattern = regexp.MustCompile(`^[0-9+\-\s]{6,20}$`)

func resolveSiteCurrency(settingService *settingsapp.Service) string {
	if settingService == nil {
		return constants.SiteCurrencyDefault
	}
	currency, err := settingService.GetSiteCurrency(constants.SiteCurrencyDefault)
	if err != nil {
		return constants.SiteCurrencyDefault
	}
	return currency
}

// ResolvePaymentExpireMinutes 解析订单与支付共同使用的付款超时时间。
func ResolvePaymentExpireMinutes(settingService *settingsapp.Service, defaultMinutes int) int {
	if defaultMinutes <= 0 {
		defaultMinutes = 15
	}
	if settingService == nil {
		return defaultMinutes
	}
	minutes, err := settingService.GetOrderPaymentExpireMinutes(defaultMinutes)
	if err != nil || minutes <= 0 {
		return defaultMinutes
	}
	return minutes
}

func resolveProductOrderSKU(productSKURepo productcontract.SKURepository, product *productdomain.Product, rawSKUID uint) (*productdomain.ProductSKU, error) {
	if product == nil || product.ID == 0 {
		return nil, ErrProductNotAvailable
	}
	if productSKURepo == nil {
		return nil, ErrProductSKUInvalid
	}
	if rawSKUID > 0 {
		sku, err := productSKURepo.GetByID(rawSKUID)
		if err != nil {
			return nil, err
		}
		if sku == nil || sku.ProductID != product.ID || !sku.IsActive {
			return nil, ErrProductSKUInvalid
		}
		return sku, nil
	}

	activeSKUs, err := productSKURepo.ListByProduct(product.ID, true)
	if err != nil {
		return nil, err
	}
	if len(activeSKUs) == 1 {
		return &activeSKUs[0], nil
	}
	if len(activeSKUs) == 0 {
		return nil, ErrProductSKUInvalid
	}
	return nil, ErrProductSKURequired
}

func ValidateAndNormalizeShippingAddress(input jsonmap.JSON, addressLookup addresscontract.Lookup) (jsonmap.JSON, error) {
	if len(input) == 0 {
		return nil, ErrShippingAddressRequired
	}

	normalized := jsonmap.JSON{
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
		"detail_address": strings.TrimSpace(toShippingText(input["detail_address"])),
	}

	requiredKeys := []string{
		"receiver_name",
		"receiver_phone",
		"province_code",
		"city_code",
		"district_code",
		"township_code",
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
	if addressLookup == nil {
		return nil, ErrShippingAddressInvalid
	}

	province, ok := addressLookup.GetProvince(toShippingText(normalized["province_code"]))
	if !ok {
		return nil, ErrShippingAddressInvalid
	}
	city, ok := addressLookup.GetCity(toShippingText(normalized["city_code"]))
	if !ok || city.ProvinceCode != province.Code {
		return nil, ErrShippingAddressInvalid
	}
	district, ok := addressLookup.GetDistrict(toShippingText(normalized["district_code"]))
	if !ok || district.ProvinceCode != province.Code || district.CityCode != city.Code {
		return nil, ErrShippingAddressInvalid
	}
	township, ok := addressLookup.GetTownship(toShippingText(normalized["township_code"]))
	if !ok || township.ProvinceCode != province.Code || township.CityCode != city.Code || township.DistrictCode != district.Code {
		return nil, ErrShippingAddressInvalid
	}

	normalized["province"] = province.Name
	normalized["city"] = city.Name
	normalized["district"] = district.Name
	normalized["township"] = township.Name

	return normalized, nil
}

func toShippingText(value interface{}) string {
	if value == nil {
		return ""
	}
	text, _ := value.(string)
	return text
}
