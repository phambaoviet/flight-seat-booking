package db

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
func TestCreateHoldTx_SeatNotFound(t *testing.T) {
	store := NewStore(testDB)

	user := randomUser(t)

	arg := CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    randomUUID(), // Non-existent seat ID
		HoldToken: randomUUID(),
	}

	// Use a non-existent seat ID
	result, err := store.CreateHoldTx(context.Background(), arg)
	require.Error(t, err)
	require.Empty(t, result)
	require.ErrorIs(t, err, ErrSeatNotFound)
}
func TestCreateHoldTx_SeatAlreadyBooked(t *testing.T) {
	store := NewStore(testDB)

	user := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	// Create a confirmed booking for the seat
	bookingArg := CreateBookingParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		BookingCode: randomString(10),
		Status:    "CONFIRMED",
	}
	booking, err := store.CreateBooking(context.Background(), bookingArg)
	require.NoError(t, err)
	require.NotEmpty(t, booking)

	// Attempt to create a hold for the already booked seat
	holdArg := CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	result, err := store.CreateHoldTx(context.Background(), holdArg)
	require.Error(t, err)
	require.Empty(t, result)
	require.ErrorIs(t, err, ErrSeatAlreadyBooked)

	// Verify that no active seat hold was created
	_, err = store.GetActiveSeatHold(context.Background(), seat.ID)
	require.ErrorIs(t, err, pgx.ErrNoRows)

}

func TestCreateHoldTx_SeatAlreadyOnHold(t *testing.T) {
	store := NewStore(testDB)

	user1 := randomUser(t)
	user2 := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	// Create an active seat hold for the seat by user1
	holdArg1 := CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user1.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	hold1, err := store.CreateHoldTx(context.Background(), holdArg1)
	require.NoError(t, err)
	require.NotEmpty(t, hold1)

	// Attempt to create a hold for the same seat by user2
	holdArg2 := CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user2.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	result, err := store.CreateHoldTx(context.Background(), holdArg2)
	require.Error(t, err)
	require.Empty(t, result)
	require.ErrorIs(t, err, ErrSeatAlreadyOnHold)

	// Verify that the active seat hold is still held by user1
	activeHold, err := store.GetActiveSeatHold(context.Background(), seat.ID)
	require.NoError(t, err)
	require.NotEmpty(t, activeHold)
	require.Equal(t, hold1.ID, activeHold.ID)
	require.Equal(t, hold1.UserID, activeHold.UserID)
}
func TestCreateHoldTx_ExpiredHold(t *testing.T) {
	store := NewStore(testDB)
	
	user1 := randomUser(t)
	user2 := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	// Create an expired seat hold for the seat by user1
	holdArg1 := CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user1.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	hold1, err := store.CreateHoldTx(context.Background(), holdArg1)
	require.NoError(t, err)
	require.NotEmpty(t, hold1)

	// Manually expire the hold by updating the expires_at field
	createdTime := time.Now().Add(-20 * time.Minute) 
	expiredTime := time.Now().Add(-10 * time.Minute) 
	_, err = store.db.Exec(context.Background(), "UPDATE seat_holds SET created_at = $1, expires_at = $2 WHERE id = $3", createdTime , expiredTime , hold1.ID)
	require.NoError(t, err)
	

	// Attempt to create a hold for the same seat by user2
	holdArg2 := CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user2.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	result, err := store.CreateHoldTx(context.Background(), holdArg2)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Verify that the active seat hold is now held by user2
	activeHold, err := store.GetActiveSeatHold(context.Background(), seat.ID)
	require.NoError(t, err)
	require.NotEmpty(t, activeHold)
	require.Equal(t, result.ID, activeHold.ID)
	require.Equal(t, result.UserID, activeHold.UserID)
}
func TestCreateHoldTx_ConcurrentHold(t *testing.T) {
	n := 10 // Number of concurrent goroutines to simulate
	errs := make(chan error, n)
	var wg sync.WaitGroup
	store := NewStore(testDB)
	
	// Barrier: Release all goroutines at the same time
	start := make(chan struct{})
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)
	seatID := seat.ID // Use the same seat ID for all goroutines
	for i:=0; i < n; i++ {
		wg.Add(1)
		go func(userID pgtype.UUID) {
			defer wg.Done()
			<-start // Wait for the signal to start

			
			
			holdArg := CreateSeatHoldParams{
				ID:        randomUUID(),
				UserID:    userID,
				SeatID:    seatID,
				HoldToken: randomUUID(),
			}
			_, err := store.CreateHoldTx(context.Background(), holdArg)
			
			errs <- err
			
		}(randomUser(t).ID)
	}

	close(start) // Signal all goroutines to start
	wg.Wait()    // Wait for all goroutines to finish
	close(errs)  // Close the error channel

	var successCount, seatAlreadyOnHoldCount int
	for err := range errs {
		if err == nil {
			successCount++
		} else {
			if errors.Is(err, ErrSeatAlreadyOnHold) {
				seatAlreadyOnHoldCount++
			}
		}
	}
	
	require.Equal(t, 1, successCount, "Exactly one goroutine should succeed in creating a hold")
	require.Equal(t, n-1, seatAlreadyOnHoldCount, "The rest of the goroutines should receive ErrSeatAlreadyOnHold")
	count, err := store.CountActiveHold(context.Background(), seatID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}