package postgres

import (
	"context"
	"fmt"

	"github.com/embedded-market/backend/services/analytics/internal/domain"
)

type eventRepository struct{ db pgxConn }

func (r *eventRepository) Create(ctx context.Context, event *domain.Event) error {
	_, err := r.db.Exec(ctx, `INSERT INTO analytics_events (id, account_id, session_id, product_id, event_type, path, referrer, query, user_agent, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, event.ID, event.AccountID, event.SessionID, event.ProductID, event.EventType, event.Path, event.Referrer, event.Query, event.UserAgent, event.CreatedAt)
	if err != nil { return fmt.Errorf("create analytics event: %w", err) }
	return nil
}

func (r *eventRepository) Overview(ctx context.Context, days int32) (domain.Overview, error) {
	item := domain.Overview{Days: days}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*), COUNT(DISTINCT session_id), COUNT(*) FILTER (WHERE event_type = 'product_view'), COUNT(*) FILTER (WHERE event_type = 'search'), COUNT(*) FILTER (WHERE event_type = 'add_to_cart'), COUNT(*) FILTER (WHERE event_type = 'begin_checkout'), COUNT(*) FILTER (WHERE event_type = 'purchase') FROM analytics_events WHERE created_at >= NOW() - ($1::int * INTERVAL '1 day')`, days).Scan(&item.TotalEvents, &item.UniqueSessions, &item.ProductViews, &item.Searches, &item.AddToCarts, &item.Checkouts, &item.Purchases); err != nil {
		return domain.Overview{}, fmt.Errorf("analytics overview: %w", err)
	}
	return item, nil
}

func (r *eventRepository) TopProducts(ctx context.Context, days, limit int32) ([]domain.TopProduct, error) {
	rows, err := r.db.Query(ctx, `SELECT product_id, COUNT(*) FILTER (WHERE event_type = 'product_view') AS views, COUNT(*) FILTER (WHERE event_type = 'add_to_cart') AS add_to_carts, COUNT(*) FILTER (WHERE event_type = 'purchase') AS purchases, (COUNT(*) FILTER (WHERE event_type = 'purchase') * 3.0 + COUNT(*) FILTER (WHERE event_type = 'add_to_cart') * 1.5 + COUNT(*) FILTER (WHERE event_type = 'product_view') * 0.2)::float8 AS conversion_score FROM analytics_events WHERE product_id IS NOT NULL AND created_at >= NOW() - ($1::int * INTERVAL '1 day') GROUP BY product_id ORDER BY conversion_score DESC, purchases DESC, views DESC LIMIT $2`, days, limit)
	if err != nil { return nil, fmt.Errorf("top analytics products: %w", err) }
	defer rows.Close()
	items := make([]domain.TopProduct, 0)
	for rows.Next() {
		var item domain.TopProduct
		if err := rows.Scan(&item.ProductID, &item.Views, &item.AddToCarts, &item.Purchases, &item.ConversionScore); err != nil { return nil, fmt.Errorf("scan top product: %w", err) }
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { return nil, fmt.Errorf("iterate top products: %w", err) }
	return items, nil
}
