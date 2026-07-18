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

> Bohdаn4k:
gitops-manifests/
│
├── 📁 bootstrap/                         # Шаблон App-of-Apps (Серце GitOps)
│   ├── root-application.yaml             # Головний маніфест ArgoCD (керує іншими аплікаціями)
│   ├── 01-platform-apps.yaml             # Пакети безпеки, мережі та автоскейлінгу
│   ├── 02-datastores-apps.yaml           # Всі бази даних, черги та CDC
│   ├── 03-observability-apps.yaml        # Весь моніторинг, логування та eBPF-трейсинг
│   └── 04-core-services-apps.yaml        # Твої 15-17 бізнес-мікросервісів
│
├── 📁 01-platform/                       # Рівень 1: Мережа, Безпека та Залізо в K8s
│   ├── 📁 cilium/
│   │   ├── Application.yaml              # Маніфест ArgoCD для Cilium
│   │   ├── values.yaml                   # eBPF mode (kube-proxy replacement), Hubble enabled
│   │   └── clustermesh-policy.yaml       # Конфіг з'єднання кількох кластерів через eBPF
│   ├── 📁 istio/  Тільки для мікросервісів
│   │   ├── base-app.yaml                 # Istio Custom Resource Definitions (CRDs)
│   │   ├── istiod-app.yaml               # Control Plane (Strict mTLS політики)
│   │   └── gateway-envoy.yaml            # Ingress / API Gateway на базі Envoy (JWT validation, Rate Limiting)
│   ├── 📁 karpenter/
│   │   ├── Application.yaml
│   │   ├── nodepool-dev-spot.yaml        # Правила для дешевих EC2 Spot-інстансів
│   │   └── ec2nodeclass-bottlerocket.yaml# Специфікація під AWS Bottlerocket OS
│   ├── 📁 cert-manager/
│   │   ├── Application.yaml
│   │   └── cluster-issuers.yaml          # Let's Encrypt конфігурація для автоматичних TLS-сертифікатів
│   ├── 📁 vault-eso/
│   │   ├── external-secrets-operator.yaml# Оператор для зв'язку з HashiCorp Vault
│   │   └── cluster-secret-store.yaml     # Глобальний конект до Vault з використанням AWS IAM (IRSA)
│   ├── 📁 kyverno/
│   │   ├── Application.yaml
│   │   └── cluster-policies.yaml         # Заборона root-контейнерів + перевірка підписів Cosign
│   └── 📁 falco/
│       ├── Application.yaml              # Рантайм-захист на eBPF
│       └── falco-rules.yaml              # Справи безпеки (алерти в Slack, якщо хтось запустив bash в подій)
│
├── 📁 02-datastores/                     # Рівень 2: Бази даних, Черги та CDC (Stateful Layer)
│   ├── 📁 cloudnative-pg/                # PostgreSQL Оператор
│   │   ├── operator.yaml
│   │   ├── cluster-postgres-ha.yaml      # 3 репліки Postgres з автоматичним бекапом на AWS S3
│   │   └── debezium-cdc-connector.yaml   # Читання WAL-логів та стрімінг змін у Кафку
│   ├── 📁 strimzi-kafka/                 # Kafka Оператор
│   │   ├── operator.yaml
│   │   ├── kafka-cluster-kraft.yaml      # Kafka кластер у режимі KRaft (без ZooKeeper)
│   │   └── kafka-topics.yaml             # Декларативний опис топіків (orders, billing, delivery)
│   ├── 📁 redis-operator/
│   │   ├── operator.yaml
│   │   └── redis-cluster-cache.yaml      # Redis Cluster для кешування сесій та товарів
│   └── 📁 clickhouse-altinity/           # ClickHouse Оператор для аналітики
│       ├── operator.yaml
│       └── clickhouse-analytics-db.yaml  # Кластер для збереження гігантських аналітичних даних
│

> Bohdаn4k:
├── 📁 03-observability/                  # Рівень 3: Моніторинг, Трейсинг та Логування
│   ├── 📁 kube-prometheus-stack/         # Prometheus Operator + Grafana
│   │   ├── Application.yaml
│   │   └── values.yaml                   # Тюнінг Prometheus, додавання Thanos Sidecar
│   ├── 📁 thanos/
│   │   ├── compactor.yaml                # Даунсемплінг (стиснення) старих метрик в S3
│   │   ├── store-gateway.yaml            # Доступ Grafana до історичних даних в S3
│   │   └── querier.yaml                  # Глобальна точка збору метрик
│   ├── 📁 loki-distributed/
│   │   ├── Application.yaml              # Мікросервісна архітектура зберігання логів
│   │   └── promtail-daemonset.yaml       # Агент збору логів з кожної Bottlerocket-ноди
│   ├── 📁 opentelemetry-collector/
│   │   ├── otel-collector-config.yaml    # Приймає трейси від додатків і шле в Tempo
│   │   └── tempo-distributed.yaml        # Розподілений трейсинг (Tempo мікросервіси)
│   └── 📁 defectdojo/
│       ├── Application.yaml
│       └── values.yaml                   # Агрегатор уразливостей (сюди стікаються звіти від Trivy/Snyk)
│
├── 📁 04-core-services/                  # Рівень 4: Конфігурація твоїх 15-17 мікросервісів
│   ├── 📁 orders/
│   │   ├── Application.yaml              # Реєстрація сервісу в ArgoCD
│   │   ├── rollout.yaml                  # Маніфест Flagger / Argo Rollouts (Canary на базі Istio метрик)
│   │   ├── secret-claims.yaml            # Опис секретів, які ESO має дістати з Vault
│   │   ├── values-dev.yaml               # Налаштування для Dev (1 репліка, CPU limits менші)
│   │   └── values-prod.yaml              # Налаштування для Prod (HPA, 5+ реплік)
│   ├── 📁 cart/
│   │   ├── Application.yaml
│   │   └── ...
│   └── 📁 gateway-bff/                   # Backend-for-Frontend сервіс
│
└── 📁 shared-charts/                     # Рівень 5: Єдиний шаблон (Umbrella Chart)
    └── 📁 microservice-base/             # Щоб не дублювати YAML для 17 сервісів
        ├── Chart.yaml
        └── templates/
            ├── deployment.yaml           # Дефолтний деплоймент (якщо не Canary)
            ├── service.yaml              # K8s внутрішній сервіс
            ├── virtual-service.yaml      # Маршрутизація Istio Traffic Splitting
            ├── authorization-policy.yaml # Обмеження Istio (хто до кого має доступ)
            └── servicemonitor.yaml       # Конфіг для Prometheus, щоб збирав метрики автоматично

Етап 1: Платформа (Foundation)

Це база. Без них кластер не буде безпечним та керованим.

    Cilium: Встановлюємо першим, бо він відповідає за мережу та eBPF-інфраструктуру.

    Vault + External Secrets Operator: Потрібні, щоб інші компоненти могли отримати доступ до секретів (наприклад, для баз даних).

    Cert-Manager: Необхідний для генерації TLS-сертифікатів (зокрема для Istio та Ingress).

    Karpenter: Після того, як мережа працює, додаємо автоскейлер, щоб кластер міг динамічно масштабуватися.

    Kyverno: Впроваджуємо політики безпеки, щоб контролювати, що і як запускається в кластері.

Етап 2: Спостережливість (Observability)

Це має бути готовим до моменту запуску бізнес-логіки, щоб бачити, що відбувається.

    Kube-prometheus-stack: База для метрик.

    Loki-distributed: Збір логів.

    Tempo + OpenTelemetry Collector: Для трейсингу.

    Thanos: Якщо плануєш довгострокове зберігання метрик в S3.

    DefectDojo: Сюди будуть стікатися дані про вразливості, коли почнеш деплоїти сервіси.

Етап 3: Сервісний шар (Connectivity & Security)

    Istio (Control Plane + Gateway): Тепер, коли база є, налаштовуємо Service Mesh для мікросервісів.

    Falco: Рантайм-захист. Налаштовуємо на фінальному етапі інфраструктури.

Етап 4: Дані (Stateful Layer)

Найскладніший етап, бо вимагає коректного налаштування Persistent Volumes.

    CloudNative-PG (Postgres): Починаємо з бази.

    Redis-operator: Кешування.

    Strimzi-kafka: Черги.

    Clickhouse: Для аналітичних даних.

Етап 5: Бізнес-сервіси (Core Services)

    Shared Charts (Library): Спочатку деплоїмо твою бібліотеку чартів, щоб ArgoCD знав, звідки брати шаблони.

    Core Microservices (15-17 аплікацій): Деплоїмо їх порціями (наприклад, спочатку gateway-bff та orders, потім все інше).





1. Що має бути у values.yaml (Налаштування СТЕКУ)

Це конфігурація "інфраструктурного рівня". Ви змінюєте ці параметри лише тоді, коли треба змінити поведінку самої системи моніторингу.

    prometheus.prometheusSpec:

        resources (requests/limits).

        retention, retentionSize.

        storageSpec (SC, розмір диска).

        replicas (кількість подів).

        remoteWrite (налаштування відправки у зовнішні системи).

        thanos (конфіг sidecar-контейнера).

        nodeSelector, tolerations (куди садити поди прометея).

    grafana:

        ingress (хости, TLS).

        persistence (якщо зберігаємо базу дашбордів).

        sidecar (налаштування автоматичного підхоплення дашбордів).

    alertmanager.alertmanagerSpec:

        Глобальні налаштування маршрутизації (SMTP, Slack, PagerDuty, Webhooks).

    kubeStateMetrics / nodeExporter:

        Вмикання/вимикання та налаштування лімітів ресурсів.

2. Що має бути ОЦКРЕМИМИ маніфестами (Контент моніторингу)

Це дані, які належать вашим додаткам або специфічним бізнес-задачам. Вони не повинні залежати від версії kube-prometheus-stack.

    ServiceMonitor / PodMonitor:

        Створюється поруч із вашим додатком (наприклад, у репозиторії my-api).

        Чому: Розробник сам знає, на якому ендпоінті віддаються метрики.

    PrometheusRule (Ваші алерти):

        Створюється окремо. Повинні мати мітки, за якими їх знаходить прометей:
        YAML

        labels:
          release: prometheus-stack # Щоб пром "побачив" правило
          role: alert-rules         # Щоб правило попало у відповідний селектор

    GrafanaDashboard:

        Якщо ви використовуєте ConfigMap як джерело для Grafana (через sidecar), просто створюйте окремі ConfigMap з міткою grafana_dashboard: "1".

    AlertmanagerConfig:

        Специфічна маршрутизація (якщо команда А хоче отримувати алерти в Slack, а команда Б — на пошту). Ви не повинні редагувати головний alertmanager.config у values.yaml заради цього.

    Probe:

        Специфічні перевірки через Blackbox Exporter (наприклад, перевірка доступності зовнішнього API вашого партнера).

Як тільки ти створюєш Ingress маніфест, контролер автоматично перечитує його і оновлює свою конфігурацію (nginx.conf) всередині себе.
Як він знає, куди відправити? В маніфесті Ingress написано: "Якщо Host == grafana.daanggo.com, відправляй на Сервіс grafana".

Трафік: Браузер -> Load Balancer IP -> Nginx Pod -> Service Virtual IP (10.43.x.x) -> Kernel/iptables -> Grafana Pod IP.

Google Artifact Registry (ПОВНИЙ ШЛЯХ)
europe-central2-docker.pkg.dev / [ПРОЄКТ] / [РЕПОЗИТОРІЙ] / [НАЗВА_ПАКЕТА] : [ТЕГ]
image: europe-central2-docker.pkg.dev/PROJECT_ID/REPOSITORY/IMAGE_NAME:TAG




apiVersion: v1
kind: ServiceAccount
metadata:
  name: postgres-backup-sa
  namespace: default
  annotations:
    iam.gke.io/gcp-service-account: pg-backup-sa@YOUR_PROJECT_ID.iam.gserviceaccount.com

Покроковий процес виконання 

    Запит: Ваша програма (наприклад, скрипт бекапу PostgreSQL) всередині пода намагається звернутися до API Google Cloud.

    Перехоплення: GKE на вузлі (node) автоматично підміняє стандартний шлях до облікових даних. Замість того, щоб шукати секрет, бібліотека клієнта GCP звертається до локального Metadata Server (спеціальний сервіс, який працює на кожному вузлі GKE).

    Перевірка: Metadata Server бачить, що запит йде від пода, який використовує postgres-backup-sa. Він перевіряє анотацію в Kubernetes API.

    Обмін (Token Exchange): Metadata Server бере токен від Kubernetes, звертається до IAM API Google, і каже: "Цей под підтвердив свою особу як KSA, видай мені короткочасний OIDC-токен для GSA pg-backup-sa".

    Доступ: Програма отримує токен і виконує свою дію (наприклад, заливає бекап у бакет) від імені GSA.