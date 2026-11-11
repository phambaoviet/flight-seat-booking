package service

import (
	"context"
	"errors"
	sqlcdb "flight-booking/db/sqlc"
	"flight-booking/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
    store UserStore
}

type RegisterUserParams struct {
	Name        string 
	Email       string
	Password    string
	PhoneNumber string
}

func NewAuthService(store UserStore) *AuthService {
    return &AuthService{store: store}
}


func (s *AuthService) RegisterUser(ctx context.Context, req RegisterUserParams) (sqlcdb.User, error) {
	// Check if the user already exists
	_, err := s.store.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return sqlcdb.User{}, domain.ErrUserAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return sqlcdb.User{}, err
	}
	// hashpassword
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return sqlcdb.User{}, err
	}

	user, err := s.store.CreateUser(ctx, sqlcdb.CreateUserParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		PhoneNumber:  req.PhoneNumber,
	})
	if err != nil {
		return sqlcdb.User{}, err
	}

	return user, nil

}