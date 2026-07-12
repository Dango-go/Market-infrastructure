-- name: ReplaceProductSpecsClear :exec
DELETE FROM product_specs WHERE product_id = $1;
