package store

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"

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

func (s *Store) ConfirmBookingTx(ctx context.Context, arg ConfirmBookingTxParams, flightID pgtype.UUID) (sqlcdb.Booking, error) {
	var result sqlcdb.Booking
	
	err := s.execTx(ctx, func(q *sqlcdb.Queries) error {
		var err error
		// Lock seat to prevent concurrent booking
		_, err = q.LockSeatForHold(ctx, sqlcdb.LockSeatForHoldParams{
			ID:       arg.SeatID,
			FlightID: flightID,
		}) 
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrSeatNotFound
			}
			return err
		}

		// GetConfrimBooking to check if the seat is already booked
		_, err = q.GetConfirmedBooking(ctx, arg.SeatID)
		if err == nil {
			return domain.ErrSeatAlreadyBooked
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		
		// Convert seat hold to booking
		hold, err := q.ConvertSeatHold(ctx, sqlcdb.ConvertSeatHoldParams{
			SeatID:    arg.SeatID,
			UserID:    arg.UserID,
			HoldToken: arg.HoldToken,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrSeatHoldNotFound
			}
			return err
		}
		
		// Call convert
		book, err := q.CreateBookingFromHold(ctx, sqlcdb.CreateBookingFromHoldParams{
			ID:          arg.BookingID,
			ID_2: 	   		hold.ID,
			UserID:      arg.UserID,
			BookingCode: arg.BookingCode,
		})
		if err != nil {
			return err
		}
		
		result = book
		return nil
	})
	if err != nil {
		return sqlcdb.Booking{}, err
	}
	return result, nil
}
