-- name: InsertUser :exec
INSERT INTO lastfm_users (discord_id, username)
VALUES (?1, ?2, ?3)
ON CONFLICT(discord_id) DO UPDATE SET username = excluded.username;

-- name: GetUsersByGuild :many
SELECT discord_id, username FROM lastfm_users
WHERE guild_id = ?1;

-- name: DeleteUser :exec
DELETE FROM lastfm_users
WHERE guild_id = ?1 AND discord_id = ?2;

-- name: GetUserByDiscordID :one
SELECT username FROM lastfm_users
WHERE discord_id = ?1
LIMIT 1;

-- name: UpdateUsername :exec
UPDATE lastfm_users
SET username = ?3
WHERE discord_id = ?1;

-- name: GetDiscordByUsername :one
SELECT discord_id, username
FROM lastfm_users
WHERE username = ?1
LIMIT 1;

