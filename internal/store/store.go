package store

import (
	"context"
	sqlcdb "flight-booking/db/sqlc"

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
