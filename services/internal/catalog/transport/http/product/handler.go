package product

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/logger"
)

type Handler struct {
	uc      usecase.ProductUsecase
	catRepo repository.CategoryRepository
	itemRepo repository.SellableItemRepository
	unitRepo repository.UnitRepository
}

func NewHandler(uc usecase.ProductUsecase, catRepo repository.CategoryRepository, itemRepo repository.SellableItemRepository, unitRepo repository.UnitRepository) *Handler {
	return &Handler{uc: uc, catRepo: catRepo, itemRepo: itemRepo, unitRepo: unitRepo}
}

func (h *Handler) Create(c echo.Context) error {
	var req createProductReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" || req.CategoryID == 0 {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	merchantID := httputil.GetMerchantID(c)
	result, err := h.uc.Create(c.Request().Context(), usecase.CreateProductParams{
		MerchantID:  merchantID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusCreated, result)
}

func (h *Handler) BulkCreate(c echo.Context) error {
	var req createBulkReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if req.Name == "" || req.CategoryID == 0 {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}
	if len(req.Items) == 0 {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrMissingFields)
	}

	merchantID := httputil.GetMerchantID(c)

	// Create product first
	product, err := h.uc.Create(c.Request().Context(), usecase.CreateProductParams{
		MerchantID:  merchantID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		return mapError(c, err)
	}

	// Create each sellable item
	type itemResult struct {
		ID    int64   `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
	created := make([]itemResult, 0, len(req.Items))
	for _, item := range req.Items {
		if item.Name == "" || item.UnitID == 0 {
			continue
		}
		// We use the sellable item usecase via the handler's itemRepo directly
		si := &domain.SellableItem{
			ProductID:      product.ID,
			Name:           item.Name,
			UnitID:         item.UnitID,
			Price:          item.Price,
			TrackInventory: item.TrackInventory,
			ImageURL:       item.ImageURL,
			Status:         domain.SellableItemStatusActive,
		}
		now := time.Now().UTC()
		si.CreatedAt = now
		si.UpdatedAt = now
		if err := h.itemRepo.Create(c.Request().Context(), si); err != nil {
			logger.Error("product.BulkCreate: failed to create item", "error", err.Error())
			continue
		}
		created = append(created, itemResult{ID: si.ID, Name: si.Name, Price: si.Price})
	}

	return httputil.WriteJSON(c, http.StatusCreated, map[string]interface{}{
		"product": product,
		"items":   created,
	})
}

func (h *Handler) BulkUpdate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var req updateBulkReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	// Update product
	var status *domain.ProductStatus
	if req.Status != nil {
		s := domain.ProductStatus(*req.Status)
		status = &s
	}
	product, err := h.uc.Update(c.Request().Context(), usecase.UpdateProductParams{
		ID:          id,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Status:      status,
	})
	if err != nil {
		return mapError(c, err)
	}

	// Replace items if provided
	type itemResult struct {
		ID    int64   `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
	var created []itemResult
	if req.Items != nil {
		items := *req.Items
		// Delete existing items
		existing, err := h.itemRepo.ListByProduct(c.Request().Context(), id)
		if err == nil {
			for _, it := range existing {
				h.itemRepo.Delete(c.Request().Context(), it.ID)
			}
		}
		// Create new items
		now := time.Now().UTC()
		for _, item := range items {
			if item.Name == "" || item.UnitID == 0 {
				continue
			}
			si := &domain.SellableItem{
				ProductID:      product.ID,
				Name:           item.Name,
				UnitID:         item.UnitID,
				Price:          item.Price,
				TrackInventory: item.TrackInventory,
				ImageURL:       item.ImageURL,
				Status:         domain.SellableItemStatusActive,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := h.itemRepo.Create(c.Request().Context(), si); err != nil {
				logger.Error("product.BulkUpdate: failed to create item", "error", err.Error())
				continue
			}
			created = append(created, itemResult{ID: si.ID, Name: si.Name, Price: si.Price})
		}
	}

	return httputil.WriteJSON(c, http.StatusOK, map[string]interface{}{
		"product": product,
		"items":   created,
	})
}

func (h *Handler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	result, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) List(c echo.Context) error {
	p := httputil.ParsePageParams(c)
	search := c.QueryParam("search")
	merchantID := httputil.GetMerchantID(c)

	products, meta, err := h.uc.List(c.Request().Context(), merchantID, search, p)
	if err != nil {
		logger.Error("product.List failed", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}

	// Enrich: category name, price range, sellable items with units
	type unitResp struct {
		ID   int64  `json:"id"`
		Code string `json:"code"`
		Name string `json:"name"`
	}

	type itemResp struct {
		ID             int64    `json:"id"`
		Name           string   `json:"name"`
		Unit           unitResp `json:"unit"`
		Price          float64  `json:"price"`
		TrackInventory bool     `json:"trackInventory"`
		Status         string   `json:"status"`
	}

	type categoryResp struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	type priceRangeResp struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	}

	type productResp struct {
		ID          int64          `json:"id"`
		Name        string         `json:"name"`
		Description string         `json:"description,omitempty"`
		ImageURL    string         `json:"imageUrl,omitempty"`
		Status      string         `json:"status"`
		Category    categoryResp   `json:"category"`
		PriceRange  priceRangeResp `json:"priceRange"`
		Items       []itemResp     `json:"items"`
		CreatedAt   string         `json:"createdAt"`
		UpdatedAt   string         `json:"updatedAt"`
	}

	// Build category map
	categories, err := h.catRepo.ListByMerchant(c.Request().Context(), merchantID)
	if err != nil {
		logger.Error("product.List: failed to load categories", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
	catMap := make(map[int64]string)
	for _, cat := range categories {
		catMap[cat.ID] = cat.Name
	}

	// Build product ID list for batch item fetch
	prodIDs := make([]int64, len(products))
	for i, p := range products {
		prodIDs[i] = p.ID
	}

	// Fetch all items + units for these products
	itemMap := make(map[int64][]itemResp)
	for _, pid := range prodIDs {
		items, err := h.itemRepo.ListByProduct(c.Request().Context(), pid)
		if err != nil {
			logger.Error("product.List: failed to load items", "error", err.Error())
			continue
		}
		// Fetch units for items in this product
		unitCache := make(map[int64]unitResp)
		for _, it := range items {
			if _, ok := unitCache[it.UnitID]; !ok {
				u, err := h.unitRepo.GetByID(c.Request().Context(), it.UnitID)
				if err == nil {
					unitCache[it.UnitID] = unitResp{ID: u.ID, Code: u.Code, Name: u.Name}
				}
			}
		}
		var respItems []itemResp
		for _, it := range items {
			respItems = append(respItems, itemResp{
				ID:             it.ID,
				Name:           it.Name,
				Unit:           unitCache[it.UnitID],
				Price:          it.Price,
				TrackInventory: it.TrackInventory,
				Status:         string(it.Status),
			})
		}
		itemMap[pid] = respItems
	}

	result := make([]productResp, 0, len(products))
	for _, p := range products {
		items := itemMap[p.ID]
		if items == nil {
			items = []itemResp{}
		}
		minP, maxP := 0.0, 0.0
		for i, it := range items {
			if i == 0 || it.Price < minP {
				minP = it.Price
			}
			if i == 0 || it.Price > maxP {
				maxP = it.Price
			}
		}
		catName := catMap[p.CategoryID]

		result = append(result, productResp{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			ImageURL:    p.ImageURL,
			Status:      string(p.Status),
			Category:    categoryResp{ID: p.CategoryID, Name: catName},
			PriceRange:  priceRangeResp{Min: minP, Max: maxP},
			Items:       items,
			CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return httputil.WritePaginated(c, http.StatusOK, result, *meta)
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	var req updateProductReq
	if err := c.Bind(&req); err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}

	result, err := h.uc.Update(c.Request().Context(), usecase.UpdateProductParams{
		ID:          id,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Status:      req.Status,
	})
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, map[string]string{"message": "product deleted"})
}

func (h *Handler) Restore(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.WriteError(c, http.StatusBadRequest, httputil.ErrInvalidBody)
	}
	result, err := h.itemRepo.Restore(c.Request().Context(), id)
	if err != nil {
		return mapError(c, err)
	}
	return httputil.WriteJSON(c, http.StatusOK, result)
}

func mapError(c echo.Context, err error) error {
	switch err {
	case domain.ErrProductNotFound:
		return httputil.WriteError(c, http.StatusNotFound, err)
	case domain.ErrCategoryNotFound:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	case domain.ErrInvalidProduct:
		return httputil.WriteError(c, http.StatusBadRequest, err)
	default:
		logger.Error("product handler error", "error", err.Error())
		return httputil.WriteError(c, http.StatusInternalServerError, domain.ErrInternal)
	}
}
