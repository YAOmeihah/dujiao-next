package application

import (
	"strings"

	addresscontract "github.com/dujiao-next/internal/modules/addressdivision/contract"
)

// Service 提供五级行政区划查询与校验能力。
type Service struct {
	repo addresscontract.Lookup
}

func NewService(repo addresscontract.Lookup) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListProvinces() ([]addresscontract.Division, error) {
	if s == nil || s.repo == nil {
		return []addresscontract.Division{}, nil
	}
	return s.repo.ListProvinces(), nil
}

func (s *Service) ListCities(provinceCode string) ([]addresscontract.Division, error) {
	if s == nil || s.repo == nil {
		return []addresscontract.Division{}, nil
	}
	return s.repo.ListCities(strings.TrimSpace(provinceCode)), nil
}

func (s *Service) ListDistricts(cityCode string) ([]addresscontract.Division, error) {
	if s == nil || s.repo == nil {
		return []addresscontract.Division{}, nil
	}
	return s.repo.ListDistricts(strings.TrimSpace(cityCode)), nil
}

func (s *Service) ListTownships(districtCode string) ([]addresscontract.Division, error) {
	if s == nil || s.repo == nil {
		return []addresscontract.Division{}, nil
	}
	return s.repo.ListTownships(strings.TrimSpace(districtCode)), nil
}

func (s *Service) GetProvince(code string) (addresscontract.Division, bool) {
	if s == nil || s.repo == nil {
		return addresscontract.Division{}, false
	}
	return s.repo.GetProvince(strings.TrimSpace(code))
}

func (s *Service) GetCity(code string) (addresscontract.Division, bool) {
	if s == nil || s.repo == nil {
		return addresscontract.Division{}, false
	}
	return s.repo.GetCity(strings.TrimSpace(code))
}

func (s *Service) GetDistrict(code string) (addresscontract.Division, bool) {
	if s == nil || s.repo == nil {
		return addresscontract.Division{}, false
	}
	return s.repo.GetDistrict(strings.TrimSpace(code))
}

func (s *Service) GetTownship(code string) (addresscontract.Division, bool) {
	if s == nil || s.repo == nil {
		return addresscontract.Division{}, false
	}
	return s.repo.GetTownship(strings.TrimSpace(code))
}
