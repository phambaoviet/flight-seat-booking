-- name: LockSeatForHold :one
SELECT *
FROM seats
WHERE id = $1
  AND flight_id = $2
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

-- name: CountActiveHold :one
SELECT COUNT(*)
FROM seat_holds
WHERE seat_id = $1
AND expires_at > CURRENT_TIMESTAMP;

-- name: ExpireOldSeatHolds :exec
UPDATE seat_holds
SET status = 'EXPIRED'
WHERE seat_id = $1
AND expires_at <= CURRENT_TIMESTAMP
AND status = 'ACTIVE';  

-- name: ReleaseSeatHold :one
UPDATE seat_holds
SET 
    status = 'RELEASED',
    released_at = CURRENT_TIMESTAMP
WHERE id = $1
AND status = 'ACTIVE'
AND expires_at > CURRENT_TIMESTAMP
RETURNING *;

-- name: GetSeatHold :one
SELECT *
FROM seat_holds
WHERE id = $1;

-- name: ConvertSeatHold :one
UPDATE seat_holds
SET status = 'CONVERTED'
WHERE seat_id = $1
AND user_id = $2
AND hold_token = $3
AND status = 'ACTIVE'
AND expires_at > CURRENT_TIMESTAMP
RETURNING *;

-- name: CreateBookingFromHold :one
INSERT INTO bookings (id, user_id, seat_id, booking_code, status)
SELECT $1, seat_holds.user_id, seat_holds.seat_id, $4, 'CONFIRMED'
FROM seat_holds
WHERE seat_holds.id  = $2
AND seat_holds.user_id = $3
AND seat_holds.status = 'CONVERTED'
RETURNING *;