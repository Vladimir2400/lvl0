# WB-Service

Демонстрационный микросервис для обработки заказов с использованием Kafka, PostgreSQL и in-memory кэша.

## 📋 Содержание

- [Обзор](#обзор)
- [Архитектура](#архитектура)
- [Технологии](#технологии)
- [Требования](#требования)
- [Установка и запуск](#установка-и-запуск)
- [Конфигурация](#конфигурация)
- [API Endpoints](#api-endpoints)
- [Тестирование](#тестирование)
- [Структура проекта](#структура-проекта)
- [Особенности реализации](#особенности-реализации)
- [Производительность](#производительность)
- [Развертывание](#развертывание)
- [Лицензия](#лицензия)

## 🎯 Обзор

WB-Service - это микросервис для обработки заказов в режиме реального времени. Сервис получает данные о заказах из Kafka, сохраняет их в PostgreSQL, кэширует в памяти для быстрого доступа и предоставляет REST API для получения информации о заказах.

### Основные возможности

- ✅ Потребление сообщений из Kafka в реальном времени
- ✅ Валидация входящих данных перед сохранением
- ✅ Хранение данных в PostgreSQL с поддержкой связей (delivery, payment, items)
- ✅ LRU кэш с TTL для быстрого доступа к данным
- ✅ REST API для получения заказов по UID
- ✅ Восстановление кэша из БД при перезапуске
- ✅ Graceful shutdown для корректного завершения работы
- ✅ Веб-интерфейс для поиска заказов
- ✅ Генератор тестовых данных
- ✅ Покрытие тестами 80%+

## 🏗️ Архитектура

### Диаграмма компонентов

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            WB-Service                                   │
│                                                                         │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐           │
│  │  HTTP Server │     │ Kafka        │     │  Database    │           │
│  │  (Gin)       │     │  Consumer    │     │  Layer       │           │
│  │              │     │              │     │  (GORM)      │           │
│  │  - REST API  │     │  - Subscribe │     │              │           │
│  │  - Health    │     │  - Validate  │     │  - Create    │           │
│  │  - Web UI    │     │  - Process   │     │  - Read      │           │
│  └──────┬───────┘     └──────┬───────┘     └──────┬───────┘           │
│         │                    │                    │                   │
│         │                    │                    │                   │
│         ├────────────────────┴────────────────────┤                   │
│         │                                         │                   │
│         ▼                                         ▼                   │
│  ┌────────────────────────────────────────────────────┐               │
│  │          LRU Cache (In-Memory)                     │               │
│  │  - TTL Support                                     │               │
│  │  - Auto Eviction                                   │               │
│  │  - Thread-Safe                                     │               │
│  └────────────────────────────────────────────────────┘               │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
           ▲                    ▲                    ▲
           │                    │                    │
           │                    │                    │
    ┌──────┴──────┐     ┌───────┴────────┐   ┌──────┴──────┐
    │ HTTP Client │     │ Kafka Topic    │   │ PostgreSQL  │
    │             │     │ "orders"       │   │ Database    │
    └─────────────┘     └────────────────┘   └─────────────┘
                               ▲
                               │
                        ┌──────┴──────┐
                        │ Data        │
                        │ Generator   │
                        └─────────────┘
```

### Поток данных

#### 1. Обработка входящих заказов (Kafka → DB → Cache)

```
Kafka Message → Consumer.ReadMessage()
                     ↓
              JSON.Unmarshal()
                     ↓
              Validator.Validate()
                     ↓ (valid)
              Database.Create()
                     ↓ (success)
              Cache.Set()
                     ↓
              Kafka.Commit()
```

#### 2. Обработка HTTP запросов (Cache-Aside Pattern)

```
HTTP GET /order/{uid}
         ↓
    Cache.Get(uid)
         ↓
    ┌────┴────┐
    │  Hit?   │
    └────┬────┘
         │
    ┌────┴────────────┐
    ▼                 ▼
  (Yes)              (No)
    │                 │
    │            DB.GetOrder(uid)
    │                 ↓
    │            Cache.Set(uid, order)
    │                 │
    └────────┬────────┘
             ▼
       Return JSON
```

#### 3. Graceful Shutdown

```
SIGINT/SIGTERM
      ↓
Cancel Context
      ↓
   ┌──┴───────────────┐
   ▼                  ▼
Kafka Consumer    HTTP Server
stops reading     waits for
                  active requests
   │                  │
   │                  ▼
   │            Shutdown (30s timeout)
   │                  │
   └──────┬───────────┘
          ▼
    Close DB Connections
          ↓
    Exit Clean
```

## 🛠️ Технологии

### Backend
- **Go 1.23+** - основной язык программирования
- **Gin** - HTTP веб-фреймворк для REST API
- **GORM** - ORM для работы с базой данных
- **segmentio/kafka-go** - клиент для Apache Kafka

### Инфраструктура
- **PostgreSQL 13** - реляционная база данных
- **Apache Kafka** - брокер сообщений для event streaming
- **Zookeeper** - координация Kafka кластера
- **Docker Compose** - контейнеризация инфраструктуры

### Инструменты разработки
- **Docker** - контейнеризация
- **Make** - автоматизация сборки и тестирования
- **golangci-lint** - линтер для Go кода
- **goimports** - форматирование и сортировка импортов

## 📦 Требования

### Обязательные
- Go 1.23 или выше
- Docker 20.10+
- Docker Compose 1.29+
- Make (опционально, но рекомендуется)

### Для разработки
- Git
- golangci-lint (для линтинга)
- goimports (для форматирования)

## 🚀 Установка и запуск

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd WB-service
```

### 2. Запуск инфраструктуры

```bash
# Запуск PostgreSQL, Kafka и Zookeeper
make docker-up

# Или напрямую через docker-compose
docker-compose up -d

# Проверка статуса контейнеров
make docker-status
```

Инфраструктура будет доступна:
- PostgreSQL: `localhost:5434`
- Kafka: `localhost:9092`
- Zookeeper: `localhost:2181`

### 3. Установка зависимостей

```bash
make deps

# Или напрямую
go mod download
go mod tidy
```

### 4. Запуск приложения

```bash
# Через Make
make run

# Или напрямую
go run main.go
```

Сервис будет доступен: `http://localhost:8080`

### 5. Генерация тестовых данных

```bash
# Сгенерировать 10 заказов с интервалом 1 секунда
make generate-data

# Или с кастомными параметрами
go run cmd/generator/main.go -count=50 -delay=500ms

# Все параметры генератора
go run cmd/generator/main.go \
  -brokers=localhost:9092 \
  -topic=orders \
  -count=100 \
  -delay=1s
```

### 6. Проверка работы

```bash
# Health check
curl http://localhost:8080/health

# Получение заказа (после генерации данных)
curl http://localhost:8080/order/{order_uid}

# Или откройте веб-интерфейс
open http://localhost:8080
```

## ⚙️ Конфигурация

Все настройки конфигурируются через переменные окружения. Если переменная не задана, используется значение по умолчанию.

### База данных

| Переменная | Описание | Значение по умолчанию |
|-----------|----------|----------------------|
| `DB_HOST` | Хост PostgreSQL | `127.0.0.1` |
| `DB_PORT` | Порт PostgreSQL | `5434` |
| `DB_USER` | Имя пользователя БД | `wb_user` |
| `DB_PASSWORD` | Пароль БД | `wb_password` |
| `DB_NAME` | Имя базы данных | `wb_db` |
| `DB_SSL_MODE` | Режим SSL | `disable` |

### Kafka

| Переменная | Описание | Значение по умолчанию |
|-----------|----------|----------------------|
| `KAFKA_BROKERS` | Адреса брокеров Kafka | `localhost:9092` |
| `KAFKA_TOPIC` | Топик для заказов | `orders` |
| `KAFKA_GROUP_ID` | ID группы потребителей | `order-group` |
| `KAFKA_MIN_BYTES` | Минимальный размер батча | `10000` (10KB) |
| `KAFKA_MAX_BYTES` | Максимальный размер батча | `10000000` (10MB) |

### HTTP сервер

| Переменная | Описание | Значение по умолчанию |
|-----------|----------|----------------------|
| `SERVER_HOST` | Хост сервера | `""` (все интерфейсы) |
| `SERVER_PORT` | Порт сервера | `8080` |

### Кэш

| Переменная | Описание | Значение по умолчанию |
|-----------|----------|----------------------|
| `CACHE_MAX_SIZE` | Максимальный размер кэша | `1000` элементов |
| `CACHE_TTL` | Время жизни элемента | `3600` секунд (1 час) |

### Пример конфигурации

```bash
# .env файл
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=myuser
export DB_PASSWORD=mypassword
export DB_NAME=mydb

export KAFKA_BROKERS=kafka1:9092,kafka2:9092
export KAFKA_TOPIC=orders
export KAFKA_GROUP_ID=order-service-group

export SERVER_PORT=8080

export CACHE_MAX_SIZE=2000
export CACHE_TTL=7200

# Запуск с конфигурацией
source .env
go run main.go
```

## 🌐 API Endpoints

### GET /order/{order_uid}

Получение заказа по уникальному идентификатору.

**Параметры:**
- `order_uid` (path) - уникальный идентификатор заказа

**Пример запроса:**
```bash
curl http://localhost:8080/order/b563feb7b2b84b6test
```

**Успешный ответ (200 OK):**
```json
{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389232,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}
```

**Ошибка - заказ не найден (404 Not Found):**
```json
{
  "error": "record not found"
}
```

**Ошибка базы данных (500 Internal Server Error):**
```json
{
  "error": "database error"
}
```

### GET /health

Проверка состояния сервиса.

**Пример запроса:**
```bash
curl http://localhost:8080/health
```

**Ответ (200 OK):**
```json
{
  "status": "ok"
}
```

### GET /

Веб-интерфейс для поиска заказов. Открывается в браузере:
```bash
open http://localhost:8080
```

## 🧪 Тестирование

### Запуск тестов

```bash
# Все тесты
make test

# Тесты с покрытием
make test-cover

# Подробный вывод + HTML отчет покрытия
make test-verbose

# Проверка минимального покрытия 80%
make test-coverage-check
```

### Тесты отдельных пакетов

```bash
make test-models       # Тесты моделей
make test-config       # Тесты конфигурации
make test-cache        # Тесты кэша
make test-validator    # Тесты валидатора
make test-repository   # Тесты репозитория
make test-generator    # Тесты генератора
make test-main         # Тесты HTTP handlers
```

### Бенчмарки и профилирование

```bash
# Бенчмарки производительности кэша
make benchmark

# Профилирование CPU
make profile-cpu

# Профилирование памяти
make profile-mem
```

### Линтинг

```bash
# Запуск линтера (требует golangci-lint)
make lint
```

### Интеграционные тесты

```bash
# Запуск интеграционных тестов (требует Docker)
make test-integration
```

## 📂 Структура проекта

```
WB-service/
├── cmd/                        # Исполняемые приложения
│   └── generator/             # Генератор тестовых данных
│       ├── main.go           # Kafka producer с fake data
│       └── main_test.go      # Тесты генератора
│
├── config/                    # Конфигурация
│   ├── config.go             # Загрузка настроек из ENV
│   └── config_test.go        # Тесты конфигурации
│
├── database/                  # База данных
│   ├── database.go           # Инициализация GORM
│   └── database_test.go      # Тесты подключения к БД
│
├── internal/                  # Внутренние пакеты
│   ├── cache/                # LRU кэш с TTL
│   │   ├── lru_cache.go
│   │   └── lru_cache_test.go
│   │
│   ├── interfaces/           # Интерфейсы для DI
│   │   └── interfaces.go
│   │
│   ├── repository/           # Слой доступа к данным
│   │   ├── database.go
│   │   └── database_test.go
│   │
│   └── validator/            # Валидация бизнес-правил
│       ├── order_validator.go
│       └── order_validator_test.go
│
├── kafka/                     # Kafka consumer
│   ├── consumer.go           # Обработка сообщений
│   └── consumer_test.go      # Тесты consumer
│
├── models/                    # Модели данных
│   ├── models.go             # GORM модели (Order, Delivery, Payment, Item)
│   └── models_test.go        # Тесты моделей
│
├── producer/                  # Kafka producer (устаревший)
│   └── producer.go
│
├── web/                       # Веб-интерфейс
│   └── index.html            # HTML страница для поиска заказов
│
├── docker-compose.yml         # Инфраструктура (Kafka, PostgreSQL)
├── schema.sql                 # SQL схема базы данных
├── Makefile                   # Команды для сборки и тестирования
├── go.mod                     # Go модуль и зависимости
├── go.sum                     # Хеши зависимостей
├── main.go                    # Точка входа приложения
├── main_test.go              # Тесты HTTP handlers
└──  README.md                  # Документация (этот файл)

```

## ⚡ Особенности реализации

### 1. LRU Cache с TTL

Реализован кэш на основе двух структур данных:
- **HashMap** для O(1) поиска
- **Doubly Linked List** для отслеживания порядка использования

**Возможности:**
- Автоматическое удаление наименее используемых элементов
- TTL (Time To Live) с фоновой очисткой устаревших элементов
- Thread-safe операции с использованием `sync.RWMutex`
- Warm-up из БД при старте

**Код:** `internal/cache/lru_cache.go`

### 2. Graceful Shutdown

Корректное завершение работы при получении SIGINT/SIGTERM:
1. Остановка Kafka consumer (через context cancellation)
2. Завершение обработки активных HTTP запросов (30s timeout)
3. Закрытие соединений с базой данных
4. Логирование всех этапов

**Код:** `main.go:115-143`

### 3. Cache-Aside Pattern

Паттерн кэширования для оптимизации доступа к данным:
1. Проверка кэша при запросе
2. При промахе - загрузка из БД
3. Сохранение результата в кэш
4. Возврат данных клиенту

**Код:** `main.go:24-58`

### 4. Валидация данных

Многоуровневая валидация заказов:
- **Структурная** - корректность JSON
- **Типовая** - соответствие типов данных
- **Бизнес-правила** - положительные суммы, корректные email/phone и т.д.

Невалидные сообщения логируются и коммитятся (не вызывают бесконечные retry).

**Код:** `internal/validator/order_validator.go`

### 5. Repository Pattern

Использование интерфейсов для абстракции доступа к данным:
- Упрощает тестирование (можно использовать mock)
- Позволяет менять реализацию БД без изменения бизнес-логики
- Следует принципам SOLID (Dependency Inversion)

**Код:** `internal/interfaces/interfaces.go`, `internal/repository/database.go`

### 6. Обработка ошибок Kafka

Стратегия обработки сообщений:
- ✅ **Commit** - успешная обработка и сохранение
- ✅ **Commit** - ошибка парсинга JSON (невалидный формат)
- ✅ **Commit** - ошибка валидации (невалидные данные)
- ❌ **No Commit** - ошибка БД (будет retry)

**Код:** `kafka/consumer.go:60-111`

## 📊 Производительность

### Кэш

- **Время доступа:** O(1)
- **Throughput:** ~10,000 req/sec (на типичном железе)
- **Memory usage:** Зависит от CACHE_MAX_SIZE и размера Order объектов
- **Hit rate:** 90%+ для повторяющихся запросов

### База данных

- **Индексы:** на `order_uid` для быстрого поиска
- **Connection pooling:** управляется GORM
- **Preload:** загрузка связанных сущностей одним запросом
- **Transactions:** для консистентности данных

### HTTP Server

- **Framework:** Gin (высокопроизводительный роутер)
- **JSON сериализация:** стандартный `encoding/json`
- **Keep-alive:** поддержка persistent connections
- **Concurrency:** горутины для каждого запроса

## 🚢 Развертывание

### Docker (рекомендуется)

1. **Сборка Docker образа:**

```bash
docker build -t wb-service:latest .
```

2. **Запуск с docker-compose:**

```bash
docker-compose up -d
```

### Production

**Рекомендации для production:**

1. **Используйте внешние сервисы:**
   - Managed PostgreSQL (AWS RDS, Google Cloud SQL)
   - Managed Kafka (Confluent Cloud, AWS MSK)

2. **Настройте переменные окружения:**
   ```bash
   export DB_HOST=production-db.example.com
   export DB_PORT=5432
   export KAFKA_BROKERS=kafka1.prod:9092,kafka2.prod:9092
   export CACHE_MAX_SIZE=10000
   export CACHE_TTL=3600
   ```

3. **Настройте reverse proxy:**
   - nginx или Traefik для SSL termination
   - Rate limiting
   - Request buffering

4. **Включите SSL/TLS:**
   - PostgreSQL: `DB_SSL_MODE=require`
   - Kafka: настройте SSL/SASL аутентификацию

5. **Настройте мониторинг:**
   - Prometheus метрики
   - Grafana дашборды
   - Alertmanager для алертов
   - ELK/Loki для логов

6. **Horizontal scaling:**
   - Запустите несколько инстансов сервиса
   - Load balancer перед HTTP серверами
   - Kafka consumer group для распределенной обработки

7. **Безопасность:**
   - Используйте secrets management (Vault, AWS Secrets Manager)
   - Настройте network policies
   - Регулярно обновляйте зависимости
   - Сканируйте Docker образы на уязвимости

### Kubernetes

**Пример deployment.yaml:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wb-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: wb-service
  template:
    metadata:
      labels:
        app: wb-service
    spec:
      containers:
      - name: wb-service
        image: wb-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-secrets
              key: host
        - name: KAFKA_BROKERS
          value: "kafka.default.svc.cluster.local:9092"
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```
