# AutoImport KZ

Production-quality monolithic web app for Kazakhstan car import cost estimation.

## Tech Stack
- Go (net/http + chi)
- PostgreSQL
- html/template
- bcrypt + session cookies

## Features
- Auth: register, login, logout
- Cars catalog and details
- Import cost calculator with breakdown
- Calculation history and warnings
- Car purchase flow (buy car + purchases history)
- Purchase statuses: created -> paid -> delivered
- Dashboard page with key metrics
- Admin UI actions: edit/delete for cars and countries
- Admin car image upload and image_url support
- Admin CRUD for countries and cars
- Middleware: request logging, auth, admin role

## Project Structure
- `cmd/server/main.go`
- `internal/config`
- `internal/handlers`
- `internal/services`
- `internal/repositories`
- `internal/models`
- `internal/middleware`
- `internal/templates`
- `static/styles`
- `migrations`
- `docs`

## Setup
1. Create database:
```sql
CREATE DATABASE autoimport;
```

2. Apply migrations:
```bash
psql -d autoimport -f migrations/001_init.sql
psql -d autoimport -f migrations/002_purchases.sql
psql -d autoimport -f migrations/003_cars_image_url.sql
```

3. Set environment variables (defaults shown):
```bash
ADDR=:8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/autoimport?sslmode=disable
SESSION_KEY=change-me-in-env
SESSION_NAME=autoimport_session
STATIC_DIR=static
TEMPLATE_DIR=internal/templates
DEFAULT_ADMIN_EMAIL=admin@demo.kz
DEFAULT_ADMIN_PASSWORD=Admin123!
```

4. Install deps and run:
```bash
go mod tidy
go run ./cmd/server
```

Open: http://localhost:8080

## Default Admin
- Email: `admin@demo.kz`
- Password: `Admin123!`

## API List (Web Routes)
### Auth + User
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `GET /me`

### Cars
- `GET /cars`
- `GET /cars/{id}`
- `POST /cars/{id}/buy`
- `GET /purchases`
- `GET /purchases/{id}`
- `POST /purchases/{id}/pay`
- `POST /admin/purchases/{id}/deliver`
- `GET /dashboard`
- `GET /admin/cars`
- `POST /admin/cars`
- `POST /admin/cars/{id}/update`
- `POST /admin/cars/{id}/delete`
- `PUT /admin/cars/{id}`
- `DELETE /admin/cars/{id}`

### Calculator + Countries
- `GET /calculator`
- `POST /calculator/calculate`
- `GET /history`
- `GET /admin/countries`
- `POST /admin/countries`
- `POST /admin/countries/{id}/update`
- `POST /admin/countries/{id}/delete`
- `PUT /admin/countries/{id}`
- `DELETE /admin/countries/{id}`

## Commit Plan (Suggested)
Member A (Auth + user):
- Commit 1: Add user model, user repo, auth service, session middleware wiring
- Commit 2: Add auth handlers, profile page, login/register templates

Member B (Cars):
- Commit 1: Add car model + repo + service + public car handlers
- Commit 2: Add admin car CRUD handlers + templates

Member C (Calculator + Countries):
- Commit 1: Add country model + repo + service + admin countries templates
- Commit 2: Add calculation service + handlers + history view + warnings

## Diagrams
- `docs/erd.puml` (ERD)
- `docs/class.puml` (UML Class)
- `docs/usecase.puml` (Use Case)
- `docs/dataflow.mmd` (Data Flow)
