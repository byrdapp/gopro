

-- name: UpdateBookingStatus :exec
UPDATE bookings SET accepted = $2, completed = $3, task = $4 WHERE id = $1;

-- name: DeleteBooking :exec
DELETE FROM bookings WHERE id = $1;

-- name: GetBookingsByMediaUID :many
SELECT * FROM bookings WHERE media_id = $1 ORDER BY created_at DESC;

-- name: GetUser :one
SELECT * FROM profiles WHERE id = $1 LIMIT 1;

-- name: CreateProfile :one
INSERT INTO profiles (user_id, pro_level)
    VALUES ($1, $2) RETURNING id;

-- name: CreateBooking :one
INSERT INTO bookings (media_id, task, price, credits, date_start, date_end, lat, lng)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;

-- name: AcceptBooking :exec
UPDATE bookings SET accepted = $2 WHERE photographer_id = $1;

-- name: ListBookingsByUser :many
SELECT
    bookings.task,
    bookings.credits,
    bookings.price,
    bookings.created_at,
    bookings.accepted,
    bookings.completed,
    profiles.pro_level,
    profiles.user_id
FROM
    bookings
    LEFT JOIN profiles ON bookings.media_id = profiles.user_id
ORDER BY
    bookings.created_at DESC,
    bookings.accepted DESC
LIMIT 5;
