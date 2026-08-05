package addresshttp

import (
	"strings"

	addressapp "github.com/dujiao-next/internal/modules/addressdivision/application"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	addresses *addressapp.Service
}

func NewHandler(addresses *addressapp.Service) *Handler {
	return &Handler{addresses: addresses}
}

func (h *Handler) GetProvinces(c *gin.Context) {
	if h == nil || h.addresses == nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	rows, err := h.addresses.ListProvinces()
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, rows)
}

func (h *Handler) GetCities(c *gin.Context) {
	h.respondChildren(c, strings.TrimSpace(c.Query("province_code")), func(code string) (interface{}, error) {
		return h.addresses.ListCities(code)
	})
}

func (h *Handler) GetDistricts(c *gin.Context) {
	h.respondChildren(c, strings.TrimSpace(c.Query("city_code")), func(code string) (interface{}, error) {
		return h.addresses.ListDistricts(code)
	})
}

func (h *Handler) GetTownships(c *gin.Context) {
	h.respondChildren(c, strings.TrimSpace(c.Query("district_code")), func(code string) (interface{}, error) {
		return h.addresses.ListTownships(code)
	})
}

func (h *Handler) respondChildren(c *gin.Context, parentCode string, fetch func(string) (interface{}, error)) {
	if h == nil || h.addresses == nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal_error", nil)
		return
	}
	if parentCode == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	rows, err := fetch(parentCode)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, rows)
}

func RegisterPublicRoutes(group *gin.RouterGroup, handler *Handler) {
	group.GET("/address/provinces", handler.GetProvinces)
	group.GET("/address/cities", handler.GetCities)
	group.GET("/address/districts", handler.GetDistricts)
	group.GET("/address/townships", handler.GetTownships)
}
