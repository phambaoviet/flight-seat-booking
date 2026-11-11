package store

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	*sqlcdb.Queries
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db:      db,
		Queries: sqlcdb.New(db),
	}
}
func (s *Store) execTx(ctx context.Context, fn func(*sqlcdb.Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	q := sqlcdb.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}
func (s *Store) ReleaseHoldTx(ctx context.Context, holdID pgtype.UUID,) (sqlcdb.SeatHold, error) {
	var result sqlcdb.SeatHold
	err := s.execTx(ctx, func(q *sqlcdb.Queries) error {
		var err error
		result, err = q.ReleaseSeatHold(ctx, holdID)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrSeatHoldNotFound
		}
		if err != nil {
			return err
		}
		return nil
	})
	return result, err
}