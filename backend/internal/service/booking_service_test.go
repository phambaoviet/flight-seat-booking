package service

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

type MockBookingStore struct {
	ConfirmBookingTxFunc func(ctx context.Context, arg domain.ConfirmBookingParams, flightID pgtype.UUID) (sqlcdb.Booking, error)
}

func (m *MockBookingStore) ConfirmBookingTx(ctx context.Context, arg domain.ConfirmBookingParams, flightID pgtype.UUID) (sqlcdb.Booking, error) {
	if m.ConfirmBookingTxFunc != nil {
		return m.ConfirmBookingTxFunc(ctx, arg, flightID)
	}
	return sqlcdb.Booking{}, nil
}

func TestConfirmBooking_Success(t *testing.T) {
	seatID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	userID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	holdToken := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	flightID := pgtype.UUID{Bytes: uuid.New(), Valid: true}

	storeCalled := false

	mockStore := &MockBookingStore{
		ConfirmBookingTxFunc: func(
			ctx context.Context,
			arg domain.ConfirmBookingParams,
			gotFlightID pgtype.UUID,
		) (sqlcdb.Booking, error) {
			storeCalled = true

			assert.Equal(t, seatID, arg.SeatID)
			assert.Equal(t, userID, arg.UserID)
			assert.Equal(t, holdToken, arg.HoldToken)
			assert.Equal(t, flightID, gotFlightID)

			assert.True(t, arg.BookingID.Valid)
			assert.NotEmpty(t, arg.BookingCode)

			return sqlcdb.Booking{
				ID:          arg.BookingID,
				UserID:      arg.UserID,
				SeatID:      arg.SeatID,
				BookingCode: arg.BookingCode,
			}, nil
		},
	}

	service := NewBookingService(mockStore)

	booking, err := service.ConfirmBooking(
		context.Background(),
		ConfirmBookingRequest{
			SeatID:    seatID,
			UserID:    userID,
			HoldToken: holdToken,
			FlightID:  flightID,
		},
	)

	assert.NoError(t, err)
	assert.True(t, storeCalled)
	assert.Equal(t, userID, booking.UserID)
	assert.Equal(t, seatID, booking.SeatID)
	assert.True(t, booking.ID.Valid)
	assert.NotEmpty(t, booking.BookingCode)
}

func TestConfirmBooking_Error(t *testing.T) {
	seatID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	userID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	holdToken := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	flightID := pgtype.UUID{Bytes: uuid.New(), Valid: true}

	expectedErr := errors.New("database error")

	mockStore := &MockBookingStore{
		ConfirmBookingTxFunc: func(
			ctx context.Context,
			arg domain.ConfirmBookingParams,
			gotFlightID pgtype.UUID,
		) (sqlcdb.Booking, error) {
			return sqlcdb.Booking{}, expectedErr
		},
	}

	service := NewBookingService(mockStore)

	_, err := service.ConfirmBooking(
		context.Background(),
		ConfirmBookingRequest{
			SeatID:    seatID,
			UserID:    userID,
			HoldToken: holdToken,
			FlightID:  flightID,
		},
	)

	assert.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
}

func TestConfirmBooking_SeatAlreadyBooked(t *testing.T) {
	seatID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	userID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	holdToken := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	flightID := pgtype.UUID{Bytes: uuid.New(), Valid: true}

	mockStore := &MockBookingStore{
		ConfirmBookingTxFunc: func(
			ctx context.Context,
			arg domain.ConfirmBookingParams,
			gotFlightID pgtype.UUID,
		) (sqlcdb.Booking, error) {
			return sqlcdb.Booking{}, domain.ErrSeatAlreadyBooked
		},
	}

	service := NewBookingService(mockStore)

	_, err := service.ConfirmBooking(
		context.Background(),
		ConfirmBookingRequest{
			SeatID:    seatID,
			UserID:    userID,
			HoldToken: holdToken,
			FlightID:  flightID,
		},
	)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrSeatAlreadyBooked)
}

func TestConfirmBooking_HoldNotFound(t *testing.T) {
	seatID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	userID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	holdToken := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	flightID := pgtype.UUID{Bytes: uuid.New(), Valid: true}

	mockStore := &MockBookingStore{
		ConfirmBookingTxFunc: func(
			ctx context.Context,
			arg domain.ConfirmBookingParams,
			gotFlightID pgtype.UUID,
		) (sqlcdb.Booking, error) {
			return sqlcdb.Booking{}, domain.ErrSeatHoldNotFound
		},
	}

	service := NewBookingService(mockStore)

	_, err := service.ConfirmBooking(
		context.Background(),
		ConfirmBookingRequest{
			SeatID:    seatID,
			UserID:    userID,
			HoldToken: holdToken,
			FlightID:  flightID,
		},
	)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrSeatHoldNotFound)
}
