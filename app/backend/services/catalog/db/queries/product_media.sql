-- name: ReplaceProductMediaClear :exec
DELETE FROM product_media WHERE product_id = $1;
