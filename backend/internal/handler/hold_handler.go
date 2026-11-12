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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgtype"
)

type HoldService interface {
	CreateHold(
		ctx context.Context,
		req service.CreateHoldRequest,
	) (sqlcdb.SeatHold, error)
	ReleaseHold(
		ctx context.Context,
		holdID pgtype.UUID,
		userID pgtype.UUID,
	) (sqlcdb.SeatHold, error)
}

type HoldHandler struct {
	HoldService HoldService
	
}

func NewHoldHandler(svc HoldService) *HoldHandler {
	return &HoldHandler{
		HoldService: svc,
	}
}

type createHoldHTTPRequest struct {
	SeatID string `json:"seat_id"`
}

type releaseHoldHTTPRequest struct {
    HoldID string `json:"hold_id"`
}
func (h *HoldHandler) CreateHold(w http.ResponseWriter, r *http.Request) {
	var req createHoldHTTPRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// Convert the seatID to pgtype.UUID
	seatID, err := uuid.Parse(req.SeatID)
	if err != nil {
		http.Error(w, "Invalid seat_id", http.StatusBadRequest)
		return
	}
	pgSeatID := pgtype.UUID{
		Bytes: seatID,
		Valid: true,
	}

	// Convert the userID to pgtype.UUID
	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID , err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	pgUserID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	// Convert Flight ID to pgtype.UUID
	flightIDStr := chi.URLParam(r, "flight_id")	
	flightID, err := uuid.Parse(flightIDStr)
	if err != nil {
		http.Error(w, "Invalid flight_id", http.StatusBadRequest)
		return
	}
	pgFlightID := pgtype.UUID{
		Bytes: flightID,
		Valid: true,
	}

	// Call the service to create a hold
	createHoldReq := service.CreateHoldRequest{
		UserID: pgUserID,
		SeatID: pgSeatID,
		FlightID: pgFlightID,
	}
	hold, err := h.HoldService.CreateHold(r.Context(), createHoldReq)
	if errors.Is(err, domain.ErrSeatAlreadyOnHold) {
		http.Error(w, "Seat is already held", http.StatusConflict)
		return
	} else if errors.Is(err, domain.ErrSeatAlreadyBooked) {
		http.Error(w, "Seat is already booked", http.StatusConflict)
		return
	} else if errors.Is(err, domain.ErrSeatNotFound) {
		http.Error(w, "Seat not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to create hold", http.StatusInternalServerError)
		return
	}  else if !hold.ID.Valid || !hold.UserID.Valid || !hold.SeatID.Valid || !hold.HoldToken.Valid {
		http.Error(w, "Invalid hold data", http.StatusInternalServerError)
		return
	} else if !hold.CreatedAt.Valid || !hold.ExpiresAt.Valid {
		http.Error(w, "Invalid hold timestamps", http.StatusInternalServerError)
		return
	} else if hold.ExpiresAt.Time.Before(hold.CreatedAt.Time) {
		http.Error(w, "Hold expiration time is before creation time", http.StatusInternalServerError)
		return
	} 
	// Return the created hold as JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(hold)

}

func (h *HoldHandler) ReleaseHold(w http.ResponseWriter, r *http.Request) {
	var req releaseHoldHTTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	holdID, err := uuid.Parse(req.HoldID)
	if err != nil {
		http.Error(w, "Invalid hold_id", http.StatusBadRequest)
		return
	}
	pgHoldID := pgtype.UUID{
		Bytes: holdID,
		Valid: true,
	}

	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	pgUserID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	releasedHold, err := h.HoldService.ReleaseHold(r.Context(), pgHoldID, pgUserID)
	if errors.Is(err, domain.ErrSeatHoldNotFound) {
		http.Error(w, "Seat hold not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, domain.ErrUnauthorizedRelease) {
		http.Error(w, "Unauthorized to release this hold", http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, "Failed to release hold", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(releasedHold)
}