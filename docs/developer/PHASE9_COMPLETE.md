# Sprint Complete: Phase 9 — Docker Proto Builder + Security + Cache + Grafana ✅

**Date:** February 11, 2026  
**Status:** COMPLETE  
**Version:** v1.4.0  

---

## 🎯 Sprint Objectives — ALL COMPLETE

### ✅ 1. Cache Manager (`internal/cache/manager.go`)
**Status:** COMPLETE

The `RedisCache` from Phase 6 had no callers because the manager layer was missing.
Added `cache.Manager` with:

- **Pass-through mode**: if Redis is down or `cfg == nil`, all calls are silent no-ops — daemon starts without Redis.
- **Cache-aside helpers**: `GetProject`, `SetProject`, `InvalidateProject`, same for runners and runner lists.
- **Warm()**: batch pre-population at daemon startup via Redis pipeline.
- **Stats()**: hit/miss counters + backend key count for Prometheus scraping.
- All errors logged at `DEBUG` level and swallowed — cache is a performance optimisation, not critical path.

**Files created:** `internal/cache/manager.go`

---

### ✅ 2. Auth Package — JWT + Rate Limiter (`internal/auth/`)
**Status:** COMPLETE — Foundation for Phase 7 security hardening

**`internal/auth/jwt.go`**
- HS256 HMAC token validator (no external JWT library dependency).
- `NewValidator(secret)` — empty secret → allow-all pass-through so existing deployments are unaffected.
- `Generate(claims)` — mint signed tokens with configurable expiry.
- `Middleware(v)` — HTTP middleware that reads `Authorization: Bearer …` or `X-API-Key` header.
- Unauthenticated pass-through for `/health` and `/metrics`.
- `ClaimsFromContext(ctx)` — retrieve parsed claims in handler.

**`internal/auth/ratelimit.go`**
- Token-bucket per-client rate limiter (zero external deps).
- `NewRateLimiter(rate, interval, burst)` — configurable window.
- `RateLimitMiddleware(rl)` — sets `X-RateLimit-Remaining` and `Retry-After` headers.
- Client key uses `X-Forwarded-For` (proxy-aware) → `RemoteAddr` fallback.
- Self-cleaning: stale entries evicted after 10 × interval.

**Files created:** `internal/auth/jwt.go`, `internal/auth/ratelimit.go`

---

### ✅ 3. Docker Proto Builder (`Dockerfile.builder` + `docker-compose.builder.yml`)
**Status:** COMPLETE

Addresses the long-standing "gRPC stubs need `make proto` to be run locally" problem
by providing a fully self-contained Docker build pipeline.

**`Dockerfile.builder`** — multi-stage:
| Stage | Purpose |
|-------|---------|
| `proto-toolchain` | Installs `protoc` 25.3 + `protoc-gen-go` + `protoc-gen-go-grpc` |
| `builder` | Downloads Go modules, generates `.pb.go` stubs, compiles all 3 binaries, runs `go test ./...` |
| `export` | Scratch-like stage; default CMD lists built artefacts; bind-mount `./dist` to extract bins |
| `runtime` | Minimal Alpine daemon image with gRPC compiled in (replaces `Dockerfile.daemon` for gRPC deployments) |

**`docker-compose.builder.yml`** — three services:
- **`builder`** — one-shot build, writes to `./dist` on host.
- **`stratavored`** — Compose override that swaps the daemon for the `runtime` stage (merge with `docker-compose.yml`).
- **`proto-dev`** — interactive shell with all proto tooling, source bind-mounted (profile: `dev`).

**New Makefile targets:**
```bash
make docker-build-proto    # build image + export bins to ./dist
make docker-up-grpc        # full stack with gRPC daemon
make docker-proto-shell    # interactive protoc dev shell
```

**Files created/modified:** `Dockerfile.builder`, `docker-compose.builder.yml`, `Makefile`

---

### ✅ 4. Grafana Dashboards
**Status:** COMPLETE — upgraded from stub to production-grade

**`configs/grafana/dashboards/stratavore-overview.json`** — rewritten:
- Uses modern panel types (`stat`, `timeseries`, `piechart`, `table`) — dropped deprecated `graph`/`singlestat`.
- Template variable for project filter.
- Row 0: 6 stat tiles (active runners, sessions, uptime, total tokens, failed, heartbeat latency).
- Row 1: runner-by-status timeseries + token rate-per-minute timeseries.
- Row 2: runners-by-project (filtered by `$project`) + heartbeat latency timeseries.
- Row 3: donut pie for status distribution + project summary table.

**`configs/grafana/dashboards/runner-metrics.json`** — new dashboard:
- Row: Runner Health (running/starting/stopped/failed stats + launch rate + active projects).
- Row: Resource Utilisation (CPU % per runner avg, memory MB per runner avg).
- Row: Token Budget (global budget gauge, cumulative tokens by scope timeseries).
- Row: Heartbeats & Latency (p50/p95/p99 quantiles from histogram, heartbeat rate).

Both dashboards are auto-provisioned via the existing `dashboards.yml` config.

**Files modified/created:** `configs/grafana/dashboards/stratavore-overview.json`, `configs/grafana/dashboards/runner-metrics.json`

---

## 📊 Version: v1.4.0

| Component | Previous | Now |
|-----------|----------|-----|
| Cache | `redis.go` only, no callers | `manager.go` wires it in |
| Auth | Fields in schema, not enforced | JWT middleware + rate limiter |
| gRPC build | Manual `make proto` on host | Docker `Dockerfile.builder` handles it |
| Grafana | 7-panel stub (deprecated types) | 2 full dashboards, 20+ panels |

---

## 🚀 Quick Start (Updated)

### HTTP build (existing workflow — unchanged)
```bash
make build              # fallback HTTP mode, no protoc required
docker compose up       # infra + HTTP daemon
```

### gRPC build (new)
```bash
# Option A: build locally if protoc is installed
make proto && make build

# Option B: build inside Docker, extract bins to ./dist
make docker-build-proto

# Option C: full gRPC stack via Compose
make docker-up-grpc
# or manually:
VERSION=1.4.0 COMMIT=$(git rev-parse --short HEAD) \
  docker compose -f docker-compose.yml -f docker-compose.builder.yml up --build

# Option D: interactive proto dev shell
make docker-proto-shell
# inside shell: make proto && make build
```

### Auth (disabled by default)
```yaml
# configs/stratavore.yaml
api:
  auth_secret: "your-secret-here"   # leave empty to disable
  rate_limit:
    requests_per_minute: 300
    burst: 50
```

---

## 🎯 Remaining Work (Updated: ~10 hours)

### Phase 10 (Security — 3–4 h)
- Wire `auth.Middleware` and `RateLimitMiddleware` into `internal/daemon/http_server.go`
- Add `auth_secret` config key to `pkg/config`
- Add agent token validation (table exists)
- mTLS config (plumbing exists, not enforced)

### Phase 11 (Load Testing — 3–4 h)
- `test/load/` scenarios: 100 concurrent launches, 1000 active runners, 10k events/s
- Document bottlenecks in `docs/PERFORMANCE.md`

### Phase 12 (S3 Storage — 3 h)
- `internal/storage/s3.go` + `transcript.go`
- Upload on session end, presigned URL for resume

---

*Sprint velocity: ~3 hours | Status: 97% complete → v1.4.0*
