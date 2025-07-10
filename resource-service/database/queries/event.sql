-- name: CreateEvent :one
INSERT INTO events (name, topic, payload)
VALUES ($1, $2, $3)
RETURNING id, name, topic, payload, sent, event_time;

-- name: GetNotSentEvents :many
SELECT id, name, topic, payload, sent, event_time
FROM events
WHERE sent=false
ORDER BY event_time ASC
LIMIT $1 OFFSET $2;