package store

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateHoldTx(ctx context.Context, arg sqlcdb.CreateSeatHoldParams) (sqlcdb.SeatHold, error) {
	var result sqlcdb.SeatHold
	err := s.execTx(ctx, func(q *sqlcdb.Queries) error {
		var err error
		_, err = q.LockSeat(ctx, arg.SeatID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrSeatNotFound
			}
			return err
		}
		_, err = q.GetConfirmedBooking(ctx, arg.SeatID)
		if err == nil {
			return domain.ErrSeatAlreadyBooked
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		_, err = q.GetActiveSeatHold(ctx, arg.SeatID)
		if err == nil {
			return domain.ErrSeatAlreadyOnHold
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
