package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)
type ConfirmBookingTxParams struct {
    SeatID      pgtype.UUID
    UserID      pgtype.UUID
    HoldToken   pgtype.UUID
    BookingID   pgtype.UUID
    BookingCode string
}
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
func (s *Store) ConfirmBookingTx(ctx context.Context, arg ConfirmBookingTxParams) (Booking, error) {
	var result Booking
	err := s.execTx(ctx, func(q *Queries) error {
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
		hold, err := q.GetActiveHoldByToken(ctx, GetActiveHoldByTokenParams{
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
		bookingArg := CreateBookingParams{
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
