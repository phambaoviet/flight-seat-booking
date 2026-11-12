package middleware

import (
	"flight-booking/internal/domain"
	"flight-booking/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockTokenService struct {
	GenerateFunc func(userID string) (string, error)
	VerifyFunc   func(token string) (*service.Claims, error)
}

func (m *MockTokenService) Generate(userID string) (string, error) {
	return m.GenerateFunc(userID)
}

func (m *MockTokenService) Verify(token string) (*service.Claims, error) {
	return m.VerifyFunc(token)
}
func TestMissingAuthorization(t *testing.T) {
	// Create a mock TokenService that tracks if Verify is called
	verifyCalled := false
	mockTokenService := &MockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			verifyCalled = true
			return nil, nil
		},
	}

	// Create the middleware
	authMiddleware := NewAuthMiddleware(mockTokenService)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Create a handler that should not be called
	handler := authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when Authorization header is missing")
	}))

	handler.ServeHTTP(w, req)

	// Ensure that the Verify function was not called
	if verifyCalled {
		t.Error("Verify function should not be called when Authorization header is missing")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
func TestInvalidBearerToken(t *testing.T) {
	// Create a mock TokenService that tracks if Verify is called
	verifyCalled := false
	mockTokenService := &MockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			verifyCalled = true
			return nil, nil
		},
	}

	// Create the middleware
	authMiddleware := NewAuthMiddleware(mockTokenService)

	// Create a request with an invalid Authorization header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidTokenFormat")
	w := httptest.NewRecorder()

	handler := authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when Authorization header is invalid")
	}))

	handler.ServeHTTP(w, req)
	// Ensure that the Verify function was not called
	if verifyCalled {
		t.Error("Verify function should not be called with invalid Authorization format")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestEmptyToken(t *testing.T) {
	// Create a mock TokenService that tracks if Verify is called
	verifyCalled := false
	mockTokenService := &MockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			verifyCalled = true
			return nil, nil
		},
	}

	authMiddleware := NewAuthMiddleware(mockTokenService)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	handler := authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when token is empty")
	}))

	handler.ServeHTTP(w, req)
	// Ensure that the Verify function was not called
	if verifyCalled {
		t.Error("Verify function should not be called when token is empty")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestInvalidJWTToken(t *testing.T) {
	mockTokenService := &MockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			return nil, domain.ErrInvalidToken
		},
	}

	authMiddleware := NewAuthMiddleware(mockTokenService)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	w := httptest.NewRecorder()

	handler := authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when token is invalid")
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestValidJWTToken(t *testing.T) {
	mockTokenService := &MockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{UserID: "12345"}, nil
		},
	}

	authMiddleware := NewAuthMiddleware(mockTokenService)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	w := httptest.NewRecorder()

	// Create a handler that checks if the user ID is set in the context
	handlerCalled := false
	handler := authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		userID, ok := GetUserID(r.Context())
		if !ok || userID != "12345" {
			t.Error("User ID not set correctly in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Fatal("next handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestInvalidClaims(t *testing.T) {
	mockTokenService := &MockTokenService{
		VerifyFunc: func(token string) (*service.Claims, error) {
			return &service.Claims{UserID: ""}, nil
		},
	}

	authMiddleware := NewAuthMiddleware(mockTokenService)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	w := httptest.NewRecorder()

	handler := authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when claims are invalid")
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
