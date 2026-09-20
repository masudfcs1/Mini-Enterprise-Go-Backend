# Mini Enterprise Go Backend (Chi + Prisma ORM + PostgreSQL)

A modular, scalable, layered REST API backend in Go inspired by clean enterprise architectural patterns.

---

## Highlights & Features

- **Modular Layered Architecture**: Independent modules (`user`, `auth`) with strict separation of concerns (`route` ➔ `handler` ➔ `service` ➔ `repository` ➔ `prisma client`).
- **Dual-Token System (Access + Refresh)**:
  - **Access Token (15m)**: Short-lived HMAC-SHA256 JWT signed with `JWT_ACCESS_SECRET`, carrying user identity and role.
  - **Refresh Token (7d / 30d)**: Long-lived token signed with `JWT_REFRESH_SECRET`, statefully tracked in PostgreSQL (`RefreshToken` table) with **Token Rotation** and instant revocation on logout.
  - **Remember Me**: Configurable extended lifetime (`30d`).
- **Flexible Token Transport**:
  - `AUTH_TOKEN_TRANSPORT=cookie`: Automatically manages secure `HttpOnly` cookies (`access_token`, `refresh_token`) for modern web frontends (Next.js/React).
  - Transparent fallback to `Authorization: Bearer <token>` for mobile apps, cURL, and automated tests.
- **Prisma Role-Based Access Control (RBAC)**:
  - PostgreSQL enum `Role { USER, ADMIN, SUPER_ADMIN }`.
  - Roles embedded in JWT claims.
  - `RequireRole("ADMIN", "SUPER_ADMIN")` middleware guarding administrative actions (e.g. deleting users).
- **Pino-Style Colorful Terminal Logger**:
  - Built on `rs/zerolog` + `mattn/go-colorable`.
  - Color-coded badges for HTTP verbs, status codes (2xx green, 3xx cyan, 4xx yellow, 5xx red), latency, and ASCII startup banner.
- **Rate Limiting**: Built on `go-chi/httprate` (100 req/min globally; strict 10 req/min on authentication endpoints).
- **PgBouncer / Neon Connection Pooler Ready**: Disables conflicting prepared statements in transaction pooling mode (`&pgbouncer=true`).

---

## Architecture Flow

```text
HTTP Request (Header: Bearer <token> OR Cookie: access_token)
     │
     ▼
Global Middlewares (Pino Logger, Recoverer, CORS, Rate Limiter)
     │
     ▼
Global Router (internal/router/router.go)
     │
     ▼
Module Routes & RBAC Guards (internal/user/route.go, internal/auth/route.go)
     │
     ▼
Handlers (Cookie & Payload negotiation, JSON envelopes)
     │
     ▼
Services (Token generation, rotation, password hashing with BCrypt)
     │
     ▼
Repositories (PostgreSQL queries & RefreshToken lifecycle via Prisma)
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
│   │   ├── dto.go           # Request/Response DTOs with Role field
│   │   ├── repository.go    # Data access layer interfacing Prisma
│   │   ├── service.go       # Business logic & conflict validation
│   │   ├── handler.go       # HTTP handlers decoding requests & returning envelopes
│   │   ├── route.go         # Module route mounting with RBAC guards
│   │   └── user_test.go     # Unit tests with in-memory repository mock
│   │
│   └── auth/
│       ├── dto.go           # Register, Login, Refresh, Logout, and Me DTOs
│       ├── repository.go    # Auth and RefreshToken database queries
│       ├── service.go       # BCrypt hashing, dual-token issuance & rotation
│       ├── handler.go       # Cookie management (Set-Cookie / Clear) and endpoints
│       ├── route.go         # Auth route mounting with rate limits & JWT guards
│       └── auth_test.go     # Dual-token and rotation unit tests
│
├── pkg/
│   ├── config/
│   │   └── config.go        # Configuration loader with duration parsing
│   ├── errors/
│   │   └── errors.go        # Domain errors and centralized HTTP error handler
│   ├── jwt/
│   │   ├── jwt.go           # Dual-token (access/refresh) signing & validation
│   │   └── jwt_test.go      # JWT unit tests
│   ├── logger/
│   │   └── logger.go        # Pino-style colorful zero-alloc console/JSON logger
│   ├── middleware/
│   │   ├── middleware.go    # Recovery, CORS, Timeout, JSON Content-Type
│   │   ├── auth.go          # JWT Bearer & Cookie authentication middleware
│   │   ├── rbac.go          # Role-Based Access Control guard middleware
│   │   ├── rbac_test.go     # RBAC unit tests
│   │   └── ratelimit.go     # IP-based rate limiting
│   └── response/
│       └── response.go      # Standardized JSON response envelope
│
├── prisma/
│   └── schema.prisma        # Prisma schema (Role enum, User, RefreshToken)
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

## Environment Variables

| Variable | Default / Description |
| :--- | :--- |
| `PORT` | `8080` |
| `ENV` | `development` (or `production`) |
| `DATABASE_URL` | PostgreSQL connection string (`&pgbouncer=true` for poolers) |
| `JWT_ACCESS_SECRET` | Secret key used to sign Access Tokens (min. 32 chars) |
| `JWT_REFRESH_SECRET` | Secret key used to sign Refresh Tokens (min. 32 chars) |
| `AUTH_TOKEN_SECRET` | Secret key for auth challenges / HMAC |
| `COOKIE_SECRET` | Secret key for cookie signing / encryption |
| `JWT_ACCESS_EXPIRES` | Access token lifespan (e.g. `15m`) |
| `JWT_REFRESH_EXPIRES` | Standard refresh token lifespan (e.g. `7d`) |
| `JWT_REFRESH_REMEMBER_EXPIRES` | Remember-me refresh token lifespan (e.g. `30d`) |
| `JWT_ISSUER` | Token issuer (e.g. `enterprise-backend`) |
| `JWT_AUDIENCE` | Token audience (e.g. `enterprise-api`) |
| `AUTH_TOKEN_TRANSPORT` | `cookie` (sets HttpOnly cookies) or `bearer` (header only) |

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
| `POST` | `/api/v1/auth/register` | Register account (supports optional role) | Public | 10 req/min |
| `POST` | `/api/v1/auth/login` | Log in (supports `"remember_me": true`) | Public | 10 req/min |
| `POST` | `/api/v1/auth/refresh` | Rotate access & refresh tokens | Cookie / Body | 10 req/min |
| `POST` | `/api/v1/auth/logout` | Revoke session & clear cookies | Public | 100 req/min |
| `GET` | `/api/v1/auth/me` | Fetch authenticated user profile | Bearer / Cookie | 100 req/min |

### User Module (`/api/v1/users`)
| Method | Endpoint | Description | Auth | Required Role |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/users` | List all users | Public | - |
| `POST` | `/api/v1/users` | Create user | Public | - |
| `GET` | `/api/v1/users/{id}` | Get user by ID | Public | - |
| `PATCH` | `/api/v1/users/{id}` | Update user fields | Public | - |
| `DELETE` | `/api/v1/users/{id}` | Delete user | Bearer / Cookie | **ADMIN** or **SUPER_ADMIN** |

---

## Example Requests

### 1. Register with Role
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "secretPassword123",
    "name": "Admin User",
    "role": "ADMIN"
  }'
```
Response sets `access_token` and `refresh_token` as secure `HttpOnly` cookies and returns:
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsIn...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsIn...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": "f626fa50-...",
      "email": "admin@example.com",
      "name": "Admin User",
      "role": "ADMIN"
    }
  }
}
```

### 2. Login with Remember Me
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "secretPassword123",
    "remember_me": true
  }'
```

### 3. Rotate Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<REFRESH_TOKEN_HERE>"
  }'
```
*(Or simply call `curl -b cookies.txt -X POST http://localhost:8080/api/v1/auth/refresh` when using cookie transport).*

### 4. Access Protected Profile
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

### 5. RBAC Enforcement on Delete User
```bash
# Permitted when bearer token belongs to an ADMIN or SUPER_ADMIN:
curl -X DELETE http://localhost:8080/api/v1/users/<USER_ID> \
  -H "Authorization: Bearer <ADMIN_ACCESS_TOKEN>"

# If a standard USER attempts this action, returns HTTP 403 Forbidden:
# {"success":false,"error":"forbidden: role 'USER' lacks required permissions"}
```