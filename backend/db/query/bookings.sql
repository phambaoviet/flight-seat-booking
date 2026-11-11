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
AND status = 'ACTIVE'
AND expires_at > CURRENT_TIMESTAMP;

-- name: DeleteSeatHold :exec
DELETE FROM seat_holds
WHERE id = $1;

-- name: CancelBooking :one
UPDATE bookings
SET status = 'CANCELLED'
WHERE id = $1
AND booking_code = $2
AND user_id = $3
AND status = 'CONFIRMED'
RETURNING *;