package service

import (
	"strings"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
)

// AddressService 提供五级行政区划查询与校验能力。
type AddressService struct {
	repo repository.AddressDivisionRepository
}

func NewAddressService(repo repository.AddressDivisionRepository) *AddressService {
	return &AddressService{repo: repo}
}

func (s *AddressService) ListProvinces() ([]models.AddressDivision, error) {
	if s == nil || s.repo == nil {
		return []models.AddressDivision{}, nil
	}
	return s.repo.ListProvinces(), nil
}

func (s *AddressService) ListCities(provinceCode string) ([]models.AddressDivision, error) {
	if s == nil || s.repo == nil {
		return []models.AddressDivision{}, nil
	}
	return s.repo.ListCities(strings.TrimSpace(provinceCode)), nil
}

func (s *AddressService) ListDistricts(cityCode string) ([]models.AddressDivision, error) {
	if s == nil || s.repo == nil {
		return []models.AddressDivision{}, nil
	}
	return s.repo.ListDistricts(strings.TrimSpace(cityCode)), nil
}

func (s *AddressService) ListTownships(districtCode string) ([]models.AddressDivision, error) {
	if s == nil || s.repo == nil {
		return []models.AddressDivision{}, nil
	}
	return s.repo.ListTownships(strings.TrimSpace(districtCode)), nil
}

func (s *AddressService) ListVillages(townshipCode string) ([]models.AddressDivision, error) {
	if s == nil || s.repo == nil {
		return []models.AddressDivision{}, nil
	}
	return s.repo.ListVillages(strings.TrimSpace(townshipCode)), nil
}

func (s *AddressService) GetProvince(code string) (models.AddressDivision, bool) {
	if s == nil || s.repo == nil {
		return models.AddressDivision{}, false
	}
	return s.repo.GetProvince(strings.TrimSpace(code))
}

func (s *AddressService) GetCity(code string) (models.AddressDivision, bool) {
	if s == nil || s.repo == nil {
		return models.AddressDivision{}, false
	}
	return s.repo.GetCity(strings.TrimSpace(code))
}

func (s *AddressService) GetDistrict(code string) (models.AddressDivision, bool) {
	if s == nil || s.repo == nil {
		return models.AddressDivision{}, false
	}
	return s.repo.GetDistrict(strings.TrimSpace(code))
}

func (s *AddressService) GetTownship(code string) (models.AddressDivision, bool) {
	if s == nil || s.repo == nil {
		return models.AddressDivision{}, false
	}
	return s.repo.GetTownship(strings.TrimSpace(code))
}

func (s *AddressService) GetVillage(code string) (models.AddressDivision, bool) {
	if s == nil || s.repo == nil {
		return models.AddressDivision{}, false
	}
	return s.repo.GetVillage(strings.TrimSpace(code))
}
