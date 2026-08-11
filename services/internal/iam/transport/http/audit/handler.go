package audit

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/transport/http/common"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type Handler struct {
	authService usecase.AuthUsecase
}

func NewHandler(authService usecase.AuthUsecase) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) ListLogins(c echo.Context) error {
	raw := c.QueryParam("userIds")
	if raw == "" {
		return common.WriteError(c, http.StatusBadRequest, common.ErrMissingFields)
	}

	parts := strings.Split(raw, ",")
	userIDs := make([]int64, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil {
			return common.WriteError(c, http.StatusBadRequest, common.ErrInvalidID)
		}
		userIDs = append(userIDs, id)
	}

	p := httputil.ParsePageParams(c)
	data, meta, err := h.authService.ListLoginAudits(c.Request().Context(), userIDs, p)
	if err != nil {
		return common.WriteError(c, http.StatusInternalServerError, err)
	}
	return httputil.WritePaginated(c, http.StatusOK, data, *meta)
}
