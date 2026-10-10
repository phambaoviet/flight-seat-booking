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
	CreateHoldTxFunc func(ctx context.Context, arg sqlcdb.CreateSeatHoldParams) (sqlcdb.SeatHold, error)
}

func (m *mockHoldStore) CreateHoldTx(ctx context.Context, arg sqlcdb.CreateSeatHoldParams) (sqlcdb.SeatHold, error) {
	return m.CreateHoldTxFunc(ctx, arg)
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
		UserID:    expectedHold.UserID,
		SeatID:    expectedHold.SeatID,
	}
	// create a mock store that returns the expected hold
	mockStore := &mockHoldStore{
		CreateHoldTxFunc: func(ctx context.Context, gotArg sqlcdb.CreateSeatHoldParams) (sqlcdb.SeatHold, error) {
			assert.Equal(t, req.UserID, gotArg.UserID)
        	assert.Equal(t, req.SeatID, gotArg.SeatID)
			assert.True(t, gotArg.ID.Valid)
        	assert.True(t, gotArg.HoldToken.Valid)
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
		UserID:    randomUUID(),
		SeatID:    randomUUID(),
	}
	// create a mock store that returns an error
	mockStore := &mockHoldStore{
		CreateHoldTxFunc: func(ctx context.Context, gotArg sqlcdb.CreateSeatHoldParams) (sqlcdb.SeatHold, error) {
			assert.Equal(t, req.UserID, gotArg.UserID)
			assert.Equal(t, req.SeatID, gotArg.SeatID)
			assert.True(t, gotArg.ID.Valid)
			assert.True(t, gotArg.HoldToken.Valid)
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
