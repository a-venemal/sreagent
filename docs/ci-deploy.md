# SREAgent — CI/CD & Deployment Guide

> Last updated: 2026-04-04

---

## Table of Contents

1. [CI Pipeline (GitHub Actions)](#ci-pipeline)
2. [Dockerfile (Multi-stage Build)](#dockerfile)
3. [Entrypoint & Startup Flow](#entrypoint)
4. [Configuration Variables](#configuration-variables)
5. [Kubernetes Deployment](#kubernetes-deployment)
6. [Local Development](#local-development)
7. [Build & Release Workflow](#build--release-workflow)

---

## CI Pipeline

**File**: `.github/workflows/docker-build.yml`

### Trigger Rules

| Event | Condition | Behavior |
|-------|-----------|----------|
| `push` | Branch `main` | Build + push `:latest` tag |
| `push` | Tag `v*` (e.g. `v1.2.3`) | Build + push `:v1.2.3`, `:1.2`, `:1`, `:latest` |
| `pull_request` | Target `main` | Build only (no push) — CI validation |

### Jobs

The pipeline runs 3 jobs:

#### Job 1: `test` — Go Unit Tests
```yaml
runs-on: ubuntu-latest
steps:
  - checkout
  - setup-go (version from go.mod, with cache)
  - go test ./... -timeout 120s
```
> Note: Currently passes vacuously (zero `*_test.go` files exist).

#### Job 2: `typecheck` — Frontend TypeScript Check
```yaml
runs-on: ubuntu-latest
steps:
  - checkout
  - setup-node 20 (cache: npm, key: web/package-lock.json)
  - npm ci (in web/)
  - npm run typecheck (in web/)
```

#### Job 3: `build-and-push` — Multi-arch Docker Image
```yaml
runs-on: ubuntu-latest
needs: [test, typecheck]   # runs after both jobs pass
steps:
  - checkout
  - setup QEMU (for arm64 cross-compilation)
  - setup Docker Buildx
  - login to Docker Hub (skip on PR)
  - docker/metadata-action → generate tags
  - docker/build-push-action:
      context: .
      file: deploy/docker/Dockerfile
      platforms: linux/amd64, linux/arm64
      push: true (false on PR)
      cache: GitHub Actions cache (GHA)
      build-args:
        BUILD_VERSION=${{ github.ref_name }}
        BUILD_COMMIT=${{ github.sha }}
  - (optional) update Docker Hub README
```

### Required GitHub Secrets

| Secret | Purpose |
|--------|---------|
| `DOCKERHUB_USERNAME` | Docker Hub login username |
| `DOCKERHUB_TOKEN` | Docker Hub access token |

### Image Naming

The image name is `${{ secrets.DOCKERHUB_USERNAME }}/sreagent`.

Tag examples:
- Push to `main` → `user/sreagent:latest`
- Push tag `v1.2.3` → `user/sreagent:v1.2.3`, `user/sreagent:1.2`, `user/sreagent:1`, `user/sreagent:latest`
- PR #42 → `user/sreagent:pr-42` (built but not pushed)

---

## Dockerfile

**File**: `deploy/docker/Dockerfile`

### Stages

| Stage | Base | Purpose |
|-------|------|---------|
| `backend` | `golang:1.24-alpine` | Compile Go binary with ldflags |
| `frontend` | `node:20-alpine` | Build Vue 3 SPA |
| `final` | `alpine:3.20` | Minimal runtime image |

### Build Args

| Arg | Default | Description |
|-----|---------|-------------|
| `BUILD_VERSION` | `dev` | Injected as `-X main.Version` via ldflags |
| `BUILD_COMMIT` | `unknown` | Injected as `-X main.Commit` via ldflags |

### What Goes Into the Final Image

```
/app/
├── sreagent              # Go binary (statically linked, CGO_ENABLED=0)
├── web/dist/             # Built Vue SPA assets
├── configs/config.yaml   # Copy of config.example.yaml
├── entrypoint.sh         # Startup script
└── logs/                 # Empty dir (app logs to stdout by default)
```

### Runtime Dependencies

Installed via `apk add --no-cache`:
- `ca-certificates` — TLS cert trust store
- `tzdata` — timezone support (Asia/Shanghai by default)
- `curl` — healthcheck probe
- `bash` — entrypoint script
- `mysql-client` — DB creation at startup

### Healthcheck

```dockerfile
HEALTHCHECK --interval=15s --timeout=3s --start-period=15s \
  CMD curl -f http://localhost:8080/healthz || exit 1
```

### Building Locally

```bash
# Basic build
make docker-build

# With version info
docker build \
  --build-arg BUILD_VERSION=v1.0.0 \
  --build-arg BUILD_COMMIT=$(git rev-parse HEAD) \
  -t sreagent:v1.0.0 \
  -f deploy/docker/Dockerfile .
```

---

## Entrypoint

**File**: `deploy/docker/entrypoint.sh`

### Startup Sequence

1. **Wait for MySQL** — TCP probe to `${DB_HOST}:${DB_PORT}`, max 60 retries (2s interval = 2 min timeout)
2. **Create database** — `CREATE DATABASE IF NOT EXISTS` via `mysql` CLI:
   - First try with application credentials (`SREAGENT_DATABASE_*`)
   - If that fails and `MYSQL_ROOT_PASSWORD` is set, retry with root
   - If both fail, assume DB already exists and continue
3. **Start server** — `exec ./sreagent --config configs/config.yaml`
   - golang-migrate runs automatically on startup (embedded SQL migrations)
   - Admin user `admin/admin123` is seeded if no users exist

### Entrypoint Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SREAGENT_DATABASE_HOST` | `127.0.0.1` | MySQL host |
| `SREAGENT_DATABASE_PORT` | `3306` | MySQL port |
| `SREAGENT_DATABASE_USERNAME` | `sreagent` | MySQL user |
| `SREAGENT_DATABASE_PASSWORD` | `sreagent` | MySQL password |
| `SREAGENT_DATABASE_DATABASE` | `sreagent` | Database name |
| `MYSQL_ROOT_PASSWORD` | (empty) | Optional root password for DB creation fallback |

---

## Configuration Variables

### Config File (`configs/config.yaml`)

All values can be overridden by environment variables with the `SREAGENT_` prefix using Viper's `AutomaticEnv()`. Dots become underscores: `database.host` → `SREAGENT_DATABASE_HOST`.

```yaml
server:
  host: "0.0.0.0"           # SREAGENT_SERVER_HOST
  port: 8080                 # SREAGENT_SERVER_PORT
  mode: "debug"              # SREAGENT_SERVER_MODE ("debug" | "release")

database:
  driver: "mysql"            # SREAGENT_DATABASE_DRIVER
  host: "127.0.0.1"         # SREAGENT_DATABASE_HOST
  port: 3306                 # SREAGENT_DATABASE_PORT
  username: "sreagent"       # SREAGENT_DATABASE_USERNAME
  password: "change-me"     # SREAGENT_DATABASE_PASSWORD
  database: "sreagent"       # SREAGENT_DATABASE_DATABASE
  charset: "utf8mb4"        # SREAGENT_DATABASE_CHARSET
  max_idle_conns: 10         # SREAGENT_DATABASE_MAX_IDLE_CONNS
  max_open_conns: 100        # SREAGENT_DATABASE_MAX_OPEN_CONNS
  max_lifetime: 3600         # SREAGENT_DATABASE_MAX_LIFETIME

redis:
  host: "127.0.0.1"         # SREAGENT_REDIS_HOST
  port: 6379                 # SREAGENT_REDIS_PORT
  password: ""               # SREAGENT_REDIS_PASSWORD
  db: 0                      # SREAGENT_REDIS_DB
  pool_size: 100             # SREAGENT_REDIS_POOL_SIZE

jwt:
  secret: "change-me"       # SREAGENT_JWT_SECRET
  expire: 86400              # SREAGENT_JWT_EXPIRE (seconds)
  issuer: "sreagent"         # SREAGENT_JWT_ISSUER

log:
  level: "info"              # SREAGENT_LOG_LEVEL
  format: "json"             # SREAGENT_LOG_FORMAT
  output: "stdout"           # SREAGENT_LOG_OUTPUT
  file: "logs/sreagent.log"  # SREAGENT_LOG_FILE

engine:
  enabled: true              # SREAGENT_ENGINE_ENABLED
  sync_interval: 30          # SREAGENT_ENGINE_SYNC_INTERVAL (seconds)
```

### Manually Read Environment Variables

These are read via `os.Getenv()` directly, NOT through Viper:

| Variable | Description | Example |
|----------|-------------|---------|
| `SREAGENT_SECRET_KEY` | AES-256-GCM master key for encrypting sensitive DB fields. 64 hex chars = 32 bytes. | `a1b2c3...` (64 chars) |
| `SREAGENT_DB_DEBUG` | Enable GORM SQL debug logging (`"true"` to enable) | `"false"` |
| `CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins | `"https://sreagent.example.com"` |

### AI & Lark Configuration

AI and Lark credentials are **NOT** in config files or environment variables. They are stored
encrypted (AES-256-GCM) in the `system_settings` database table and managed through
the Web UI at **Settings → AI Config** and **Settings → Lark Bot**.

---

## Kubernetes Deployment

### Directory Structure

```
deploy/kubernetes/
├── 00-namespace.yaml           # Namespace: sreagent
├── app/
│   ├── configmap.yaml          # Embedded config.yaml
│   ├── secret.yaml             # 4 keys: db-password, redis-password, jwt-secret, secret-key
│   ├── deployment.yaml         # 1 replica, rolling update, init containers
│   ├── service.yaml            # ClusterIP 80→8080
│   ├── ingress.yaml            # NGINX Ingress with TLS
│   └── hpa.yaml                # HPA min=1, max=3 (CPU 80%)
├── mysql/                      # MySQL 8.0 StatefulSet + configmap + secret
├── redis/                      # Redis 7 StatefulSet + secret
├── helm/                       # (empty — reserved for future)
└── kustomize/                  # (empty — reserved for future)
```

### Deployment Order

```bash
# 1. Create namespace
kubectl apply -f deploy/kubernetes/00-namespace.yaml

# 2. Deploy dependencies
kubectl apply -f deploy/kubernetes/mysql/
kubectl apply -f deploy/kubernetes/redis/

# 3. Wait for MySQL and Redis to be ready
kubectl -n sreagent wait --for=condition=ready pod -l app=mysql --timeout=120s
kubectl -n sreagent wait --for=condition=ready pod -l app=redis --timeout=60s

# 4. Create secrets (edit base64 values first!)
kubectl apply -f deploy/kubernetes/app/secret.yaml

# 5. Deploy application
kubectl apply -f deploy/kubernetes/app/configmap.yaml
kubectl apply -f deploy/kubernetes/app/deployment.yaml
kubectl apply -f deploy/kubernetes/app/service.yaml
kubectl apply -f deploy/kubernetes/app/ingress.yaml
kubectl apply -f deploy/kubernetes/app/hpa.yaml
```

### Secrets (Base64-encoded)

Edit `deploy/kubernetes/app/secret.yaml` before applying:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sreagent-secret
  namespace: sreagent
type: Opaque
data:
  db-password: <base64>        # → SREAGENT_DATABASE_PASSWORD
  redis-password: <base64>     # → SREAGENT_REDIS_PASSWORD
  jwt-secret: <base64>         # → SREAGENT_JWT_SECRET
  secret-key: <base64>         # → SREAGENT_SECRET_KEY (64 hex chars)
```

Generate base64 values:
```bash
echo -n 'your-db-password' | base64
echo -n 'your-redis-password' | base64
echo -n 'your-jwt-secret' | base64
echo -n 'your-64-char-hex-key' | base64
```

### Init Containers

The deployment includes 2 init containers:
1. `wait-for-mysql` — `busybox:1.36`, polls `nc -z mysql-svc 3306`
2. `wait-for-redis` — `busybox:1.36`, polls `nc -z redis 6379`

### Environment Variables Injected

```yaml
env:
  - name: SREAGENT_DATABASE_PASSWORD  # from secret: db-password
  - name: SREAGENT_REDIS_PASSWORD     # from secret: redis-password
  - name: SREAGENT_JWT_SECRET         # from secret: jwt-secret
  - name: SREAGENT_SECRET_KEY         # from secret: secret-key
  - name: TZ                          # "Asia/Shanghai"
  - name: SREAGENT_DB_DEBUG           # "false"
  - name: CORS_ALLOWED_ORIGINS        # "https://sreagent.example.com"
```

### Resource Limits

```yaml
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 512Mi
```

### Probes

| Probe | Path | Initial Delay | Period |
|-------|------|---------------|--------|
| Liveness | `GET /healthz` | 20s | 15s |
| Readiness | `GET /healthz` | 10s | 5s |

### Scaling

- Default: 1 replica (single instance — alert engine state machine is in-memory)
- HPA: min=1, max=3, target CPU=80%
- **Caution**: Multi-replica requires distributed locking for the alert engine (Phase 2)

### Rolling Update Strategy

```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxUnavailable: 0   # Zero downtime
    maxSurge: 1
```

### Triggering a Rolling Restart on ConfigMap Change

Since we don't use Helm templating, manually update the annotation:
```bash
kubectl -n sreagent annotate deployment sreagent --overwrite \
  checksum/config=$(sha256sum deploy/kubernetes/app/configmap.yaml | cut -d' ' -f1)
```

### Updating the Image

```bash
# After CI pushes a new tag
kubectl -n sreagent set image deployment/sreagent \
  sreagent=ghcr.io/sreagent/sreagent:v1.2.3
```

---

## Local Development

### Prerequisites

- Go 1.24+
- Node 20+
- MySQL 8.0 (or use Docker)
- Redis 7 (or use Docker)
- (Optional) `air` for hot reload: `go install github.com/air-verse/air@latest`
- (Optional) `golangci-lint` for linting

### Quick Start

```bash
# 1. Start dependencies
make docker-up     # Starts MySQL + Redis in Docker containers

# 2. Copy config
cp configs/config.example.yaml configs/config.yaml
# Edit configs/config.yaml with your local settings

# 3. Run backend (with hot reload)
make dev           # Uses air for hot reload
# OR
make run           # Build and run once

# 4. Run frontend (separate terminal)
make web-install   # First time only
make web-dev       # Starts Vite dev server with HMR
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make help` | Show all targets with descriptions |
| `make build` | Build Go binary to `bin/sreagent` |
| `make run` | Build and run the server |
| `make dev` | Run with hot reload (requires `air`) |
| `make test` | Run Go tests with coverage |
| `make lint` | Run `golangci-lint` |
| `make fmt` | Format Go code (`go fmt` + `goimports`) |
| `make tidy` | `go mod tidy` |
| `make web-install` | Install frontend npm dependencies |
| `make web-dev` | Start frontend dev server |
| `make web-build` | Build frontend for production |
| `make docker-up` | Start MySQL + Redis containers |
| `make docker-down` | Stop MySQL + Redis containers |
| `make docker-build` | Build Docker image locally |
| `make db-migrate` | Run database migrations (builds and runs the binary) |
| `make clean` | Remove `bin/`, `web/dist/`, `web/node_modules/` |
| `make all` | `tidy` + `fmt` + `build` + `web-build` |

### Default Admin Credentials

On first startup with an empty database, the server seeds:
- Username: `admin`
- Password: `admin123`

**Change this immediately in production.**

---

## Build & Release Workflow

### Development Flow

1. Create feature branch from `main`
2. Develop and test locally (`make dev` + `make web-dev`)
3. Push branch → create PR → CI runs `test` + `typecheck` + build (no push)
4. Review and merge to `main` → CI builds + pushes `:latest`

### Release Flow

1. Ensure `main` is stable
2. Create and push a semver tag:
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```
3. CI builds + pushes: `:v1.2.3`, `:1.2`, `:1`, `:latest`
4. Update K8s deployment:
   ```bash
   kubectl -n sreagent set image deployment/sreagent \
     sreagent=ghcr.io/sreagent/sreagent:v1.2.3
   ```

### Rollback

```bash
# Check deployment history
kubectl -n sreagent rollout history deployment/sreagent

# Rollback to previous revision
kubectl -n sreagent rollout undo deployment/sreagent

# Or rollback to specific revision
kubectl -n sreagent rollout undo deployment/sreagent --to-revision=3
```
