-- name: ReplaceCompatibilityRulesClear :exec
DELETE FROM compatibility_rules WHERE product_id = $1;
