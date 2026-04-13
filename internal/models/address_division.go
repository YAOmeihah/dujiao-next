package models

// AddressDivision 表示一个行政区划节点。
type AddressDivision struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	ProvinceCode string `json:"province_code,omitempty"`
	CityCode     string `json:"city_code,omitempty"`
	DistrictCode string `json:"district_code,omitempty"`
	TownshipCode string `json:"township_code,omitempty"`
}
