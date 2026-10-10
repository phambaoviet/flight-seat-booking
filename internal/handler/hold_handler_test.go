package handler

import (
	"context"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

type mockHoldService struct {
	CreateHoldFunc func(
        ctx context.Context,
        req service.CreateHoldRequest,
    ) (sqlcdb.SeatHold, error)
}

func (m *mockHoldService) CreateHold(ctx context.Context, arg service.CreateHoldRequest) (sqlcdb.SeatHold, error) {
	return m.CreateHoldFunc(ctx, arg)
}


func TestCreateHold_InvalidSeatID(t *testing.T) {
	invalidJSON := `{"seat_id": "invalid-uuid", "user_id": "invalid-uuid"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/holds",
		strings.NewReader(invalidJSON),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	serviceCalled := false

	// Create a mock service that sets serviceCalled to true if called
	mockSvc := &mockHoldService{
		CreateHoldFunc: func(
            ctx context.Context,
            gotArg service.CreateHoldRequest,
        ) (sqlcdb.SeatHold, error) {
            serviceCalled = true
            return sqlcdb.SeatHold{}, nil
        },
	}
	handler := NewHoldHandler(mockSvc)
    handler.CreateHold(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.False(t, serviceCalled)
}
func TestCreateHold_InvalidUserID(t *testing.T){
	invalidJSON := `{"seat_id": "550e8400-e29b-41d4-a716-446655440000", "user_id": "invalid-uuid"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/holds",
		strings.NewReader(invalidJSON),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	serviceCalled := false

	// Create a mock service that sets serviceCalled to true if called
	mockSvc := &mockHoldService{
		CreateHoldFunc: func(
			ctx context.Context,
			gotArg service.CreateHoldRequest,
		) (sqlcdb.SeatHold, error) {
			serviceCalled = true
			return sqlcdb.SeatHold{}, nil
		},
	}	
	
	// Create a handler with the mock service
	handler := NewHoldHandler(mockSvc)
	handler.CreateHold(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.False(t, serviceCalled)
}
func TestCreateHold_InvalidJSON(t *testing.T) {
	invalidJSON := `{"seat_id": "550e8400-e29b-41d4-a716-446655440000", "user_id": }` 

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/holds",
		strings.NewReader(invalidJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()

	// Create a mock service that sets serviceCalled to true if called
	serviceCalled := false
	mockSvc := &mockHoldService{
		CreateHoldFunc: func(
			ctx context.Context,
			gotArg service.CreateHoldRequest,
		) (sqlcdb.SeatHold, error) {
			serviceCalled = true
			return sqlcdb.SeatHold{}, nil
		},
	}

	handler := NewHoldHandler(mockSvc)
	handler.CreateHold(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.False(t, serviceCalled)
}

func TestCreateHold_ServiceError(t *testing.T) {
	validJSON := `{
		"seat_id": "550e8400-e29b-41d4-a716-446655440000",
		"user_id": "123e4567-e89b-12d3-a456-426614174000"
	}`
	
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/holds",
		strings.NewReader(validJSON),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	serviceCalled := false
	// Create a mock service that returns an error
	mockSvc := &mockHoldService{
		CreateHoldFunc: func(
			ctx context.Context,
			gotArg service.CreateHoldRequest,
		) (sqlcdb.SeatHold, error) {
			serviceCalled = true
			return sqlcdb.SeatHold{}, assert.AnError
		},
	}

	handler := NewHoldHandler(mockSvc)
	handler.CreateHold(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.True(t, serviceCalled)
}
func TestCreateHold_Success(t *testing.T) {
	validJSON := `{
		"seat_id": "550e8400-e29b-41d4-a716-446655440000",
		"user_id": "123e4567-e89b-12d3-a456-426614174000"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/holds",
		strings.NewReader(validJSON),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	serviceCalled := false
	expectedHold := sqlcdb.SeatHold{
		ID:        pgtype.UUID{Bytes: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), Valid: true},
		UserID:    pgtype.UUID{Bytes: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), Valid: true},
		HoldToken: pgtype.UUID{Bytes: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(10 * time.Minute), Valid: true},
	}

	// Create a mock service that returns the expected hold
	mockSvc := &mockHoldService{
		CreateHoldFunc: func(
			ctx context.Context,
			gotArg service.CreateHoldRequest,
		) (sqlcdb.SeatHold, error) {
			serviceCalled = true
			return expectedHold, nil
		},
	}

	handler := NewHoldHandler(mockSvc)
	handler.CreateHold(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.True(t, serviceCalled)
}