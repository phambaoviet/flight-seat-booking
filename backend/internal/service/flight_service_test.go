package service

import (
	"context"
	"errors"
	"flight-booking/internal/duffel"
	"fmt"
	"testing"
	"time"
)

type mockDuffelClient struct {
	searchFlightsFunc func(
		ctx context.Context,
		req duffel.DuffelSearchParams,
	) (*duffel.FlightOfferResponse, error)
}

func (m *mockDuffelClient) SearchFlights(
	ctx context.Context,
	req duffel.DuffelSearchParams,
) (*duffel.FlightOfferResponse, error) {
	if m.searchFlightsFunc != nil {
		return m.searchFlightsFunc(ctx, req)
	}
	return nil, fmt.Errorf("SearchFlights method not implemented in mock")
}
func TestNormalizeSearchFlightsRequest(t *testing.T) {
	// Define test cases
	testcases := []struct {
		name     string
		input    SearchFlightsRequest
		expected SearchFlightsRequest
	}{
		{
			name: "normalize origin and destination with lowercase",
			input: SearchFlightsRequest{
				Origin:      "jfk",
				Destination: "lax",
				Departure:   "2024-07-01",
				Return:      "2024-07-10",
			},
			expected: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   "2024-07-01",
				Return:      "2024-07-10",
			},
		},
		{
			name: "normalize origin and destination with mixed case and spaces",
			input: SearchFlightsRequest{
				Origin:      "  jFk  ",
				Destination: "  LaX  ",
				Departure:   "2024-07-01",
				Return:      "2024-07-10",
			},
			expected: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   "2024-07-01",
				Return:      "2024-07-10",
			},
		},
		{
			name: "normalize empty and whitespace fields",
			input: SearchFlightsRequest{
				Origin:      "   ",
				Destination: "",
				Departure:   "",
				Return:      " ",
			},
			expected: SearchFlightsRequest{
				Origin:      "",
				Destination: "",
				Departure:   "",
				Return:      "",
			},
		},
		{
			name: "normalize departure and return dates",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   " 2024-07-01 ",
				Return:      " 2024-07-10 ",
			},
			expected: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   "2024-07-01",
				Return:      "2024-07-10",
			},
		},
	}

	// Run the test cases
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeSearchFlightsRequest(tc.input)
			if result != tc.expected {
				t.Fatalf("expected %+v, got %+v", tc.expected, result)
			}
		})
	}

}

func TestValidateSearchFlightsRequest(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().In(loc)

	valiDep := now.AddDate(0, 0, 30).Format("2006-01-02")  // 30 days in the future
	pastDep := now.AddDate(0, 0, -30).Format("2006-01-02") // 30 days in the past
	laterRet := now.AddDate(0, 0, 60).Format("2006-01-02") // 60 days in the future
	testcases := []struct {
		name        string
		input       SearchFlightsRequest
		expectError bool
	}{
		{
			name: "valid request",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   valiDep,
				TripType:    "one-way",
				Adults:      1,
				Children:    0,
				Infants:     0,
			},
			expectError: false,
		},
		{
			name: "missing origin",
			input: SearchFlightsRequest{
				Origin:      "",
				Destination: "LAX",
				Departure:   "2024-07-01",
			},
			expectError: true,
		},
		{
			name: "missing destination",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "",
				Departure:   "2024-07-01",
			},
			expectError: true,
		},
		{
			name: "missing departure",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   "",
			},
			expectError: true,
		},
		{
			name: "invalid trip type",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   "2024-07-01",
				TripType:    "invalid-type",
			},
			expectError: true,
		},
		{
			name: "return date before departure",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   valiDep,
				Return:      pastDep,
				Adults:      1,
				TripType:    "round-trip",
			},
			expectError: true,
		},
		{
			name: "infants exceed adults",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   valiDep,
				TripType:    "one-way",
				Adults:      1,
				Infants:     2,
			},
			expectError: true,
		},
		{
			name: "origin and destination are the same",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "JFK",
				Departure:   valiDep,
			},
			expectError: true,
		},
		{
			name: "valid round-trip request",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   valiDep,
				Return:      laterRet,
				TripType:    "round-trip",
				Adults:      1,
				Children:    1,
				Infants:     0,
			},
			expectError: false,
		},
		{
			name: "negative children count",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   valiDep,
				TripType:    "one-way",
				Adults:      1,
				Children:    -1,
			},
			expectError: true,
		},
		{
			name: "negative infants count",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   valiDep,
				TripType:    "one-way",
				Adults:      1,
				Infants:     -1,
			},
			expectError: true,
		},
		{
			name: "no adults",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   valiDep,
				TripType:    "one-way",
				Adults:      0,
			},
			expectError: true,
		},
		{
			name: "valid request with whitespace",
			input: SearchFlightsRequest{
				Origin:      "  JFK  ",
				Destination: "  LAX  ",
				Departure:   " " + valiDep + " ",
				TripType:    "one-way",
				Adults:      1,
			},
			expectError: false,
		},
		{
			name: "invalid departure format",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   "07-01-2024", // Invalid format
				TripType:    "one-way",
				Adults:      1,
			},
			expectError: true,
		},
		{
			name: "invalid return format",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   valiDep,
				Return:      "07-10-2024", // Invalid format
				TripType:    "round-trip",
				Adults:      1,
			},
			expectError: true,
		},
		{
			name: "round-trip missing return date",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   valiDep,
				Return:      "",
				TripType:    "round-trip",
				Adults:      1,
			},
			expectError: true,
		},
		{
			name: "one-way with return date provided",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   valiDep,
				Return:      laterRet,
				TripType:    "one-way",
				Adults:      1,
			},
			expectError: true, // Return date must not be provided for one-way flights
		},
		{
			name: "departure date in the past",
			input: SearchFlightsRequest{
				Origin:      "JFK",
				Destination: "LAX",
				Departure:   pastDep,
				TripType:    "one-way",
				Adults:      1,
			},
			expectError: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			normalized := normalizeSearchFlightsRequest(tc.input)
			err := validateSearchFlightsRequest(normalized)
			if tc.expectError && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}
		})
	}
}

func TestFlightService_SearchFlights_Success(t *testing.T) {

	// create a departure date 30 days in the future
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	departure := time.Now().In(loc).AddDate(0, 0, 30).Format("2006-01-02")

	// Create a mock Duffel client that returns a successful response
	mockClient := &mockDuffelClient{
		searchFlightsFunc: func(ctx context.Context, req duffel.DuffelSearchParams) (*duffel.FlightOfferResponse, error) {
			// Verify request mapping
			if req.Origin != "JFK" {
				t.Errorf("expected origin JFK, got %s", req.Origin)
			}

			if req.Destination != "LAX" {
				t.Errorf("expected destination LAX, got %s", req.Destination)
			}

			if req.Departure != departure {
				t.Errorf("expected departure %s, got %s", departure, req.Departure)
			}

			if req.Adults != 1 {
				t.Errorf("expected 1 adult, got %d", req.Adults)
			}

			if req.TripType != duffel.TripType(OneWay) {
				t.Errorf("expected trip type one-way, got %s", req.TripType)
			}
			return &duffel.FlightOfferResponse{

				Data: duffel.DuffelOfferRequestData{
					ID: "orq_test",
					Offers: []duffel.DuffelOffer{
						{
							ID:            "offer_1",
							TotalAmount:   "500.00",
							TotalCurrency: "USD",
						},
					},
				},
			}, nil
		},
	}

	// Create the flight service with the mock client
	service := NewFlightService(mockClient)

	// Prepare a valid search request
	req := SearchFlightsRequest{
		Origin:      "JFK",
		Destination: "LAX",
		Departure:   departure,
		TripType:    "one-way",
		Adults:      1,
	}

	// Call the SearchFlights method
	resp, err := service.SearchFlights(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Validate the response
	if len(resp.FlightOffers) != 1 {
		t.Fatalf("expected 1 flight offer, got %d", len(resp.FlightOffers))
	}
	if resp.FlightOffers[0].ID != "offer_1" {
		t.Fatalf("expected offer ID 'offer_1', got '%s'", resp.FlightOffers[0].ID)
	}
	if resp.FlightOffers[0].TotalAmount != "500.00" {
		t.Fatalf("expected total amount '500.00', got '%s'", resp.FlightOffers[0].TotalAmount)
	}
	if resp.FlightOffers[0].TotalCurrency != "USD" {
		t.Fatalf("expected total currency 'USD', got '%s'", resp.FlightOffers[0].TotalCurrency)
	}
}

func TestFlightService_SearchFlights_DuffelError(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}

	departure := time.Now().In(loc).
		AddDate(0, 0, 30).
		Format("2006-01-02")
	var errDuffelUnavailable = errors.New("Duffel API unavailable")
	mockClient := &mockDuffelClient{
		searchFlightsFunc: func(ctx context.Context, req duffel.DuffelSearchParams) (*duffel.FlightOfferResponse, error) {

			return nil, errDuffelUnavailable
		},
	}

	service := NewFlightService(mockClient)

	req := SearchFlightsRequest{
		Origin:      "JFK",
		Destination: "LAX",
		Departure:   departure,
		TripType:    "one-way",
		Adults:      1,
	}

	resp, err := service.SearchFlights(context.Background(), req)
	if !errors.Is(err, errDuffelUnavailable) {
		t.Fatalf("expected Duffel error, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
}
func TestFlightService_SearchFlights_NilResponse(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}

	departure := time.Now().In(loc).
		AddDate(0, 0, 30).
		Format("2006-01-02")
	mockClient := &mockDuffelClient{
		searchFlightsFunc: func(ctx context.Context, req duffel.DuffelSearchParams) (*duffel.FlightOfferResponse, error) {
			return nil, nil // Simulate empty response
		},
	}

	service := NewFlightService(mockClient)

	req := SearchFlightsRequest{
		Origin:      "JFK",
		Destination: "LAX",
		Departure:   departure,
		TripType:    "one-way",
		Adults:      1,
	}

	resp, err := service.SearchFlights(context.Background(), req)

	if err == nil {
		t.Fatalf("expected error due to empty Duffel response, got nil")
	}
	if resp != nil {
		t.Fatalf("expected nil response due to empty Duffel response, got %+v", resp)
	}
}

func TestFlightService_SearchFlights_InvalidRequest(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}

	departure := time.Now().In(loc).
		AddDate(0, 0, 30).
		Format("2006-01-02")
	service := NewFlightService(nil) // Duffel client won't be called due to validation failure

	req := SearchFlightsRequest{
		Origin:      "JFK",
		Destination: "JFK", // Invalid: same as origin
		Departure:   departure,
		TripType:    "one-way",
		Adults:      1,
	}

	resp, err := service.SearchFlights(context.Background(), req)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
	if resp != nil {
		t.Fatalf("expected nil response due to validation error, got %+v", resp)
	}
}
