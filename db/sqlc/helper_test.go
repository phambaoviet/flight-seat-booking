package db

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)


const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var airlines = []string{"VN", "VJ", "9G"}
var airports = []string{"SGN", "HAN", "DAD", "CXR", "PQC", "VCA", "DLI"}

// Base Random
func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}
func randomEmail() string {
	return randomString(10) + "@email.com"
}
func randomUUID() pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid.New(),
		Valid: true,
	}
}

// Random Data 
func randomFlightNumber() string {
	airline := airlines[rand.Intn(len(airlines))]
	flightNum := rand.Intn(990) + 10
	return fmt.Sprintf("%s%d", airline, flightNum)
}
func randomSeatNumber() (string, int) {
	row := rand.Intn(30) + 1
	letter := []string{"A", "B", "C", "D", "E", "F"}
	col := letter[rand.Intn(len(letter))]
	seatNum := fmt.Sprintf("%d%s-%s", row, col, randomString(4))
	return seatNum, row
}
func randomSeatClass(row int) string {
	switch {
	case row <= 3:
		return "First"
	case row <= 8:
		return "Business"
	default:
		return "Economy"
	}
}
func randomAirport() (string, string) {
	dep := airports[rand.Intn(len(airports))]
	arr := airports[rand.Intn(len(airports))]
	for arr == dep {
		arr = airports[rand.Intn(len(airports))]
	}
	return dep, arr
}
func randomSeatPrice(seatClass string) int64 {
	switch seatClass {
	case "First":
		return int64(4000000 + rand.Intn(2000000))
	case "Business":
		return int64(2000000 + rand.Intn(1500000))
	default: 
		return int64(800000 + rand.Intn(700000))
	}
}
// Random Flight
func randomFlight(t *testing.T) Flight {
	dep, arr := randomAirport()
	now := time.Now().Add(24 * time.Hour)
	boardingAt := now.Add(-40 * time.Minute)
	arrivalAt := now.Add(2 * time.Hour)
	arg := CreateFlightParams{
		ID:               randomUUID(),
		FlightNumber:     randomFlightNumber(),
		DepartureAirport: dep,
		ArrivalAirport:   arr,
		BoardingAt:       pgtype.Timestamptz{Time: boardingAt, Valid: true},
		DepartureAt:      pgtype.Timestamptz{Time: now, Valid: true},
		ArrivalAt:        pgtype.Timestamptz{Time: arrivalAt, Valid: true},
	}

	flight, err := NewStore(testDB).CreateFlight(context.Background(), arg)
	if err != nil {
		t.Fatalf("failed to create flight: %v", err)
	}
	return flight
}
// Random Seat
func randomSeat(t *testing.T, flightID pgtype.UUID) Seat {
	store := NewStore(testDB)
	seatNum, row := randomSeatNumber()
	seatClass := randomSeatClass(row)
	arg := CreateSeatParams{
		ID:       randomUUID(),
		FlightID: flightID,
		SeatNumber: seatNum,
		SeatClass:  randomSeatClass(row),
		Price:      randomSeatPrice(seatClass),
	}
	seat, err := store.CreateSeat(context.Background(), arg)
	if err != nil {
		t.Fatalf("failed to create seat: %v", err)
	}
	return seat
}
// Random User 
func randomUser(t *testing.T) User {
	store := NewStore(testDB)
	
	arg := CreateUserParams{
		ID:          randomUUID(),
		Name:        randomString(10),
		Email:       randomEmail(),
		PhoneNumber: fmt.Sprintf("09%s", randomString(8)),
	}	
	user, err := store.CreateUser(context.Background(), arg)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}	
