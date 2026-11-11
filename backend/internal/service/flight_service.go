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
	CabinClass  string
	TripType    TripType
}

type SearchFlightsResponse struct {
	FlightOffers []FlightOffer `json:"flight_offers"`
}

type FlightOffer struct {
	ID            string        `json:"id"`
	TotalAmount   string        `json:"total_amount"`
	TotalCurrency string        `json:"total_currency"`
	Slices        []FlightSlice `json:"slices,omitempty"`
}

type FlightService interface {
	SearchFlights(
		ctx context.Context,
		req SearchFlightsRequest,
	) (*SearchFlightsResponse, error)
}

type DuffelClient interface {
	SearchFlights(ctx context.Context, req duffel.DuffelSearchParams) (*duffel.FlightOfferResponse, error)
}

type flightServiceImpl struct {
	duffelClient DuffelClient
}

type FlightSlice struct {
	Origin      ServiceAirport
	Destination ServiceAirport
	Segments    []FlightSegment
}
type ServiceCarrier struct {
	IATA string
	Name string
}
type ServiceAircraft struct {
	IATA string
	Name string
}
type ServiceAirport struct {
	IATA string
	Name string
}
type FlightSegment struct {
	ID               string
	Duration         string
	MarketingCarrier ServiceCarrier
	Aircraft         ServiceAircraft
	Origin           ServiceAirport
	Destination      ServiceAirport
	Departure        string
	Arrival          string
}

func NewFlightService(duffelClient DuffelClient) *flightServiceImpl {
	return &flightServiceImpl{
		duffelClient: duffelClient,
	}
}

func (s *flightServiceImpl) SearchFlights(ctx context.Context, req SearchFlightsRequest) (*SearchFlightsResponse, error) {
	// Normalize the request before validation
	req = normalizeSearchFlightsRequest(req)

	// Validate the request
	if err := validateSearchFlightsRequest(req); err != nil {
		return nil, fmt.Errorf("invalid search flights request: %w", err)
	}

	// Build the Duffel search parameters
	duffelReq := duffel.DuffelSearchParams{
		Origin:      req.Origin,
		Destination: req.Destination,
		Departure:   req.Departure,
		Return:      req.Return,
		Adults:      req.Adults,
		Children:    req.Children,
		Infants:     req.Infants,
		CabinClass:  req.CabinClass,
		TripType:    duffel.TripType(req.TripType),
	}

	// Call the Duffel API to search for flights
	duffelResp, err := s.duffelClient.SearchFlights(ctx, duffelReq)

	if err != nil {
		return nil, fmt.Errorf("failed to search flights with Duffel: %w", err)
	}
	if duffelResp == nil {
		return nil, errors.New("Duffel returned an empty response")
	}

	// Map the Duffel response to the service response
	flightOffers := mapDuffelOffers(duffelResp.Data.Offers)

	// Return the response
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
func normalizeSearchFlightsRequest(req SearchFlightsRequest) SearchFlightsRequest {

	req.Origin = strings.ToUpper(strings.TrimSpace(req.Origin))
	req.Destination = strings.ToUpper(strings.TrimSpace(req.Destination))

	req.Departure = strings.TrimSpace(req.Departure)
	req.Return = strings.TrimSpace(req.Return)

	return req
}
func mapDuffelOffers(offers []duffel.DuffelOffer) []FlightOffer {
	flightOffers := make([]FlightOffer, 0, len(offers))
	for _, offer := range offers {
		flightOffers = append(flightOffers, FlightOffer{
			ID:            offer.ID,
			TotalAmount:   offer.TotalAmount,
			TotalCurrency: offer.TotalCurrency,
			Slices:        mapDuffelSlices(offer.Slices),
		})
	}
	return flightOffers
}
func mapDuffelSegments(segments []duffel.DuffelSegment) []FlightSegment {
	flightSegments := make([]FlightSegment, 0, len(segments))
	for _, segment := range segments {
		flightSegments = append(flightSegments, FlightSegment{
			ID:       segment.ID,
			Duration: segment.Duration,
			MarketingCarrier: ServiceCarrier{
				IATA: segment.MarketingCarrier.IATA,
				Name: segment.MarketingCarrier.Name,
			},
			Aircraft: ServiceAircraft{
				IATA: segment.Aircraft.IATA,
				Name: segment.Aircraft.Name,
			},
			Origin: ServiceAirport{
				IATA: segment.Origin.IATA,
				Name: segment.Origin.Name,
			},
			Destination: ServiceAirport{
				IATA: segment.Destination.IATA,
				Name: segment.Destination.Name,
			},
			Departure: segment.Departure,
			Arrival:   segment.Arrival,
		})
	}
	return flightSegments
}
func mapDuffelSlices(offers []duffel.DuffelOfferSlice) []FlightSlice {
	flightSlices := make([]FlightSlice, 0, len(offers))
	for _, slice := range offers {
		flightSlices = append(flightSlices, FlightSlice{
			Origin: ServiceAirport{
				IATA: slice.Origin.IATA,
				Name: slice.Origin.Name,
			},
			Destination: ServiceAirport{
				IATA: slice.Destination.IATA,
				Name: slice.Destination.Name,
			},
			Segments: mapDuffelSegments(slice.Segments),
		})
	}
	return flightSlices
}
