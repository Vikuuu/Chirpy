-- name: CreateChirpForUser :one
INSERT INTO chirp (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING id, created_at, updated_at, body, user_id;

-- name: GetChirps :many
SELECT id, created_at, updated_at, body, user_id
FROM chirp
ORDER BY created_at ASC;

-- name: GetChirpsForAuthor :many
SELECT id, created_at, updated_at, body, user_id
FROM chirp
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT id, created_at, updated_at, body, user_id
FROM chirp
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirp
WHERE user_id = $1 AND id = $2;
