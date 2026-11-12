package handler

import (
	"context"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"
	"flight-booking/internal/middleware"
	"flight-booking/internal/service"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
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
type mockTokenService struct {
    VerifyFunc func(token string) (*service.Claims, error)
}

func (m *mockTokenService) Generate(userID string) (string, error) {
    return "", nil
}

func (m *mockTokenService) Verify(token string) (*service.Claims, error) {
    return m.VerifyFunc(token)
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
func TestCreateHold_InvalidUserID(t *testing.T) {
	invalidJSON := `{"seat_id": "550e8400-e29b-41d4-a716-446655440000"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/flights/550e8400-e29b-41d4-a716-446655440001/hold",
		strings.NewReader(invalidJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-test-token")


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
	holdHandler := NewHoldHandler(mockSvc)
	mockTokenService := &mockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{
				UserID: "invalid-uuid",
			}, nil
		},
	}
	authMiddleware := middleware.NewAuthMiddleware(mockTokenService)

	router := chi.NewRouter()
	router.With(authMiddleware.Authenticate).
		Post("/api/v1/flights/{flight_id}/hold", holdHandler.CreateHold)

	router.ServeHTTP(rr, req)
	

	// Check the response status code
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
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
		"seat_id": "550e8400-e29b-41d4-a716-446655440000"
	}`
	

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/flights/550e8400-e29b-41d4-a716-446655440001/hold",
		strings.NewReader(validJSON),
	)
	req.Header.Set("Content-Type", "application/json") 
	req.Header.Set("Authorization", "Bearer valid-test-token")

	rr := httptest.NewRecorder()
	serviceCalled := false

	// Create a mock service that returns an error
	mockSvc := &mockHoldService{
        CreateHoldFunc: func(
            ctx context.Context,
            gotArg service.CreateHoldRequest,
        ) (sqlcdb.SeatHold, error) {
            serviceCalled = true

            assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000",
                gotArg.UserID.String())
            assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001",
                gotArg.FlightID.String())

            return sqlcdb.SeatHold{}, assert.AnError
        },
    }
	holdHandler := NewHoldHandler(mockSvc)

	mockTokenService := &mockTokenService{
        VerifyFunc: func(token string) (*service.Claims, error) {
            assert.Equal(t, "valid-test-token", token)

            return &service.Claims{
                UserID: "123e4567-e89b-12d3-a456-426614174000",
            }, nil
        },
    }
	authMiddleware := middleware.NewAuthMiddleware(mockTokenService)
	
	router := chi.NewRouter()

	router.With(authMiddleware.Authenticate).
    Post("/api/v1/flights/{flight_id}/hold", holdHandler.CreateHold)

	router.ServeHTTP(rr, req)
	
	
	// Check the response status code	
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.True(t, serviceCalled)
}
func TestCreateHold_Success(t *testing.T) {
	validJSON := `{
		"seat_id": "550e8400-e29b-41d4-a716-446655440000"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/flights/550e8400-e29b-41d4-a716-446655440001/hold",
		strings.NewReader(validJSON),
	)
	req.Header.Set("Content-Type", "application/json")	
	req.Header.Set("Authorization", "Bearer valid-test-token")


	rr := httptest.NewRecorder()
	serviceCalled := false
	now := time.Now()
	expectedHold := sqlcdb.SeatHold{
		ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
		SeatID:    pgtype.UUID{ Bytes: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), Valid: true},
		UserID:    pgtype.UUID{Bytes: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), Valid: true},
		HoldToken: pgtype.UUID{Bytes: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: now.Add(10 * time.Minute), Valid: true},
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

	holdHandler := NewHoldHandler(mockSvc)
	mockTokenService := &mockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{	
				UserID: "123e4567-e89b-12d3-a456-426614174000"}, nil
		},
	}
	authMiddleware := middleware.NewAuthMiddleware(mockTokenService)

	router := chi.NewRouter()
	router.With(authMiddleware.Authenticate).
		Post("/api/v1/flights/{flight_id}/hold", holdHandler.CreateHold)

	router.ServeHTTP(rr, req)
	fmt.Println("Response status:", rr.Code)
	fmt.Println("Response body:", rr.Body.String())
	fmt.Println("Service called:", serviceCalled)
	// Check the response status code
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.True(t, serviceCalled)
}
func TestCreateHold_SeatNotFound(t *testing.T) {
	serviceCalled := false
	mockService := &mockHoldService{
		CreateHoldFunc: func(ctx context.Context, req service.CreateHoldRequest) (sqlcdb.SeatHold, error) {
			serviceCalled = true
			return sqlcdb.SeatHold{}, domain.ErrSeatNotFound
		},
	}

	holdHandler := NewHoldHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/flights/550e8400-e29b-41d4-a716-446655440001/hold", strings.NewReader(`{
		"seat_id": "550e8400-e29b-41d4-a716-446655440000"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-test-token")
	rr := httptest.NewRecorder()

	mockTokenService := &mockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{
				UserID: "123e4567-e89b-12d3-a456-426614174000",
			}, nil
		},
	}
	authMiddleware := middleware.NewAuthMiddleware(mockTokenService)
	router := chi.NewRouter()
	router.With(authMiddleware.Authenticate).
		Post("/api/v1/flights/{flight_id}/hold", holdHandler.CreateHold)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.True(t, serviceCalled)
}
func TestCreateHold_SeatAlreadyBooked(t *testing.T) {
	serviceCalled := false
	mockService := &mockHoldService{
		CreateHoldFunc: func(ctx context.Context, req service.CreateHoldRequest) (sqlcdb.SeatHold, error) {
			serviceCalled = true
			return sqlcdb.SeatHold{}, domain.ErrSeatAlreadyBooked
		},
	}

	holdHandler := NewHoldHandler(mockService)

	req := httptest.NewRequest(http.MethodPost,"/api/v1/flights/550e8400-e29b-41d4-a716-446655440001/hold", strings.NewReader(`{
		"seat_id": "550e8400-e29b-41d4-a716-446655440000"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-test-token")
	rr := httptest.NewRecorder()

	mockTokenService := &mockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{
				UserID: "123e4567-e89b-12d3-a456-426614174000",
			}, nil
		},
	}
	authMiddleware := middleware.NewAuthMiddleware(mockTokenService)
	router := chi.NewRouter()
	router.With(authMiddleware.Authenticate).
		Post("/api/v1/flights/{flight_id}/hold", holdHandler.CreateHold)
	router.ServeHTTP(rr, req)
	

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.True(t, serviceCalled)
}
func TestCreateHold_SeatAlreadyOnHold(t *testing.T) {
	serviceCalled := false
	mockService := &mockHoldService{
		CreateHoldFunc: func(ctx context.Context, req service.CreateHoldRequest) (sqlcdb.SeatHold, error) {
			serviceCalled = true
			return sqlcdb.SeatHold{}, domain.ErrSeatAlreadyOnHold
		},
	}

	holdHandler := NewHoldHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/flights/550e8400-e29b-41d4-a716-446655440001/hold", strings.NewReader(`{
		"seat_id": "550e8400-e29b-41d4-a716-446655440000"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-test-token")
	rr := httptest.NewRecorder()

	mockTokenService := &mockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{
				UserID: "123e4567-e89b-12d3-a456-426614174000",
			}, nil
		},
	}
	authMiddleware := middleware.NewAuthMiddleware(mockTokenService)
	router := chi.NewRouter()
	router.With(authMiddleware.Authenticate).
		Post("/api/v1/flights/{flight_id}/hold", holdHandler.CreateHold)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.True(t, serviceCalled)
}
func TestCreateHold_InternalServerError(t *testing.T) {
	serviceCalled := false
	mockService := &mockHoldService{
		CreateHoldFunc: func(ctx context.Context, req service.CreateHoldRequest) (sqlcdb.SeatHold, error) {
			serviceCalled = true
			return sqlcdb.SeatHold{}, assert.AnError
		},
	}

	holdHandler := NewHoldHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/flights/550e8400-e29b-41d4-a716-446655440001/hold", strings.NewReader(`{
		"seat_id": "550e8400-e29b-41d4-a716-446655440000"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-test-token")


	rr := httptest.NewRecorder()
	mockTokenService := &mockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{
				UserID: "123e4567-e89b-12d3-a456-426614174000",
			}, nil
		},
	}
	authMiddleware := middleware.NewAuthMiddleware(mockTokenService)
	router := chi.NewRouter()
    router.With(authMiddleware.Authenticate).
        Post("/api/v1/flights/{flight_id}/hold", holdHandler.CreateHold)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.True(t, serviceCalled)
}
