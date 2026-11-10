package handler

import (
	"context"
	"errors"
	"flight-booking/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockFlightService struct {
	SearchFlightsFunc func(ctx context.Context, req service.SearchFlightsRequest) (*service.SearchFlightsResponse, error)
}

func (m *mockFlightService) SearchFlights(ctx context.Context, req service.SearchFlightsRequest) (*service.SearchFlightsResponse, error) {
	if m.SearchFlightsFunc != nil {
		return m.SearchFlightsFunc(ctx, req)
	}
	return nil, nil
}

func TestFlightHandler_SearchFlight_Success(t *testing.T) {
	// Set the location to Asia/Ho_Chi_Minh
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}

	departure := time.Now().In(loc).
		AddDate(0, 0, 30).
		Format("2006-01-02")

	called := false
	// Create a mock flight service
	mockService := &mockFlightService{
		SearchFlightsFunc: func(ctx context.Context, req service.SearchFlightsRequest) (*service.SearchFlightsResponse, error) {
			called = true
			return &service.SearchFlightsResponse{
				FlightOffers: []service.FlightOffer{
					{ID: "1", TotalAmount: "100.00", TotalCurrency: "USD"},
					{ID: "2", TotalAmount: "200.00", TotalCurrency: "USD"},
				},
			}, nil
		},
	}

	// Create a flight handler with the mock service
	handler := NewFlightHandler(mockService)

	// Create a HTTP POST request with the test data
	req := httptest.NewRequest(http.MethodPost, "/api/v1/flights/search", strings.NewReader(`{
		"origin": "JFK",
		"destination": "LAX",
		"departure": "`+departure+`",
		"return": "",
		"adults": 1,
		"children": 0,
		"infants": 0,
		"trip_type": "one-way"
	}`))
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the SearchFlights handler
	handler.SearchFlights(rr, req)
	if !called {
		t.Fatal("expected FlightService to be called")
	}
	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Check the response body
	expectedBody := `{"flight_offers":[{"id":"1","total_amount":"100.00","total_currency":"USD"},{"id":"2","total_amount":"200.00","total_currency":"USD"}]}`
	if strings.TrimSpace(rr.Body.String()) != expectedBody {
		t.Errorf("expected response body %s, got %s", expectedBody, rr.Body.String())
	}

}
func TestFlightHandler_SearchFlights_InvalidMethod(t *testing.T) {
	called := false
	// Create a mock flight service
	mockService := &mockFlightService{
		SearchFlightsFunc: func(ctx context.Context, req service.SearchFlightsRequest) (*service.SearchFlightsResponse, error) {
			return &service.SearchFlightsResponse{
				FlightOffers: []service.FlightOffer{
					{ID: "1", TotalAmount: "100.00", TotalCurrency: "USD"},
				},
			}, nil
		},
	}

	// Create a flight handler with the mock service
	handler := NewFlightHandler(mockService)

	// Create a HTTP GET request (invalid method)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/flights/search", nil)

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the SearchFlights handler
	handler.SearchFlights(rr, req)
	if called {
		t.Fatal("FlightService should not be called for invalid HTTP method")
	}
	// Check the response status code
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status code %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestFlightHandler_SearchFlights_InvalidJSON(t *testing.T) {
	called := false
	// Create a mock flight service
	mockService := &mockFlightService{
		SearchFlightsFunc: func(ctx context.Context, req service.SearchFlightsRequest) (*service.SearchFlightsResponse, error) {
			called = true
			return nil, nil
		},
	}

	// Create a flight handler with the mock service
	handler := NewFlightHandler(mockService)

	// Create a HTTP POST request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/flights/search", strings.NewReader(`{"origin": "JFK", "destination": "LAX", "departure": "2024-12-01",`))
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the SearchFlights handler
	handler.SearchFlights(rr, req)
	if called {
		t.Fatal("FlightService should not be called for invalid JSON")
	}
	// Check the response status code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestFlightHandler_SearchFlights_ServiceError(t *testing.T) {
	// Set the location to Asia/Ho_Chi_Minh
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}

	departure := time.Now().In(loc).
		AddDate(0, 0, 30).
		Format("2006-01-02")

	called := false
	// Create a mock flight service that returns an error
	mockService := &mockFlightService{
		SearchFlightsFunc: func(ctx context.Context, req service.SearchFlightsRequest) (*service.SearchFlightsResponse, error) {
			called = true
			return nil, errors.New("service error")
		},
	}

	// Create a flight handler with the mock service
	handler := NewFlightHandler(mockService)

	// Create a HTTP POST request with valid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/flights/search", strings.NewReader(`{
		"origin": "JFK",
		"destination": "LAX",
		"departure": "`+departure+`",
		"return": "",
		"adults": 1,
		"children": 0,
		"infants": 0,
		"trip_type": "one-way"
	}`))
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the SearchFlights handler
	handler.SearchFlights(rr, req)
	if !called {
		t.Fatal("expected FlightService to be called")
	}
	// Check the response status code
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}
