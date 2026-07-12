-- name: ListWishlistByAccount :many
SELECT account_id, product_id, added_at
FROM wishlist_items
WHERE account_id = $1
ORDER BY added_at DESC;
