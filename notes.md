SERVICE_NAME=analytics-service
ENVIRONMENT=development
LOG_LEVEL=debug
HTTP_PORT=8094
# Якщо клієнт довго відправляє дані або сервіс довго відповідає, з'єднання обірветься, щоб не забивати пам'ять.
HTTP_READ_TIMEOUT=5s
HTTP_WRITE_TIMEOUT=10s
# Час на завершення сервісу
SHUTDOWN_TIMEOUT=10s
# Після увімкнення в кластері сервіс звертається до кор днс для отримання адреси. Якщо адресам відразу  @localhost:5432, він відкриває мережевий канал (TCP-з'єднання) до цієї адреси на порт 5432. Авторизація: Він посилає логін і пароль, які були в тому ж самому DSN-рядку. Після цього сервіс починає надсилати туди SQL-запити
POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/analytics?sslmode=disable  # +
POSTGRES_MAX_CONNS=10
POSTGRES_MIN_CONNS=2
# Сокети живуть довго. Але якщо між сервісом і базою стоїть якийсь мережевий проміжний пристрій (наприклад, Firewall чи Load Balancer). параметр з максимальним часом одного унікального сокету в системі між бд та сервісом.
POSTGRES_MAX_CONN_LIFETIME=1h
# Параметр каже сервісу: "Слухай, якщо база не відповіла за 5 секунд, не чекай вічно — здавайся, пиши помилку в логи і спробуй ще раз трохи пізніше".
POSTGRES_CONNECT_TIMEOUT=5s
# 
JWT_PUBLIC_KEY_PEM=
JWT_PUBLIC_KEY_FILE=../../deployments/dev/keys/jwt_public.pem  # +
# Хто саме видав
JWT_ISSUER=embedded-market-auth
# audience
JWT_AUDIENCE=embedded-market



          securityContext:      readonly fs
            allowPrivilegeEscalation: false  
            readOnlyRootFilesystem: true    
            capabilities:
              drop: ["ALL"]  

# Поки startupProbe не поверне успіх (200 OK), усі інші проби (liveness та readiness) вимкнені. Як тільки startupProbe відповів успіхом, він вимикається назавжди, і контроль передається liveness та readiness.
          startupProbe:
            httpGet:
              path: /healthz
              port: 8094
            failureThreshold: 30  Це ліміт помилок. Kubernetes дозволяє сервісу "помилитися" (не відповісти або повернути помилку) 30 разів поспіль.
            periodSeconds: 10   Це частота. Kubernetes стукає в твої двері (/healthz або /readyz) кожні 10 секунд.

failureThreshold: 30 # Даємо 30 спроб
periodSeconds: 10   # Кожні 10 секунд

# "Якщо 3 перевірки поспіль провалені — вбивай контейнер". З інтервалом X
failureThreshold: 3 

# Kubernetes надсилає HTTP-запит до твого сервісу на /healthz.
timeoutSeconds: 3



      affinity:
 
        nodeAffinity:
 
          preferredDuringSchedulingIgnoredDuringExecution:  назва стратегії. preferred дає команду дозволу якщо вільних нод немає, запустити под на рандомній
            - weight: 100
              preference:  Умови ноди
                matchExpressions:
                  - key: topology.kubernetes.io/zone
                    operator: In
                    values:
                      - 

topologySpreadConstraints: - рівномірне розподілення подів по нодах (кількість подів == кількість нод розподілення)

maxSkew: 1 дозволяє, щоб на одній ноді було 2 поди, а на іншій 1 (різниця 1). Це прийнятно.
        - maxSkew: 1   # Diff between count pods on different nodes
 whenUnsatisfiable: DoNotSchedule  # Дія якщо система не може розподілити

# labelSelector: значення по яких йде фільтрація та порівняння подів для дій афініті
      topologySpreadConstraints:
        - maxSkew: 1    
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: DoNotSchedule  # ScheduleAnyway alternative
          labelSelector:
            matchLabels:
              app: payment

