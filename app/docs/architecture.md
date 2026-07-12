# Embedded Electronics Store - System Architecture

> A production-grade e-commerce platform for embedded electronics: MCUs, dev boards,
> sensors, modules, power supplies, tools, cables, kits, and accessories.
> The platform emphasizes rich product data, compatibility guidance, inventory control,
> multi-warehouse fulfillment, and delivery tracking.

This document is the single source of architectural truth for the project. It defines the
service boundaries, ownership rules, event contracts, and build order before the rest of
the codebase grows around it.

---

## 1. Architectural Principles

- Domain-Driven Design: each service is a bounded context with its own ubiquitous
  language and its own database.
- Clean Architecture: `domain` is independent, `application` depends only on `domain`,
  and outer layers implement ports.
- Event-Driven Architecture: services exchange state changes asynchronously through Kafka.
- API-first access: synchronous reads go through the gateway and service APIs; no direct
  cross-service database access.
- Idempotency first: every consumer must handle duplicated events safely.

### Dependency Rule

```text
transport (Gin)         repository (pgx/sqlc)        infrastructure (kafka/jwt/minio)
        \                        |                         /
         \                       |                        /
          -------->   application (use cases)   <--------
                                   |
                                   v
                                domain
                   (entities, value objects, ports, errors)
```

`domain` and `application` import no framework code. HTTP, database, Kafka, Redis, and
storage clients live only in outer layers and are wired in `cmd/<service>/main.go`.

---

## 2. Monorepo Layout

```text
embedded-market/
├── backend/
│   ├── go.mod                       # single Go module
│   ├── pkg/                         # shared, cross-service libraries
│   │   ├── config/
│   │   ├── logger/
│   │   ├── httpx/
│   │   ├── middleware/
│   │   ├── postgres/
│   │   ├── redis/
│   │   ├── kafka/
│   │   ├── events/
│   │   ├── validator/
│   │   └── apperr/
│   └── services/
│       ├── auth/                    # identity, sessions, OAuth, JWKS
│       ├── user/                    # customer profiles, addresses, preferences
│       ├── catalog/                 # products, categories, brands, specs, media
│       ├── inventory/               # stock, warehouses, reservations
│       ├── pricing/                 # prices, discounts, promotions, tax rules
│       ├── cart/                    # carts, items, checkout snapshots
│       ├── order/                   # order lifecycle, cancellations, returns
│       ├── payment/                 # payment intents, captures, refunds
│       ├── shipment/                # shipping methods, labels, tracking, delivery
│       ├── search/                  # search index, facets, autocomplete
│       ├── notification/            # email, SMS, push, templates, event fan-out
│       └── gateway/                 # API gateway / BFF
├── frontend/
│   └── storefront/                  # customer-facing web app
├── proto/                           # Kafka event schema contracts
├── docs/
└── scripts/
```

### Per-service internal layout

```text
services/<svc>/
├── cmd/<svc>/main.go
├── internal/
│   ├── config/config.go
│   ├── domain/
│   ├── application/
│   ├── repository/postgres/
│   ├── infrastructure/
│   └── transport/http/
└── db/
    ├── migrations/
    ├── queries/
    └── sqlc.yaml
```

---

## 3. Service Catalog & Ownership

| # | Service | Bounded context | Owns (Postgres tables) | Cache | Storage |
|---|---------|-----------------|------------------------|-------|---------|
| 1 | **auth** | Identity & access | accounts, sessions, oauth_identities, outbox | - | - |
| 2 | **user** | Customer profile | profiles, addresses, preferences, wishlist_items, outbox | - | avatars -> object storage |
| 3 | **catalog** | Product catalog | products, categories, brands, product_specs, product_media, compatibility_rules, outbox | - | media metadata only |
| 4 | **inventory** | Stock control | warehouses, stock_items, stock_movements, reservations, outbox | Redis for hot stock reads | - |
| 5 | **pricing** | Pricing engine | prices, price_lists, discounts, promotions, tax_rules, outbox | Redis for active price lookup | - |
| 6 | **cart** | Shopping cart | carts, cart_items, cart_snapshots, outbox | Redis for active carts optional | - |
| 7 | **order** | Order lifecycle | orders, order_items, order_status_history, returns, outbox | - | - |
| 8 | **payment** | Money movement | payment_intents, transactions, refunds, payment_attempts, outbox | - | - |
| 9 | **shipment** | Fulfillment & delivery | shipments, shipment_items, carriers, tracking_events, delivery_addresses, outbox | - | label files -> object storage optional |
| 10 | **search** | Discovery | search_documents, facets, suggestions, popularity_counters, outbox | Redis | - |
| 11 | **notification** | Messaging | notifications, templates, delivery_attempts, subscriptions, outbox | - | - |
| 12 | **gateway** | Edge / BFF | stateless | Redis (rate limit, cache) | - |

### Data Ownership

Data ownership is absolute. A service reads another service's data only through:

1. Kafka events that it has projected into its own read model, or
2. a synchronous API call brokered by the gateway.

No shared tables. No cross-schema joins. No exception.

---

## 4. Domain Shape

The storefront is a real electronics shop, so the domain is centered on product depth,
inventory truth, and delivery confidence.

### Catalog Highlights

- Electronics-grade product pages with full technical specs.
- Compatibility data for MCU families, dev boards, sensors, and power adapters.
- Rich media, datasheets, manuals, pinouts, and bundle suggestions.

### Commerce Highlights

- Exact stock visibility by warehouse.
- Reservation-based checkout to avoid overselling.
- Price rules, promotions, and tax handling isolated from catalog data.
- Shipment tracking integrated into order state.

### Embedded-specific Differentiators

- Compatibility rules for "works with" and "requires" relationships.
- Suggested accessories and starter kits.
- BOM-friendly product metadata for engineering buyers.

---

## 5. Eventing Model

Every event uses a versioned envelope defined in `pkg/events`.

```jsonc
{
  "id": "uuid",
  "type": "order.created",
  "version": 1,
  "source": "order-service",
  "subject": "uuid",
  "correlation_id": "uuid",
  "occurred_at": "RFC3339",
  "data": {}
}
```

### Core Topics

| Topic | Producer | Primary consumers |
|-------|----------|-------------------|
| `user.registered` | auth | user, notification |
| `user.addresses.updated` | user | order, shipment |
| `product.created` | catalog | search, notification |
| `product.updated` | catalog | search |
| `product.compatibility.updated` | catalog | search, notification |
| `stock.reserved` | inventory | order, search |
| `stock.released` | inventory | order, search |
| `stock.adjusted` | inventory | search |
| `price.changed` | pricing | cart, search, notification |
| `promotion.started` | pricing | search, notification |
| `cart.checked_out` | cart | order, inventory |
| `order.created` | order | payment, shipment, notification |
| `order.paid` | payment | order, shipment, notification |
| `payment.failed` | payment | order, notification |
| `shipment.created` | shipment | order, notification, search |
| `shipment.tracked` | shipment | order, notification |
| `shipment.delivered` | shipment | order, notification |
| `review.created` | notification or a future review service | search, catalog |

### Delivery Guarantees

- At-least-once delivery.
- Idempotent consumers keyed on envelope `id`.
- Transactional outbox for all state-changing producers.
- Consumers persist processed-event markers or use natural upserts.

---

## 6. Cross-Cutting Conventions

- Primary keys: `uuid`, preferably time-ordered where it helps indexes.
- Timestamps: `created_at`, `updated_at`, optional `deleted_at` for soft delete.
- Pagination: offset pagination with `page` and `page_size`, max 100.
- Filtering: explicit allow-lists only.
- Validation: `go-playground/validator` with stable 422 responses.
- Errors: stable envelope with `code`, `message`, and `details`.
- Observability: propagate `X-Request-Id` and `X-Correlation-Id` everywhere.
- Auth: RS256 JWT access tokens verified locally by every service and the gateway.
- Refresh tokens: opaque, hashed, session-backed.
- Config: 12-factor environment-based config only.

---

## 7. Gateway And Frontend

### Gateway

The gateway is the only public backend edge.

Responsibilities:

- JWT verification and request context propagation.
- Public API aggregation for the storefront.
- Rate limiting and request shaping.
- Optional BFF endpoints for the product page and checkout flow.

### Storefront

The frontend is a separate web app, not a microservice.

Suggested pages:

- Home
- Category listing
- Search results
- Product page
- Cart
- Checkout
- Delivery options
- Order tracking
- Account and order history
- Admin entry points if needed later

The storefront should be SEO-friendly and optimized for product discovery and conversion.

---

## 8. Build Order

1. Auth
2. User
3. Catalog
4. Inventory
5. Pricing
6. Cart
7. Order
8. Payment
9. Shipment
10. Search
11. Notification
12. Gateway

The storefront can be built in parallel against mocked or gateway-backed APIs, but the
checkout flow should only be finalized once `cart`, `order`, `payment`, and `shipment`
contracts are stable.

---

## 9. Implementation Notes

- Each service is built end-to-end: entities, use cases, repositories, handlers, routes,
  config, migrations, and queries.
- Search and notification are read-heavy and should be event-driven projections rather
  than source-of-truth writers for commerce state.
- Compatibility can start as catalog metadata and evolve into a dedicated service later if
  the product line becomes complex enough.



auth-serviceРеєстрація, логін, OAuth, сесії, JWT, ролі.
Це вже є.

user-serviceПрофіль клієнта.
Адреси доставки.
Телефони, налаштування, wishlist, збережені кошики.

catalog-serviceТовари, категорії, бренди, характеристики, фото, slug, SEO-дані.
Для embedded-ніші тут же можна зберігати технічні спеки, сумісність, datasheets.

inventory-serviceСклади, залишки, резервування, списання.
Дуже важливо для реального магазину.

pricing-serviceБазові ціни, знижки, промокоди, акції, tax/currency rules.
Відокремити від catalog корисно, щоб ціни можна було міняти незалежно.

cart-serviceКошики, items, saved cart, checkout snapshot.

order-serviceОформлення замовлення.
Стейт-машина замовлення.
Cancel, fulfillment, returns initiation.

payment-servicePayment intents, capture, refund.
Інтеграція з платіжним провайдером.
Не змішувати з order.

shipment-serviceДоставка, способи доставки, трекінг, labels, tracking events.
Саме це ти окремо згадував, тому сервіс обов’язковий.

search-serviceПовнотекстовий пошук, фільтри, фасети, популярність.
Може будувати read-model через події з catalog/inventory/pricing.

notification-serviceEmail, SMS, Telegram/push, шаблони, тригерні повідомлення.
Підтвердження замовлення, статуси доставки, промо, відновлення кошика.

compatibility-service або recommendation-serviceЦе твоя “фішка” для embedded-магазину.
Наприклад: “цей модуль сумісний з цим контролером”, “це живлення підходить під цей борд”, “додай кабель/адаптер/датчик, який реально пасує”.