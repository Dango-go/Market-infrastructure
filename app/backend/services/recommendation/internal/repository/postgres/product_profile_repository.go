package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/recommendation/internal/domain"
	"github.com/google/uuid"
)

type productProfileRepository struct{ db pgxConn }

func (r *productProfileRepository) Upsert(ctx context.Context, item *domain.ProductProfile) error {
	_, err := r.db.Exec(ctx, `INSERT INTO product_profiles (product_id, slug, category_slug, brand_slug, tags_csv, price_cents, available, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT (product_id) DO UPDATE SET slug = EXCLUDED.slug, category_slug = EXCLUDED.category_slug, brand_slug = EXCLUDED.brand_slug, tags_csv = EXCLUDED.tags_csv, price_cents = EXCLUDED.price_cents, available = EXCLUDED.available, updated_at = EXCLUDED.updated_at`, item.ProductID, item.Slug, item.CategorySlug, item.BrandSlug, tagsToCSV(item.Tags), item.PriceCents, item.Available, item.CreatedAt, item.UpdatedAt)
	if err != nil { return fmt.Errorf("upsert profile: %w", err) }
	return nil
}

func (r *productProfileRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*domain.ProductProfile, error) {
	var item domain.ProductProfile
	var tagsCSV string
	if err := r.db.QueryRow(ctx, `SELECT product_id, slug, category_slug, brand_slug, tags_csv, price_cents, available, created_at, updated_at FROM product_profiles WHERE product_id = $1`, productID).Scan(&item.ProductID, &item.Slug, &item.CategorySlug, &item.BrandSlug, &tagsCSV, &item.PriceCents, &item.Available, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrProfileNotFound }
		return nil, fmt.Errorf("get profile: %w", err)
	}
	item.Tags = csvToTags(tagsCSV)
	return &item, nil
}

func (r *productProfileRepository) Related(ctx context.Context, productID uuid.UUID, limit int32) ([]domain.Recommendation, error) {
	rows, err := r.db.Query(ctx, `SELECT p2.product_id,
		(
			CASE WHEN p1.category_slug <> '' AND p1.category_slug = p2.category_slug THEN 4 ELSE 0 END +
			CASE WHEN p1.brand_slug <> '' AND p1.brand_slug = p2.brand_slug THEN 3 ELSE 0 END +
			CASE WHEN ABS(p1.price_cents - p2.price_cents) <= 5000 THEN 1 ELSE 0 END +
			COALESCE((SELECT SUM(CASE e.type WHEN 'view' THEN 0.2 WHEN 'favorite' THEN 1.5 WHEN 'cart' THEN 2 WHEN 'purchase' THEN 3 ELSE 0 END) FROM recommendation_events e WHERE e.product_id = p2.product_id AND e.created_at >= NOW() - INTERVAL '30 days'), 0)
		)::float8 AS score,
		CASE
			WHEN p1.category_slug <> '' AND p1.category_slug = p2.category_slug AND p1.brand_slug <> '' AND p1.brand_slug = p2.brand_slug THEN 'same_category_and_brand'
			WHEN p1.category_slug <> '' AND p1.category_slug = p2.category_slug THEN 'same_category'
			WHEN p1.brand_slug <> '' AND p1.brand_slug = p2.brand_slug THEN 'same_brand'
			ELSE 'similar_price'
		END AS reason
	FROM product_profiles p1
	JOIN product_profiles p2 ON p1.product_id <> p2.product_id
	WHERE p1.product_id = $1 AND p2.available = TRUE
	ORDER BY score DESC, p2.updated_at DESC
	LIMIT $2`, productID, limit)
	if err != nil { return nil, fmt.Errorf("related recommendations: %w", err) }
	defer rows.Close()
	items := make([]domain.Recommendation, 0)
	for rows.Next() {
		var item domain.Recommendation
		if err := rows.Scan(&item.ProductID, &item.Score, &item.Reason); err != nil { return nil, fmt.Errorf("scan related recommendation: %w", err) }
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate related recommendations: %w", err) }
	return items, nil
}

func (r *productProfileRepository) Trending(ctx context.Context, limit int32) ([]domain.Recommendation, error) {
	rows, err := r.db.Query(ctx, `SELECT p.product_id, COALESCE(SUM(CASE e.type WHEN 'view' THEN 0.2 WHEN 'favorite' THEN 1.5 WHEN 'cart' THEN 2 WHEN 'purchase' THEN 3 ELSE 0 END), 0)::float8 AS score, 'trending' AS reason FROM product_profiles p LEFT JOIN recommendation_events e ON e.product_id = p.product_id AND e.created_at >= NOW() - INTERVAL '14 days' WHERE p.available = TRUE GROUP BY p.product_id ORDER BY score DESC, p.updated_at DESC LIMIT $1`, limit)
	if err != nil { return nil, fmt.Errorf("trending recommendations: %w", err) }
	defer rows.Close()
	items := make([]domain.Recommendation, 0)
	for rows.Next() {
		var item domain.Recommendation
		if err := rows.Scan(&item.ProductID, &item.Score, &item.Reason); err != nil { return nil, fmt.Errorf("scan trending recommendation: %w", err) }
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate trending recommendations: %w", err) }
	return items, nil
}
