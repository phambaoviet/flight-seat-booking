-- name: CreateBooking :one
INSERT INTO bookings (
    id,
    user_id,
    seat_id,
    booking_code,
    status
)
VALUES (
    $1,
    $2,
    $3, 
    $4,
    $5
)
RETURNING *;