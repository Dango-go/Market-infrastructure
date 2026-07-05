package application

import (
	"context"
	"strings"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/services/review/internal/domain"
	"github.com/google/uuid"
)

type ReviewUseCase struct{ Deps }

func NewReviewUseCase(d Deps) *ReviewUseCase { return &ReviewUseCase{Deps: d} }

type CreateReviewInput struct {
	ProductID uuid.UUID
	Rating    int32
	Title     string
	Body      string
}

type UpdateReviewInput struct {
	Rating int32
	Title  string
	Body   string
}

func (uc *ReviewUseCase) Create(ctx context.Context, accountID uuid.UUID, input CreateReviewInput) (ReviewView, error) {
	if err := validateReview(input.Rating, input.Body); err != nil {
		return ReviewView{}, err
	}
	now := uc.Clock.Now()
	review := domain.NewReview(uc.IDs.New(), input.ProductID, accountID, input.Rating, strings.TrimSpace(input.Title), strings.TrimSpace(input.Body), now)
	if err := uc.Store.Reviews().Create(ctx, review); err != nil {
		return ReviewView{}, err
	}
	return toReviewView(review), nil
}

func (uc *ReviewUseCase) Update(ctx context.Context, accountID, reviewID uuid.UUID, input UpdateReviewInput) (ReviewView, error) {
	if err := validateReview(input.Rating, input.Body); err != nil {
		return ReviewView{}, err
	}
	review, err := uc.Store.Reviews().GetByID(ctx, reviewID)
	if err != nil {
		return ReviewView{}, err
	}
	if review.AccountID != accountID {
		return ReviewView{}, apperr.Forbidden("review_access_denied", "you cannot modify this review")
	}
	review.Update(input.Rating, strings.TrimSpace(input.Title), strings.TrimSpace(input.Body), uc.Clock.Now())
	if err := uc.Store.Reviews().Update(ctx, review); err != nil {
		return ReviewView{}, err
	}
	return toReviewView(review), nil
}

func (uc *ReviewUseCase) Delete(ctx context.Context, accountID, reviewID uuid.UUID) error {
	review, err := uc.Store.Reviews().GetByID(ctx, reviewID)
	if err != nil {
		return err
	}
	if review.AccountID != accountID {
		return apperr.Forbidden("review_access_denied", "you cannot delete this review")
	}
	return uc.Store.Reviews().Delete(ctx, reviewID)
}

func (uc *ReviewUseCase) ListByProductID(ctx context.Context, productID uuid.UUID, limit, offset int32) ([]ReviewView, int64, error) {
	items, total, err := uc.Store.Reviews().ListByProductID(ctx, productID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]ReviewView, 0, len(items))
	for i := range items {
		out = append(out, toReviewView(&items[i]))
	}
	return out, total, nil
}

func (uc *ReviewUseCase) ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]ReviewView, int64, error) {
	items, total, err := uc.Store.Reviews().ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]ReviewView, 0, len(items))
	for i := range items {
		out = append(out, toReviewView(&items[i]))
	}
	return out, total, nil
}

func (uc *ReviewUseCase) GetSummaryByProductID(ctx context.Context, productID uuid.UUID) (ReviewSummaryView, error) {
	summary, err := uc.Store.Reviews().GetSummaryByProductID(ctx, productID)
	if err != nil {
		return ReviewSummaryView{}, err
	}
	return ReviewSummaryView(summary), nil
}

func validateReview(rating int32, body string) error {
	if rating < 1 || rating > 5 {
		return apperr.Invalid("invalid_rating", "rating must be between 1 and 5")
	}
	if strings.TrimSpace(body) == "" {
		return apperr.Invalid("invalid_body", "body is required")
	}
	return nil
}

func toReviewView(item *domain.Review) ReviewView {
	return ReviewView{ID: item.ID, ProductID: item.ProductID, AccountID: item.AccountID, Rating: item.Rating, Title: item.Title, Body: item.Body, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
