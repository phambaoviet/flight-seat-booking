-- name: LockSeat :one
SELECT * 
FROM seats
WHERE id = $1
FOR UPDATE;

-- name: GetConfirmedBooking :one
SELECT *
FROM bookings
WHERE seat_id = $1
AND status = 'CONFIRMED';

-- name: GetActiveSeatHold :one
SELECT *
FROM seat_holds
WHERE seat_id = $1
AND expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: CreateSeatHold :one
INSERT INTO seat_holds (id, user_id, seat_id, hold_token, expires_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP + INTERVAL '10 minutes')
RETURNING *;