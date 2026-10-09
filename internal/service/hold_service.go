package service

import (
	"context"
	sqlcdb "flight-booking/db/sqlc"
)

type HoldStore interface {
	CreateHoldTx(ctx context.Context, arg sqlcdb.CreateSeatHoldParams,) (sqlcdb.SeatHold, error)
}
type HoldService struct {
	store HoldStore
}

func NewHoldService(store HoldStore) *HoldService {
	return &HoldService{
		store: store,
	}
}
func (s *HoldService) CreateHold(ctx context.Context, arg sqlcdb.CreateSeatHoldParams) (sqlcdb.SeatHold, error) {
	hold, err := s.store.CreateHoldTx(ctx, arg)
	if err != nil {
		return sqlcdb.SeatHold{}, err
	}
	return hold, nil
}
