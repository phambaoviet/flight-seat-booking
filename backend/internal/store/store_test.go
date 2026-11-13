package store

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"

	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestCreateHoldTx(t *testing.T) {
	testStore := NewStore(testDB)

	user := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	arg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}

	result, err := testStore.CreateHoldTx(context.Background(), arg, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	require.Equal(t, arg.ID, result.ID)
	require.Equal(t, arg.UserID, result.UserID)
	require.Equal(t, arg.SeatID, result.SeatID)
	require.Equal(t, arg.HoldToken, result.HoldToken)
}
func TestCreateHoldTx_SeatNotFound(t *testing.T) {
	testStore := NewStore(testDB)

	user := randomUser(t)

	arg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    randomUUID(), // Non-existent seat ID
		HoldToken: randomUUID(),
	}

	// Use a non-existent seat ID
	result, err := testStore.CreateHoldTx(context.Background(), arg, randomUUID())
	require.Error(t, err)
	require.Empty(t, result)
	require.ErrorIs(t, err, domain.ErrSeatNotFound)
}
func TestCreateHoldTx_SeatAlreadyBooked(t *testing.T) {
	testStore := NewStore(testDB)

	user := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	// Create a confirmed booking for the seat
	bookingArg := sqlcdb.CreateBookingParams{
		ID:          randomUUID(),
		UserID:      user.ID,
		SeatID:      seat.ID,
		BookingCode: randomString(10),
		Status:      "CONFIRMED",
	}
	booking, err := testStore.CreateBooking(context.Background(), bookingArg)
	require.NoError(t, err)
	require.NotEmpty(t, booking)

	// Attempt to create a hold for the already booked seat
	holdArg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	result, err := testStore.CreateHoldTx(context.Background(), holdArg, flight.ID)
	require.Error(t, err)
	require.Empty(t, result)
	require.ErrorIs(t, err, domain.ErrSeatAlreadyBooked)

	// Verify that no active seat hold was created
	_, err = testStore.GetActiveSeatHold(context.Background(), seat.ID)
	require.ErrorIs(t, err, pgx.ErrNoRows)

}

func TestCreateHoldTx_SeatAlreadyOnHold(t *testing.T) {
	testStore := NewStore(testDB)

	user1 := randomUser(t)
	user2 := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	// Create an active seat hold for the seat by user1
	holdArg1 := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user1.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	hold1, err := testStore.CreateHoldTx(context.Background(), holdArg1, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, hold1)

	// Attempt to create a hold for the same seat by user2
	holdArg2 := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user2.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	result, err := testStore.CreateHoldTx(context.Background(), holdArg2, flight.ID)
	require.Error(t, err)
	require.Empty(t, result)
	require.ErrorIs(t, err, domain.ErrSeatAlreadyOnHold)

	// Verify that the active seat hold is still held by user1
	activeHold, err := testStore.GetActiveSeatHold(context.Background(), seat.ID)
	require.NoError(t, err)
	require.NotEmpty(t, activeHold)
	require.Equal(t, hold1.ID, activeHold.ID)
	require.Equal(t, hold1.UserID, activeHold.UserID)
}
func TestCreateHoldTx_ExpiredHold(t *testing.T) {
	testStore := NewStore(testDB)

	user1 := randomUser(t)
	user2 := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	// Create an expired seat hold for the seat by user1
	holdArg1 := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user1.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	hold1, err := testStore.CreateHoldTx(context.Background(), holdArg1, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, hold1)

	// Manually expire the hold by updating the expires_at field
	createdTime := time.Now().Add(-20 * time.Minute)
	expiredTime := time.Now().Add(-10 * time.Minute)
	_, err = testStore.db.Exec(context.Background(), "UPDATE seat_holds SET created_at = $1, expires_at = $2 WHERE id = $3", createdTime, expiredTime, hold1.ID)
	require.NoError(t, err)

	// Attempt to create a hold for the same seat by user2
	holdArg2 := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user2.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	result, err := testStore.CreateHoldTx(context.Background(), holdArg2, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Verify that the active seat hold is now held by user2
	activeHold, err := testStore.GetActiveSeatHold(context.Background(), seat.ID)
	require.NoError(t, err)
	require.NotEmpty(t, activeHold)
	require.Equal(t, result.ID, activeHold.ID)
	require.Equal(t, result.UserID, activeHold.UserID)
}
func TestCreateHoldTx_ConcurrentHold(t *testing.T) {
	n := 10 // Number of concurrent goroutines to simulate
	errs := make(chan error, n)
	var wg sync.WaitGroup
	testStore := NewStore(testDB)

	// Barrier: Release all goroutines at the same time
	start := make(chan struct{})
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)
	seatID := seat.ID // Use the same seat ID for all goroutines
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(userID pgtype.UUID) {
			defer wg.Done()
			<-start // Wait for the signal to start

			holdArg := sqlcdb.CreateSeatHoldParams{
				ID:        randomUUID(),
				UserID:    userID,
				SeatID:    seatID,
				HoldToken: randomUUID(),
			}
			_, err := testStore.CreateHoldTx(context.Background(), holdArg, flight.ID)

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
			if errors.Is(err, domain.ErrSeatAlreadyOnHold) {
				seatAlreadyOnHoldCount++
			}
		}
	}

	require.Equal(t, 1, successCount, "Exactly one goroutine should succeed in creating a hold")
	require.Equal(t, n-1, seatAlreadyOnHoldCount, "The rest of the goroutines should receive ErrSeatAlreadyOnHold")
	count, err := testStore.CountActiveHold(context.Background(), seatID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}
func TestConfirmBookingTx(t *testing.T) {
	testStore := NewStore(testDB)
	// create user
	user := randomUser(t)

	// create flight
	flight := randomFlight(t)

	// create seat
	seat := randomSeat(t, flight.ID)
	holdArg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}

	result, err := testStore.CreateHoldTx(context.Background(), holdArg, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	confirmArg := domain.ConfirmBookingParams{
		BookingID:   randomUUID(),
		UserID:      user.ID,
		SeatID:      seat.ID,
		HoldToken:   holdArg.HoldToken,
		BookingCode: randomString(10),
	}

	booking, err := testStore.ConfirmBookingTx(context.Background(), confirmArg, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, booking)

	holdAfter, err := testStore.GetActiveHoldByToken(context.Background(), sqlcdb.GetActiveHoldByTokenParams{
		SeatID:    seat.ID,
		UserID:    user.ID,
		HoldToken: holdArg.HoldToken,
	})
	require.Error(t, err)
	require.Empty(t, holdAfter)
	require.ErrorIs(t, err, pgx.ErrNoRows)

	require.Equal(t, confirmArg.BookingID, booking.ID)
	require.Equal(t, confirmArg.UserID, booking.UserID)
	require.Equal(t, confirmArg.SeatID, booking.SeatID)
	require.Equal(t, confirmArg.BookingCode, booking.BookingCode)
	require.Equal(t, "CONFIRMED", booking.Status)
}
func TestConfirmBookingTx_HoldNotFound(t *testing.T) {
	testStore := NewStore(testDB)
	// create user, flight, and seat
	user := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	// Call ConfirmBookingTx with a non-existent hold token
	confirmArg := domain.ConfirmBookingParams{
		BookingID:   randomUUID(),
		UserID:      user.ID,
		SeatID:      seat.ID,
		HoldToken:   randomUUID(), // Non-existent hold token
		BookingCode: randomString(10),
	}

	result, err := testStore.ConfirmBookingTx(context.Background(), confirmArg, flight.ID)
	require.Error(t, err)
	require.Empty(t, result)
	require.ErrorIs(t, err, domain.ErrSeatHoldNotFound)
}
func TestConfirmBookingTx_ExpiredHold(t *testing.T) {
	testStore := NewStore(testDB)
	// create user
	user := randomUser(t)
	// create flight
	flight := randomFlight(t)
	// create seat
	seat := randomSeat(t, flight.ID)
	// create Hold
	holdArg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}
	result, err := testStore.CreateSeatHold(context.Background(), holdArg)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Set the hold expiration time to the past
	createdTime := time.Now().Add(-20 * time.Minute)
	expiredTime := time.Now().Add(-10 * time.Minute)
	_, err = testStore.db.Exec(context.Background(), "UPDATE seat_holds SET created_at = $1, expires_at = $2 WHERE id = $3", createdTime, expiredTime, holdArg.ID)
	require.NoError(t, err)

	// Call ConfirmBooking
	confirmArg := domain.ConfirmBookingParams{
		BookingID:   randomUUID(),
		UserID:      user.ID,
		SeatID:      seat.ID,
		HoldToken:   holdArg.HoldToken,
		BookingCode: randomString(10),
	}
	booking, err := testStore.ConfirmBookingTx(context.Background(), confirmArg, flight.ID)
	require.ErrorIs(t, err, domain.ErrSeatHoldNotFound)
	require.Empty(t, booking)
}
func TestConfirmBookingTx_AlreadyConfirmed(t *testing.T) {
	testStore := NewStore(testDB)
	// create user
	user := randomUser(t)

	// create flight
	flight := randomFlight(t)

	// create seat
	seat := randomSeat(t, flight.ID)
	holdArg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}

	result, err := testStore.CreateHoldTx(context.Background(), holdArg, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Create a first confirmBooking with success
	bookingArg1 := domain.ConfirmBookingParams{
		SeatID:      seat.ID,
		UserID:      user.ID,
		HoldToken:   holdArg.HoldToken,
		BookingID:   randomUUID(),
		BookingCode: randomString(10),
	}
	booking1, err := testStore.ConfirmBookingTx(context.Background(), bookingArg1, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, booking1)

	// Prepare a second confirmation request using the same hold
	bookingArg2 := bookingArg1
	bookingArg2.BookingID = randomUUID()
	bookingArg2.BookingCode = randomString(10)
	booking2, err := testStore.ConfirmBookingTx(context.Background(), bookingArg2, flight.ID)
	require.ErrorIs(t, err, domain.ErrSeatAlreadyBooked)
	require.Empty(t, booking2)
}
func TestConfirmBookingTx_Rollback(t *testing.T) {
	testStore := NewStore(testDB)
	// create user
	user := randomUser(t)

	// create flight
	flight := randomFlight(t)

	// create seat
	seat := randomSeat(t, flight.ID)
	holdArg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}

	result, err := testStore.CreateHoldTx(context.Background(), holdArg, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	hold, err := testStore.GetActiveHoldByToken(context.Background(), sqlcdb.GetActiveHoldByTokenParams{
		SeatID:    seat.ID,
		UserID:    user.ID,
		HoldToken: holdArg.HoldToken,
	})
	require.NoError(t, err)
	require.NotEmpty(t, hold)

	otherSeat := randomSeat(t, flight.ID)

	bookingCode := randomString(10)
	_, err = testStore.CreateBooking(context.Background(), sqlcdb.CreateBookingParams{
		ID:          randomUUID(),
		UserID:      user.ID,
		SeatID:      otherSeat.ID,
		BookingCode: bookingCode,
		Status:      "CONFIRMED",
	})
	require.NoError(t, err)

	// Try to confirm the booking with a duplicate booking code
	confirmArg := domain.ConfirmBookingParams{
		SeatID:      seat.ID,
		UserID:      user.ID,
		HoldToken:   holdArg.HoldToken,
		BookingID:   randomUUID(),
		BookingCode: bookingCode,
	}
	booking, err := testStore.ConfirmBookingTx(context.Background(), confirmArg, flight.ID)
	require.Error(t, err)
	require.Empty(t, booking)

	holdAfter, err := testStore.GetActiveHoldByToken(
		context.Background(),
		sqlcdb.GetActiveHoldByTokenParams{
			SeatID:    seat.ID,
			UserID:    user.ID,
			HoldToken: holdArg.HoldToken,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, holdAfter)

}
func TestReleaseHoldTx_Success(t *testing.T) {
	testStore := NewStore(testDB)

	user := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	holdArg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}

	result, err := testStore.CreateHoldTx(context.Background(), holdArg, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	releasedHold, err := testStore.ReleaseHoldTx(context.Background(), holdArg.ID, user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, releasedHold)
	require.Equal(t, holdArg.ID, releasedHold.ID)
	require.Equal(t, "RELEASED", releasedHold.Status)
	require.True(t, releasedHold.ReleasedAt.Valid)

}

func TestReleaseHoldTx_NotFound(t *testing.T) {
	testStore := NewStore(testDB)

	nonExistentHoldID := randomUUID()

	releasedHold, err := testStore.ReleaseHoldTx(context.Background(), nonExistentHoldID, randomUUID())
	require.Error(t, err)
	require.Empty(t, releasedHold)
	require.ErrorIs(t, err, domain.ErrSeatHoldNotFound)
}
func TestReleaseHoldTx_ExpiredHold(t *testing.T) {
	testStore := NewStore(testDB)

	user := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	holdArg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}

	result, err := testStore.CreateHoldTx(context.Background(), holdArg, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Manually expire the hold by updating the expires_at field
	createdTime := time.Now().Add(-20 * time.Minute)
	expiredTime := time.Now().Add(-10 * time.Minute)
	_, err = testStore.db.Exec(context.Background(), "UPDATE seat_holds SET created_at = $1, expires_at = $2 WHERE id = $3", createdTime, expiredTime, holdArg.ID)
	require.NoError(t, err)

	releasedHold, err := testStore.ReleaseHoldTx(context.Background(), holdArg.ID, user.ID)
	require.Error(t, err)
	require.Empty(t, releasedHold)
	require.ErrorIs(t, err, domain.ErrSeatHoldNotFound)

	// Verify database state
	storedHold, err := testStore.GetSeatHold(context.Background(), holdArg.ID)
	require.NoError(t, err)

	require.Equal(t, "ACTIVE", storedHold.Status)
	require.False(t, storedHold.ReleasedAt.Valid)
}

func TestCancelBooking_Success(t *testing.T) {
	testStore := NewStore(testDB)

	user := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	// Create a confirmed booking for the seat
	bookingArg := sqlcdb.CreateBookingParams{
		ID:          randomUUID(),
		UserID:      user.ID,
		SeatID:      seat.ID,
		BookingCode: randomString(10),
		Status:      "CONFIRMED",
	}
	booking, err := testStore.CreateBooking(context.Background(), bookingArg)
	require.NoError(t, err)
	require.NotEmpty(t, booking)

	cancelledBooking, err := testStore.CancelBookingTx(context.Background(), booking.ID, booking.BookingCode, user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, cancelledBooking)
	require.Equal(t, "CANCELLED", cancelledBooking.Status)
}

func TestCancelBooking_NotFound(t *testing.T) {
	testStore := NewStore(testDB)

	nonExistentBookingID := randomUUID()
	nonExistentBookingCode := randomString(10)
	nonExistentUserID := randomUUID()

	cancelledBooking, err := testStore.CancelBookingTx(context.Background(), nonExistentBookingID, nonExistentBookingCode, nonExistentUserID)
	require.Error(t, err)
	require.Empty(t, cancelledBooking)
	require.ErrorIs(t, err, domain.ErrBookingNotFound)
}

func TestCancelBooking_AlreadyCancelled(t *testing.T) {
	testStore := NewStore(testDB)

	user := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	// Create a confirmed booking for the seat
	bookingArg := sqlcdb.CreateBookingParams{
		ID:          randomUUID(),
		UserID:      user.ID,
		SeatID:      seat.ID,
		BookingCode: randomString(10),
		Status:      "CONFIRMED",
	}
	booking, err := testStore.CreateBooking(context.Background(), bookingArg)
	require.NoError(t, err)
	require.NotEmpty(t, booking)

	// Cancel the booking
	cancelledBooking, err := testStore.CancelBookingTx(context.Background(), booking.ID, booking.BookingCode, user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, cancelledBooking)
	require.Equal(t, "CANCELLED", cancelledBooking.Status)

	// Attempt to cancel the already cancelled booking
	cancelledBookingAgain, err := testStore.CancelBookingTx(context.Background(), booking.ID, booking.BookingCode, user.ID)
	require.Error(t, err)
	require.Empty(t, cancelledBookingAgain)
	require.ErrorIs(t, err, domain.ErrBookingNotFound)
}
func TestCancelBooking_NotOwner(t *testing.T) {
	testStore := NewStore(testDB)

	user1 := randomUser(t)
	user2 := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	// Create a confirmed booking for the seat by user1
	bookingArg := sqlcdb.CreateBookingParams{
		ID:          randomUUID(),
		UserID:      user1.ID,
		SeatID:      seat.ID,
		BookingCode: randomString(10),
		Status:      "CONFIRMED",
	}
	booking, err := testStore.CreateBooking(context.Background(), bookingArg)
	require.NoError(t, err)
	require.NotEmpty(t, booking)

	// Attempt to cancel the booking by user2 (not the owner)
	cancelledBooking, err := testStore.CancelBookingTx(context.Background(), booking.ID, booking.BookingCode, user2.ID)
	require.Error(t, err)
	require.Empty(t, cancelledBooking)
	require.ErrorIs(t, err, domain.ErrBookingNotFound)
}
func TestCreateHoldTx_ConcurrentRequests(t *testing.T) {
	var unexpectedErrors []error
	n := 20 // Number of concurrent goroutines to simulate
	errs := make(chan error, n)
	var wg sync.WaitGroup
	testStore := NewStore(testDB)

	// Barrier: Release all goroutines at the same time
	start := make(chan struct{})
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)
	seatID := seat.ID // Use the same seat ID for all goroutines
	userIDs := make([]pgtype.UUID, n)

	// Generate random user IDs for each goroutine
	for i := 0; i < n; i++ {
		userIDs[i] = randomUser(t).ID
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(userID pgtype.UUID) {
			defer wg.Done()
			<-start // Wait for the signal to start

			holdArg := sqlcdb.CreateSeatHoldParams{
				ID:        randomUUID(),
				UserID:    userID,
				SeatID:    seatID,
				HoldToken: randomUUID(),
			}
			_, err := testStore.CreateHoldTx(context.Background(), holdArg, flight.ID)

			errs <- err

		}(userIDs[i])
	}

	close(start) // Signal all goroutines to start
	wg.Wait()    // Wait for all goroutines to finish
	close(errs)  // Close the error channel

	var successCount, seatAlreadyOnHoldCount int

	for err := range errs {
		if err == nil {
			successCount++
		} else if errors.Is(err, domain.ErrSeatAlreadyOnHold) {
			seatAlreadyOnHoldCount++
		} else {
			unexpectedErrors = append(unexpectedErrors, err)
		}
	}

	require.Empty(t, unexpectedErrors)

	require.Equal(t, 1, successCount, "Exactly one goroutine should succeed in creating a hold")
	require.Equal(t, n-1, seatAlreadyOnHoldCount, "The rest of the goroutines should receive ErrSeatAlreadyOnHold")
	count, err := testStore.CountActiveHold(context.Background(), seatID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestConfirmBookingTx_ConcurrentRequests(t *testing.T) {
	n := 20 // Number of concurrent goroutines to simulate
	errs := make(chan error, n)
	var wg sync.WaitGroup
	testStore := NewStore(testDB)

	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)
	user := randomUser(t)

	holdArg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}

	result, err := testStore.CreateHoldTx(context.Background(), holdArg, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // Wait for the signal to start

			confirmArg := domain.ConfirmBookingParams{
				BookingID:   randomUUID(),
				UserID:      user.ID,
				SeatID:      seat.ID,
				HoldToken:   holdArg.HoldToken,
				BookingCode: randomString(10),
			}

			_, err := testStore.ConfirmBookingTx(context.Background(), confirmArg, flight.ID)
			errs <- err
		}()
	}

	close(start) // Signal all goroutines to start
	wg.Wait()    // Wait for all goroutines to finish
	close(errs)  // Close the error channel

	var successCount, seatAlreadyBookedCount int
	var unexpectedErrors []error
	for err := range errs {
		if err == nil {
			successCount++
		} else if errors.Is(err, domain.ErrSeatAlreadyBooked) {
			seatAlreadyBookedCount++
		} else {
			unexpectedErrors = append(unexpectedErrors, err)
		}
	}

	require.Empty(t, unexpectedErrors)
	require.Equal(t, 1, successCount)
	require.Equal(t, n-1, seatAlreadyBookedCount)
}

func TestReleaseHoldTx_NotOwner(t *testing.T) {
	testStore := NewStore(testDB)

	user1 := randomUser(t)
	user2 := randomUser(t)
	flight := randomFlight(t)
	seat := randomSeat(t, flight.ID)

	holdArg := sqlcdb.CreateSeatHoldParams{
		ID:        randomUUID(),
		UserID:    user1.ID,
		SeatID:    seat.ID,
		HoldToken: randomUUID(),
	}

	result, err := testStore.CreateHoldTx(context.Background(), holdArg, flight.ID)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	releasedHold, err := testStore.ReleaseHoldTx(context.Background(), holdArg.ID, user2.ID)
	require.Error(t, err)
	require.Empty(t, releasedHold)
	require.ErrorIs(t, err, domain.ErrSeatHoldNotFound)

	// Verify that the hold is still active in the database
	storedHold, err := testStore.GetSeatHold(context.Background(), holdArg.ID)
	require.NoError(t, err)
	require.NotEmpty(t, storedHold)
	require.Equal(t, "ACTIVE", storedHold.Status)
}
