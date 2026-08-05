package addresshttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	addressapp "github.com/dujiao-next/internal/modules/addressdivision/application"
	addresscontract "github.com/dujiao-next/internal/modules/addressdivision/contract"

	"github.com/gin-gonic/gin"
)

type addressTestLookup struct {
	dataset addresscontract.Dataset
}

func (l *addressTestLookup) ListProvinces() []addresscontract.Division {
	return l.dataset.Provinces
}

func (l *addressTestLookup) ListCities(provinceCode string) []addresscontract.Division {
	return filterAddressTestDivisions(l.dataset.Cities, func(item addresscontract.Division) bool {
		return item.ProvinceCode == provinceCode
	})
}

func (l *addressTestLookup) ListDistricts(cityCode string) []addresscontract.Division {
	return filterAddressTestDivisions(l.dataset.Districts, func(item addresscontract.Division) bool {
		return item.CityCode == cityCode
	})
}

func (l *addressTestLookup) ListTownships(districtCode string) []addresscontract.Division {
	return filterAddressTestDivisions(l.dataset.Townships, func(item addresscontract.Division) bool {
		return item.DistrictCode == districtCode
	})
}

func (l *addressTestLookup) GetProvince(code string) (addresscontract.Division, bool) {
	return findAddressTestDivision(l.dataset.Provinces, code)
}

func (l *addressTestLookup) GetCity(code string) (addresscontract.Division, bool) {
	return findAddressTestDivision(l.dataset.Cities, code)
}

func (l *addressTestLookup) GetDistrict(code string) (addresscontract.Division, bool) {
	return findAddressTestDivision(l.dataset.Districts, code)
}

func (l *addressTestLookup) GetTownship(code string) (addresscontract.Division, bool) {
	return findAddressTestDivision(l.dataset.Townships, code)
}

func filterAddressTestDivisions(items []addresscontract.Division, keep func(addresscontract.Division) bool) []addresscontract.Division {
	result := make([]addresscontract.Division, 0)
	for _, item := range items {
		if keep(item) {
			result = append(result, item)
		}
	}
	return result
}

func findAddressTestDivision(items []addresscontract.Division, code string) (addresscontract.Division, bool) {
	for _, item := range items {
		if item.Code == code {
			return item, true
		}
	}
	return addresscontract.Division{}, false
}

func newAddressTestHandler() *Handler {
	addresses := addressapp.NewService(&addressTestLookup{dataset: addresscontract.Dataset{
		Provinces: []addresscontract.Division{{Code: "33", Name: "浙江省"}},
		Cities:    []addresscontract.Division{{Code: "3301", Name: "杭州市", ProvinceCode: "33"}},
		Districts: []addresscontract.Division{{Code: "330106", Name: "西湖区", ProvinceCode: "33", CityCode: "3301"}},
		Townships: []addresscontract.Division{{Code: "330106001", Name: "西溪街道", ProvinceCode: "33", CityCode: "3301", DistrictCode: "330106"}},
	}})
	return NewHandler(addresses)
}

func TestGetTownshipsByDistrictCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newAddressTestHandler()
	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api/v1/public/address/townships", h.GetTownships)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/address/townships?district_code=330106", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "西溪街道") {
		t.Fatalf("expected township in response, got %s", w.Body.String())
	}
}

func TestGetCitiesRequiresProvinceCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newAddressTestHandler()
	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api/v1/public/address/cities", h.GetCities)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/address/cities", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"status_code":400`) {
		t.Fatalf("expected business bad request, got %s", w.Body.String())
	}
}
