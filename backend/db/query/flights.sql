-- name: CreateFlight :one
INSERT INTO flights (
    id,
    flight_number, 
    departure_airport, 
    arrival_airport, 
    boarding_at,
    departure_at,
    arrival_at
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;