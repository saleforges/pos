package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
)

type mockShiftRepo struct {
	shifts       map[int64]*domain.Shift
	seq          int64
	cashPayments float64
	err          error
}

func (m *mockShiftRepo) Create(_ context.Context, s *domain.Shift) error {
	if m.err != nil {
		return m.err
	}
	if m.shifts == nil {
		m.shifts = make(map[int64]*domain.Shift)
	}
	m.seq++
	s.ID = m.seq
	m.shifts[s.ID] = s
	return nil
}

func (m *mockShiftRepo) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Shift, error) {
	if m.err != nil {
		return nil, m.err
	}
	s, ok := m.shifts[id]
	if !ok || s.MerchantID != merchantID {
		return nil, domain.ErrShiftNotFound
	}
	return s, nil
}

func (m *mockShiftRepo) GetOpenByBranch(_ context.Context, merchantID, branchID int64) (*domain.Shift, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, s := range m.shifts {
		if s.MerchantID == merchantID && s.BranchID == branchID && s.Status == domain.ShiftStatusOpen {
			return s, nil
		}
	}
	return nil, domain.ErrShiftNotFound
}

func (m *mockShiftRepo) Update(_ context.Context, s *domain.Shift) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.shifts[s.ID]; !ok {
		return domain.ErrShiftNotFound
	}
	m.shifts[s.ID] = s
	return nil
}

func (m *mockShiftRepo) List(_ context.Context, merchantID, branchID int64) ([]domain.Shift, error) {
	var result []domain.Shift
	for _, s := range m.shifts {
		if s.MerchantID != merchantID {
			continue
		}
		if branchID > 0 && s.BranchID != branchID {
			continue
		}
		result = append(result, *s)
	}
	return result, nil
}

func (m *mockShiftRepo) SumCashPayments(_ context.Context, _, _ int64, _, _ time.Time) (float64, error) {
	return m.cashPayments, nil
}

func TestShiftUsecase_Open(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("opens a shift for the branch", func(t *testing.T) {
		uc := NewShiftUsecase(&mockShiftRepo{})
		shift, err := uc.Open(ctx, OpenShiftParams{MerchantID: 1, BranchID: 1, OpenedBy: 5, StartingCash: 100000})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if shift.Status != domain.ShiftStatusOpen {
			t.Errorf("expected status open, got %s", shift.Status)
		}
		if shift.StartingCash != 100000 {
			t.Errorf("expected startingCash 100000, got %f", shift.StartingCash)
		}
	})

	t.Run("rejected when a shift is already open for the branch", func(t *testing.T) {
		repo := &mockShiftRepo{}
		uc := NewShiftUsecase(repo)
		if _, err := uc.Open(ctx, OpenShiftParams{MerchantID: 1, BranchID: 1, OpenedBy: 5, StartingCash: 100000}); err != nil {
			t.Fatalf("unexpected error on first open: %v", err)
		}
		_, err := uc.Open(ctx, OpenShiftParams{MerchantID: 1, BranchID: 1, OpenedBy: 6, StartingCash: 50000})
		if err != domain.ErrShiftAlreadyOpen {
			t.Errorf("expected ErrShiftAlreadyOpen, got %v", err)
		}
	})

	t.Run("different branches can each have their own open shift", func(t *testing.T) {
		repo := &mockShiftRepo{}
		uc := NewShiftUsecase(repo)
		if _, err := uc.Open(ctx, OpenShiftParams{MerchantID: 1, BranchID: 1, OpenedBy: 5, StartingCash: 100000}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := uc.Open(ctx, OpenShiftParams{MerchantID: 1, BranchID: 2, OpenedBy: 5, StartingCash: 100000}); err != nil {
			t.Errorf("expected branch 2 to open independently, got %v", err)
		}
	})
}

func TestShiftUsecase_Close(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("computes expected cash and variance from cash sales", func(t *testing.T) {
		repo := &mockShiftRepo{cashPayments: 50000}
		uc := NewShiftUsecase(repo)
		opened, err := uc.Open(ctx, OpenShiftParams{MerchantID: 1, BranchID: 1, OpenedBy: 5, StartingCash: 100000})
		if err != nil {
			t.Fatalf("unexpected error opening: %v", err)
		}

		closed, err := uc.Close(ctx, CloseShiftParams{ShiftID: opened.ID, MerchantID: 1, ClosedBy: 6, ActualCash: 145000})
		if err != nil {
			t.Fatalf("unexpected error closing: %v", err)
		}
		if closed.Status != domain.ShiftStatusClosed {
			t.Errorf("expected status closed, got %s", closed.Status)
		}
		if closed.ExpectedCash == nil || *closed.ExpectedCash != 150000 {
			t.Errorf("expected expectedCash 150000 (100000 starting + 50000 cash sales), got %v", closed.ExpectedCash)
		}
		if closed.ActualCash == nil || *closed.ActualCash != 145000 {
			t.Errorf("expected actualCash 145000, got %v", closed.ActualCash)
		}
		if closed.Variance == nil || *closed.Variance != -5000 {
			t.Errorf("expected variance -5000 (145000-150000), got %v", closed.Variance)
		}
		if closed.ClosedBy == nil || *closed.ClosedBy != 6 {
			t.Errorf("expected closedBy 6, got %v", closed.ClosedBy)
		}
	})

	t.Run("closing again is rejected", func(t *testing.T) {
		repo := &mockShiftRepo{}
		uc := NewShiftUsecase(repo)
		opened, _ := uc.Open(ctx, OpenShiftParams{MerchantID: 1, BranchID: 1, OpenedBy: 5, StartingCash: 100000})
		if _, err := uc.Close(ctx, CloseShiftParams{ShiftID: opened.ID, MerchantID: 1, ClosedBy: 6, ActualCash: 100000}); err != nil {
			t.Fatalf("unexpected error on first close: %v", err)
		}
		_, err := uc.Close(ctx, CloseShiftParams{ShiftID: opened.ID, MerchantID: 1, ClosedBy: 6, ActualCash: 100000})
		if err != domain.ErrShiftAlreadyClosed {
			t.Errorf("expected ErrShiftAlreadyClosed, got %v", err)
		}
	})

	t.Run("negative actual cash rejected", func(t *testing.T) {
		repo := &mockShiftRepo{}
		uc := NewShiftUsecase(repo)
		opened, _ := uc.Open(ctx, OpenShiftParams{MerchantID: 1, BranchID: 1, OpenedBy: 5, StartingCash: 100000})
		_, err := uc.Close(ctx, CloseShiftParams{ShiftID: opened.ID, MerchantID: 1, ClosedBy: 6, ActualCash: -1})
		if err != domain.ErrInvalidShift {
			t.Errorf("expected ErrInvalidShift, got %v", err)
		}
	})

	t.Run("unknown shift returns not found", func(t *testing.T) {
		uc := NewShiftUsecase(&mockShiftRepo{})
		_, err := uc.Close(ctx, CloseShiftParams{ShiftID: 999, MerchantID: 1, ClosedBy: 6, ActualCash: 100000})
		if err != domain.ErrShiftNotFound {
			t.Errorf("expected ErrShiftNotFound, got %v", err)
		}
	})
}

func TestShiftUsecase_GetActive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("returns not found when no shift is open", func(t *testing.T) {
		uc := NewShiftUsecase(&mockShiftRepo{})
		_, err := uc.GetActive(ctx, 1, 1)
		if err != domain.ErrShiftNotFound {
			t.Errorf("expected ErrShiftNotFound, got %v", err)
		}
	})

	t.Run("returns the open shift", func(t *testing.T) {
		repo := &mockShiftRepo{}
		uc := NewShiftUsecase(repo)
		opened, _ := uc.Open(ctx, OpenShiftParams{MerchantID: 1, BranchID: 1, OpenedBy: 5, StartingCash: 100000})
		active, err := uc.GetActive(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if active.ID != opened.ID {
			t.Errorf("expected shift %d, got %d", opened.ID, active.ID)
		}
	})
}
