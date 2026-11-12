package service

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlcdb "flight-booking/db/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

type mockHoldStore struct {
	CreateHoldTxFunc func(ctx context.Context, arg sqlcdb.CreateSeatHoldParams, flightID pgtype.UUID,) (sqlcdb.SeatHold, error)
	ReleaseHoldTxFunc func(ctx context.Context, holdID pgtype.UUID, userID pgtype.UUID) (sqlcdb.SeatHold, error) 
}

func (m *mockHoldStore) CreateHoldTx(ctx context.Context, arg sqlcdb.CreateSeatHoldParams, flightID pgtype.UUID) (sqlcdb.SeatHold, error) {
	return m.CreateHoldTxFunc(ctx, arg, flightID)
}
func (m *mockHoldStore) ReleaseHoldTx(ctx context.Context, holdID pgtype.UUID, userID pgtype.UUID) (sqlcdb.SeatHold, error) {
    return m.ReleaseHoldTxFunc(ctx, holdID, userID)
}
func randomUUID() pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid.New(),
		Valid: true,
	}
}

func TestCreateHold_Success(t *testing.T) {
	now := time.Now()
	// create a mock store that returns a successful result
	expectedHold := sqlcdb.SeatHold{
		ID:        randomUUID(),
		UserID:    randomUUID(),
		SeatID:    randomUUID(),
		HoldToken: randomUUID(),
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: now.Add(10 * time.Minute), Valid: true}, // 10 minutes later
	}

	req := CreateHoldRequest{
		UserID: expectedHold.UserID,
		SeatID: expectedHold.SeatID,
		FlightID: randomUUID(),
	}
	// create a mock store that returns the expected hold
	mockStore := &mockHoldStore{
		CreateHoldTxFunc: func(ctx context.Context, gotArg sqlcdb.CreateSeatHoldParams, flightID pgtype.UUID,) (sqlcdb.SeatHold, error) {
			assert.Equal(t, req.UserID, gotArg.UserID)
			assert.Equal(t, req.SeatID, gotArg.SeatID)
			assert.True(t, gotArg.ID.Valid)
			assert.True(t, gotArg.HoldToken.Valid)
			assert.Equal(t, req.FlightID, flightID)
			return expectedHold, nil
		},
	}
	service := NewHoldService(mockStore)
	// call CreateHold and expect a successful result
	hold, err := service.CreateHold(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, expectedHold, hold)

}
func TestCreateHoldError(t *testing.T) {
	expectedErr := errors.New("database error")
	req := CreateHoldRequest{
		UserID: randomUUID(),
		SeatID: randomUUID(),
		FlightID: randomUUID(),
	}
	// create a mock store that returns an error
	mockStore := &mockHoldStore{
		CreateHoldTxFunc: func(ctx context.Context, gotArg sqlcdb.CreateSeatHoldParams, flightID pgtype.UUID) (sqlcdb.SeatHold, error) {
			assert.Equal(t, req.UserID, gotArg.UserID)
			assert.Equal(t, req.SeatID, gotArg.SeatID)
			assert.True(t, gotArg.ID.Valid)
			assert.True(t, gotArg.HoldToken.Valid)
			assert.Equal(t, req.FlightID, flightID)
			return sqlcdb.SeatHold{}, expectedErr
		},
	}
	// create a hold service with the mock store
	service := NewHoldService(mockStore)
	// call CreateHold and expect an error
	hold, err := service.CreateHold(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, sqlcdb.SeatHold{}, hold)
	
}

func TestReleaseHold_Success(t *testing.T) {
	now := time.Now()
	// create a mock store that returns a successful result
	expectedHold := sqlcdb.SeatHold{
		ID:        randomUUID(),
		UserID:    randomUUID(),
		SeatID:    randomUUID(),
		HoldToken: randomUUID(),
		Status:   "RELEASED",
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: now.Add(10 * time.Minute), Valid: true}, // 10 minutes later
	}

	req := ReleaseHoldRequest{
		HoldID: expectedHold.ID,
		UserID: expectedHold.UserID,
	}
	storeCalled := false
	// create a mock store that returns the expected hold
	mockStore := &mockHoldStore{
		ReleaseHoldTxFunc: func(ctx context.Context, holdID pgtype.UUID, userID pgtype.UUID) (sqlcdb.SeatHold, error) {
			storeCalled = true
			assert.Equal(t, req.HoldID, holdID)
			assert.Equal(t, req.UserID, userID)
			return expectedHold, nil
		},
	}
	service := NewHoldService(mockStore)
	// call ReleaseHold and expect a successful result
	hold, err := service.ReleaseHold(context.Background(), req.HoldID, req.UserID)
	assert.NoError(t, err)
	assert.Equal(t, expectedHold, hold)
	assert.True(t, storeCalled, "ReleaseHoldTxFunc was not called")

}
func TestReleaseHold_InternalError(t *testing.T) {
	expectedErr := errors.New("database error")
	req := ReleaseHoldRequest{
		HoldID: randomUUID(),
		UserID: randomUUID(),
	}
	// create a mock store that returns an error
	mockStore := &mockHoldStore{
		ReleaseHoldTxFunc: func(ctx context.Context, holdID pgtype.UUID, userID pgtype.UUID) (sqlcdb.SeatHold, error) {
			assert.Equal(t, req.HoldID, holdID)
			assert.Equal(t, req.UserID, userID)
			return sqlcdb.SeatHold{}, expectedErr
		},
	}
	// create a hold service with the mock store
	service := NewHoldService(mockStore)
	// call ReleaseHold and expect an error
	hold, err := service.ReleaseHold(context.Background(), req.HoldID, req.UserID)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, sqlcdb.SeatHold{}, hold)
	
}

func TestReleaseHold_NotFoundError(t *testing.T) {
	expectedErr := errors.New("hold not found")
	req := ReleaseHoldRequest{
		HoldID: randomUUID(),
		UserID: randomUUID(),
	}
	// create a mock store that returns an error
	mockStore := &mockHoldStore{
		ReleaseHoldTxFunc: func(ctx context.Context, holdID pgtype.UUID, userID pgtype.UUID) (sqlcdb.SeatHold, error) {
			assert.Equal(t, req.HoldID, holdID)
			assert.Equal(t, req.UserID, userID)
			return sqlcdb.SeatHold{}, expectedErr
		},
	}
	// create a hold service with the mock store
	service := NewHoldService(mockStore)
	// call ReleaseHold and expect an error
	hold, err := service.ReleaseHold(context.Background(), req.HoldID, req.UserID)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, sqlcdb.SeatHold{}, hold)
	
}