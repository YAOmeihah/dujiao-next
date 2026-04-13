package public

import (
	"strings"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetProvinces(c *gin.Context) {
	if h.AddressService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	rows, err := h.AddressService.ListProvinces()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, rows)
}

func (h *Handler) GetCities(c *gin.Context) {
	h.respondAddressChildren(c, strings.TrimSpace(c.Query("province_code")), func(code string) (interface{}, error) {
		return h.AddressService.ListCities(code)
	})
}

func (h *Handler) GetDistricts(c *gin.Context) {
	h.respondAddressChildren(c, strings.TrimSpace(c.Query("city_code")), func(code string) (interface{}, error) {
		return h.AddressService.ListDistricts(code)
	})
}

func (h *Handler) GetTownships(c *gin.Context) {
	h.respondAddressChildren(c, strings.TrimSpace(c.Query("district_code")), func(code string) (interface{}, error) {
		return h.AddressService.ListTownships(code)
	})
}

func (h *Handler) GetVillages(c *gin.Context) {
	h.respondAddressChildren(c, strings.TrimSpace(c.Query("township_code")), func(code string) (interface{}, error) {
		return h.AddressService.ListVillages(code)
	})
}

func (h *Handler) respondAddressChildren(c *gin.Context, parentCode string, fetch func(string) (interface{}, error)) {
	if h.AddressService == nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	if parentCode == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	rows, err := fetch(parentCode)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, rows)
}
