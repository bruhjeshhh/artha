# Backend developer guide — Rent & Cost Analyzer

Everything a backend dev should know about this project: architecture, APIs, database, config, and how to extend it.

---

## 1. Architecture at a glance

- **Monorepo**: one Go module, multiple services in `cmd/`.
- **Shared DB**: one PostgreSQL; each service connects with `DB_URL` and owns specific tables.
- **No API gateway**: the CLI calls each service by port (8081–8087).
- **HTTP JSON**: services expose REST-style endpoints; no OpenAPI (yet), no auth.

```
                    ┌─────────────┐
                    │     CLI     │  (terminal UI)
                    └──────┬──────┘
                           │ HTTP (localhost:8081–8087)
     ┌─────────────────────┼─────────────────────┐
     │                     │                     │
     ▼                     ▼                     ▼
┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ...
│  user   │  │ rental  │  │ grocery │  │transport │  (7 services)
│ :8081   │  │ :8082   │  │ :8083   │  │ :8084   │
└────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘
     │            │            │            │
     └────────────┴────────────┴────────────┴──────────► PostgreSQL (5433)
```

---

## 2. Repo layout

```
artha/
├── go.mod, go.sum           # Module rent-cost-analyzer, deps (e.g. lib/pq)
├── Dockerfile               # Multi-stage: build all binaries, run one per container
├── docker-compose.yml       # postgres + 7 services
├── Makefile                 # setup, build, run, run-all, db-*, clean
├── setup.sh                 # Start postgres, go mod download
│
├── cmd/                     # All runnables (one main per dir)
│   ├── cli/                 # CLI client (calls services via HTTP)
│   ├── user-service/
│   ├── rental-service/
│   ├── grocery-service/
│   ├── transport-service/
│   ├── inflation-service/
│   ├── geospatial-service/
│   └── cost-prediction-service/
│
├── pkg/                     # Shared, importable by any service
│   └── models/
│       └── types.go        # UserProfile, RentalListing, CostAnalysis, GroceryItem, etc.
│
├── internal/                 # Private to this module
│   └── db/
│       └── conn.go         # DB_URL / default conn string, db.Open()
│
└── docs/
    └── BACKEND.md           # This file
```

**Conventions**

- **`cmd/<name>/main.go`**: one service or app; minimal logic, wire handlers and start server.
- **`pkg/models`**: DTOs and shared structs; used by services and CLI (for request/response).
- **`internal/db`**: DB connection only; no table definitions (those live in each service).

---

## 3. Service ports and ownership

| Port  | Service              | Owns DB tables           | Depends on        |
|-------|----------------------|--------------------------|-------------------|
| 8081  | user-service         | `users`                  | postgres          |
| 8082  | rental-service       | `rental_listings`        | postgres          |
| 8083  | grocery-service      | `groceries`              | postgres          |
| 8084  | transport-service    | `transport_routes`       | postgres          |
| 8085  | inflation-service    | `inflation_data`         | postgres          |
| 8086  | geospatial-service   | (none; reads `rental_listings`) | postgres, rental |
| 8087  | cost-prediction-service | (none; stateless)     | none              |

**Important**: `rental_listings` is created and seeded by **rental-service**. Geospatial only reads it, so start rental before (or with) geospatial.

---

## 4. API reference (what the CLI uses)

Base URL for local: `http://localhost:PORT`. All JSON request/response unless noted.

### User service (8081)

| Method | Path    | Description        | Body / Params | Response |
|--------|---------|--------------------|---------------|----------|
| GET    | /profile | Get current profile (id=1) | — | 200 UserProfile or 404 `{"error":"no profile"}` |
| POST   | /profile | Create/update profile (upsert id=1) | JSON: name, income, family_size, preferred_locale, commute_distance | 201 UserProfile |
| GET    | /health | Liveness           | — | 200 |

### Rental service (8082)

| Method | Path               | Description           | Params | Response |
|--------|--------------------|-----------------------|--------|----------|
| GET    | /listings          | Top 10 listings by rent | — | `{ "listings": [ RentalListing, ... ] }` |
| GET    | /listings/summary   | Count fair vs overpriced | — | `{ "fair": N, "overpriced": N }` |
| GET    | /compare           | Compare two localities | `loc1`, `loc2` | `{ "locality1", "locality2", "analysis1", "analysis2" }` (CostAnalysis each) |
| GET    | /cost-burden       | Burden % by locality   | `income` (required) | `{ "income", "localities": [ { locality, avg_rent, total, burden_pct } ] }` |
| GET    | /health            | Liveness               | — | 200 |

### Grocery service (8083)

| Method | Path   | Description        | Response |
|--------|--------|--------------------|----------|
| GET    | /items | All items + totals | `{ "items": [ { item, price, source } ], "total_basket", "monthly_estimate" }` |
| GET    | /health | Liveness         | 200 |

### Transport service (8084)

| Method | Path     | Description        | Params   | Response |
|--------|----------|--------------------|----------|----------|
| GET    | /route   | Route from→to      | `from`, `to` | `{ "found", "route?", "daily_cost?", "monthly_cost?" }` |
| GET    | /isochrone | Destinations from a locality | `from` | `{ "from", "destinations": [ { to_locality, distance_km, fare, time_zone } ] }` |
| GET    | /health  | Liveness           | —        | 200 |

### Inflation service (8085)

| Method | Path    | Description     | Response |
|--------|---------|-----------------|----------|
| GET    | /data   | All inflation rows | `{ "data": [ { month, category, rate } ] }` |
| GET    | /summary | Avg overall + trend | `{ "average_overall_inflation", "trend" }` |
| GET    | /health | Liveness        | 200 |

### Geospatial service (8086)

| Method | Path    | Description     | Params   | Response |
|--------|---------|-----------------|----------|----------|
| GET    | /heatmap | Rent by locality (for heatmap) | — | `{ "localities": [ { locality, avg_rent, count, intensity } ] }` |
| GET    | /nearby  | Localities “near” one (by distance) | `locality` | `{ "center", "nearby": [ { locality, distance_km, lat, lon } ] }` |
| GET    | /health  | Liveness        | —        | 200 |

### Cost-prediction service (8087)

| Method | Path    | Description     | Body    | Response |
|--------|---------|-----------------|---------|----------|
| POST   | /predict | Predict monthly costs | JSON: UserProfile (name, income, family_size, preferred_locale, commute_distance) | `{ user, income, rent, groceries, transport, total, cost_burden, confidence, feature_importance }` |
| GET    | /health | Liveness        | —       | 200 |

---

## 5. Database

- **Engine**: PostgreSQL 15.
- **Local port**: **5433** (host) → 5432 (container).
- **Connection**: `DB_URL` env var, or default in `internal/db`:  
  `host=localhost port=5433 user=postgres password=postgres dbname=rentanalyzer sslmode=disable`  
  In Docker, services use `host=postgres port=5432 ...`.

### Schema (by service)

**user-service** — `users`

- `id` INTEGER PRIMARY KEY DEFAULT 1 (single global profile)
- `name`, `income`, `family_size`, `preferred_locale`, `commute_distance`

**rental-service** — `rental_listings`

- `id` SERIAL, `locality`, `rent`, `bedrooms`, `sqft`, `classification`, `distance`, `lat`, `lon`

**grocery-service** — `groceries`

- `id` SERIAL, `item`, `price`, `source`

**transport-service** — `transport_routes`

- `id` SERIAL, `from_locality`, `to_locality`, `distance`, `fare`

**inflation-service** — `inflation_data`

- `id` SERIAL, `month`, `rate`, `category`

Tables are created in each service’s `main` on startup (`CREATE TABLE IF NOT EXISTS ...`). Seed logic runs once (e.g. “if count == 0 then insert mock data”). There are no migrations; schema changes = code change + redeploy.

---

## 6. Configuration

| Variable        | Used by        | Purpose |
|-----------------|----------------|---------|
| `DB_URL`        | All DB-using services | PostgreSQL connection string (required in Docker) |
| `SERVICES_HOST` | CLI only       | Host for 8081–8087 (default `localhost`) |

Ports are fixed in code (8081–8087). To change them you’d update each `ListenAndServe` and the CLI’s `baseURL(port)` (and optionally env).

---

## 7. Running and debugging

- **Full stack**: `make run-all` (Docker Compose), then `make run` (CLI).
- **Postgres only**: `make db-start` (or `docker-compose up -d postgres`).
- **Local binaries**: `make build` → `./bin/<service-name>`. Run each in a terminal or background; ensure postgres is up and `DB_URL` points to it (e.g. `host=localhost port=5433 ...`).
- **Logs**: `docker-compose logs -f <service>` or stdout of each binary.
- **Health**: every service has `GET /health` → 200. Use for readiness in Docker/Kubernetes later.

No debugger config in repo; run services with `go run ./cmd/<service>` and use Delve or breakpoints as usual.

---

## 8. Adding or changing behavior

**New endpoint in an existing service**

1. In `cmd/<service>/main.go`, add `http.HandleFunc("/path", handlePath)`.
2. Implement `handlePath`: parse query/body, use `conn` (DB) if needed, `json.NewEncoder(w).Encode(...)`, set `Content-Type: application/json` and status codes.
3. If the CLI should use it, add an HTTP call in `cmd/cli/main.go` and wire it to a menu option.

**New shared type**

- Add the struct in `pkg/models/types.go` with JSON tags. Use in service handlers and CLI.

**New service**

1. Add `cmd/<new-service>/main.go` (DB init, seed if needed, handlers, `ListenAndServe(":808X")`).
2. Add the binary to `Dockerfile` and a service in `docker-compose.yml` with `DB_URL` and `depends_on: postgres`.
3. In `Makefile` add a build line and, if you want, a run-all target or doc.
4. From CLI, call `http://localhost:808X/...` (and optionally introduce a `SERVICES_HOST`-style base URL for the new port).

**Changing schema**

- Change the `CREATE TABLE` (and any seed) in the owning service. For existing data you’d add one-off migration logic or manual SQL; this project currently relies on “empty DB” or reseeding.

---

## 9. Dependencies and patterns

- **go.mod**: `github.com/lib/pq` for Postgres. No router (stdlib `net/http`), no config library.
- **Errors**: services use `http.Error(w, err.Error(), 4xx/5xx)` or return a JSON `{"error": "..."}`. CLI prints "❌ Error: ..." and returns from the handler.
- **IDs**: user profile is a single row `id=1`. Others use SERIAL.
- **Concurrency**: one handler per request; no global locks. `sql.DB` is safe for concurrent use.

---

## 10. Quick checklist for a new backend dev

1. Run `make run-all` then `make run` and click through the CLI menu to see which service backs which feature.
2. Read `pkg/models/types.go` and `internal/db/conn.go`.
3. Skim one DB-backed service (e.g. `cmd/rental-service/main.go`) and one stateless one (`cmd/cost-prediction-service/main.go`).
4. Add a trivial `GET /ping` (or use `/health`) and call it from the CLI or `curl`.
5. Change one response shape in a service and update the CLI to match.

Once you’re comfortable with that, you can add endpoints, new services, or new fields in `pkg/models` and the DB as needed.
