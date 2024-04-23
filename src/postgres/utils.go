package postgres

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var Pool *pgxpool.Pool

var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func MakePgxPool(connStr string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	return pgxpool.NewWithConfig(context.Background(), cfg)
}

func timeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Millisecond*500)
}
