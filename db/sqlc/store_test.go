package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateHoldTx (t *testing.T) {
	store := NewStore(testDB)

	user := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	arg := CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}

	result, err := store.CreateHoldTx(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	require.Equal(t, arg.ID, result.ID)
	require.Equal(t, arg.UserID, result.UserID)
	require.Equal(t, arg.SeatID, result.SeatID)
	require.Equal(t, arg.HoldToken, result.HoldToken)
}