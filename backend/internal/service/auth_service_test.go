package service

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"
	"testing"

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

	svc := NewAuthService(mockStore)
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

	svc := NewAuthService(mockStore)
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

	svc := NewAuthService(mockStore)
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

	svc := NewAuthService(mockStore)
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