package postgres

import (
	"diploma/src/document"
	"diploma/src/messages"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UpdateRepository struct {
	roomName string
	pool     *pgxpool.Pool
}

func NewUpdateRepository(roomName string) *UpdateRepository {
	return &UpdateRepository{roomName: roomName, pool: Pool}
}

var _ document.UpdateRepository = (*UpdateRepository)(nil)

func (r *UpdateRepository) StoreUpdate(message *messages.UpdateMessage) error {
	ctx, cancel := timeoutCtx()
	defer cancel()

	statement := builder.Insert("updates").
		Columns("room_name", "client_id", "clock", "content").
		Values(r.roomName, message.ClientID, message.Clock, message.Data)

	sql, args, err := statement.ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *UpdateRepository) GetUpdates(request *messages.SyncReqMessage) ([][]byte, error) {
	// resulting query is seq scan, which is terribly slow in long sessions (I guess). Possible solutions:
	// 1) Optimize somehow with subqueries, cursors etc.
	// 2) Implement Epochs (requires custom TypeScript provider on frontend)
	// 3) Implement compression of updates to limit num of rows to seqscan from
	ctx, cancel := timeoutCtx()
	defer cancel()

	clientIDs := make([]uint64, 0, len(request.StateVector))

	statement := builder.Select("content").From("updates").Where(sq.Eq{"room_name": r.roomName})

	filter := sq.Or{}

	for _, vc := range request.StateVector {
		clientIDs = append(clientIDs, vc.ClientID)

		filter = append(filter, sq.And{sq.Eq{"client_id": vc.ClientID}, sq.GtOrEq{"clock": vc.Clock}})
	}

	filter = append(filter, sq.NotEq{"client_id": clientIDs})
	statement = statement.Where(filter)

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	updates, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) ([]byte, error) {
		var content pgtype.UndecodedBytes
		err := row.Scan(&content)
		if err != nil {
			return nil, err
		}
		return content, nil
	})
	if err != nil {
		return nil, err
	}

	return updates, nil
}

type RemovalRepository struct {
	roomName string
	pool     *pgxpool.Pool
}

func NewRemovalRepository(roomName string) *RemovalRepository {
	return &RemovalRepository{roomName: roomName, pool: Pool}
}

var _ document.RemovalRepository = (*RemovalRepository)(nil)

func (r *RemovalRepository) StoreRemoval(data []byte) error {
	ctx, cancel := timeoutCtx()
	defer cancel()

	statement := builder.Insert("removals").
		Columns("room_name", "content").
		Values(r.roomName, data)

	sql, args, err := statement.ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}
func (r *RemovalRepository) GetRemovals() ([][]byte, error) {
	ctx, cancel := timeoutCtx()
	defer cancel()

	statement := builder.Select("content").From("removals").Where(sq.Eq{"room_name": r.roomName})

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	removals, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) ([]byte, error) {
		var content pgtype.UndecodedBytes
		err := row.Scan(&content)
		if err != nil {
			return nil, err
		}
		return content, nil
	})
	if err != nil {
		return nil, err
	}

	return removals, nil
}
