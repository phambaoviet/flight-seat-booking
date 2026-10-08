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

-- name: GetActiveHoldByToken :one
SELECT *
FROM seat_holds
WHERE seat_id = $1
    AND user_id = $2
    AND hold_token = $3
    AND expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: DeleteSeatHold :exec
DELETE FROM seat_holds
WHERE id = $1;