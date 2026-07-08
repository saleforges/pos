package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/adapter/storage/minio"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type ImageHandler struct {
	storage *minio.Client
}

func NewImageHandler(storage *minio.Client) *ImageHandler {
	return &ImageHandler{storage: storage}
}

func (h *ImageHandler) Upload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	src, err := file.Open()
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	folder := c.FormValue("folder")
	if folder == "" {
		folder = "uploads"
	}
	folder = strings.Trim(folder, "/")

	url, err := h.storage.Upload(c.Request().Context(), folder, src, file.Size, contentType)
	if err != nil {
		return httputil.WriteError(c, http.StatusInternalServerError, err)
	}

	return httputil.WriteJSON(c, http.StatusCreated, map[string]string{"url": url})
}
