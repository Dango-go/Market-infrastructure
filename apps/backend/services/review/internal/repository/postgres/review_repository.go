package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/review/internal/domain"
	"github.com/google/uuid"
)

type reviewRepository struct{ db pgxConn }

func (r *reviewRepository) Create(ctx context.Context, review *domain.Review) error {
	_, err := r.db.Exec(ctx, `INSERT INTO reviews (id, product_id, account_id, rating, title, body, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, review.ID, review.ProductID, review.AccountID, review.Rating, review.Title, review.Body, review.CreatedAt, review.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrReviewAlreadyExists
		}
		return fmt.Errorf("create review: %w", err)
	}
	return nil
}

func (r *reviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	return r.getOne(ctx, `SELECT id, product_id, account_id, rating, title, body, created_at, updated_at FROM reviews WHERE id = $1`, id)
}

func (r *reviewRepository) ListByProductID(ctx context.Context, productID uuid.UUID, limit, offset int32) ([]domain.Review, int64, error) {
	rows, err := r.db.Query(ctx, `SELECT id, product_id, account_id, rating, title, body, created_at, updated_at, COUNT(*) OVER() AS total_count FROM reviews WHERE product_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, productID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list product reviews: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Review, 0)
	var total int64
	for rows.Next() {
		var item domain.Review
		if err := rows.Scan(&item.ID, &item.ProductID, &item.AccountID, &item.Rating, &item.Title, &item.Body, &item.CreatedAt, &item.UpdatedAt, &total); err != nil {
			return nil, 0, fmt.Errorf("scan product review: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate product reviews: %w", err)
	}
	return items, total, nil
}

func (r *reviewRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]domain.Review, int64, error) {
	rows, err := r.db.Query(ctx, `SELECT id, product_id, account_id, rating, title, body, created_at, updated_at, COUNT(*) OVER() AS total_count FROM reviews WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, accountID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list account reviews: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Review, 0)
	var total int64
	for rows.Next() {
		var item domain.Review
		if err := rows.Scan(&item.ID, &item.ProductID, &item.AccountID, &item.Rating, &item.Title, &item.Body, &item.CreatedAt, &item.UpdatedAt, &total); err != nil {
			return nil, 0, fmt.Errorf("scan account review: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate account reviews: %w", err)
	}
	return items, total, nil
}

func (r *reviewRepository) Update(ctx context.Context, review *domain.Review) error {
	tag, err := r.db.Exec(ctx, `UPDATE reviews SET rating = $2, title = $3, body = $4, updated_at = $5 WHERE id = $1`, review.ID, review.Rating, review.Title, review.Body, review.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrReviewNotFound
	}
	return nil
}

func (r *reviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM reviews WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrReviewNotFound
	}
	return nil
}

func (r *reviewRepository) GetSummaryByProductID(ctx context.Context, productID uuid.UUID) (domain.ReviewSummary, error) {
	item := domain.ReviewSummary{ProductID: productID}
	if err := r.db.QueryRow(ctx, `SELECT COALESCE(COUNT(*), 0), COALESCE(AVG(rating)::float8, 0), COUNT(*) FILTER (WHERE rating = 1), COUNT(*) FILTER (WHERE rating = 2), COUNT(*) FILTER (WHERE rating = 3), COUNT(*) FILTER (WHERE rating = 4), COUNT(*) FILTER (WHERE rating = 5) FROM reviews WHERE product_id = $1`, productID).Scan(&item.TotalReviews, &item.AverageRating, &item.Rating1Count, &item.Rating2Count, &item.Rating3Count, &item.Rating4Count, &item.Rating5Count); err != nil {
		return domain.ReviewSummary{}, fmt.Errorf("get review summary: %w", err)
	}
	return item, nil
}

func (r *reviewRepository) getOne(ctx context.Context, query string, arg any) (*domain.Review, error) {
	var item domain.Review
	if err := r.db.QueryRow(ctx, query, arg).Scan(&item.ID, &item.ProductID, &item.AccountID, &item.Rating, &item.Title, &item.Body, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if isNoRows(err) {
			return nil, domain.ErrReviewNotFound
		}
		return nil, fmt.Errorf("get review: %w", err)
	}
	return &item, nil
}
