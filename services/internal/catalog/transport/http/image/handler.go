package image

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/adapter/storage/minio"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type Handler struct {
	store *minio.Client
}

func NewHandler(store *minio.Client) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Upload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	src, err := file.Open()
	if err != nil {
		logger.Error("image.Upload: failed to open file", "error", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		logger.Error("image.Upload: failed to read file", "error", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	merchantID := httputil.GetMerchantID(c)
	folder := path.Join("catalog", "merchant", fmt.Sprintf("%d", merchantID))

	url, err := h.store.Upload(c.Request().Context(), folder, bytes.NewReader(data), int64(len(data)), contentType)
	if err != nil {
		logger.Error("image.Upload: store upload failed", "error", err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	return httputil.WriteJSON(c, http.StatusCreated, map[string]string{"url": url})
}
