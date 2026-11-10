package handler

import (
	"context"
	"flight-booking/internal/service"
)

type FlightService interface {
	SearchFlights(
		ctx context.Context,
		req service.SearchFlightsRequest,
	) (*service.SearchFlightsResponse, error)
}

type FlightHandler struct {
	FlightService FlightService
}

