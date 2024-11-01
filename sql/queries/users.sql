-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING id, created_at, updated_at, email, is_chirpy_red; 

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT id, created_at, updated_at, email, hashed_password, is_chirpy_red
FROM users
WHERE email = $1;

-- name: EditUser :one
UPDATE users
SET email = $1, hashed_password = $2, updated_at = $3
WHERE id = $4
RETURNING email;

-- name: UpgradeUserToRed :exec
UPDATE users
SET is_chirpy_red = TRUE
WHERE id = $1
RETURNING is_chirpy_red;
