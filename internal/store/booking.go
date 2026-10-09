package store

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ConfirmBookingTxParams struct {
	SeatID      pgtype.UUID
	UserID      pgtype.UUID
	HoldToken   pgtype.UUID
	BookingID   pgtype.UUID
	BookingCode string
}

func (s *Store) ConfirmBookingTx(ctx context.Context, arg ConfirmBookingTxParams) (sqlcdb.Booking, error) {
	var result sqlcdb.Booking
	err := s.execTx(ctx, func(q *sqlcdb.Queries) error {
		var err error
		// Lock seat to prevent concurrent booking
		_, err = q.LockSeat(ctx, arg.SeatID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrSeatNotFound
			}
			return err
		}
		// Check if the seat is already booked
		hold, err := q.GetActiveHoldByToken(ctx, sqlcdb.GetActiveHoldByTokenParams{
			SeatID:    arg.SeatID,
			UserID:    arg.UserID,
			HoldToken: arg.HoldToken,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrHoldNotFound
			}
			return err
		}
		// Check if the seat is already booked
		bookingArg := sqlcdb.CreateBookingParams{
			ID:          arg.BookingID,
			UserID:      hold.UserID,
			SeatID:      hold.SeatID,
			BookingCode: arg.BookingCode,
			Status:      "CONFIRMED",
		}

		result, err = q.CreateBooking(ctx, bookingArg)
		if err != nil {
			return err
		}
		// Delete the seat hold after successful booking
		err = q.DeleteSeatHold(ctx, hold.ID)
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}
