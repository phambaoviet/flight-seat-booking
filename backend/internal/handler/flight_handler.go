package handler

import (
	"context"
	"encoding/json"
	"flight-booking/internal/service"
	"net/http"
)


type TripType string

const (
	OneWay    TripType = "one-way"
	RoundTrip TripType = "round-trip"
)

type SearchFlightsRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Departure   string `json:"departure"`
	Return      string `json:"return"`
	Adults      int	 	`json:"adults"`
	Children    int	 	`json:"children"`
	Infants     int	 	`json:"infants"`
	TripType    TripType 	`json:"trip_type"`
}

type FlightService interface {
    SearchFlights(
        ctx context.Context,
        req service.SearchFlightsRequest,
    ) (*service.SearchFlightsResponse, error)
}

type FlightOffer struct {
	ID            string `json:"id"`
	TotalAmount   string `json:"total_amount"`
	TotalCurrency string `json:"total_currency"`
}

type FlightHandler struct {
	flightService FlightService
}

func NewFlightHandler(flightService FlightService) *FlightHandler {
	return &FlightHandler{
		flightService: flightService,
	}
}


func (h *FlightHandler) SearchFlights(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read JSON
	var req SearchFlightsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Call the flight service to search for flights
	resp, err := h.flightService.SearchFlights(r.Context(), service.SearchFlightsRequest{
		Origin:      req.Origin,
		Destination: req.Destination,
		Departure:   req.Departure,
		Return:      req.Return,
		Adults:      req.Adults,
		Children:    req.Children,
		Infants:     req.Infants,
		TripType:    service.TripType(req.TripType),
	})

	if err != nil {
		http.Error(w, "Failed to search flights: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the response as JSON

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)


	if _, err := w.Write(jsonBytes); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}

}