package sync

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type Handler struct {
	uc usecase.SyncUsecase
}

func NewHandler(uc usecase.SyncUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Sync(c echo.Context) error {
	var req syncReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	merchantID := httputil.GetMerchantID(c)

	params := usecase.SyncParams{
		MerchantID: merchantID,
		BranchID:   req.BranchID,
	}
	if req.LastSync != nil {
		t, err := time.Parse(time.RFC3339Nano, *req.LastSync)
		if err == nil {
			params.LastSync = &t
		}
	}
	params.Products = req.Products
	params.Items = req.Items
	params.Categories = req.Categories

	result, err := h.uc.Sync(c.Request().Context(), params)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}

	return httputil.WriteJSON(c, http.StatusOK, syncResp{
		SyncToken:  result.SyncToken,
		Products:   result.Products,
		Items:      result.Items,
		Categories: result.Categories,
		Units:      result.Units,
		Barcodes:   result.Barcodes,
	})
}
