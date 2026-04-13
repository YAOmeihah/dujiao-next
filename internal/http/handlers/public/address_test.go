package public

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/provider"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
)

func newAddressTestHandler() *Handler {
	addressService := service.NewAddressService(repository.NewAddressDivisionRepository(repository.AddressDivisionDataset{
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
	return &Handler{Container: &provider.Container{AddressService: addressService}}
}

func TestGetVillagesByTownshipCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newAddressTestHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api/v1/address/villages", h.GetVillages)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/address/villages?township_code=330106001", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "文一社区") {
		t.Fatalf("expected village in response, got %s", body)
	}
}

func TestGetCitiesRequiresProvinceCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newAddressTestHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/api/v1/address/cities", h.GetCities)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/address/cities", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"status_code":`+`400`) && !strings.Contains(body, `"status_code":400`) {
		t.Fatalf("expected business bad request, got %s", body)
	}
	if !strings.Contains(body, `"msg"`) {
		t.Fatalf("expected error message, got %s", body)
	}
	if response.CodeBadRequest != 400 {
		t.Fatalf("unexpected bad request code constant: %d", response.CodeBadRequest)
	}
}
