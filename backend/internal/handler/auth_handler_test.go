package handler

import (
	"context"
	"encoding/json"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"
	"flight-booking/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockAuthService struct {
	RegisterUserFunc func(ctx context.Context, req service.RegisterUserParams) (sqlcdb.User, error)
	LoginUserFunc    func(ctx context.Context, email, password string) (string, error)
}

func (m *mockAuthService) LoginUser(ctx context.Context, email, password string) (string, error) {
    if m.LoginUserFunc != nil {
        return m.LoginUserFunc(ctx, email, password)
    }
    return "", nil
}

func (m *mockAuthService) RegisterUser(ctx context.Context, req service.RegisterUserParams) (sqlcdb.User, error) {
	if m.RegisterUserFunc != nil {
		return m.RegisterUserFunc(ctx, req)
	}
	return sqlcdb.User{}, nil
}

func TestRegister_Success(t *testing.T) {
	mockSvc := &mockAuthService{
		RegisterUserFunc: func(ctx context.Context, req service.RegisterUserParams) (sqlcdb.User, error) {
			return sqlcdb.User{
				ID: pgtype.UUID{
					Bytes: uuid.New(),
					Valid: true,
				},
				Name:        req.Name,
				Email:       req.Email,
				PhoneNumber: req.PhoneNumber,
			}, nil
		},
	}

	handler := NewAuthHandler(mockSvc)

	reqBody := `{"name":"John Doe","email":"john@example.com", "password":"password123", "phone_number":"1234567890"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp.StatusCode)
	}

	var resBody RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resBody.Name != "John Doe" || resBody.Email != "john@example.com" || resBody.PhoneNumber != "1234567890" {
		t.Fatalf("unexpected response body: %+v", resBody)
	}
}
func TestRegister_UserAlreadyExists(t *testing.T) {
	mockSvc := &mockAuthService{
		RegisterUserFunc: func(ctx context.Context, req service.RegisterUserParams) (sqlcdb.User, error) {
			return sqlcdb.User{}, domain.ErrUserAlreadyExists
		},
	}

	handler := NewAuthHandler(mockSvc)

	reqBody := `{"name":"John Doe","email":"john@example.com", "password":"password123", "phone_number":"1234567890"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected status 409 Conflict, got %d", resp.StatusCode)
	}
}
func TestRegister_InvalidMethod(t *testing.T) {
	mockSvc := &mockAuthService{}

	handler := NewAuthHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405 Method Not Allowed, got %d", resp.StatusCode)
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	mockSvc := &mockAuthService{}

	handler := NewAuthHandler(mockSvc)

	reqBody := `{"name":"John Doe","email":"john@example.com", "password":"password123", "phone_number":"1234567890"`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", resp.StatusCode)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	mockSvc := &mockAuthService{}

	handler := NewAuthHandler(mockSvc)

	reqBody := `{"name":"John Doe","email":"john@example.com", "password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", resp.StatusCode)
	}
}

func TestRegister_InvalidPassword(t *testing.T) {
	mockSvc := &mockAuthService{
		RegisterUserFunc: func(ctx context.Context, req service.RegisterUserParams) (sqlcdb.User, error) {
			return sqlcdb.User{}, domain.ErrInValidPassword
		},
	}

	handler := NewAuthHandler(mockSvc)

	reqBody := `{"name":"John Doe","email":"john@example.com", "password":"123", "phone_number":"1234567890"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", resp.StatusCode)
	}
}

func TestRegister_InternalServerError(t *testing.T) {
	mockSvc := &mockAuthService{
		RegisterUserFunc: func(ctx context.Context, req service.RegisterUserParams) (sqlcdb.User, error) {
			return sqlcdb.User{}, errors.New("internal server error")
		},
	}

	handler := NewAuthHandler(mockSvc)

	reqBody := `{"name":"John Doe","email":"john@example.com", "password":"password123", "phone_number":"1234567890"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}
