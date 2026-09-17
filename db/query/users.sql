-- name: CreateUser :one
INSERT INTO users(
    id,
    name,
    email,
    phone_number
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;