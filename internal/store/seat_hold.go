package store

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateHoldTx(ctx context.Context, arg sqlcdb.CreateSeatHoldParams) (sqlcdb.SeatHold, error) {
	var result sqlcdb.SeatHold
	err := s.execTx(ctx, func(q *sqlcdb.Queries) error {
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
