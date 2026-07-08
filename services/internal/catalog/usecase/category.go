package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type categoryUsecase struct {
	catRepo repository.CategoryRepository
}

func NewCategoryUsecase(catRepo repository.CategoryRepository) CategoryUsecase {
	return &categoryUsecase{catRepo: catRepo}
}

func (uc *categoryUsecase) Create(ctx context.Context, input CreateCategoryInput) (*domain.Category, error) {
	ctx, span := otel.StartSpan(ctx, "category.Create")
	defer span.End()

	cat := &domain.Category{
		ID:          uuid.NewString()[:14],
		MerchantID:  input.MerchantID,
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		ParentID:    input.ParentID,
		SortOrder:   input.SortOrder,
		Status:      domain.CategoryStatusActive,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := uc.catRepo.Create(ctx, cat); err != nil {
		logger.Error("category.Create: create failed", "error", err.Error())
		return nil, domain.ErrInternal
	}
	return cat, nil
}

func (uc *categoryUsecase) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	ctx, span := otel.StartSpan(ctx, "category.GetByID")
	defer span.End()

	cat, err := uc.catRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return cat, nil
}

func (uc *categoryUsecase) List(ctx context.Context, merchantID string, search string, offset, limit int) (*PaginatedResult[domain.Category], error) {
	ctx, span := otel.StartSpan(ctx, "category.List")
	defer span.End()

	total, err := uc.catRepo.Count(ctx, merchantID, search)
	if err != nil {
		return nil, err
	}

	items, err := uc.catRepo.List(ctx, merchantID, search, offset, limit)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[domain.Category]{
		Items: items,
		Meta: PaginationMeta{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}, nil
}

func (uc *categoryUsecase) Update(ctx context.Context, input UpdateCategoryInput) (*domain.Category, error) {
	ctx, span := otel.StartSpan(ctx, "category.Update")
	defer span.End()

	cat, err := uc.catRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		cat.Name = *input.Name
	}
	if input.Slug != nil {
		cat.Slug = *input.Slug
	}
	if input.Description != nil {
		cat.Description = *input.Description
	}
	if input.ParentID != nil {
		cat.ParentID = input.ParentID
	}
	if input.SortOrder != nil {
		cat.SortOrder = *input.SortOrder
	}
	if input.Status != nil {
		cat.Status = *input.Status
	}
	cat.UpdatedAt = time.Now().UTC()

	if err := uc.catRepo.Update(ctx, cat); err != nil {
		logger.Error("category.Update: update failed", "error", err.Error())
		return nil, domain.ErrInternal
	}
	return cat, nil
}

func (uc *categoryUsecase) Delete(ctx context.Context, id string) error {
	ctx, span := otel.StartSpan(ctx, "category.Delete")
	defer span.End()

	_, err := uc.catRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return uc.catRepo.Delete(ctx, id)
}
