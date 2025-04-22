# Banking Service REST API

Безопасный и эффективный REST API для банковских услуг, разработанный на Go. Сервис предоставляет управление пользователями, операции со счетами, управление картами, кредитные услуги и финансовую аналитику.

## Возможности

- **Управление пользователями**
  - Регистрация с проверкой уникальности email/username
  - Валидация email (формат, домен)
  - Валидация пароля (сложность, длина)
  - JWT-based аутентификация
  - Контроль доступа на основе ролей

- **Операции со счетами**
  - Создание и управление банковскими счетами
  - Операции по вкладам и снятию средств
  - Переводы между счетами (с транзакциями)
  - Отслеживание баланса
  - Проверка прав доступа к счетам

- **Управление картами**
  - Генерация виртуальных карт (алгоритм Луна)
  - Безопасное хранение данных карт (PGP шифрование + HMAC)
  - Хеширование CVV через bcrypt
  - Управление статусом карт (активна/заблокирована)
  - Проверка прав доступа к картам

- **Кредитные услуги**
  - Оформление и управление кредитами
  - Расчет аннуитетных платежей
  - Генерация графиков платежей
  - Автоматическая обработка платежей (каждые 12 часов)
  - Штрафы за просрочку платежей (+10% к сумме)
  - Интеграция с ЦБ РФ для получения ключевой ставки

- **Финансовая аналитика**
  - История транзакций
  - Анализ кредитной нагрузки
  - Прогнозирование баланса (до 365 дней)
  - Финансовая статистика
  - Отчеты по доходам/расходам

- **Внешние интеграции**
  - API Центрального Банка России (ключевая ставка через SOAP)
  - SMTP для email-уведомлений
  - Безопасное шифрование данных

## Технический стек

- **Язык**: Go 1.23+
- **Фреймворк**: gorilla/mux
- **База данных**: PostgreSQL 15 (Alpine) с pgcrypto
- **Аутентификация**: JWT (golang-jwt/jwt/v5)
- **Логирование**: logrus
- **Шифрование**: bcrypt, HMAC-SHA256, PGP
- **Email**: gomail.v2
- **XML парсинг**: beevik/etree
- **UUID**: google/uuid

## Структура базы данных

### Таблицы

- **users**: Данные пользователей
  - id, username, email, password, created_at, updated_at
  - Индексы по email и username

- **accounts**: Банковские счета
  - id, user_id, balance, currency, created_at, updated_at
  - Индекс по user_id

- **cards**: Данные карт
  - id, user_id, account_id, card_number (PGP), expiry_date (PGP), cvv_hash (bcrypt)
  - card_type, status, hmac, created_at, updated_at
  - Индексы по user_id и account_id

- **transactions**: История операций
  - id, from_account_id, to_account_id, amount, currency
  - description, transaction_type, created_at
  - Индексы по from_account_id и to_account_id

- **credits**: Кредиты
  - id, user_id, account_id, amount, interest_rate
  - term_months, status, created_at, updated_at
  - Индексы по user_id и account_id

- **payment_schedules**: Графики платежей
  - id, credit_id, payment_number, payment_date
  - amount, principal, interest, status, created_at
  - Индексы по credit_id и payment_date

### Конфигурация

Сервис может быть настроен через переменные окружения или конфигурационный файл:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "database": {
    "host": "localhost",
    "port": 5438,
    "user": "postgres",
    "password": "********",
    "dbname": "abi_banking"
  },
  "jwt": {
    "secret": "your-256-bit-secret",
    "expiration_time": "24h"
  },
  "scheduler": {
    "interval": "12h"
  },
  "logging": {
    "level": "info",
    "format": "json"
  },
  "cbr": {
    "wsdl": "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
    "cache_ttl": "1h"
  },
  "smtp": {
    "host": "smtp.example.com",
    "port": 587,
    "username": "noreply@example.com",
    "password": "********"
  },
  "pgp": {
    "public_key_path": "/path/to/public.key",
    "private_key_path": "/path/to/private.key",
    "passphrase": "********"
  }
}
```

Для запуска с Docker Compose:

```bash
docker-compose up -d
```

## Процессы и планировщики

- **Планировщик платежей**
  - Запуск каждые 12 часов
  - Автоматическое списание платежей
  - Обработка просроченных платежей
  - Начисление штрафов
  - Отправка уведомлений

- **Интеграция с ЦБ РФ**
  - SOAP-запросы к DailyInfoWebServ
  - Получение ключевой ставки
  - Кэширование данных
  - Обработка ошибок

- **Логирование**
  - Настраиваемые уровни (debug, info, error)
  - Структурированные логи
  - Контекстная информация
  - Ротация логов

## Структура проекта

```
.
├── cmd/                 # Точка входа приложения
├── internal/           # Внутренние пакеты
│   ├── config/        # Управление конфигурацией
│   ├── database/      # Подключение и настройка БД
│   ├── handlers/      # HTTP обработчики запросов
│   ├── integration/   # Интеграции с внешними сервисами
│   │   ├── cbr/      # Интеграция с ЦБ (SOAP)
│   │   └── smtp/     # Интеграция с email-сервисом
│   ├── middleware/    # HTTP middleware
│   ├── models/        # Модели данных
│   ├── repository/    # Репозитории БД
│   ├── router/        # Определение маршрутов
│   ├── scheduler/     # Планировщик фоновых задач
│   └── service/       # Бизнес-логика
├── migrations/        # Миграции базы данных
└── tests/            # Тестовые файлы
```

## API Endpoints

### Публичные эндпоинты

- `POST /api/v1/public/register` - Регистрация пользователя
- `POST /api/v1/public/login` - Аутентификация пользователя

### Защищенные эндпоинты

#### Счета
- `POST /api/v1/accounts` - Создание счета
- `GET /api/v1/accounts/{id}` - Получение информации о счете
- `POST /api/v1/accounts/{id}/deposit` - Внесение средств
- `POST /api/v1/accounts/{id}/withdraw` - Снятие средств
- `POST /api/v1/accounts/transfer` - Перевод между счетами
- `GET /api/v1/accounts/{id}/predict` - Прогноз баланса

#### Карты
- `POST /api/v1/cards` - Создание карты
- `GET /api/v1/cards/{id}` - Получение информации о карте
- `POST /api/v1/cards/{id}/block` - Блокировка карты
- `POST /api/v1/cards/{id}/unblock` - Разблокировка карты

#### Кредиты
- `POST /api/v1/credits` - Создание кредита
- `GET /api/v1/credits/{id}` - Получение информации о кредите
- `GET /api/v1/credits/{id}/schedule` - Получение графика платежей
- `POST /api/v1/credits/{id}/pay` - Внесение платежа

#### Аналитика
- `GET /api/v1/analytics/transactions` - Получение аналитики транзакций
- `GET /api/v1/analytics/credits` - Получение аналитики кредитов

## Функции безопасности

- JWT-based аутентификация (24 часа)
- PGP шифрование данных карт
- HMAC для целостности данных
- Хеширование паролей с помощью bcrypt
- Хеширование CVV с помощью bcrypt
- Контроль доступа на основе ролей
- Валидация входных данных
- Ограничение частоты запросов
- Защита от CORS
- Проверка прав доступа к ресурсам

## Начало работы

### Предварительные требования

- Go 1.23+
- PostgreSQL 17
- PGP ключи для шифрования карт
- Доступ к SMTP серверу
- Доступ к API ЦБ РФ

### Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/yourusername/banking-service.git
cd banking-service
```

2. Установите зависимости:
```bash
go mod download
```

3. Настройте переменные окружения:
```bash
cp .env.example .env
# Отредактируйте .env с вашей конфигурацией
```

4. Запустите миграции базы данных:
```bash
go run cmd/migrate/main.go
```

5. Запустите сервис:
```bash
go run cmd/main.go
```

### Конфигурация

Сервис может быть настроен через переменные окружения или конфигурационный файл:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "database": {
    "host": "localhost",
    "port": 5438,
    "user": "bank_user",
    "password": "********",
    "dbname": "bank_db"
  },
  "jwt": {
    "secret": "your-256-bit-secret",
    "expiration_time": "24h"
  },
  "scheduler": {
    "interval": "12h"
  },
  "logging": {
    "level": "info",
    "format": "json"
  },
  "cbr": {
    "wsdl": "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
    "cache_ttl": "1h"
  },
  "smtp": {
    "host": "smtp.example.com",
    "port": 587,
    "username": "noreply@example.com",
    "password": "********"
  },
  "pgp": {
    "public_key_path": "/path/to/public.key",
    "private_key_path": "/path/to/private.key",
    "passphrase": "********"
  }
}
```

Для безопасности рекомендуется использовать переменные окружения:

```bash
export DB_USER=bank_user
export DB_PASSWORD=your_secure_password
export JWT_SECRET=your-256-bit-secret
export SMTP_PASSWORD=your_smtp_password
export PGP_PASSPHRASE=your_pgp_passphrase
```