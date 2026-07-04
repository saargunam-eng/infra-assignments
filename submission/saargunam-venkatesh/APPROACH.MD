# Approach

> Pre-implementation overview. Details like schema, exact commands, and repo layout will be in the final PR and README.

## What I understood from the assignment

Build and deploy a small config management HTTP API in Go, backed by PostgreSQL, on a local Kubernetes cluster. The evaluation isn't just about getting the app running — it's about repeatable automation, clean infra structure, and production-style operational thinking.

---

## Architectural Decisions

- **Go + chi + pgx** — Layered into handler → service → repository → domain. Business logic stays testable without a live database.
- **Kind** — Lightweight local cluster, no extra daemon, fast to spin up and tear down.
- **Helm** — App packaged as a chart; Bitnami Postgres chart for the database rather than hand-rolling a StatefulSet.
- **Terraform** — Drives both Helm releases so the full environment is one `terraform apply`. Single source of truth for what's deployed.
- **GitHub Actions** — CI on every push: lint, test, `terraform validate`, `helm lint`.

---

## API (intended contract)

| Method | Path | Behaviour |
|--------|------|-----------|
| GET | `/ping` | Returns `pong` — used for liveness |
| GET | `/configs/:id` | Returns config or `404` |
| POST | `/configs` | Upsert — creates or overwrites |

---

## Key Operational Decisions

**Config & secrets** — Non-sensitive values via `ConfigMap` → env vars. DB password via `Secret` → `DATABASE_URL`. Nothing hardcoded. In production this would point to a secrets manager, but the app wouldn't change.

**Health checks** — `/ping` as both liveness and readiness probe. DB connection has a startup retry loop — pod restarts if Postgres is unreachable after retries, rather than silently failing.

**Observability** — Structured JSON logs (log/slog), `/metrics` endpoint (Prometheus-compatible). Startup logs surface DB connection status so failures are immediately visible.

**Reproducibility** — Everything bootstraps through a Makefile (`make up` / `make down`). No hidden manual steps. Documented in README.

---

## What I'm not building (and why)

- **HA Postgres** — Not needed for local dev. Known limitation, will note in the PR.
- **TLS** — Out of scope for local-only.
- **CD pipeline** — Bonus item, skipping to keep the core solid.

---

## Questions

Nothing blocking right now. Will flag any assumptions in the final PR as I go.
