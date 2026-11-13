package service

import (
	"context"

	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type BookingStore interface {
	ConfirmBookingTx(
		ctx context.Context,
		arg domain.ConfirmBookingParams,
		flightID pgtype.UUID,
	) (sqlcdb.Booking, error)
}

type BookingService struct {
	store BookingStore
}

type ConfirmBookingRequest struct {
	SeatID    pgtype.UUID
	UserID    pgtype.UUID
	HoldToken pgtype.UUID
	FlightID  pgtype.UUID
}

func NewBookingService(store BookingStore) *BookingService {
	return &BookingService{
		store: store,
	}
}

func (s *BookingService) ConfirmBooking(
	ctx context.Context,
	req ConfirmBookingRequest,
) (sqlcdb.Booking, error) {

	arg := domain.ConfirmBookingParams{
		SeatID:      req.SeatID,
		UserID:      req.UserID,
		HoldToken:   req.HoldToken,
		BookingID:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
		BookingCode: uuid.NewString(),
	}
	booking, err := s.store.ConfirmBookingTx(
		ctx,
		arg,
		req.FlightID,
	)
	if err != nil {
		return sqlcdb.Booking{}, err
	}
	return booking, nil
}
