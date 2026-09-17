package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	*Queries
	db *pgxpool.Pool
}
func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
		Queries: New(db),
	}
}		
func (s *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}




func (s *Store) CreateHoldTx(ctx context.Context, arg CreateSeatHoldParams) (SeatHold, error) {
	var result SeatHold
	err := s.execTx(ctx, func(q *Queries) error {
		var err error
		_, err = q.LockSeat(ctx, arg.SeatID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrSeatNotFound
			}
			return err
		}
		_, err = q.GetConfirmedBooking(ctx, arg.SeatID)
		if err == nil {
			return ErrSeatAlreadyBooked
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		_, err = q.GetActiveSeatHold(ctx, arg.SeatID)
		if err == nil {
			return ErrSeatAlreadyOnHold
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		result, err = q.CreateSeatHold(ctx, arg)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
