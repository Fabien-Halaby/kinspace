# KinSpace

**A private, structured operating system for family life.**

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)](https://go.dev)
[![Expo](https://img.shields.io/badge/Expo-SDK%2054-000020?style=flat&logo=expo&logoColor=white)](https://expo.dev)
[![PostgreSQL](https://img.shields.io/badge/Postgres-17+-336791?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![Backend CI](https://img.shields.io/github/actions/workflow/status/Fabien-Halaby/kinspace/backend.yml?label=backend%20CI&logo=github)](https://github.com/Fabien-Halaby/kinspace/actions/workflows/backend.yml)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE.md)

KinSpace centralizes communication, organization and mutual support within a closed
family network. It is not a social network in the public sense — no ads, no algorithmic
feed, no public surface. Every family owns a private, invite-only space, structured
around how families actually work: relationships, roles and shared trust.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [API](#api)
- [Testing](#testing)
- [CI/CD](#cicd)
- [Roadmap](#roadmap)
- [License](#license)

---

## Overview

Most families coordinate their lives across a patchwork of tools never designed for the
job: chat apps for conversation, cloud drives for documents, social networks for photos,
spreadsheets for everything else. Information gets fragmented and context gets lost.

KinSpace replaces that patchwork with a single, private, structured space: a family feed,
a relationship graph, shared member profiles and — in later phases — shared documents,
scheduling and a mutual-aid network built on the skills already present within the family.

## Features

| | |
|---|---|
| **Authentication** | Register / login with bcrypt-hashed passwords and signed JWTs. |
| **Private family spaces** | Create a family and invite members via a shareable invite code. |
| **Member profiles** | Name, role, profession, skills and bio for each member. |
| **Family tree** | Typed, bidirectional relationship graph (`parent`, `child`, `spouse`, `sibling`), tenant-scoped per family. |
| **Tenant isolation** | Every query is scoped to the caller's `family_id` server-side. |
| **Family calendar** *(planned)* | Events and reminders, including relationship-derived events such as birthdays. |
| **Encrypted vault** *(planned)* | Encrypted storage for sensitive documents. |
| **Real-time messaging** *(planned)* | Chat between members. |

## Tech Stack

| Component | Choice |
|---|---|
| Backend | Go 1.26, [Gin](https://gin-gonic.com/), [pgx](https://github.com/jackc/pgx) (PostgreSQL driver) |
| API | REST, versioned under `/api/v1`, JSON with unified error contract |
| Auth | [golang-jwt/jwt](https://github.com/golang-jwt/jwt) HS256, bcrypt |
| Database | PostgreSQL 17, versioned migrations embedded in the binary |
| Logging | Structured JSON via `log/slog` with request correlation ids |
| Mobile frontend | React Native ([Expo SDK 54](https://expo.dev)), TypeScript |
| Infrastructure | Docker / Docker Compose (local), GitHub Actions (CI) |

## Architecture

The backend follows a layered (hexagonal) architecture. Dependency flow points inward:
transport → application → domain. The domain layer has **zero dependencies** on Go
frameworks or the database, which keeps the core logic pure, portable and easy to test.

```mermaid
flowchart TD
    subgraph Mobile["Mobile App (Expo)"]
        M[Features: auth, family, home]
        M --> C[API client]
    end

    subgraph API["KinSpace backend"]
        direction LR
        H["httpapi (transport)<br/>Gin handlers, middleware, error mapping"]
        A["application (use cases)<br/>services, validation, tokens"]
        D["domain (core)<br/>entities, ports, rules"]
        S["storage/postgres (adapters)<br/>repositories, migrations"]
        T["token<br/>JWT issuance/verification"]
    end

    C -->|REST /api/v1| H
    H --> A --> D
    A --> S
    A --> T
    S --> DB[(PostgreSQL)]
```

Layering rules enforced by the codebase:

- `internal/domain` — entities and **ports** (repository interfaces). No imports outside the
  standard library.
- `internal/application` — use cases and business rules. Depends on domain ports and the
  token interface only.
- `internal/httpapi` — HTTP concerns: handlers, middleware, request/response DTOs, and the
  **single** place where domain errors are mapped to HTTP status codes.
- `internal/storage/postgres` — persistence adapters and the versioned migration runner.
- `cmd/api` — the composition root: wires configuration, storage, services and the HTTP
  server, and owns the process lifecycle (graceful shutdown).

Every request is scoped to a `family_id`; the backend enforces strict tenant isolation so
one family's data is never reachable through another family's session.

## Getting Started

### Prerequisites

- Go 1.26+
- Docker (for the local PostgreSQL) — or any PostgreSQL 15+
- Node.js 20+ for the mobile app

### 1. Start PostgreSQL

```bash
docker compose up -d postgres
```

### 2. Backend

```bash
cd backend
cp .env.example .env    # configure DB URL and JWT secret
make run                # runs the API on :8080
```

Or, from the repo root: `make build`, `make test`, `make lint`.

### 3. Mobile app

```bash
cd mobile
npm install
npx expo start
```

The app derives the API base URL from the Expo dev-server host automatically, or you can
pin it with `EXPO_PUBLIC_API_URL`.

## Project Structure

```
kinspace/
├── backend/
│   ├── cmd/api/                  → composition root, process lifecycle
│   └── internal/
│       ├── domain/               → entities, ports, business rules (no deps)
│       ├── application/          → use cases (auth, family, relations)
│       ├── httpapi/              → Gin router, handlers, middleware, error mapping
│       ├── storage/postgres/     → pgx repositories + versioned migrations
│       ├── token/                → JWT issuance/verification (golang-jwt)
│       ├── config/               → env loading + validation
│       └── logger/               → structured slog setup
├── mobile/
│   └── app/ + src/               → Expo Router screens and feature code
├── .github/workflows/            → backend + mobile CI
├── Makefile                      → build / test / lint / run helpers
├── docker-compose.yml            → local PostgreSQL
└── .golangci.yml                 → lint configuration
```

## API

The REST API is versioned under `/api/v1`. Every protected endpoint requires
`Authorization: Bearer <jwt>`. Errors follow a single contract: `{"error": "message"}`.

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness probe |
| `POST` | `/api/v1/auth/register` | Create an account, returns `{user, token}` |
| `POST` | `/api/v1/auth/login` | Authenticate, returns `{user, token}` |
| `GET` | `/api/v1/auth/me` | Current user profile *(protected)* |
| `POST` | `/api/v1/families` | Create a family *(protected)* |
| `POST` | `/api/v1/families/join` | Join via invite code *(protected)* |
| `GET` | `/api/v1/families/me` | Current user's family *(protected)* |
| `POST` | `/api/v1/relations` | Create a relation edge *(protected)* |
| `GET` | `/api/v1/relations` | List the family graph *(protected)* |

## Testing

The test suite has two tiers:

- **Unit tests** — services (with fake ports), token manager, config, HTTP handlers
  (with `httptest`), run on every `make test`.
- **Integration tests** — the repository adapters and the migration runner against a real
  PostgreSQL instance. Gated behind `TEST_DATABASE_URL`:

```bash
TEST_DATABASE_URL=postgres://kinspace:spacekin@localhost:5433/kinspace_test?sslmode=disable \
  make test-integration
```

## CI/CD

GitHub Actions runs on every push and pull request:

- **Backend** — `gofmt` and `go mod tidy` checks, build, `go vet`, `golangci-lint`,
  unit tests with `-race`, and integration tests against a service-backed PostgreSQL.
- **Mobile** — TypeScript type-checking and `expo lint`.

## Roadmap

- [x] Authentication (register/login, JWT)
- [x] Family creation and invite flow
- [x] Member profiles
- [x] Family tree (relationship graph)
- [ ] Feed (text and image posts)
- [ ] Mutual aid network (skills → requests)
- [ ] Calendar and events
- [ ] Encrypted family vault
- [ ] Real-time messaging

## License

Licensed under the [MIT License](LICENSE.md).
