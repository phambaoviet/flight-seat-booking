package service

import (
	"context"
	"errors"
	"flight-booking/internal/duffel"
	"fmt"
	"strings"
	"time"
)

type TripType string

const (
	OneWay    TripType = "one-way"
	RoundTrip TripType = "round-trip"
)

type SearchFlightsRequest struct {
	Origin      string
	Destination string
	Departure   string
	Return      string
	Adults      int
	Children    int
	Infants     int
	TripType    TripType
}

type SearchFlightsResponse struct {
	FlightOffers []FlightOffer `json:"flight_offers"`
}

type FlightOffer struct {
	ID            string
	TotalAmount   string
	TotalCurrency string
}

type FlightService interface {
	SearchFlights(
		ctx context.Context,
		req SearchFlightsRequest,
	) (*SearchFlightsResponse, error)
}

type FlightHandler struct {
	FlightService FlightService
}

type DuffelClient interface {
	SearchFlights(ctx context.Context, req duffel.DuffelSearchParams) (*duffel.FlightOfferResponse, error)
}

type flightServiceImpl struct {
	DuffelClient DuffelClient
}

func NewFlightService(duffelClient DuffelClient) *flightServiceImpl {
	return &flightServiceImpl{
		DuffelClient: duffelClient,
	}
}

func (s *flightServiceImpl) SearchFlights(ctx context.Context, req SearchFlightsRequest) (*SearchFlightsResponse, error) {
	duffelReq := duffel.DuffelSearchParams{
		Origin:      req.Origin,
		Destination: req.Destination,
		Departure:   req.Departure,
		Return:      req.Return,
		Adults:      req.Adults,
		Children:    req.Children,
		Infants:     req.Infants,
		TripType:    duffel.TripType(req.TripType),
	}

	duffelResp, err := s.DuffelClient.SearchFlights(ctx, duffelReq)
	if err != nil {
		return nil, err
	}

	flightOffers := make([]FlightOffer, 0, len(duffelResp.Data.Offers))
	for _, offer := range duffelResp.Data.Offers {
		flightOffers = append(flightOffers, FlightOffer{
			ID:            offer.ID,
			TotalAmount:   offer.TotalAmount,
			TotalCurrency: offer.TotalCurrency,
		})
	}

	return &SearchFlightsResponse{
		FlightOffers: flightOffers,
	}, nil
}

func validateSearchFlightsRequest(req SearchFlightsRequest) error {
	// Normalize input for validation
	origin := strings.TrimSpace(req.Origin)
	destination := strings.TrimSpace(req.Destination)
	departure := strings.TrimSpace(req.Departure)
	returnDate := strings.TrimSpace(req.Return)

	// Validate required fields
	if origin == "" {
		return errors.New("origin is required")
	}
	if destination == "" {
		return errors.New("destination is required")
	}
	if departure == "" {
		return errors.New("departure date is required")
	}

	// Validate trip type
	if req.TripType != OneWay && req.TripType != RoundTrip {
		return errors.New("invalid trip type")
	}

	// Validate passenger counts
	if req.Adults < 1 {
		return errors.New("at least one adult passenger is required")
	}
	if req.Children < 0 || req.Infants < 0 {
		return errors.New("number of children and infants cannot be negative")
	}
	if req.Infants > req.Adults {
		return errors.New("number of infants cannot exceed number of adults")
	}

	// Validate route
	if strings.EqualFold(origin, destination) {
		return errors.New("origin and destination cannot be the same")
	}

	// Validate departure date
	const dateLayout = "2006-01-02"

	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		return fmt.Errorf("failed to load timezone: %w", err)
	}

	departureTime, err := time.ParseInLocation(dateLayout, departure, loc)
	if err != nil {
		return errors.New("invalid departure date format, expected YYYY-MM-DD")
	}

	// Compare calendar dates instead of comparing with the current time.
	now := time.Now().In(loc)
	today := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, loc,
	)

	if departureTime.Before(today) {
		return errors.New("departure date cannot be in the past")
	}

	// Validate return date
	if req.TripType == RoundTrip {
		if returnDate == "" {
			return errors.New("return date is required for round-trip flights")
		}

		returnTime, err := time.ParseInLocation(dateLayout, returnDate, loc)
		if err != nil {
			return errors.New("invalid return date format, expected YYYY-MM-DD")
		}

		if returnTime.Before(departureTime) {
			return errors.New("return date cannot be before departure date")
		}
	}

	// One-way requests should not contain a return date.
	if req.TripType == OneWay && returnDate != "" {
		return errors.New("return date must not be provided for one-way flights")
	}

	return nil
}
