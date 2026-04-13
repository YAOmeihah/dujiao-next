package service

import (
	"errors"
	"testing"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
)

func newTestAddressService() *AddressService {
	return NewAddressService(repository.NewAddressDivisionRepository(repository.AddressDivisionDataset{
		Provinces: []models.AddressDivision{
			{Code: "33", Name: "浙江省"},
		},
		Cities: []models.AddressDivision{
			{Code: "3301", Name: "杭州市", ProvinceCode: "33"},
		},
		Districts: []models.AddressDivision{
			{Code: "330106", Name: "西湖区", ProvinceCode: "33", CityCode: "3301"},
		},
		Townships: []models.AddressDivision{
			{Code: "330106001", Name: "西溪街道", ProvinceCode: "33", CityCode: "3301", DistrictCode: "330106"},
		},
		Villages: []models.AddressDivision{
			{Code: "330106001001", Name: "文一社区", ProvinceCode: "33", CityCode: "3301", DistrictCode: "330106", TownshipCode: "330106001"},
		},
	}))
}

func TestAddressServiceListChildrenByParentCode(t *testing.T) {
	svc := newTestAddressService()

	cities, err := svc.ListCities("33")
	if err != nil {
		t.Fatalf("ListCities failed: %v", err)
	}
	if len(cities) != 1 || cities[0].Code != "3301" {
		t.Fatalf("unexpected cities: %+v", cities)
	}

	districts, err := svc.ListDistricts("3301")
	if err != nil {
		t.Fatalf("ListDistricts failed: %v", err)
	}
	if len(districts) != 1 || districts[0].Code != "330106" {
		t.Fatalf("unexpected districts: %+v", districts)
	}

	townships, err := svc.ListTownships("330106")
	if err != nil {
		t.Fatalf("ListTownships failed: %v", err)
	}
	if len(townships) != 1 || townships[0].Code != "330106001" {
		t.Fatalf("unexpected townships: %+v", townships)
	}

	villages, err := svc.ListVillages("330106001")
	if err != nil {
		t.Fatalf("ListVillages failed: %v", err)
	}
	if len(villages) != 1 || villages[0].Code != "330106001001" {
		t.Fatalf("unexpected villages: %+v", villages)
	}
}

func TestValidateAndNormalizeShippingAddressCanonicalizesFiveLevelAddress(t *testing.T) {
	svc := newTestAddressService()

	normalized, err := ValidateAndNormalizeShippingAddress(models.JSON{
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
		"village":        "错误社区",
		"village_code":   "330106001001",
		"detail_address": " 文三路100号 ",
	}, svc)
	if err != nil {
		t.Fatalf("ValidateAndNormalizeShippingAddress failed: %v", err)
	}

	if normalized["province"] != "浙江省" || normalized["city"] != "杭州市" {
		t.Fatalf("expected normalized province/city names, got %+v", normalized)
	}
	if normalized["township"] != "西溪街道" || normalized["village"] != "文一社区" {
		t.Fatalf("expected normalized township/village names, got %+v", normalized)
	}
	if normalized["detail_address"] != "文三路100号" {
		t.Fatalf("expected detail address to be trimmed, got %+v", normalized)
	}
}

func TestValidateAndNormalizeShippingAddressRejectsBrokenHierarchy(t *testing.T) {
	svc := newTestAddressService()

	_, err := ValidateAndNormalizeShippingAddress(models.JSON{
		"receiver_name":  "张三",
		"receiver_phone": "13800138000",
		"province_code":  "33",
		"city_code":      "3301",
		"district_code":  "330106",
		"township_code":  "330106001",
		"village_code":   "110101001001",
		"detail_address": "文三路100号",
	}, svc)
	if !errors.Is(err, ErrShippingAddressInvalid) {
		t.Fatalf("expected shipping address invalid, got %v", err)
	}
}
