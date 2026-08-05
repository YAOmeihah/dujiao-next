package application

import (
	"errors"
	"testing"

	addresscontract "github.com/dujiao-next/internal/modules/addressdivision/contract"
	addressstore "github.com/dujiao-next/internal/modules/addressdivision/infrastructure/memorystore"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

func TestNormalizeGuestPhone_CanonicalizesFormatting(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain digits", input: "13800138000", want: "13800138000"},
		{name: "spaces and dashes", input: " 138-0013 8000 ", want: "13800138000"},
		{name: "country code and parentheses", input: "(+86) 138-0013-8000", want: "+8613800138000"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeGuestPhone(tc.input)
			if err != nil {
				t.Fatalf("normalizeGuestPhone returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("normalizeGuestPhone(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func newTestAddressLookup() addresscontract.Lookup {
	return addressstore.New(addresscontract.Dataset{
		Provinces: []addresscontract.Division{{Code: "33", Name: "浙江省"}},
		Cities:    []addresscontract.Division{{Code: "3301", Name: "杭州市", ProvinceCode: "33"}},
		Districts: []addresscontract.Division{{Code: "330106", Name: "西湖区", ProvinceCode: "33", CityCode: "3301"}},
		Townships: []addresscontract.Division{{Code: "330106001", Name: "西溪街道", ProvinceCode: "33", CityCode: "3301", DistrictCode: "330106"}},
	})
}

func TestValidateAndNormalizeShippingAddressCanonicalizesFiveLevelAddress(t *testing.T) {
	normalized, err := ValidateAndNormalizeShippingAddress(jsonmap.JSON{
		"receiver_name":  " 张三 ",
		"receiver_phone": " 13800138000 ",
		"province":       "错误省份",
		"province_code":  "33",
		"city":           "错误城市",
		"city_code":      "3301",
		"district":       "错误区县",
		"district_code":  "330106",
		"township":       "错误街道",
		"township_code":  "330106001",
		"detail_address": " 文三路100号 ",
	}, newTestAddressLookup())
	if err != nil {
		t.Fatalf("ValidateAndNormalizeShippingAddress failed: %v", err)
	}
	if normalized["province"] != "浙江省" || normalized["city"] != "杭州市" {
		t.Fatalf("expected normalized province/city names, got %+v", normalized)
	}
	if normalized["township"] != "西溪街道" || normalized["detail_address"] != "文三路100号" {
		t.Fatalf("unexpected normalized address: %+v", normalized)
	}
}

func TestValidateAndNormalizeShippingAddressRejectsBrokenHierarchy(t *testing.T) {
	_, err := ValidateAndNormalizeShippingAddress(jsonmap.JSON{
		"receiver_name":  "张三",
		"receiver_phone": "13800138000",
		"province_code":  "33",
		"city_code":      "3301",
		"district_code":  "330106",
		"township_code":  "110101001",
		"detail_address": "文三路100号",
	}, newTestAddressLookup())
	if !errors.Is(err, ErrShippingAddressInvalid) {
		t.Fatalf("expected shipping address invalid, got %v", err)
	}
}
