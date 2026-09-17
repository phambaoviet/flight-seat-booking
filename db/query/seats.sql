-- name: CreateSeat :one
INSERT INTO seats (
    id,
    flight_id,
    seat_number,
    seat_class,
    price
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;