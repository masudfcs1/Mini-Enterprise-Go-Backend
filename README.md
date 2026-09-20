# Mini Enterprise Go Backend (Chi + Prisma ORM + PostgreSQL)

A modular, scalable, layered REST API backend in Go inspired by clean enterprise architectural patterns.

---

## Highlights & Features

- **Modular Architecture**: Independent domain modules (`user`, `auth`) with distinct `route` ➔ `handler` ➔ `service` ➔ `repository` ➔ `prisma` layers.
- **Pino-Style Colorful Terminal Logger**: Built on `rs/zerolog` + `go-colorable`, featuring vibrant ANSI status codes (2xx green, 3xx cyan, 4xx yellow, 5xx red), method badges, latency metrics, and an ASCII startup banner.
- **Real JWT Authentication**: Cryptographically signed access tokens via `golang-jwt/jwt/v5` and secure password hashing using `golang.org/x/crypto/bcrypt`.
- **Protected Endpoints**: Reusable JWT Bearer authentication middleware (`pkg/middleware/auth.go`).
- **Rate Limiting**: Built on `go-chi/httprate` with global 100 req/min limit and strict 10 req/min protection on auth endpoints against brute-force attacks.
- **PgBouncer & Neon Pooler Ready**: Configured for transaction-pooler environments (`&pgbouncer=true`).

---

## Architecture Flow

```text
HTTP Request
     │
     ▼
Global Middlewares (Pino Logger, Recoverer, CORS, Rate Limiter)
     │
     ▼
Global Router (internal/router/router.go)
     │
     ▼
Module Routes (internal/user/route.go, internal/auth/route.go)
     │
     ▼
Handlers (internal/user/handler.go, internal/auth/handler.go)
     │
     ▼
Services (internal/user/service.go, internal/auth/service.go)
     │
     ▼
Repositories (internal/user/repository.go, internal/auth/repository.go)
     │
     ▼
Prisma ORM Client (internal/database/db)
     │
     ▼
PostgreSQL Database (Neon / Local)
```

---

## Directory Structure

```text
go-mini-setup/
│
├── cmd/
│   └── server/
│       └── main.go          # Dependency injection & HTTP server lifecycle
│
├── internal/
│   ├── database/
│   │   ├── prisma.go        # Prisma connection and lifecycle management
│   │   └── db/              # Auto-generated Prisma Go Client
│   │
│   ├── router/
│   │   ├── router.go        # Global Chi router, middlewares & v1 registration
│   │   └── router_test.go   # Router integration tests
│   │
│   ├── user/
│   │   ├── model.go         # Domain model alias
│   │   ├── dto.go           # Request/Response DTOs and validation
│   │   ├── repository.go    # Data access layer interfacing Prisma
│   │   ├── service.go       # Business logic & conflict validation
│   │   ├── handler.go       # HTTP handlers decoding requests & returning envelopes
│   │   ├── route.go         # Module route mounting
│   │   └── user_test.go     # Unit tests with in-memory repository mock
│   │
│   └── auth/
│       ├── dto.go           # Register, Login & Me DTOs
│       ├── repository.go    # Auth database queries
│       ├── service.go       # BCrypt password hashing & JWT generation
│       ├── handler.go       # Auth HTTP handlers (Register, Login, GetMe)
│       ├── route.go         # Auth route mounting with rate limit & JWT guards
│       └── auth_test.go     # Unit tests with JWT verification
│
├── pkg/
│   ├── config/
│   │   └── config.go        # Environment and configuration loader
│   ├── errors/
│   │   └── errors.go        # Domain errors and centralized HTTP error handler
│   ├── jwt/
│   │   ├── jwt.go           # JWT token creation and claim validation
│   │   └── jwt_test.go      # JWT unit tests
│   ├── logger/
│   │   └── logger.go        # Pino-style colorful zero-alloc console/JSON logger
│   ├── middleware/
│   │   ├── middleware.go    # Recovery, CORS, Timeout, JSON Content-Type
│   │   ├── auth.go          # JWT Bearer authentication middleware
│   │   └── ratelimit.go     # IP-based rate limiting
│   └── response/
│       └── response.go      # Standardized JSON response envelope
│
├── prisma/
│   └── schema.prisma        # Prisma schema definition for PostgreSQL
│
├── .env.example             # Example environment configuration
├── .env                     # Local environment settings (git-ignored)
├── .gitignore
├── Makefile                 # Automation targets (dev, run, build, test, db-push)
├── go.mod
├── go.sum
└── README.md
```

---

## Getting Started

### 1. Environment Configuration
Verify your [`.env`](file:///c:/Users/Masud%20Rana/Desktop/go-mini-setup/.env) file:
```env
PORT=8080
ENV=development
JWT_SECRET="mini-enterprise-go-jwt-secret-key-32bytes-secure!"
DATABASE_URL="postgresql://neondb_owner:...@...-pooler...neon.tech/neondb?sslmode=require&pgbouncer=true"
```

### 2. Push Schema to PostgreSQL
```bash
make db-push
```

### 3. Start the Server
```bash
make dev
```
*(Or `make run`)*

---

## API Endpoints

### System & Discovery
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | API Root Index & endpoint directory | Public |
| `GET` | `/health` | Service health status check | Public |

### Auth Module (`/api/v1/auth`)
| Method | Endpoint | Description | Auth | Rate Limit |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/register` | Register user & get JWT | Public | 10 req/min |
| `POST` | `/api/v1/auth/login` | Authenticate & get JWT | Public | 10 req/min |
| `GET` | `/api/v1/auth/me` | Current user profile | `Bearer <token>` | 100 req/min |

### User Module (`/api/v1/users`)
| Method | Endpoint | Description | Auth | Rate Limit |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/users` | List all users | Public | 100 req/min |
| `POST` | `/api/v1/users` | Create user | Public | 100 req/min |
| `GET` | `/api/v1/users/{id}` | Get user by ID | Public | 100 req/min |
| `PATCH` | `/api/v1/users/{id}` | Update user by ID | Public | 100 req/min |
| `DELETE` | `/api/v1/users/{id}` | Delete user by ID | Public | 100 req/min |

---

## Example Requests

### 1. Register Account & Obtain JWT
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "developer@example.com",
    "password": "secretPassword123",
    "name": "Dev User"
  }'
```
Response:
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "e0b04c83-...",
      "email": "developer@example.com",
      "name": "Dev User",
      "createdAt": "2026-09-20T14:40:00Z",
      "updatedAt": "2026-09-20T14:40:00Z"
    }
  }
}
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "developer@example.com",
    "password": "secretPassword123"
  }'
```

### 3. Access Protected Route (`GET /api/v1/auth/me`)
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <your_token_here>"
```
Response:
```json
{
  "success": true,
  "message": "Current user profile fetched",
  "data": {
    "user": {
      "id": "e0b04c83-...",
      "email": "developer@example.com",
      "name": "Dev User",
      "createdAt": "2026-09-20T14:40:00Z",
      "updatedAt": "2026-09-20T14:40:00Z"
    }
  }
}
```