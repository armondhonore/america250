# Nexlayer — america250

<!-- nexlayer:meta version=1 analyzed=2026-07-04T20:19:29Z repo=https://github.com/armondhonore/america250 branch=main -->

> **For AI agents (Claude Code, Cursor, Gemini CLI, Copilot):**
> This file is the **project context** for this Nexlayer deployment — tech stack, env vars, secrets, live URL.
> For full platform detail (nexlayer.yaml schema, Dockerfile rules, CI/CD, task recipes) read **`nexlayer.skills`** in this repo.
>
> **Critical rules (full detail in `nexlayer.skills`):**
> - Inter-pod refs: `${podName:port}` only — never `localhost` or bare hostnames
> - Docker Hub images: prefix with `mirror.gcr.io/library/` — bare tags fail on the cluster
> - Secrets: set in the Nexlayer dashboard — never commit to `nexlayer.yaml` or Dockerfile
>
> **This file:** `agent-managed` sections update automatically. `user-editable` sections (Local Development Setup, Nexlayer Deployment Plan, Build Notes) are yours — preserved across re-analysis.

## Project Summary
<!-- nexlayer:section agent-managed=project_summary -->
A Go-based web application for America250, utilizing a PostgreSQL database for persistence and serving static content and templates.
<!-- nexlayer:end -->

## Technology Stack
<!-- nexlayer:section agent-managed=tech_stack -->
| Name | Kind | Version | Detected From |
|------|------|---------|---------------|
| Go | language | 1.26.4 | go.mod |
| PostgreSQL | database | latest | go.mod |
| Alpine Linux | infra | 3.20 | Dockerfile |
<!-- nexlayer:end -->

## Repository Structure
<!-- nexlayer:section agent-managed=structure_map -->
- handlers/ — HTTP request handler logic
- static/ — Static assets (CSS, JS, Images)
- templates/ — HTML templates for server-side rendering
- seed/ — Database seeding scripts
- main.go — Application entry point
<!-- nexlayer:end -->

## External Services Required
<!-- nexlayer:section agent-managed=external_deps -->
_No external services detected._
<!-- nexlayer:end -->

## Local Development Setup
<!-- nexlayer:section user-editable=local_setup -->
### Prerequisites

- Go >= 1.23
- PostgreSQL

### Environment variables

Copy `.env.example` to `.env.local` and fill in:

```
DATABASE_URL=postgresql://user:pass@localhost:5432/america250
```

### Steps

1. `go mod download` — Download Go dependencies
2. `go run main.go` — Start the server on http://localhost:8080

<!-- nexlayer:end -->

## Nexlayer Setup
<!-- nexlayer:section agent-managed=nexlayer_setup -->
### Pod Environment Variables

| Pod | Variable | Value | Kind |
|-----|----------|-------|------|
| `app` | `PORT` | `"8080"` | plain |
| `app` | `DATABASE_URL` | `"postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres.pod:5432/${POSTGRES_DB}?sslmode=disable"` | inter-pod |
| `postgres` | `POSTGRES_USER` | `america_user` | plain |
| `postgres` | `POSTGRES_PASSWORD` | `${POSTGRES_PASSWORD}` | inter-pod |
| `postgres` | `POSTGRES_DB` | `america250` | plain |
| `america250-postgres-data` | `size` | `10Gi` | plain |
| `america250-postgres-data` | `mountPath` | `/var/lib/postgresql` | plain |

### nexlayer.yaml

```yaml
application:
  name: america250
  pods:
    - name: app
      image: "registry.nexlayer.io/user_01kece1xyh817dwff7wnarhkxd/america250:19f2ee545ca"
      path: /
      servicePorts:
        - 8080
      vars:
        PORT: "8080"
        DATABASE_URL: "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres.pod:5432/${POSTGRES_DB}?sslmode=disable"
    - name: postgres
      image: mirror.gcr.io/library/postgres:16-alpine
      servicePorts:
        - 5432
      vars:
        POSTGRES_USER: america_user
        POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
        POSTGRES_DB: america250
      volumes:
        - name: america250-postgres-data
          size: 10Gi
          mountPath: /var/lib/postgresql
```
<!-- nexlayer:end -->

## Nexlayer Deployment Plan
<!-- nexlayer:section user-editable=deployment_plan -->
### Pod Topology

| Pod | Image | Port | Role |
|-----|-------|------|------|
| app | mirror.gcr.io/library/alpine:3.19 | 8080 | web |
| db | mirror.gcr.io/library/postgres:16-alpine | 5432 | database |

### Deployment notes

- The application pod communicates with the database pod using the address db.pod:5432 per Nexlayer platform rules.
- The Go binary is built using a multi-stage Dockerfile to keep the final image lightweight.

<!-- nexlayer:end -->

## Build Notes
<!-- nexlayer:section user-editable=build_notes -->
<!-- Add notes for future builds here — preserved across re-analysis -->
<!-- nexlayer:end -->

## Nexlayer Configuration
<!-- nexlayer:section agent-managed=nexlayer_config -->
**Last deployed:** 2026-07-05T15:36:04Z  
**Live URL:** https://relaxed-weasel-america250.cloud.nexlayer.ai  
**Runtime:**  · **Port:** auto-detected  
**Deploy branch:** nexlayer  

```yaml
application:
  name: america250
  pods:
    - name: app
      image: "registry.nexlayer.io/user_01kece1xyh817dwff7wnarhkxd/america250:19f2ee545ca"
      path: /
      servicePorts:
        - 8080
      vars:
        PORT: "8080"
        DATABASE_URL: "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres.pod:5432/${POSTGRES_DB}?sslmode=disable"
    - name: postgres
      image: mirror.gcr.io/library/postgres:16-alpine
      servicePorts:
        - 5432
      vars:
        POSTGRES_USER: america_user
        POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
        POSTGRES_DB: america250
      volumes:
        - name: america250-postgres-data
          size: 10Gi
          mountPath: /var/lib/postgresql
```
<!-- nexlayer:end -->

## Build History
<!-- nexlayer:section agent-managed=build_history -->
| Date | Status | Notes |
|------|--------|-------|
| 2026-07-05T15:35:04Z | analyzed | initial repo analysis |
| 2026-07-05T15:36:04Z | success | deployed https://relaxed-weasel-america250.cloud.nexlayer.ai |
<!-- nexlayer:end -->



