-- name: GetResources :many
SELECT id, name, type, url, extracted_content, raw_content, status, owner_id, created_at, updated_at
FROM resources
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: GetResourcesByOwnerID :many
SELECT id, name, type, url, extracted_content, raw_content, status, owner_id, created_at, updated_at
FROM resources
WHERE owner_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: GetUsersResourceByID :one
SELECT id, name, type, url, extracted_content, raw_content, status, owner_id, created_at, updated_at
FROM resources
WHERE id = $1 AND owner_id = $2;

-- name: CreateResource :one
INSERT INTO resources (
    name, type, url, extracted_content, raw_content, owner_id
) VALUES (
    $1, $2, $3, $4, $5,  $6
) RETURNING id, name, type, url, extracted_content, raw_content, status, owner_id, created_at, updated_at;

-- name: UpdateUsersResource :one
UPDATE resources
SET
    name = COALESCE($3, name),
    type = COALESCE($4, type),
    url = COALESCE($5, url),
    extracted_content = COALESCE($6, extracted_content),
    raw_content = COALESCE($7, raw_content),
    status = COALESCE($8, status),
    owner_id = COALESCE($9, owner_id),
    updated_at = NOW()
WHERE id = $1 AND owner_id = $2
RETURNING id, name, type, url, extracted_content, raw_content, status, owner_id, created_at, updated_at;

-- name: DeleteUsersResource :exec
DELETE FROM resources
WHERE id = $1 AND owner_id = $2;

-- name: CheckResourceOwnership :one
SELECT COUNT(*) > 0 as owned
FROM resources
WHERE id = $1 AND (owner_id = $2 OR owner_id IS NULL OR owner_id = '');

-- name: GetResourcesWithFilter :many
SELECT id, name, type, url, extracted_content, raw_content, status, owner_id, created_at, updated_at
FROM resources
WHERE
    ($1::text IS NULL OR name ILIKE '%' || $1 || '%') AND
    ($2::resource_type IS NULL OR type = $2) AND
    ($3::resource_status IS NULL OR status = $3) AND
    ($4::text IS NULL OR owner_id = $4)
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: UpdateResourceStatus :one
UPDATE resources
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, name, type, url, extracted_content, raw_content, status, owner_id, created_at, updated_at;

-- name: GetResourcesByStatus :many
SELECT id, name, type, url, extracted_content, raw_content, status, owner_id, created_at, updated_at
FROM resources
WHERE status = $1
ORDER BY created_at DESC;

-- name: GetResourcesByType :many
SELECT id, name, type, url, extracted_content, raw_content, status, owner_id, created_at, updated_at
FROM resources
WHERE type = $1
ORDER BY created_at DESC;

-- name: CountResourcesByOwner :one
SELECT COUNT(*) as count
FROM resources
WHERE owner_id = $1;

-- name: CountResourcesByStatus :one
SELECT COUNT(*) as count
FROM resources
WHERE status = $1;

-- name: GetResourcesCount :one
SELECT COUNT(*) as count
FROM resources
WHERE
    ($1::text IS NULL OR name ILIKE '%' || $1 || '%') AND
    ($2::resource_type IS NULL OR type = $2) AND
    ($3::resource_status IS NULL OR status = $3) AND
    ($4::text IS NULL OR owner_id = $4);
