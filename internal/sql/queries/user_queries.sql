-- name: CreateUser :one
INSERT INTO users (name, email, password_hash, activated)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, version;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, activated, version, created_at, updated_at
FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET 
    name = $1, 
    email = $2, 
    password_hash = $3, 
    activated = $4
WHERE id = $5 AND version = $6
RETURNING version, updated_at;