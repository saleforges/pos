package unit

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type Handler struct {
	uc usecase.UnitUsecase
}

func NewHandler(uc usecase.UnitUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) List(c echo.Context) error {
	items, err := h.uc.List(c.Request().Context())
	if err != nil {
		logger.Error("unit.List failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
	return httputil.WriteJSON(c, http.StatusOK, items)
}
