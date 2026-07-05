package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/embedded-market/backend/services/catalog/internal/domain"
	"github.com/google/uuid"
)

type productRepository struct{ db pgxConn }

func (r *productRepository) Create(ctx context.Context, p *domain.Product) error {
	const q = `INSERT INTO products (id, category_id, brand_id, slug, sku, name, short_description, description, datasheet_url, image_url, status, featured, created_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`
	if _, err := r.db.Exec(ctx, q, p.ID, p.CategoryID, p.BrandID, p.Slug, p.SKU, p.Name, p.ShortDescription, p.Description, p.DatasheetURL, p.ImageURL, string(p.Status), p.Featured, p.CreatedBy, p.CreatedAt, p.UpdatedAt); err != nil {
		if isUniqueViolation(err) { return domain.ErrSlugTaken }
		return fmt.Errorf("insert product: %w", err)
	}
	if err := r.ReplaceSpecs(ctx, p.ID, p.Specs); err != nil { return err }
	if err := r.ReplaceMedia(ctx, p.ID, p.Media); err != nil { return err }
	if err := r.ReplaceCompatibility(ctx, p.ID, p.Compatibility); err != nil { return err }
	return nil
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	p, err := r.getOne(ctx, `SELECT id, category_id, brand_id, slug, sku, name, short_description, description, datasheet_url, image_url, status, featured, created_by, created_at, updated_at FROM products WHERE id = $1`, id)
	if err != nil { return nil, err }
	return r.hydrate(ctx, p)
}

func (r *productRepository) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	p, err := r.getOne(ctx, `SELECT id, category_id, brand_id, slug, sku, name, short_description, description, datasheet_url, image_url, status, featured, created_by, created_at, updated_at FROM products WHERE slug = $1`, strings.TrimSpace(strings.ToLower(slug)))
	if err != nil { return nil, err }
	return r.hydrate(ctx, p)
}

func (r *productRepository) List(ctx context.Context, filters domain.ProductFilters, limit, offset int32) ([]*domain.Product, int64, error) {
	where := []string{"1=1"}
	args := make([]any, 0)
	idx := 1
	if q := strings.TrimSpace(filters.Query); q != "" {
		where = append(where, fmt.Sprintf("(p.name ILIKE $%d OR p.sku ILIKE $%d)", idx, idx))
		args = append(args, "%"+q+"%")
		idx++
	}
	if filters.CategorySlug != "" {
		where = append(where, fmt.Sprintf("c.slug = $%d", idx))
		args = append(args, strings.TrimSpace(strings.ToLower(filters.CategorySlug)))
		idx++
	}
	if filters.BrandSlug != "" {
		where = append(where, fmt.Sprintf("b.slug = $%d", idx))
		args = append(args, strings.TrimSpace(strings.ToLower(filters.BrandSlug)))
		idx++
	}
	if filters.Featured != nil {
		where = append(where, fmt.Sprintf("p.featured = $%d", idx))
		args = append(args, *filters.Featured)
		idx++
	}
	if filters.Status != "" {
		where = append(where, fmt.Sprintf("p.status = $%d", idx))
		args = append(args, filters.Status)
		idx++
	}
	base := " FROM products p JOIN categories c ON c.id = p.category_id JOIN brands b ON b.id = p.brand_id WHERE " + strings.Join(where, " AND ")
	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*)"+base, args...).Scan(&total); err != nil { return nil, 0, fmt.Errorf("count products: %w", err) }
	listArgs := append(append([]any{}, args...), limit, offset)
	q := fmt.Sprintf("SELECT p.id, p.category_id, p.brand_id, p.slug, p.sku, p.name, p.short_description, p.description, p.datasheet_url, p.image_url, p.status, p.featured, p.created_by, p.created_at, p.updated_at%s ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d", base, idx, idx+1)
	rows, err := r.db.Query(ctx, q, listArgs...)
	if err != nil { return nil, 0, fmt.Errorf("list products: %w", err) }
	defer rows.Close()
	items := make([]*domain.Product, 0)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil { return nil, 0, err }
		items = append(items, p)
	}
	if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("iterate products: %w", err) }
	for i := range items {
		if _, err := r.hydrateInto(ctx, items[i]); err != nil { return nil, 0, err }
	}
	return items, total, nil
}

func (r *productRepository) Update(ctx context.Context, p *domain.Product) error {
	const q = `UPDATE products SET category_id = $2, brand_id = $3, slug = $4, sku = $5, name = $6, short_description = $7, description = $8, datasheet_url = $9, image_url = $10, status = $11, featured = $12, updated_at = $13 WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, p.ID, p.CategoryID, p.BrandID, p.Slug, p.SKU, p.Name, p.ShortDescription, p.Description, p.DatasheetURL, p.ImageURL, string(p.Status), p.Featured, p.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) { return domain.ErrSlugTaken }
		return fmt.Errorf("update product: %w", err)
	}
	if tag.RowsAffected() == 0 { return domain.ErrProductNotFound }
	return nil
}

func (r *productRepository) ReplaceSpecs(ctx context.Context, productID uuid.UUID, specs []domain.ProductSpec) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM product_specs WHERE product_id = $1`, productID); err != nil { return fmt.Errorf("clear specs: %w", err) }
	for _, item := range specs {
		if _, err := r.db.Exec(ctx, `INSERT INTO product_specs (product_id, spec_key, spec_value) VALUES ($1, $2, $3)`, productID, item.Key, item.Value); err != nil { return fmt.Errorf("insert spec: %w", err) }
	}
	return nil
}

func (r *productRepository) ReplaceMedia(ctx context.Context, productID uuid.UUID, media []domain.ProductMedia) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM product_media WHERE product_id = $1`, productID); err != nil { return fmt.Errorf("clear media: %w", err) }
	for _, item := range media {
		if _, err := r.db.Exec(ctx, `INSERT INTO product_media (product_id, url, media_type, sort_order) VALUES ($1, $2, $3, $4)`, productID, item.URL, item.Type, item.SortOrder); err != nil { return fmt.Errorf("insert media: %w", err) }
	}
	return nil
}

func (r *productRepository) ReplaceCompatibility(ctx context.Context, productID uuid.UUID, items []domain.CompatibilityRule) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM compatibility_rules WHERE product_id = $1`, productID); err != nil { return fmt.Errorf("clear compatibility: %w", err) }
	for _, item := range items {
		if _, err := r.db.Exec(ctx, `INSERT INTO compatibility_rules (product_id, rule_kind, rule_value) VALUES ($1, $2, $3)`, productID, item.Kind, item.Value); err != nil { return fmt.Errorf("insert compatibility: %w", err) }
	}
	return nil
}

func (r *productRepository) getOne(ctx context.Context, q string, arg any) (*domain.Product, error) {
	p, err := scanProductRow(r.db.QueryRow(ctx, q, arg))
	if err != nil { return nil, err }
	return p, nil
}

func (r *productRepository) hydrate(ctx context.Context, p *domain.Product) (*domain.Product, error) { return r.hydrateInto(ctx, p) }
func (r *productRepository) hydrateInto(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	specRows, err := r.db.Query(ctx, `SELECT spec_key, spec_value FROM product_specs WHERE product_id = $1 ORDER BY spec_key ASC`, p.ID)
	if err != nil { return nil, fmt.Errorf("query specs: %w", err) }
	defer specRows.Close()
	p.Specs = nil
	for specRows.Next() {
		var item domain.ProductSpec
		if err := specRows.Scan(&item.Key, &item.Value); err != nil { return nil, fmt.Errorf("scan spec: %w", err) }
		p.Specs = append(p.Specs, item)
	}
	mediaRows, err := r.db.Query(ctx, `SELECT url, media_type, sort_order FROM product_media WHERE product_id = $1 ORDER BY sort_order ASC, url ASC`, p.ID)
	if err != nil { return nil, fmt.Errorf("query media: %w", err) }
	defer mediaRows.Close()
	p.Media = nil
	for mediaRows.Next() {
		var item domain.ProductMedia
		if err := mediaRows.Scan(&item.URL, &item.Type, &item.SortOrder); err != nil { return nil, fmt.Errorf("scan media: %w", err) }
		p.Media = append(p.Media, item)
	}
	compatRows, err := r.db.Query(ctx, `SELECT rule_kind, rule_value FROM compatibility_rules WHERE product_id = $1 ORDER BY rule_kind ASC, rule_value ASC`, p.ID)
	if err != nil { return nil, fmt.Errorf("query compatibility: %w", err) }
	defer compatRows.Close()
	p.Compatibility = nil
	for compatRows.Next() {
		var item domain.CompatibilityRule
		if err := compatRows.Scan(&item.Kind, &item.Value); err != nil { return nil, fmt.Errorf("scan compatibility: %w", err) }
		p.Compatibility = append(p.Compatibility, item)
	}
	return p, nil
}

func scanProduct(rows pgxRows) (*domain.Product, error) { return scanProductRow(rows) }

type pgxRows interface { Scan(dest ...any) error }

func scanProductRow(row pgxRows) (*domain.Product, error) {
	var p domain.Product
	var status string
	if err := row.Scan(&p.ID, &p.CategoryID, &p.BrandID, &p.Slug, &p.SKU, &p.Name, &p.ShortDescription, &p.Description, &p.DatasheetURL, &p.ImageURL, &status, &p.Featured, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if isNoRows(err) { return nil, domain.ErrProductNotFound }
		return nil, fmt.Errorf("scan product: %w", err)
	}
	p.Status = domain.ProductStatus(status)
	return &p, nil
}
