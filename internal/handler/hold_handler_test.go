package handler

import (
	"context"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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