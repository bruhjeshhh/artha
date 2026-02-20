# Walkthrough script — Rent & Cost Analyzer (intermediate backend)

A script you can read or use as talking points to explain this project to someone with intermediate backend experience (comfortable with HTTP APIs, SQL, Go or similar, and basic microservice ideas).

---

## Intro (30 sec)

"This is a Rent & Cost Analyzer for a specific region — think rent listings, groceries, transport, inflation, and cost predictions. The important part is how it’s built: we took a single monolith and split it into several small services, plus a CLI that talks to them over HTTP. So you get a clear example of a microservice-style backend in one repo, with a shared database and no gateway — just the CLI calling each service by port. I’ll walk through the architecture, how the pieces talk to each other, and how you’d run or change things."

---

## 1. High-level architecture (1–2 min)

"The app has two kinds of components: the CLI and the backend services.

The CLI is the only 'front end' — it’s a terminal UI. It doesn’t have a database or business logic; it just prompts the user, then does HTTP calls to the right service and prints the result. So from a backend perspective, the CLI is just another HTTP client.

On the other side we have seven services, each listening on a fixed port from 8081 to 8087. User is 8081, rental 8082, grocery 8083, transport 8084, inflation 8085, geospatial 8086, cost-prediction 8087. There’s no API gateway and no service mesh — the CLI knows each port and calls them directly. So the architecture is: one client, many services, and they all share a single PostgreSQL instance. We’re not doing database-per-service here; we’re doing table ownership. Each service owns one or two tables, and only that service writes to them. Geospatial is the odd one out — it doesn’t own any table; it only reads rental_listings, which rental-service owns. So we get a simple split of responsibilities without splitting the database."

---

## 2. Repo layout and conventions (1 min)

"The repo is a Go monorepo: one module, multiple runnables. Everything that runs lives under `cmd/`. So you have `cmd/cli`, `cmd/user-service`, `cmd/rental-service`, and so on. Each of those has a `main.go` that starts an HTTP server or, in the CLI’s case, a loop that calls those servers. That’s the standard Go layout: one main per directory under `cmd/`.

Shared code lives in `pkg` and `internal`. We use `pkg/models` for DTOs — structs like UserProfile, RentalListing, CostAnalysis — that multiple services and the CLI need. So when we say 'the rental service returns a list of RentalListing,' that type is defined once in `pkg/models/types.go`. The `internal` folder is for code we don’t want to expose outside the module; here we only have `internal/db`, which wraps the PostgreSQL connection string — either from the `DB_URL` env var or a default for local dev. So: runnables in `cmd/`, shared types in `pkg/models`, and DB connection helper in `internal/db`. No shared 'repository' or 'service' layer; each service does its own DB access."

---

## 3. Services and what they own (2 min)

"Let’s go service by service so you know who does what.

**User service, port 8081.** It owns the `users` table. In this demo we only have one profile, so we use a single row with id equals 1. You create or update it with POST to slash profile, and you read it with GET. That’s it. The CLI uses this to get the current user when it needs income or preferred locality for predictions or transport.

**Rental service, 8082.** It owns `rental_listings` — id, locality, rent, bedrooms, sqft, classification like 'fair' or 'overpriced,' distance, lat, lon. On first run it seeds mock listings. It exposes GET list listings, GET listing summary — fair vs overpriced counts — GET compare with two locality names, and GET cost-burden with an income query param. So rental is the place for anything about listings and locality-level cost.

**Grocery service, 8083.** Owns `groceries` — item, price, source. Seeds a small set of items. One endpoint: GET items, which returns the list plus a monthly estimate. Simple read-only style API.

**Transport service, 8084.** Owns `transport_routes`: from_locality, to_locality, distance, fare. GET route with from and to gives you that route and derived daily and monthly cost. GET isochrone with from gives you destinations from that locality with distance and a simple time zone label. So transport is all about routes and fares.

**Inflation service, 8085.** Owns `inflation_data` — month, rate, category. GET data returns all rows; GET summary returns average overall inflation and a short trend string. Again, read-oriented.

**Geospatial service, 8086.** This one doesn’t create tables. It only reads `rental_listings`, which rental-service created. GET heatmap returns locality-level average rent and an intensity value for visualization. GET nearby takes a locality and returns other localities with distance and coordinates. So geospatial is a read-only consumer of rental data.

**Cost-prediction service, 8087.** Stateless. No database. It has one endpoint: POST predict with a JSON body that looks like a user profile — name, income, family size, commute distance, etc. It runs a simple formula — you could think of it as a stand-in for an XGBoost model — and returns predicted rent, groceries, transport, total, cost burden, and a confidence number. So this service is pure compute; no persistence."

---

## 4. How the CLI uses the APIs (1 min)

"The CLI is just a big menu. Each menu option maps to one or more HTTP calls. For example: 'Create user profile' — it reads name, income, family size, locality, commute from stdin, then POSTs that JSON to user-service slash profile. 'Analyze rent listings' — GET rental-service slash listings, then GET slash listings slash summary, and it formats the tables in the terminal. 'Cost prediction' — GET user-service slash profile to load the current user, then POST that profile to cost-prediction slash predict and prints the result. So the CLI is the orchestration layer. It doesn’t duplicate business logic; it just gathers input, calls the right service, and formats output. The base URL for services comes from env: SERVICES_HOST, default localhost, and the ports are fixed in code. So if you run the CLI against a different host, you set SERVICES_HOST and all seven ports are assumed to be on that host."

---

## 5. Database and startup (1 min)

"We use one PostgreSQL database. Locally it’s on port 5433 so it doesn’t clash with a local Postgres. In Docker, the containers talk to the postgres container on 5432; we map 5433 on the host to 5432 in the container so you can still connect from your machine if you want. Each service that has tables runs 'create table if not exists' and then a one-time seed — usually 'if count is zero, insert mock data.' So there are no migration files; schema is in code. If you add a column, you change the create table and the seed in that service and redeploy. For a production system you’d add proper migrations, but for this demo, startup creates and seeds everything. Order matters a bit: geospatial reads rental_listings, so rental-service should be up before or with geospatial. Docker Compose handles that with depends_on."

---

## 6. Running it (30 sec)

"To run the whole thing: run 'make run-all' to start Postgres and all seven services in Docker. Wait a few seconds, then run 'make run' to start the CLI. The CLI will hit localhost 8081 through 8087. That’s it. If you want to run services locally instead of Docker: start Postgres with 'make db-start', run 'make build' to get binaries in bin slash, then run each binary — user-service, rental-service, and so on — in separate terminals or in the background. Set DB_URL to point at your Postgres if it’s not the default. Then run the CLI with 'make run'. So you have two modes: everything in Docker, or Postgres in Docker and services as local binaries."

---

## 7. What to touch when you change things (1 min)

"If you’re adding a new endpoint to an existing service: add a handler in that service’s main, register it with HandleFunc, and if the CLI should use it, add the corresponding HTTP call and menu option in the CLI. If you’re adding a new shared type — a new DTO — put it in pkg slash models slash types dot go with JSON tags so both the service and the CLI can use it. If you’re adding a whole new service: create a new cmd slash something with its own main, give it a port — say 8088 — add it to the Dockerfile and docker-compose, and from the CLI call that port. The pattern is the same: HTTP, JSON, and either own a table or stay stateless. One more thing: config. Right now we only have DB_URL for services and SERVICES_HOST for the CLI. Ports are hardcoded. If you need different ports or more env, you’d add them in the same style — read from os.Getenv and fall back to a default."

---

## 8. Wrap-up (30 sec)

"So in summary: we have a monorepo with a CLI and seven small backend services. They share one Postgres; each service owns its tables and exposes a few REST-style endpoints. The CLI orchestrates by calling those endpoints and rendering the results in the terminal. No auth, no gateway — just a clear split of responsibilities and a good base to add auth, API docs, or more services later. If you open the repo, start with BACKEND.md for the full API and schema reference, and use this script as the narrative that ties it all together."

---

## Optional: One-sentence per service (for quick reference)

- **User:** Single profile (id=1); GET/POST `/profile`.
- **Rental:** Listings, summary, compare localities, cost burden; owns `rental_listings`.
- **Grocery:** Item list and monthly estimate; owns `groceries`.
- **Transport:** Route and fare, isochrone from a locality; owns `transport_routes`.
- **Inflation:** Inflation rows and summary; owns `inflation_data`.
- **Geospatial:** Heatmap and nearby localities; reads `rental_listings` only.
- **Cost-prediction:** Stateless; POST profile, get predicted costs.

End of script.
