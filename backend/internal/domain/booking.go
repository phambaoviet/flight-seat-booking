package domain

import "github.com/jackc/pgx/v5/pgtype"

type ConfirmBookingParams struct {
    SeatID      pgtype.UUID
    UserID      pgtype.UUID
    HoldToken   pgtype.UUID
    BookingID   pgtype.UUID
    BookingCode string
}