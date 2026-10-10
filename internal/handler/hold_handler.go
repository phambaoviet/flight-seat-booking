package handler

import (
	"context"
	"encoding/json"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/service"
	"net/http"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgtype"
)

type HoldService interface {
    CreateHold(
        ctx context.Context,
        req service.CreateHoldRequest,
    ) (sqlcdb.SeatHold, error)
}

type HoldHandler struct {
	HoldService interface {
		CreateHold(ctx context.Context, req service.CreateHoldRequest) (sqlcdb.SeatHold, error)
	}
}
func NewHoldHandler(svc HoldService) *HoldHandler {
	return &HoldHandler{
		HoldService: svc,
	}
}
type createHoldHTTPRequest struct {
    SeatID string `json:"seat_id"`
	UserID string `json:"user_id"`
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
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}
	pgUserID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	// Call the service to create a hold
	createHoldReq := service.CreateHoldRequest{
		UserID: pgUserID,
		SeatID: pgSeatID,
	}
	hold, err := h.HoldService.CreateHold(r.Context(), createHoldReq)
	if err != nil {
		http.Error(w, "Failed to create hold", http.StatusInternalServerError)
		return
	}
	// Return the created hold as JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(hold)
	
}
