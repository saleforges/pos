package port

import "context"

// TxBeginner can begin a database transaction.
// The returned Tx exposes Commit and Rollback and can be used as a Querier.
type TxBeginner interface {
	Begin(ctx context.Context) (Tx, error)
}

// Tx wraps Commit/Rollback lifecycle.
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
