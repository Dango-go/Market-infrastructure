# Embedded Market

Мікросервісний магазин embedded-електроніки з gateway, auth, каталогом, кошиком, замовленнями, оплатами, доставкою, аналітикою та додатковими сервісами для пошуку, рекомендацій, відгуків і wishlist.

## 1. Призначення проєкту

Проєкт моделює повноцінний e-commerce застосунок для магазину embedded-компонентів:
- фронтенд магазину з товарами, кошиком, checkout і доставкою
- окремий `api-gateway` як вхідна точка для клієнта
- набір доменних мікросервісів з окремими зонами відповідальності
- PostgreSQL як основне сховище стану
- подієвий підхід через outbox + Kafka/Redpanda там, де це доречно

## 2. Список сервісів

### `frontend`
Вітрина магазину.

Відповідає за:
- каталог товарів
- авторизацію користувача
- роботу з кошиком
- оформлення замовлення
- відображення оплат, доставок і wishlist

Не є джерелом істини для бізнес-стану. Працює через `api-gateway`.

### `api-gateway`
Єдина публічна HTTP-точка входу для фронтенду.

Відповідає за:
- маршрутизацію запитів до внутрішніх сервісів
- JWT authentication middleware для protected маршрутів
- уніфікований зовнішній API для клієнта

### `auth-service`
Сервіс автентифікації та сесій.

Відповідає за:
- register / login / logout / refresh
- випуск access token і refresh token
- OAuth login providers
- JWKS endpoint
- rate limiting для auth flow

### `user-service`
Сервіс профілю користувача.

Відповідає за:
- профіль користувача
- preferences
- адреси доставки / billing
- bootstrap профілю після `user.registered`

Після рефакторингу більше не містить wishlist-логіку.

### `catalog-service`
Каталог товарів.

Відповідає за:
- категорії
- бренди
- товари
- базові публічні дані про продукт

### `inventory-service`
Складські залишки.

Відповідає за:
- наявність товару
- доступні залишки
- резервування / release / adjust stock

### `pricing-service`
Ціноутворення.

Відповідає за:
- поточні ціни товарів
- compare-at / promo-friendly pricing state

### `cart-service`
Кошик користувача.

Відповідає за:
- активний кошик
- додавання / оновлення / видалення позицій
- checkout кошика
- подію `cart.checked_out`

### `order-service`
Замовлення.

Відповідає за:
- створення замовлення з товарних позицій
- статуси замовлення
- історію замовлень користувача

### `payment-service`
Оплати.

Відповідає за:
- створення платежів
- confirm / fail / refund
- історію платежів

### `shipping-service`
Доставка.

Відповідає за:
- створення shipment
- tracking metadata
- статуси доставки

### `notification-service`
Нотифікації та шаблони повідомлень.

Відповідає за:
- запис та видачу notification state
- шаблони повідомлень
- подальшу основу для email / sms / push інтеграцій

### `search-service`
Пошук по каталогу.

Відповідає за:
- search endpoint для товарів
- фільтрацію / знаходження релевантних позицій

### `review-service`
Відгуки.

Відповідає за:
- створення та редагування review
- список review по товару
- summary по рейтингу

### `recommendation-service`
Рекомендації.

Відповідає за:
- trending / recommended items
- видачу добірок для storefront

### `analytics-service`
Аналітика storefront-подій.

Відповідає за:
- прийом подій типу `page_view`
- накопичення аналітичного сигналу

### `wishlist-service`
Окремий сервіс збережених товарів.

Відповідає за:
- список збережених товарів користувача
- додавання товару в wishlist
- видалення товару з wishlist
- подію `user.wishlist.updated`

## 3. Взаємодія сервісів

### Загальна карта

```mermaid
flowchart LR
    FE[Frontend] --> GW[API Gateway]

    GW --> AUTH[auth-service]
    GW --> USER[user-service]
    GW --> CATALOG[catalog-service]
    GW --> INVENTORY[inventory-service]
    GW --> PRICING[pricing-service]
    GW --> CART[cart-service]
    GW --> ORDER[order-service]
    GW --> PAYMENT[payment-service]
    GW --> SHIPPING[shipping-service]
    GW --> NOTIFY[notification-service]
    GW --> SEARCH[search-service]
    GW --> REVIEW[review-service]
    GW --> RECO[recommendation-service]
    GW --> ANALYTICS[analytics-service]
    GW --> WISHLIST[wishlist-service]

    AUTH --> PG[(PostgreSQL)]
    USER --> PG
    CATALOG --> PG
    INVENTORY --> PG
    PRICING --> PG
    CART --> PG
    ORDER --> PG
    PAYMENT --> PG
    SHIPPING --> PG
    NOTIFY --> PG
    SEARCH --> PG
    REVIEW --> PG
    RECO --> PG
    ANALYTICS --> PG
    WISHLIST --> PG

    AUTH -. rate limit .-> REDIS[(Redis)]

    AUTH -. outbox/events .-> KAFKA[(Kafka / Redpanda)]
    USER -. consume user.registered .-> KAFKA
    CART -. outbox/events .-> KAFKA
    ORDER -. outbox/events .-> KAFKA
    PAYMENT -. outbox/events .-> KAFKA
    SHIPPING -. outbox/events .-> KAFKA
    NOTIFY -. outbox/events .-> KAFKA
    WISHLIST -. outbox/events .-> KAFKA
```

## 4. Основні бізнес-воркфлоу

### 4.1 Авторизація

```text
Frontend -> API Gateway -> auth-service -> PostgreSQL
                               -> Redis (optional, rate limiting)
```

Потік:
1. користувач реєструється або логіниться
2. `auth-service` перевіряє облікові дані
3. видається access token + refresh cookie
4. frontend працює через gateway з bearer token

### 4.2 Завантаження storefront

```text
Frontend -> API Gateway
                  -> catalog-service
                  -> recommendation-service
                  -> review-service
                  -> analytics-service
```

Потік:
1. storefront запитує товари
2. окремо тягнуться категорії, рекомендації та review summary
3. frontend збирає з цього вітрину

### 4.3 Робота з кошиком

```text
Frontend -> API Gateway -> cart-service -> PostgreSQL
```

Потік:
1. користувач додає товар
2. `cart-service` створює або оновлює активний кошик
3. під час checkout кошик переходить у checked-out state
4. формується outbox event `cart.checked_out`

### 4.4 Checkout / order / payment / shipping

```text
Frontend -> API Gateway -> order-service
Frontend -> API Gateway -> payment-service
Frontend -> API Gateway -> shipping-service
Frontend -> API Gateway -> cart-service
```

Поточний demo flow:
1. frontend створює order
2. frontend створює payment
3. frontend confirm-ить payment
4. frontend створює shipment
5. frontend фіналізує cart checkout
6. frontend перечитує operational data

Примітка:
- це ще клієнтська orchestration модель
- для production-архітектури краще винести цей flow в окремий backend orchestration endpoint або checkout-service

### 4.5 Wishlist

```text
Frontend -> API Gateway -> wishlist-service -> PostgreSQL
```

Потік:
1. користувач натискає heart на товарі
2. frontend викликає `POST /api/v1/wishlist/:product_id`
3. `wishlist-service` зберігає товар у wishlist
4. при видаленні викликається `DELETE /api/v1/wishlist/:product_id`
5. сервіс публікує `user.wishlist.updated`

### 4.6 Bootstrap профілю користувача

```text
auth-service -> Kafka/Redpanda -> user-service -> PostgreSQL
```

Потік:
1. `auth-service` створює подію `user.registered`
2. `user-service` її споживає
3. створює profile/preferences bootstrap state

## 5. Хто є джерелом істини

- `auth-service`: акаунт, сесії, токени
- `user-service`: профіль, адреси, preferences
- `catalog-service`: каталог товарів
- `inventory-service`: залишки
- `pricing-service`: ціни
- `cart-service`: активний кошик
- `order-service`: замовлення
- `payment-service`: платежі
- `shipping-service`: доставка
- `wishlist-service`: збережені товари
- `review-service`: відгуки
- `analytics-service`: події аналітики

## 6. Які сервіси мають базу даних

### Сервіси / компоненти, яким потрібен стан

Ось кому реально потрібне stateful сховище:

- `postgres`
- `redis`
- `redpanda` або `kafka`-кластер

### Які прикладні сервіси використовують PostgreSQL

Цим сервісам потрібна власна схема / власна БД у PostgreSQL:
- `auth-service`
- `user-service`
- `catalog-service`
- `inventory-service`
- `pricing-service`
- `cart-service`
- `order-service`
- `shipping-service`
- `payment-service`
- `notification-service`
- `search-service`
- `review-service`
- `recommendation-service`
- `analytics-service`
- `wishlist-service`

Тобто з прикладних мікросервісів зараз **усі бекенд-сервіси, крім `api-gateway`, мають стан у PostgreSQL**.

### Кому зазвичай НЕ потрібен StatefulSet

Зазвичай як `Deployment`, а не `StatefulSet`, можна запускати:
- `api-gateway`
- `frontend`
- усі stateless Go-сервіси, якщо їх стан лежить у зовнішньому PostgreSQL/Redis/Kafka

Важливий момент:
- наявність БД у сервісу не означає, що сам сервіс треба робити `StatefulSet`
- `StatefulSet` зазвичай потрібен саме для **самих stateful infrastructure-компонентів**: `PostgreSQL`, `Redis` (залежно від режиму), `Kafka/Redpanda`
- прикладні сервіси (`auth-service`, `cart-service`, `order-service` тощо) найчастіше лишаються `Deployment`, бо їхній стан живе в зовнішній БД

## 7. Рекомендація по Kubernetes

### Як я б це розклав

**StatefulSet / stateful infra:**
- PostgreSQL
- Redis
- Kafka / Redpanda

**Deployment / stateless app layer:**
- frontend
- api-gateway
- auth-service
- user-service
- catalog-service
- inventory-service
- pricing-service
- cart-service
- order-service
- shipping-service
- payment-service
- notification-service
- search-service
- review-service
- recommendation-service
- analytics-service
- wishlist-service

## 8. Що ще варто доробити перед production-like запуском

- перевести checkout orchestration з frontend у backend
- додати справжні readiness-перевірки в `api-gateway`
- перевести `docker-compose` з `go run` на `build:` через Dockerfile
- додати e2e smoke tests
- уніфікувати observability: request id, metrics, structured logs
- визначити event consumers для recommendation/analytics/notification flows

## 9. Короткий висновок

Зараз система вже виглядає як повноцінна мікросервісна e-commerce платформа для embedded-магазину:
- є окремий auth layer
- є user domain
- є catalog / inventory / pricing
- є cart / order / payment / shipping
- є review / recommendation / analytics / wishlist
- є gateway і красивий storefront

Для DevOps-проєкту це вже дуже сильна база, яку можна розкладати в Docker, а далі в Kubernetes.
