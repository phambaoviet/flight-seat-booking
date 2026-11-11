-- name: CreateUser :one
INSERT INTO users (
    id,
    name,
    email,
    phone_number,
    password_hash
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;
-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;   
