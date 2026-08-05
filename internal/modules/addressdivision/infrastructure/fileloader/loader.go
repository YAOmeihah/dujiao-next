package fileloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	addresscontract "github.com/dujiao-next/internal/modules/addressdivision/contract"
)

const addressDivisionsDirEnv = "DJ_ADDRESS_DIVISIONS_DIR"

var (
	once          sync.Once
	cachedDataset addresscontract.Dataset
	cachedErr     error
)

type provinceRow struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type cityRow struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	ProvinceCode string `json:"provinceCode"`
}

type districtRow struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	CityCode     string `json:"cityCode"`
	ProvinceCode string `json:"provinceCode"`
}

type townshipRow struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	DistrictCode string `json:"areaCode"`
	CityCode     string `json:"cityCode"`
	ProvinceCode string `json:"provinceCode"`
}

// LoadDataset 返回外置文件中的五级行政区划数据集，并缓存默认加载结果。
func LoadDataset() (addresscontract.Dataset, error) {
	once.Do(func() {
		dir, err := resolveDatasetDir()
		if err != nil {
			cachedErr = err
			return
		}
		cachedDataset, cachedErr = LoadDatasetFromDir(dir)
	})
	return cachedDataset, cachedErr
}

// LoadDatasetFromDir 从指定目录读取五级行政区划数据文件。
func LoadDatasetFromDir(dir string) (addresscontract.Dataset, error) {
	baseDir := strings.TrimSpace(dir)
	if baseDir == "" {
		return addresscontract.Dataset{}, fmt.Errorf("address divisions dir is empty")
	}

	provinces, err := loadJSON[provinceRow](baseDir, "provinces.json")
	if err != nil {
		return addresscontract.Dataset{}, err
	}
	cities, err := loadJSON[cityRow](baseDir, "cities.json")
	if err != nil {
		return addresscontract.Dataset{}, err
	}
	districts, err := loadJSON[districtRow](baseDir, "districts.json")
	if err != nil {
		return addresscontract.Dataset{}, err
	}
	townships, err := loadJSON[townshipRow](baseDir, "townships.json")
	if err != nil {
		return addresscontract.Dataset{}, err
	}

	dataset := addresscontract.Dataset{
		Provinces: make([]addresscontract.Division, 0, len(provinces)),
		Cities:    make([]addresscontract.Division, 0, len(cities)),
		Districts: make([]addresscontract.Division, 0, len(districts)),
		Townships: make([]addresscontract.Division, 0, len(townships)),
	}
	for _, row := range provinces {
		dataset.Provinces = append(dataset.Provinces, addresscontract.Division{Code: row.Code, Name: row.Name})
	}
	for _, row := range cities {
		dataset.Cities = append(dataset.Cities, addresscontract.Division{Code: row.Code, Name: row.Name, ProvinceCode: row.ProvinceCode})
	}
	for _, row := range districts {
		dataset.Districts = append(dataset.Districts, addresscontract.Division{Code: row.Code, Name: row.Name, ProvinceCode: row.ProvinceCode, CityCode: row.CityCode})
	}
	for _, row := range townships {
		dataset.Townships = append(dataset.Townships, addresscontract.Division{Code: row.Code, Name: row.Name, ProvinceCode: row.ProvinceCode, CityCode: row.CityCode, DistrictCode: row.DistrictCode})
	}
	return dataset, nil
}

func resolveDatasetDir() (string, error) {
	if envDir := strings.TrimSpace(os.Getenv(addressDivisionsDirEnv)); envDir != "" {
		if dirExists(envDir) {
			return envDir, nil
		}
		return "", fmt.Errorf("address divisions dir from %s not found: %s", addressDivisionsDirEnv, envDir)
	}

	candidates := defaultDatasetDirCandidates()
	for _, candidate := range candidates {
		if dirExists(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("address divisions dir not found, tried: %v", candidates)
}

func defaultDatasetDirCandidates() []string {
	candidates := make([]string, 0, 3)
	if exePath, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exePath), "data", "address_divisions"))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "data", "address_divisions"),
			filepath.Join(wd, "internal", "seed", "address_divisions"),
		)
	}
	return dedupeStrings(candidates)
}

func loadJSON[T any](dir, name string) ([]T, error) {
	path := filepath.Join(dir, name)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var rows []T
	if err := json.Unmarshal(content, &rows); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}
	return rows, nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
