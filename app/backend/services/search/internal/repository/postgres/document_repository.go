package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/embedded-market/backend/services/search/internal/domain"
	"github.com/google/uuid"
)

type documentRepository struct{ db pgxConn }

func (r *documentRepository) Upsert(ctx context.Context, item *domain.Document) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO search_documents (id, product_id, slug, sku, name, short_description, category_slug, brand_slug, tags_csv, specs_text, price_cents, currency, available, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) ON CONFLICT (id) DO UPDATE SET product_id = EXCLUDED.product_id, slug = EXCLUDED.slug, sku = EXCLUDED.sku, name = EXCLUDED.name, short_description = EXCLUDED.short_description, category_slug = EXCLUDED.category_slug, brand_slug = EXCLUDED.brand_slug, tags_csv = EXCLUDED.tags_csv, specs_text = EXCLUDED.specs_text, price_cents = EXCLUDED.price_cents, currency = EXCLUDED.currency, available = EXCLUDED.available, updated_at = EXCLUDED.updated_at`, item.ID, item.ProductID, item.Slug, item.SKU, item.Name, item.ShortDescription, item.CategorySlug, item.BrandSlug, tagsToCSV(item.Tags), item.SpecsText, item.PriceCents, item.Currency, item.Available, item.CreatedAt, item.UpdatedAt); err != nil { return fmt.Errorf("upsert document: %w", err) }
	return nil
}

func (r *documentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	return r.getOne(ctx, `SELECT id, product_id, slug, sku, name, short_description, category_slug, brand_slug, tags_csv, specs_text, price_cents, currency, available, created_at, updated_at FROM search_documents WHERE id = $1`, id)
}

func (r *documentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM search_documents WHERE id = $1`, id)
	if err != nil { return fmt.Errorf("delete document: %w", err) }
	if tag.RowsAffected() == 0 { return domain.ErrDocumentNotFound }
	return nil
}

func (r *documentRepository) Search(ctx context.Context, filters domain.SearchFilters, limit, offset int32) ([]domain.Document, int64, error) {
	query := `SELECT id, product_id, slug, sku, name, short_description, category_slug, brand_slug, tags_csv, specs_text, price_cents, currency, available, created_at, updated_at, COUNT(*) OVER() AS total_count FROM search_documents WHERE 1=1`
	args := make([]any, 0, 8)
	idx := 1
	if q := strings.TrimSpace(filters.Query); q != "" {
		query += fmt.Sprintf(` AND search_vector @@ websearch_to_tsquery('simple', $%d)`, idx)
		args = append(args, q)
		idx++
	}
	if filters.CategorySlug != "" {
		query += fmt.Sprintf(` AND category_slug = $%d`, idx)
		args = append(args, filters.CategorySlug)
		idx++
	}
	if filters.BrandSlug != "" {
		query += fmt.Sprintf(` AND brand_slug = $%d`, idx)
		args = append(args, filters.BrandSlug)
		idx++
	}
	if filters.Available != nil {
		query += fmt.Sprintf(` AND available = $%d`, idx)
		args = append(args, *filters.Available)
		idx++
	}
	if strings.TrimSpace(filters.Query) != "" {
		query += fmt.Sprintf(` ORDER BY ts_rank(search_vector, websearch_to_tsquery('simple', $1)) DESC, created_at DESC`)
	} else {
		query += ` ORDER BY created_at DESC`
	}
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil { return nil, 0, fmt.Errorf("search documents: %w", err) }
	defer rows.Close()
	items := make([]domain.Document, 0)
	var total int64
	for rows.Next() {
		var item domain.Document
		var tagsCSV string
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Slug, &item.SKU, &item.Name, &item.ShortDescription, &item.CategorySlug, &item.BrandSlug, &tagsCSV, &item.SpecsText, &item.PriceCents, &item.Currency, &item.Available, &item.CreatedAt, &item.UpdatedAt, &total); err != nil { return nil, 0, fmt.Errorf("scan document: %w", err) }
		item.Tags = csvToTags(tagsCSV)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("iterate documents: %w", err) }
	return items, total, nil
}

func (r *documentRepository) Suggest(ctx context.Context, query string, limit int32) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT name FROM search_documents WHERE name ILIKE $1 OR slug ILIKE $1 ORDER BY name ASC LIMIT $2`, strings.TrimSpace(query)+"%", limit)
	if err != nil { return nil, fmt.Errorf("suggest documents: %w", err) }
	defer rows.Close()
	items := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil { return nil, fmt.Errorf("scan suggestion: %w", err) }
		items = append(items, name)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate suggestions: %w", err) }
	return items, nil
}

func (r *documentRepository) getOne(ctx context.Context, query string, arg any) (*domain.Document, error) {
	var item domain.Document
	var tagsCSV string
	if err := r.db.QueryRow(ctx, query, arg).Scan(&item.ID, &item.ProductID, &item.Slug, &item.SKU, &item.Name, &item.ShortDescription, &item.CategorySlug, &item.BrandSlug, &tagsCSV, &item.SpecsText, &item.PriceCents, &item.Currency, &item.Available, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrDocumentNotFound }
		return nil, fmt.Errorf("get document: %w", err)
	}
	item.Tags = csvToTags(tagsCSV)
	return &item, nil
}
