package service

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)
var _ UserStore = (*MockUserStore)(nil)
type MockUserStore struct {
    users map[string]sqlcdb.User

    GetUserByEmailFunc func(ctx context.Context, email string) (sqlcdb.User, error)
    CreateUserFunc     func(ctx context.Context, user sqlcdb.CreateUserParams) (sqlcdb.User, error)
}

type MockTokenService struct {
	GenerateFunc func(userID string) (string, error)
	VerifyFunc   	  func(token string) (*Claims, error)
}

func (m *MockTokenService) Generate(userID string) (string, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(userID)
	}
	return "", nil
}

func (m *MockTokenService) Verify(token string) (*Claims, error) {
	if m.VerifyFunc != nil {
		return m.VerifyFunc(token)
	}
	return nil, nil
}


func (m *MockUserStore) CreateUser(ctx context.Context, user sqlcdb.CreateUserParams,) (sqlcdb.User, error) {
    if m.CreateUserFunc != nil {
        return m.CreateUserFunc(ctx, user)
    }
    return sqlcdb.User{}, errors.New("CreateUserFunc not configured")
}
func (m *MockUserStore) GetUserByEmail(ctx context.Context, email string) (sqlcdb.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	
	return sqlcdb.User{}, errors.New("GetUserByEmailFunc not configured")
}
func TestRegisterUser_Success(t *testing.T) {
	var captured sqlcdb.CreateUserParams
	mockStore := &MockUserStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (sqlcdb.User, error) {
			return sqlcdb.User{}, pgx.ErrNoRows
		},
		CreateUserFunc: func(ctx context.Context, user sqlcdb.CreateUserParams) (sqlcdb.User, error) {
			captured = user
			return sqlcdb.User{
				ID:           user.ID,
				Name:         user.Name,
				Email:        user.Email,
				PhoneNumber:  user.PhoneNumber,
				PasswordHash: user.PasswordHash,
			}, nil
		},
	}

	mockTokenService := &MockTokenService{}

	svc := NewAuthService(mockStore, mockTokenService)	
	req := RegisterUserParams{
		Name:        "John Doe",
		Email:       "John@Example.com",
		Password:    "password123",
		PhoneNumber: "1234567890",
	}

	user, err := svc.RegisterUser(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Name != req.Name || user.Email != req.Email || user.PhoneNumber != req.PhoneNumber {
		t.Errorf("user fields do not match request")
	}

	if captured.Name != req.Name || captured.Email != req.Email || captured.PhoneNumber != req.PhoneNumber {
		t.Errorf("captured fields do not match request")
	}	
	
	// Check that the password hash is not the same as the plain password
	if captured.PasswordHash == req.Password {
		t.Errorf("password hash should not match the plain password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(captured.PasswordHash), []byte(req.Password)); err != nil {
		t.Errorf("password hash does not match the original password")
	}
}

func TestRegisterUser_EmailAlreadyExists(t *testing.T) {
	createUserCalled := false
	mockStore := &MockUserStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (sqlcdb.User, error) {
			return sqlcdb.User{
				ID:    pgtype.UUID{Valid: true},
				Name:  "Existing User",
				Email: email,
			}, nil
		},
		CreateUserFunc: func(ctx context.Context, user sqlcdb.CreateUserParams) (sqlcdb.User, error) {
			createUserCalled = true
			return sqlcdb.User{}, nil
		},
	}

	mockTokenService := &MockTokenService{}

	svc := NewAuthService(mockStore, mockTokenService)	
	req := RegisterUserParams{
		Name:        "John Doe",
		Email:       "John@Example.com",
		Password:    "password123",
		PhoneNumber: "1234567890",
	}

	_, err := svc.RegisterUser(context.Background(), req)
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}

	if createUserCalled {
		t.Fatalf("expected GetUserByEmail to be called before CreateUser")
	}

}

func TestRegisterUser_DatabaseError(t *testing.T) {
	mockStore := &MockUserStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (sqlcdb.User, error) {
			return sqlcdb.User{}, errors.New("database error")
		},
	}

	mockTokenService := &MockTokenService{}

	svc := NewAuthService(mockStore, mockTokenService)	
	req := RegisterUserParams{
		Name:        "John Doe",
		Email:       "John@Example.com",
		Password:    "password123",
		PhoneNumber: "1234567890",
	}

	_, err := svc.RegisterUser(context.Background(), req)
	if err == nil || err.Error() != "database error" {
		t.Fatalf("expected database error, got %v", err)
	}
}

func TestRegisterUser_CreateUserError(t *testing.T) {
	mockStore := &MockUserStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (sqlcdb.User, error) {
			return sqlcdb.User{}, pgx.ErrNoRows
		},
		CreateUserFunc: func(ctx context.Context, user sqlcdb.CreateUserParams) (sqlcdb.User, error) {
			return sqlcdb.User{}, errors.New("create user error")
		},
	}

	mockTokenService := &MockTokenService{}

	svc := NewAuthService(mockStore, mockTokenService)	
	req := RegisterUserParams{
		Name:        "John Doe",
		Email:       "John@Example.com",
		Password:    "password123",
		PhoneNumber: "1234567890",
	}

	_, err := svc.RegisterUser(context.Background(), req)
	if err == nil || err.Error() != "create user error" {
		t.Fatalf("expected create user error, got %v", err)
	}
}

func TestLoginUser_Success(t *testing.T) {
	const password = "password123"
	
	store := &MockUserStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (sqlcdb.User, error) {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				t.Fatalf("failed to hash password: %v", err)
			}
			return sqlcdb.User{
				ID:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
				Name:         "John Doe",
				Email:        email,
				PhoneNumber:  "1234567890",
				PasswordHash: string(hash),
			}, nil
		},
	}
	
	var capturedUserID string
	tokenService := &MockTokenService{
		GenerateFunc: func(userID string) (string, error) {
			capturedUserID = userID
			return "mocked_token", nil
		},
	}

	svc := NewAuthService(store, tokenService)

	token, err := svc.LoginUser(context.Background(), "john@example.com", password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token != "mocked_token" {
		t.Errorf("expected token 'mocked_token', got %s", token)
	}

	if capturedUserID == "" {
		t.Errorf("expected userID to be captured, but it was empty")
	}
}
func TestLoginUser_InvalidPassword(t *testing.T) {
	const password = "password123"
	
	store := &MockUserStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (sqlcdb.User, error) {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				t.Fatalf("failed to hash password: %v", err)
			}
			return sqlcdb.User{
				ID:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
				Name:         "John Doe",
				Email:        email,
				PhoneNumber:  "1234567890",
				PasswordHash: string(hash),
			}, nil
		},
	}
	
	tokenService := &MockTokenService{}

	svc := NewAuthService(store, tokenService)

	_, err := svc.LoginUser(context.Background(), "john@example.com", "wrongpassword")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUser_UserNotFound(t *testing.T) {
	store := &MockUserStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (sqlcdb.User, error) {
			return sqlcdb.User{}, pgx.ErrNoRows
		},
	}
	
	tokenService := &MockTokenService{}

	svc := NewAuthService(store, tokenService)

	_, err := svc.LoginUser(context.Background(), "john@example", "password123")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
func TestLoginUser_DatabaseError(t *testing.T) {
	store := &MockUserStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (sqlcdb.User, error) {
			return sqlcdb.User{}, errors.New("database error")
		},
	}
	
	tokenService := &MockTokenService{}

	svc := NewAuthService(store, tokenService)

	_, err := svc.LoginUser(context.Background(), "john@example.com", "password123")
	if err == nil || err.Error() != "database error" {
		t.Fatalf("expected database error, got %v", err)
	}
}

func TestLoginUser_TokenGenerationError(t *testing.T) {
	const password = "password123"
	
	store := &MockUserStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (sqlcdb.User, error) {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				t.Fatalf("failed to hash password: %v", err)
			}
			return sqlcdb.User{
				ID:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
				Name:         "John Doe",
				Email:        email,
				PhoneNumber:  "1234567890",
				PasswordHash: string(hash),
			}, nil
		},
	}
	
	tokenService := &MockTokenService{
		GenerateFunc: func(userID string) (string, error) {
			return "", errors.New("token generation error")
		},
	}

	svc := NewAuthService(store, tokenService)

	_, err := svc.LoginUser(context.Background(), "john@example.com", password)
	if err == nil || err.Error() != "token generation error" {
		t.Fatalf("expected token generation error, got %v", err)
	}
}