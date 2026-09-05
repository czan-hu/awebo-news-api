# awebo-api

Бэкенд для новостного сайта awebo. Реализует REST API по спецификации
`awebo-news/docs/openapi.yaml` (Gin + GORM + PostgreSQL, чистая архитектура:
`entities` / `dto` / `repositories` / `services` / `handlers`).

## Стек

- Go 1.25, [Gin](https://github.com/gin-gonic/gin), [GORM](https://gorm.io) + `gorm.io/driver/postgres`
- PostgreSQL 16
- Авторизация: одноразовый код на email → JWT (HS256, реализован вручную в
  `app/pkg/jwtutil`, без сторонних JWT-библиотек)
- Миграции: собственный минимальный раннер (`app/infrastructure/database/migrator.go`),
  без сторонних зависимостей — применяет `*.up.sql` из `migrations/` по порядку
  и хранит применённые версии в таблице `schema_migrations`

Никакие новые модули в `go.mod` не добавлялись — всё написано на уже
подключённых зависимостях (это было осознанное решение, так как в текущем
окружении не было доступа к рабочему Go-тулчейну, чтобы прогнать
`go mod tidy` и проверить `go.sum`).

## Структура

```
app/
  entities/        доменные объекты, которые отдаются наружу как JSON
  dto/              входные тела запросов и query-параметры
  exceptions/       единый формат ошибки { statusCode, message, errors? }
  repositories/     доступ к БД (по одному файлу на домен)
  services/         бизнес-логика, маппинг models -> entities
  handlers/         HTTP-роуты (Gin), авторизация, CORS
  infrastructure/
    config/         чтение .env
    database/       подключение к Postgres + миграции
    database/models GORM-модели
    database/migrations  *.up.sql / *.down.sql
    server/         обёртка над http.Server (graceful shutdown)
  pkg/
    jwtutil/        свой HS256 JWT
    idgen/          slug для тем форума, id для комментариев, код подтверждения
    validate/       простые regex-проверки (username, code)
```

## Запуск

### Docker Compose (рекомендуется)

```bash
docker compose up --build
```

Поднимет Postgres и API (`:8080`), миграции применятся автоматически при
старте приложения.

### Локально

```bash
# Postgres поднять отдельно, либо docker compose up db
go mod tidy   # на всякий случай, если что-то разошлось с go.sum
go run ./cmd
```

Для локального запуска без docker поменяйте `DB_HOST=db` на `DB_HOST=localhost`
в `.env`.

⚠️ Код писался и проверялся без доступа к Go-тулчейну в этом окружении —
после генерации обязательно прогоните `go build ./...` и поправьте, если
где-то разойдётся сигнатура.

## Переменные окружения (`.env`)

| Переменная | Назначение |
| --- | --- |
| `API_PREFIX` | префикс роутов, по умолчанию `/api/v1` (как в OpenAPI `servers`) |
| `SERVICE_PORT` | порт HTTP-сервера |
| `DB_*` | подключение к Postgres |
| `JWT_SECRET`, `JWT_TTL_HOURS` | подпись и время жизни токена |
| `CODE_TTL_SECONDS`, `CODE_RESEND_SECONDS` | TTL одноразового кода и антиспам-пауза между отправками |
| `MAIL_DRIVER` | `console` (код пишется в лог, удобно для разработки) или `smtp` |
| `SMTP_*` | настройки почты, если `MAIL_DRIVER=smtp` |
| `UPLOADS_DIR` | куда сохраняются загруженные файлы (аватары), отдаётся статикой по `/uploads/*` |
| `MAX_AVATAR_SIZE_MB` | лимит размера загружаемого аватара |

## Что покрыто

Все эндпоинты из `docs/openapi.yaml`: Auth (код на почту → verify →
complete-profile, logout), Users (профиль, публичная страница, liked/saved/
history, статьи автора), Articles (лента с фильтрами category/tag/author/
featured/sort, детальная, related, like/save/view), Categories, Newsletter,
Forum (темы, плоские комментарии с `parentId`, голосование с автогоном
score=1 при создании комментария), Search.

При первом запуске миграция `000008_seed_demo_content` создаёт одного
редакционного пользователя, одну статью и одну тему форума — чтобы
фронтенд не был пустым сразу после `docker compose up`.

### Роли и публикация статей

У `users` есть системная роль (`role`, отдельно от публичного `title` —
строки вида «Автор · Технологии»): `user` (по умолчанию), `creator`,
`editor`, `admin`. Публиковать статьи через `POST /articles` может любая
роль, для которой `entities.CanPublishArticles` возвращает true —
сейчас это `creator`/`editor`/`admin`.

Апгрейд `user → creator` идёт через модерацию, не самообслуживанием:

1. `POST /users/me/creator-request` — пользователь подаёт заявку (требует
   заполненного профиля; повторно подать нельзя, пока предыдущая заявка
   не рассмотрена — see `idx_creator_requests_one_pending`).
2. `GET /users/me/creator-request` — статус своей последней заявки.
3. `GET /admin/creator-requests` — очередь необработанных заявок, доступна
   только роли `admin`.
4. `POST /admin/creator-requests/{id}/approve` — роль заявителя становится
   `creator`. `.../reject` — заявка закрывается без изменения роли.

Для роли `editor`/`admin` (и любых будущих ролей) отдельной админки нет —
выдаются вручную в БД. Для локальной разработки миграция
`000010_seed_admin_account` создаёт демо-админа `admin@awebo.example`
(логин обычным флоу — код на почту, в `MAIL_DRIVER=console` смотрите в
логах контейнера `api`).

## Дизайн-решения, которые стоит знать

- **Теги статей и контент-блоки** хранятся как `jsonb`, а не через
  промежуточные таблицы — тегов немного, а фильтрация `tags @> '["x"]'`
  с GIN-индексом достаточно быстрая для этого масштаба.
- **История чтения** — это отдельная таблица `article_views`
  `(article_id, user_id)` с обновлением `viewed_at` при повторном визите
  (upsert), а не лог всех просмотров.
- **Верификационный код** хранится в открытом виде с TTL — это OTP на
  6 цифр с коротким временем жизни, а не пароль, усложнять хешированием
  не было смысла.
- **JWT — самописный HS256**, а не `golang-jwt`, чтобы не тянуть новую
  зависимость в окружении без доступа к интернету/Go-тулчейну для
  `go mod tidy`.
