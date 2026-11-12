package store

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Store) CreateHoldTx(ctx context.Context, arg sqlcdb.CreateSeatHoldParams, flightID pgtype.UUID) (sqlcdb.SeatHold, error) {
	var result sqlcdb.SeatHold
	err := s.execTx(ctx, func(q *sqlcdb.Queries) error {
		var err error
		// Lock the seat to prevent concurrent holds
		_, err = q.LockSeatForHold(ctx, sqlcdb.LockSeatForHoldParams{
			ID:   arg.SeatID,
			FlightID: flightID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrSeatNotFound
			}
			return err
		}

		// Expire old seat holds for the seat
		err = q.ExpireOldSeatHolds(ctx, arg.SeatID)
		if err != nil {
			return err
		}

		// Check if the seat is already booked
		_, err = q.GetConfirmedBooking(ctx, arg.SeatID)
		if err == nil {
			return domain.ErrSeatAlreadyBooked
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		// Check if the seat is already on hold
		_, err = q.GetActiveSeatHold(ctx, arg.SeatID)
		if err == nil {
			return domain.ErrSeatAlreadyOnHold
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		
		// Create the seat hold
		result, err = q.CreateSeatHold(ctx, arg)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
