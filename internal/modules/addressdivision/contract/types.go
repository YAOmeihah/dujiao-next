package contract

// Division 表示一个行政区划节点。
type Division struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	ProvinceCode string `json:"province_code,omitempty"`
	CityCode     string `json:"city_code,omitempty"`
	DistrictCode string `json:"district_code,omitempty"`
	TownshipCode string `json:"township_code,omitempty"`
}

// Dataset 是五级行政区划的静态数据快照。
type Dataset struct {
	Provinces []Division
	Cities    []Division
	Districts []Division
	Townships []Division
}

// Lookup 提供行政区划查询能力。
type Lookup interface {
	ListProvinces() []Division
	ListCities(provinceCode string) []Division
	ListDistricts(cityCode string) []Division
	ListTownships(districtCode string) []Division
	GetProvince(code string) (Division, bool)
	GetCity(code string) (Division, bool)
	GetDistrict(code string) (Division, bool)
	GetTownship(code string) (Division, bool)
}
