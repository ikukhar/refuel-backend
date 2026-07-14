# Refuel Backend

REST API backend for a mobile fitness and nutrition app. Calculates daily calorie targets, distributes macros across meals, and picks recipes — all personalised to the user's metrics and activity load.

## Stack

- **Go 1.22+** — Gin (HTTP), GORM (ORM), golang-jwt/v5, zerolog, viper
- **PostgreSQL** — auto-migrated via GORM (5 models)
- **Admin panel** — HTMX + Pico CSS (server‑side rendered Go templates)
- **Tests** — testify, go.uber.org/mock, testcontainers-go (real Postgres in Docker)

## Architecture

```
main.go  →  repository layer (concrete, *gorm.DB)
          →  service layer    (interfaces)  →  handler layer
          →  router.Setup()  →  http.Server (graceful shutdown)
```

All services depend on repository **interfaces** (`service/interfaces.go`), not concrete implementations — making the service layer fully unit‑testable with generated mocks.

## Key Logic

### Nutrition engine (`service/nutrition.go`)

1. **BMR** — Mifflin‑St Jeor equation: `10·W + 6.25·H − 5·A + (±5/‑161)`
2. **TDEE** — BMR × 1.2 (sedentary multiplier)
3. **Activity load** — exponentially decaying window over 3 days (today 100%, yesterday 50%, −2d 25%, −3d ~12%)
4. **Macros** — 30 % protein / 25 % fat / 45 % carbs
5. **Meal distribution** — user‑defined meal periods with custom `calories_percent`; falls back to built‑in defaults.
6. **Recipe selection** — greedy fill with shuffled recipes until ≈85 % of meal target; excludes recently used recipes.

If the user has no weight / height / age / gender, a flat 2000 kcal fallback is used.

### Auth

- **Access** (short‑lived) + **Refresh** (long‑lived) JWT tokens
- **Refresh rotation** — every refresh increments `TokenVersion` on the user model, instantly revoking all previously issued tokens
- bcrypt for password hashing

### Activity ingestion

Idempotent via `source_id` unique constraint — re‑sending the same external activity returns the existing record instead of creating a duplicate.

### Admin panel

Protected with HTTP Basic Auth. Built with server‑side Go templates + HTMX for smooth navigation without a JS framework.

## Modules

```
internal/
├── config/         — .env loading via viper
├── model/          — GORM models (User, Activity, DailyNutrition, Recipe, MealPeriod, MealType)
├── repository/     — data access (GORM queries)
├── service/        — business logic (nutrition engine, auth, etc.)
│   └── mocks/      — generated mock implementations
├── handler/        — HTTP handlers (Gin)
│   └── admin/      — admin panel handlers
├── middleware/     — JWT auth, Basic auth, request logging
├── router/         — Gin route setup
└── testutil/       — test helpers (ptr converters)
```

## API Endpoints

| Route | Auth | Description |
|---|---|---|
| `GET /api/v1/health` | — | Health check |
| `GET /api/v1/meal-periods/default` | — | Built‑in default meal periods |
| `POST /api/v1/auth/register` | — | Register |
| `POST /api/v1/auth/login` | — | Login |
| `POST /api/v1/auth/refresh` | — | Refresh tokens |
| `GET /api/v1/user/profile` | JWT | Get profile + meal periods |
| `PUT /api/v1/user/profile` | JWT | Update profile |
| `GET /api/v1/activities` | JWT | List activities |
| `POST /api/v1/activities` | JWT | Create activity |
| `GET /api/v1/nutrition/today` | JWT | Today's nutrition plan |
| `POST /api/v1/meal-periods` | JWT | Upsert meal periods |
| `GET /admin/*` | Basic | Admin panel |

## Configuration

via `.env` file (see `.env.example`):

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=refuel
JWT_SECRET=your-secret-key
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=72h
APP_PORT=8080
ADMIN_USER=admin
ADMIN_PASS=admin
```

## Running

```bash
make run       # start via Air (hot reload)
make build     # compile
make test      # run all tests
make docker    # docker compose up
```

## Tests

- **Unit** — service + handler tests with gomock mocks
- **Integration** — repository tests against a real Postgres spun up by testcontainers-go
- **API** — full HTTP tests with the real router and mocked repositories

```bash
go test ./...
```

## Development

Requires Go 1.22+ and a running PostgreSQL instance. Migrations are handled by GORM's `AutoMigrate` (no external migration tool needed in dev).
