package service

import (
	"context"
	sqlcdb "flight-booking/db/sqlc"
)
type UserStore interface {
	GetUserByEmail(ctx context.Context, email string) (sqlcdb.User, error)
	CreateUser(ctx context.Context, user sqlcdb.CreateUserParams) (sqlcdb.User, error)
}