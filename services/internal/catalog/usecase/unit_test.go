package usecase

import (
	"context"
	"testing"
)

func TestUnitUsecase_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("returns all units", func(t *testing.T) {
		uc := NewUnitUsecase(newMockUnitRepo())
		units, err := uc.List(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(units) != 2 {
			t.Errorf("expected 2 units, got %d", len(units))
		}
	})
}
