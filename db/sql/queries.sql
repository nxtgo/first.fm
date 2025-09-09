-- name: GetUser :one
SELECT lastfm_username
FROM users
WHERE discord_id = ?;

-- name: GetUserByUsername :one
SELECT discord_id, lastfm_username
FROM users
WHERE lastfm_username = ?;

-- name: ListUsers :many
SELECT discord_id, lastfm_username
FROM users;

-- name: GetUserCount :one
SELECT COUNT(*) AS count
FROM users;

-- name: UpsertUser :exec
INSERT INTO users (discord_id, lastfm_username)
VALUES (?, ?)
ON CONFLICT(discord_id) DO UPDATE
SET lastfm_username = excluded.lastfm_username;

-- name: DeleteUser :exec
DELETE FROM users
WHERE discord_id = ?;
