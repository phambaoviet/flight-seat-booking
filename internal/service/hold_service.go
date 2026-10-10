package service

import (
	"context"
	sqlcdb "flight-booking/db/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type HoldStore interface {
	CreateHoldTx(ctx context.Context, arg sqlcdb.CreateSeatHoldParams,) (sqlcdb.SeatHold, error)
}
type HoldService struct {
	store HoldStore
}

type CreateHoldRequest struct {
	UserID pgtype.UUID 
	SeatID pgtype.UUID 
}



func NewHoldService(store HoldStore) *HoldService {
	return &HoldService{
		store: store,
	}
}
func (s *HoldService) CreateHold(ctx context.Context, req CreateHoldRequest) (sqlcdb.SeatHold, error) {
	arg := sqlcdb.CreateSeatHoldParams{
		ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
		UserID:    req.UserID,
		SeatID:    req.SeatID,
		HoldToken: pgtype.UUID{Bytes: uuid.New(), Valid: true},
	}
	hold, err := s.store.CreateHoldTx(ctx, arg)
	if err != nil {
		return sqlcdb.SeatHold{}, err
	}
	return hold, nil
}
