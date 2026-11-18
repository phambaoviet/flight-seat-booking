package handler

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"
	"flight-booking/internal/middleware"
	"flight-booking/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

type MockBookingService struct {
	ConfirmBookingFunc func(ctx context.Context, req service.ConfirmBookingRequest) (sqlcdb.Booking, error)
	CancelBookingFunc  func(ctx context.Context, req service.CancelBookingRequest) (sqlcdb.Booking, error)
}

func (m *MockBookingService) ConfirmBooking(ctx context.Context, req service.ConfirmBookingRequest) (sqlcdb.Booking, error) {
	if m.ConfirmBookingFunc != nil {
		return m.ConfirmBookingFunc(ctx, req)
	}

	panic("ConfirmBookingFunc is not configured")
}
func (m *MockBookingService) CancelBooking(ctx context.Context, req service.CancelBookingRequest) (sqlcdb.Booking, error) {
	if m.CancelBookingFunc != nil {
		return m.CancelBookingFunc(ctx, req)
	}
	panic("CancelBookingFunc is not configured")
}

// Test cases for ConfirmBooking
func TestBooking_Success(t *testing.T) {
	// Create a mock booking service
	userID := uuid.New().String()
	seatID := uuid.New().String()
	holdToken := uuid.New().String()
	flightID := uuid.New().String()
	bookingID := uuid.New()

	// Create a mock booking service
	serviceCalled := false
	mockService := &MockBookingService{
		ConfirmBookingFunc: func(ctx context.Context, req service.ConfirmBookingRequest) (sqlcdb.Booking, error) {
			serviceCalled = true
			assert.Equal(t, userID, uuid.UUID(req.UserID.Bytes).String())
			assert.Equal(t, seatID, uuid.UUID(req.SeatID.Bytes).String())
			assert.Equal(t, holdToken, uuid.UUID(req.HoldToken.Bytes).String())
			assert.Equal(t, flightID, uuid.UUID(req.FlightID.Bytes).String())
			return sqlcdb.Booking{
				ID:          pgtype.UUID{Bytes: bookingID, Valid: true},
				UserID:      req.UserID,
				BookingCode: "test-booking-code",
			}, nil
		},
	}

	handler := NewBookingHandler(mockService)

	recorder := httptest.NewRecorder()

	requestBody := `{
		"seat_id": "` + seatID + `",
		"hold_token": "` + holdToken + `",
		"flight_id": "` + flightID + `"
	}`
	req := httptest.NewRequest("POST", "/api/v1/holds/confirm", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserID(req.Context(), userID)
	req = req.WithContext(ctx)

	// Call the handler
	handler.ConfirmBooking(recorder, req)

	// Check the response
	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.True(t, serviceCalled, "Expected ConfirmBooking to be called")

}

// table-driven tests for different error scenarios
func TestBooking_RequestValidation(t *testing.T) {
	userID := uuid.New().String()
	seatID := uuid.New().String()
	holdToken := uuid.New().String()
	flightID := uuid.New().String()

	serviceCalled := false

	testCases := []struct {
		name           string
		userID         string
		requestBody    string
		method         string
		expectedStatus int
	}{
		{
			name:   "Invalid seat_id",
			userID: userID,
			requestBody: `{
				"seat_id": "invalid-uuid",
				"hold_token": "` + holdToken + `",
				"flight_id": "` + flightID + `"
			}`,
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid hold_token",
			userID: userID,
			requestBody: `{
				"seat_id": "` + seatID + `",
				"hold_token": "invalid-uuid",
				"flight_id": "` + flightID + `"
			}`,
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid flight_id",
			userID: userID,
			requestBody: `{
				"seat_id": "` + seatID + `",
				"hold_token": "` + holdToken + `",
				"flight_id": "invalid-uuid"
			}`,
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid Method",
			userID: userID,
			requestBody: `{
				"seat_id": "` + seatID + `",
				"hold_token": "` + holdToken + `",
				"flight_id": "` + flightID + `"
			}`,
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Invalid JSON",
			userID: userID,
			requestBody: `{
				"seat_id": "` + seatID + `",
				"hold_token": "` + holdToken + `",
				"flight_id": "` + flightID + `"
			`, // Missing closing brace
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Missing userID in context",
			userID: "",
			requestBody: `{
				"seat_id": "` + seatID + `",
				"hold_token": "` + holdToken + `",
				"flight_id": "` + flightID + `"
			}`,
			method:         http.MethodPost,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockBookingService{
				ConfirmBookingFunc: func(ctx context.Context, req service.ConfirmBookingRequest) (sqlcdb.Booking, error) {
					serviceCalled = true
					return sqlcdb.Booking{}, nil
				},
			}
			handler := NewBookingHandler(mockService)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/api/v1/holds/confirm", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			if tc.userID != "" {
				ctx := middleware.SetUserID(req.Context(), tc.userID)
				req = req.WithContext(ctx)
			}

			handler.ConfirmBooking(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.False(t, serviceCalled, "Expected ConfirmBooking not to be called for invalid requests")
		})
	}
}

func TestBooking_ServiceErrors(t *testing.T) {
	serviceCalled := false
	userID := uuid.New().String()
	seatID := uuid.New().String()
	holdToken := uuid.New().String()
	flightID := uuid.New().String()

	testCases := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "Seat already booked",
			serviceError:   domain.ErrSeatAlreadyBooked,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Hold not found",
			serviceError:   domain.ErrSeatHoldNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Hold token mismatch",
			serviceError:   domain.ErrHoldTokenMismatch,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Seat not found",
			serviceError:   domain.ErrSeatNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "internal server error",
			serviceError:   domain.ErrBookingNotFound,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockBookingService{
				ConfirmBookingFunc: func(ctx context.Context, req service.ConfirmBookingRequest) (sqlcdb.Booking, error) {
					serviceCalled = true
					return sqlcdb.Booking{}, tc.serviceError
				},
			}
			handler := NewBookingHandler(mockService)

			recorder := httptest.NewRecorder()
			requestBody := `{
				"seat_id": "` + seatID + `",
				"hold_token": "` + holdToken + `",
				"flight_id": "` + flightID + `"
			}`
			req := httptest.NewRequest("POST", "/api/v1/holds/confirm", strings.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.SetUserID(req.Context(), userID)
			req = req.WithContext(ctx)

			handler.ConfirmBooking(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.True(t, serviceCalled)
		})
	}
}

// Test cases for CancelBooking
func TestCancelBooking_Success(t *testing.T) {
	// Create a mock booking service
	bookingID := uuid.New()
	bookingCode := "test-booking-code"
	userID := uuid.New().String()

	serviceCalled := false
	mockService := &MockBookingService{
		CancelBookingFunc: func(ctx context.Context, req service.CancelBookingRequest) (sqlcdb.Booking, error) {
			serviceCalled = true
			assert.Equal(t, bookingID.String(), uuid.UUID(req.BookingID.Bytes).String())
			assert.Equal(t, bookingCode, req.BookingCode)
			assert.Equal(t, userID, uuid.UUID(req.UserID.Bytes).String())
			return sqlcdb.Booking{
				ID:          pgtype.UUID{Bytes: bookingID, Valid: true},
				UserID:      pgtype.UUID{Bytes: uuid.MustParse(userID), Valid: true},
				BookingCode: bookingCode,
			}, nil
		},
	}

	handler := NewBookingHandler(mockService)

	recorder := httptest.NewRecorder()

	requestBody := `{
		"booking_id": "` + bookingID.String() + `",
		"booking_code": "` + bookingCode + `"
	}`
	req := httptest.NewRequest("POST", "/api/v1/bookings/cancel", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserID(req.Context(), userID)
	req = req.WithContext(ctx)

	// Call the handler
	handler.CancelBooking(recorder, req)

	// Check the response
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.True(t, serviceCalled, "Expected CancelBooking to be called")
}

func TestCancelBooking_RequestValidation(t *testing.T) {
	bookingID := uuid.New()
	bookingCode := "test-booking-code"
	userID := uuid.New().String()

	testCases := []struct {
		name           string
		userID         string
		requestBody    string
		method         string
		expectedStatus int
	}{
		{
			name:   "Invalid booking_id",
			userID: userID,
			requestBody: `{
				"booking_id": "invalid-uuid",
				"booking_code": "` + bookingCode + `"
			}`,
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid Method",
			userID: userID,
			requestBody: `{
				"booking_id": "` + bookingID.String() + `",
				"booking_code": "` + bookingCode + `"
			}`,
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Invalid JSON",
			userID: userID,
			requestBody: `{
				"booking_id": "` + bookingID.String() + `",
				"booking_code": "` + bookingCode + `"
			`, // Missing closing brace
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Missing userID in context",
			userID: "",
			requestBody: `{
				"booking_id": "` + bookingID.String() + `",
				"booking_code": "` + bookingCode + `"
			}`,
			method:         http.MethodPost,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Missing booking_code",
			userID: userID,
			requestBody: `{
				"booking_id": "` + bookingID.String() + `"
			}`,
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serviceCalled := false
			mockService := &MockBookingService{
				CancelBookingFunc: func(ctx context.Context, req service.CancelBookingRequest) (sqlcdb.Booking, error) {
					serviceCalled = true
					return sqlcdb.Booking{}, nil
				},
			}
			handler := NewBookingHandler(mockService)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/api/v1/bookings/cancel", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			if tc.userID != "" {
				ctx := middleware.SetUserID(req.Context(), tc.userID)
				req = req.WithContext(ctx)
			}

			handler.CancelBooking(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.False(t, serviceCalled, "Expected CancelBooking not to be called for invalid requests")
		})
	}
}

func TestCancelBooking_ServiceErrors(t *testing.T) {
	bookingID := uuid.New()
	bookingCode := "test-booking-code"
	userID := uuid.New().String()

	testCases := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "Booking not found",
			serviceError:   domain.ErrBookingNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "internal server error",
			serviceError:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serviceCalled := false
			mockService := &MockBookingService{
				CancelBookingFunc: func(ctx context.Context, req service.CancelBookingRequest) (sqlcdb.Booking, error) {
					serviceCalled = true
					return sqlcdb.Booking{}, tc.serviceError
				},
			}
			handler := NewBookingHandler(mockService)

			recorder := httptest.NewRecorder()
			requestBody := `{
				"booking_id": "` + bookingID.String() + `",
				"booking_code": "` + bookingCode + `"
			}`
			req := httptest.NewRequest("POST", "/api/v1/bookings/cancel", strings.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.SetUserID(req.Context(), userID)
			req = req.WithContext(ctx)

			handler.CancelBooking(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.True(t, serviceCalled)
		})
	}
}
