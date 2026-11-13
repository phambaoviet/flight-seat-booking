package handler

import (
	"context"
	"encoding/json"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"
	"flight-booking/internal/middleware"
	"flight-booking/internal/service"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type BookingService interface {
	ConfirmBooking(ctx context.Context, req service.ConfirmBookingRequest) (sqlcdb.Booking, error)
}

type BookingHandler struct {
	bookingService BookingService
}

func NewBookingHandler(bookingService BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

type confirmBookingHTTPRequest struct {
	SeatID    string `json:"seat_id"`
	HoldToken string `json:"hold_token"`
	FlightID  string `json:"flight_id"`
}

func (h *BookingHandler) ConfirmBooking(w http.ResponseWriter, r *http.Request) {
	var req confirmBookingHTTPRequest
	// Check if method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// Get userID from context
	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// parse userID from string to uuid
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	seatID, err := uuid.Parse(req.SeatID)
	if err != nil {
		http.Error(w, "Invalid seat_id", http.StatusBadRequest)
		return
	}

	holdToken, err := uuid.Parse(req.HoldToken)
	if err != nil {
		http.Error(w, "Invalid hold_token", http.StatusBadRequest)
		return
	}

	flightID, err := uuid.Parse(req.FlightID)
	if err != nil {
		http.Error(w, "Invalid flight_id", http.StatusBadRequest)
		return
	}

	bookingReq := service.ConfirmBookingRequest{
		SeatID:    pgtype.UUID{Bytes: seatID, Valid: true},
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		HoldToken: pgtype.UUID{Bytes: holdToken, Valid: true},
		FlightID:  pgtype.UUID{Bytes: flightID, Valid: true},
	}

	booking, err := h.bookingService.ConfirmBooking(r.Context(), bookingReq)
	if errors.Is(err, domain.ErrSeatNotFound) {
		http.Error(w, "Seat not found", http.StatusNotFound)
		return
	} else if errors.Is(err, domain.ErrSeatAlreadyBooked) {
		http.Error(w, "Seat already booked", http.StatusConflict)
		return
	} else if errors.Is(err, domain.ErrSeatHoldNotFound) {
		http.Error(w, "Hold not found", http.StatusNotFound)
		return
	}  else if errors.Is(err, domain.ErrHoldTokenMismatch) {
		http.Error(w, "Hold token mismatch", http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}