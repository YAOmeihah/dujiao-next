package memorystore

import (
	"sort"
	"strings"

	addresscontract "github.com/dujiao-next/internal/modules/addressdivision/contract"
)

type store struct {
	provinces        []addresscontract.Division
	citiesByProvince map[string][]addresscontract.Division
	districtsByCity  map[string][]addresscontract.Division
	townshipsByDist  map[string][]addresscontract.Division
	provinceByCode   map[string]addresscontract.Division
	cityByCode       map[string]addresscontract.Division
	districtByCode   map[string]addresscontract.Division
	townshipByCode   map[string]addresscontract.Division
}

// New 创建一个基于静态快照的仓储。
func New(dataset addresscontract.Dataset) addresscontract.Lookup {
	repo := &store{
		provinces:        cloneDivisions(dataset.Provinces),
		citiesByProvince: make(map[string][]addresscontract.Division),
		districtsByCity:  make(map[string][]addresscontract.Division),
		townshipsByDist:  make(map[string][]addresscontract.Division),
		provinceByCode:   make(map[string]addresscontract.Division),
		cityByCode:       make(map[string]addresscontract.Division),
		districtByCode:   make(map[string]addresscontract.Division),
		townshipByCode:   make(map[string]addresscontract.Division),
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
	for key := range repo.citiesByProvince {
		sortDivisions(repo.citiesByProvince[key])
	}
	for key := range repo.districtsByCity {
		sortDivisions(repo.districtsByCity[key])
	}
	for key := range repo.townshipsByDist {
		sortDivisions(repo.townshipsByDist[key])
	}
	return repo
}

func (r *store) ListProvinces() []addresscontract.Division {
	return cloneDivisions(r.provinces)
}

func (r *store) ListCities(provinceCode string) []addresscontract.Division {
	return cloneDivisions(r.citiesByProvince[strings.TrimSpace(provinceCode)])
}

func (r *store) ListDistricts(cityCode string) []addresscontract.Division {
	return cloneDivisions(r.districtsByCity[strings.TrimSpace(cityCode)])
}

func (r *store) ListTownships(districtCode string) []addresscontract.Division {
	return cloneDivisions(r.townshipsByDist[strings.TrimSpace(districtCode)])
}

func (r *store) GetProvince(code string) (addresscontract.Division, bool) {
	row, ok := r.provinceByCode[strings.TrimSpace(code)]
	return row, ok
}

func (r *store) GetCity(code string) (addresscontract.Division, bool) {
	row, ok := r.cityByCode[strings.TrimSpace(code)]
	return row, ok
}

func (r *store) GetDistrict(code string) (addresscontract.Division, bool) {
	row, ok := r.districtByCode[strings.TrimSpace(code)]
	return row, ok
}

func (r *store) GetTownship(code string) (addresscontract.Division, bool) {
	row, ok := r.townshipByCode[strings.TrimSpace(code)]
	return row, ok
}

func cloneDivisions(rows []addresscontract.Division) []addresscontract.Division {
	if len(rows) == 0 {
		return []addresscontract.Division{}
	}
	cloned := make([]addresscontract.Division, len(rows))
	copy(cloned, rows)
	return cloned
}

func sortDivisions(rows []addresscontract.Division) {
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Code < rows[j].Code
	})
}
