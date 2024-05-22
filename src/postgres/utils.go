package postgres

import (
	"collab/src/config"
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var Pool *pgxpool.Pool

var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func MakePgxPool(cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.ConnURL)
	if err != nil {
		return nil, err
	}

	poolCfg.ConnConfig.User = cfg.Login
	poolCfg.ConnConfig.Password = cfg.Password
	poolCfg.ConnConfig.Database = cfg.DBName

	return pgxpool.NewWithConfig(context.Background(), poolCfg)
}

func timeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Millisecond*500)
}
