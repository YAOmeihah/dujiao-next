package repository

import (
	"sort"
	"strings"

	"github.com/dujiao-next/internal/models"
)

// AddressDivisionDataset 是五级行政区划的静态数据快照。
type AddressDivisionDataset struct {
	Provinces []models.AddressDivision
	Cities    []models.AddressDivision
	Districts []models.AddressDivision
	Townships []models.AddressDivision
	Villages  []models.AddressDivision
}

// AddressDivisionRepository 提供内存中的行政区划查询。
type AddressDivisionRepository interface {
	ListProvinces() []models.AddressDivision
	ListCities(provinceCode string) []models.AddressDivision
	ListDistricts(cityCode string) []models.AddressDivision
	ListTownships(districtCode string) []models.AddressDivision
	ListVillages(townshipCode string) []models.AddressDivision
	GetProvince(code string) (models.AddressDivision, bool)
	GetCity(code string) (models.AddressDivision, bool)
	GetDistrict(code string) (models.AddressDivision, bool)
	GetTownship(code string) (models.AddressDivision, bool)
	GetVillage(code string) (models.AddressDivision, bool)
}

type addressDivisionRepository struct {
	provinces        []models.AddressDivision
	citiesByProvince map[string][]models.AddressDivision
	districtsByCity  map[string][]models.AddressDivision
	townshipsByDist  map[string][]models.AddressDivision
	villagesByTown   map[string][]models.AddressDivision
	provinceByCode   map[string]models.AddressDivision
	cityByCode       map[string]models.AddressDivision
	districtByCode   map[string]models.AddressDivision
	townshipByCode   map[string]models.AddressDivision
	villageByCode    map[string]models.AddressDivision
}

// NewAddressDivisionRepository 创建一个基于静态快照的仓储。
func NewAddressDivisionRepository(dataset AddressDivisionDataset) AddressDivisionRepository {
	repo := &addressDivisionRepository{
		provinces:        cloneDivisions(dataset.Provinces),
		citiesByProvince: make(map[string][]models.AddressDivision),
		districtsByCity:  make(map[string][]models.AddressDivision),
		townshipsByDist:  make(map[string][]models.AddressDivision),
		villagesByTown:   make(map[string][]models.AddressDivision),
		provinceByCode:   make(map[string]models.AddressDivision),
		cityByCode:       make(map[string]models.AddressDivision),
		districtByCode:   make(map[string]models.AddressDivision),
		townshipByCode:   make(map[string]models.AddressDivision),
		villageByCode:    make(map[string]models.AddressDivision),
	}

	sortDivisions(repo.provinces)
	for _, row := range repo.provinces {
		repo.provinceByCode[row.Code] = row
	}
	for _, row := range cloneDivisions(dataset.Cities) {
		repo.cityByCode[row.Code] = row
		repo.citiesByProvince[row.ProvinceCode] = append(repo.citiesByProvince[row.ProvinceCode], row)
	}
	for _, row := range cloneDivisions(dataset.Districts) {
		repo.districtByCode[row.Code] = row
		repo.districtsByCity[row.CityCode] = append(repo.districtsByCity[row.CityCode], row)
	}
	for _, row := range cloneDivisions(dataset.Townships) {
		repo.townshipByCode[row.Code] = row
		repo.townshipsByDist[row.DistrictCode] = append(repo.townshipsByDist[row.DistrictCode], row)
	}
	for _, row := range cloneDivisions(dataset.Villages) {
		repo.villageByCode[row.Code] = row
		repo.villagesByTown[row.TownshipCode] = append(repo.villagesByTown[row.TownshipCode], row)
	}
	for key := range repo.citiesByProvince {
		sortDivisions(repo.citiesByProvince[key])
	}
	for key := range repo.districtsByCity {
		sortDivisions(repo.districtsByCity[key])
	}
	for key := range repo.townshipsByDist {
		sortDivisions(repo.townshipsByDist[key])
	}
	for key := range repo.villagesByTown {
		sortDivisions(repo.villagesByTown[key])
	}
	return repo
}

func (r *addressDivisionRepository) ListProvinces() []models.AddressDivision {
	return cloneDivisions(r.provinces)
}

func (r *addressDivisionRepository) ListCities(provinceCode string) []models.AddressDivision {
	return cloneDivisions(r.citiesByProvince[strings.TrimSpace(provinceCode)])
}

func (r *addressDivisionRepository) ListDistricts(cityCode string) []models.AddressDivision {
	return cloneDivisions(r.districtsByCity[strings.TrimSpace(cityCode)])
}

func (r *addressDivisionRepository) ListTownships(districtCode string) []models.AddressDivision {
	return cloneDivisions(r.townshipsByDist[strings.TrimSpace(districtCode)])
}

func (r *addressDivisionRepository) ListVillages(townshipCode string) []models.AddressDivision {
	return cloneDivisions(r.villagesByTown[strings.TrimSpace(townshipCode)])
}

func (r *addressDivisionRepository) GetProvince(code string) (models.AddressDivision, bool) {
	row, ok := r.provinceByCode[strings.TrimSpace(code)]
	return row, ok
}

func (r *addressDivisionRepository) GetCity(code string) (models.AddressDivision, bool) {
	row, ok := r.cityByCode[strings.TrimSpace(code)]
	return row, ok
}

func (r *addressDivisionRepository) GetDistrict(code string) (models.AddressDivision, bool) {
	row, ok := r.districtByCode[strings.TrimSpace(code)]
	return row, ok
}

func (r *addressDivisionRepository) GetTownship(code string) (models.AddressDivision, bool) {
	row, ok := r.townshipByCode[strings.TrimSpace(code)]
	return row, ok
}

func (r *addressDivisionRepository) GetVillage(code string) (models.AddressDivision, bool) {
	row, ok := r.villageByCode[strings.TrimSpace(code)]
	return row, ok
}

func cloneDivisions(rows []models.AddressDivision) []models.AddressDivision {
	if len(rows) == 0 {
		return []models.AddressDivision{}
	}
	cloned := make([]models.AddressDivision, len(rows))
	copy(cloned, rows)
	return cloned
}

func sortDivisions(rows []models.AddressDivision) {
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Code < rows[j].Code
	})
}
