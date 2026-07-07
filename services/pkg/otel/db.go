package otel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type TracedPool struct {
	*pgxpool.Pool
	tracer trace.Tracer
}

func NewTracedPool(pool *pgxpool.Pool) *TracedPool {
	return &TracedPool{Pool: pool, tracer: tracerProvider()}
}

func (p *TracedPool) beginSpan(ctx context.Context, name, query string) (context.Context, trace.Span) {
	ctx, span := p.tracer.Start(ctx, name,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", query),
		),
	)
	return ctx, span
}

func (p *TracedPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	ctx, span := p.beginSpan(ctx, "db.Exec", sql)
	defer span.End()
	tag, err := p.Pool.Exec(ctx, sql, args...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	return tag, err
}

func (p *TracedPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	ctx, span := p.beginSpan(ctx, "db.Query", sql)
	defer span.End()
	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	return rows, err
}

func (p *TracedPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	ctx, span := p.beginSpan(ctx, "db.QueryRow", sql)
	defer span.End()
	return &tracedRow{row: p.Pool.QueryRow(ctx, sql, args...), span: span}
}

func (p *TracedPool) Begin(ctx context.Context) (*TracedTx, error) {
	ctx, span := p.tracer.Start(ctx, "db.Begin",
		trace.WithAttributes(attribute.String("db.system", "postgresql")),
	)
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, err
	}
	return &TracedTx{tx: tx, tracer: p.tracer, outer: span}, nil
}

type TracedTx struct {
	tx     pgx.Tx
	tracer trace.Tracer
	outer  trace.Span
}

func (t *TracedTx) beginSpan(ctx context.Context, name, query string) (context.Context, trace.Span) {
	ctx, span := t.tracer.Start(ctx, name,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", query),
		),
	)
	return ctx, span
}

func (t *TracedTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	ctx, span := t.beginSpan(ctx, "db.TxExec", sql)
	defer span.End()
	tag, err := t.tx.Exec(ctx, sql, args...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	return tag, err
}

func (t *TracedTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	ctx, span := t.beginSpan(ctx, "db.TxQuery", sql)
	defer span.End()
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	return rows, err
}

func (t *TracedTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	ctx, span := t.beginSpan(ctx, "db.TxQueryRow", sql)
	defer span.End()
	return &tracedRow{row: t.tx.QueryRow(ctx, sql, args...), span: span}
}

func (t *TracedTx) Commit(ctx context.Context) error {
	ctx, span := t.tracer.Start(ctx, "db.Commit",
		trace.WithAttributes(attribute.String("db.system", "postgresql")),
	)
	defer span.End()
	defer t.outer.End()
	if err := t.tx.Commit(ctx); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (t *TracedTx) Rollback(ctx context.Context) error {
	ctx, span := t.tracer.Start(ctx, "db.Rollback",
		trace.WithAttributes(attribute.String("db.system", "postgresql")),
	)
	defer span.End()
	defer t.outer.End()
	return t.tx.Rollback(ctx)
}

type tracedRow struct {
	row  pgx.Row
	span trace.Span
}

func (r *tracedRow) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	if err != nil && err != pgx.ErrNoRows {
		r.span.SetStatus(codes.Error, err.Error())
	}
	return err
}
