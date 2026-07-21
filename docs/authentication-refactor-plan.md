# Authentication & Session Architecture

## Goals

* Use `HttpOnly Cookie` instead of `localStorage`.
* Support multi-merchant and multi-branch active context.
* Keep frontend authentication logic minimal.
* Support logout per device, token revocation, and refresh token rotation.
* Use JWT signed by IAM and verified through JWKS.
* Store active context inside JWT claims.
* Avoid `X-Merchant-Id` and `X-Branch-Id` headers.

---

# High Level Architecture

```text
Browser
    ↓
IAM Service
    ↓
Access Token (JWT)
    ↓
JWKS
    ↓
Catalog Service
Inventory Service
Sales Service
```

Only IAM manages:

* Login
* Refresh Token
* Session
* Switch Context
* Logout

Other services only:

1. Verify JWT via JWKS.
2. Read claims.
3. Authorize request.

---

# Session Store

## Interface

```go
type SessionStore interface {
    Create(ctx context.Context, session *Session) error
    Get(ctx context.Context, id string) (*Session, error)
    Update(ctx context.Context, session *Session) error
    Delete(ctx context.Context, id string) error
}
```

Implementations:

```text
SessionStore
├── InMemorySessionStore
└── RedisSessionStore
```

### Development

```yaml
auth:
  session_store: memory
```

### Production

```yaml
auth:
  session_store: redis
```

---

# Session Schema

```text
auth_sessions
-------------
id (uuid)
user_id
refresh_token_hash
active_user_role_id
user_agent
ip_address
last_used_at
expires_at
revoked_at
created_at
updated_at
```

Redis:

```text
session:{session_id}
```

Example:

```json
{
  "id": "sess_abc123",
  "userId": 3,
  "activeUserRoleId": 16,
  "expiresAt": "2026-08-14T00:00:00Z"
}
```

---

# JWT Claims

We use abbreviated claims to reduce token size.

| Claim | Description    |
| ----- | -------------- |
| sub   | User ID        |
| sid   | Session ID     |
| rid   | User Role ID   |
| mid   | Merchant ID    |
| bid   | Branch ID      |
| iss   | Token issuer   |
| aud   | Token audience |
| exp   | Expired At     |
| iat   | Issued At      |

---

## Merchant User

```json
{
  "sub": "3",
  "sid": "sess_abc123",
  "rid": 16,
  "mid": 1,
  "bid": 2,
  "iss": "https://iam.saleforges.com",
  "aud": "saleforges-api",
  "iat": 1783996400,
  "exp": 1784000000
}
```

---

## Platform Admin

```json
{
  "sub": "1",
  "sid": "sess_abc123",
  "rid": 1,
  "iss": "https://iam.saleforges.com",
  "aud": "saleforges-api",
  "iat": 1783996400,
  "exp": 1784000000
}
```

No merchant or branch context is required.

---

# Token Lifetime

| Token         | Lifetime |
| ------------- | -------- |
| Access Token  | 1 hour   |
| Refresh Token | 30 days  |
| Session       | 30 days  |

Recommended:

* Access Token: 1 hour
* Refresh Rotation: Enabled
* Sliding Session: Enabled

---

# Cookie Strategy

## Access Token

```text
HttpOnly : true
Secure    : true
SameSite  : Lax
Domain    : .saleforges.com
MaxAge    : 1 hour
```

## Refresh Token

```text
HttpOnly : true
Secure    : true
SameSite  : Strict
Domain    : .saleforges.com
MaxAge    : 30 days
```

---

# Login Flow

```text
POST /auth/login
↓
Create Session
↓
Generate Access Token
↓
Generate Refresh Token
↓
Set Cookies
```

Response:

```http
Set-Cookie: access_token=...
Set-Cookie: refresh_token=...
```

---


---

# Active Context

Current context is stored in:

```text
active_user_role_id
```

inside the session.

Every access token contains:

```text
rid
mid
bid
```

No tenant headers are required.

---

# Switch Context

```http
POST /auth/switch-context
```

Request:

```json
{
  "userRoleId": 16
}
```

Flow:

```text
Validate assignment
↓
Update session.active_user_role_id
↓
Generate new access token
↓
Replace access_token cookie
```

Refresh token remains unchanged.

---

# Refresh Flow

```text
Access Token Expired
↓
POST /auth/refresh
↓
Validate Refresh Token
↓
Load Session
↓
Generate Access Token
↓
Rotate Refresh Token
↓
Update Cookies
```

---

# Logout

```text
POST /auth/logout
↓
Delete Session
↓
Clear Cookies
```

---

# Frontend Configuration

Frontend:

```text
https://app.saleforges.com
```

Backend:

```text
https://api.saleforges.com
```

Axios:

```javascript
axios.create({
  baseURL: 'https://api.saleforges.com',
  withCredentials: true,
});
```

No token management is needed.

No Authorization header is needed.

No localStorage is needed.

---

# CORS

```go
AllowOrigins = []string{
    "https://app.saleforges.com",
}

AllowCredentials = true
```

Do not use:

```go
AllowOrigins = []string{"*"}
```

---

# Benefits

## Security

* Tokens are not accessible from JavaScript.
* Reduced XSS impact.
* Refresh token revocation.
* Logout per device.
* Refresh token rotation.
* Better tenant isolation.

## Frontend

* No token storage.
* No Authorization header handling.
* No tenant headers.
* Minimal authentication logic.

## Backend

* JWT + JWKS.
* Active context in token.
* Session revocation support.
* Multi-device support.
* Multi-branch support.
* Simpler authorization.

---

# Implementation Plan

**Total steps: ~13** | Hanya backend (`services/`) — frontend tidak tersentuh.

## Phase 1: Session Management (IAM)

1. **Session domain** — update struct `Session` di `domain/user.go` sesuai skema baru (uuid id, refresh_token_hash, active_user_role_id, user_agent, ip_address, dll)
2. **Session store interface** — `Create`, `Get`, `Update`, `Delete` di `port/session_store.go`
3. **In-memory session store** — implementasi di `adapter/session/memory/` (pakai `sync.Map`)
4. **Update AuthUsecase** — integrasi session: login buat session, logout hapus session, refresh validasi session

## Phase 2: JWT & Cookie (IAM)

5. **New JWT claims** — ganti `user_id`, `role_name`, `permissions[]` → `sub`, `sid`, `rid`, `mid`, `bid`. Hapus permissions dari token.
6. **Cookie middleware** — baca access_token dari cookie `HttpOnly`, bukan `Authorization: Bearer`
7. **Login handler** → `Set-Cookie` (access + refresh)
8. **Refresh handler** + rotation (ganti refresh token tiap dipakai)
9. **Switch-context endpoint** — `POST /auth/switch-context`
10. **Logout handler** — hapus session + clear cookies

## Phase 3: JWKS Verification (Shared)

11. **`pkg/jwks/`** — shared JWKS client: fetch public key, verify RS256 JWT, extract claims. Satu fungsi: `Verify(r *http.Request) (*Claims, error)`

## Phase 4: Other Services

12. **Merchant** — ganti `adapter/iam/client.go` (introspect HTTP) → pakai `pkg/jwks`. Hapus `X-Merchant-Id`, baca `mid` dari claims.
13. **Catalog** — tambah `pkg/jwks` auth middleware. Hapus `X-Merchant-Id`, baca `mid` dari claims.

## Skipped (YAGNI)

| Item | Alasan |
|------|--------|
| Redis session store | Pakai in-memory dulu, tambah kalau butuh multi-instance |
| Sliding session | Tambah kalau diminta |
| Inventory/Sales service | Belum ada |
| Rate limit refresh | Belum perlu |

---

## Status: ✅ SELESAI (2026-07-14)

Semua 13 step telah diimplementasikan:

### Commits
| Commit | Deskripsi |
|--------|-----------|
| `afd94de` | Session management infrastructure (11 files) |
| `685b340` | New JWT claims sub/sid/rid/mid/bid (3 files) |
| `f5f6f2a` | Cookie-based auth middleware + handlers (2 files) |
| `5edfb35` | Switch-context endpoint (3 files) |
| `c114396` | Shared JWKS + merchant/catalog refactor (8 files) |

### Changed Files (Total: 27 files)

| Area | Files |
|------|-------|
| **IAM domain** | `domain/user.go`, `domain/errors.go` |
| **IAM port** | `port/session_store.go` **NEW**, `port/token_signer.go` |
| **IAM adapter** | `adapter/security/jwt.go`, `adapter/session/memory/store.go` **NEW**, `adapter/repository/postgres/staff_repository.go` |
| **IAM usecase** | `usecase/auth.go`, `usecase/auth_test.go` |
| **IAM handlers** | `transport/http/handler/auth_handler.go`, `transport/http/middleware/middleware.go`, `transport/http/router.go` |
| **IAM bootstrap** | `bootstrap/bootstrap.go` |
| **Shared** | `pkg/jwks/verifier.go` **NEW**, `pkg/httputil/merchant.go` |
| **Merchant** | `transport/http/middleware/auth.go`, `transport/http/router.go`, `bootstrap/bootstrap.go` |
| **Catalog** | `transport/http/router.go`, `bootstrap/bootstrap.go`, `cmd/catalog/main.go` |
| **Bruno** | `Auth/Switch-Context.bru` **NEW**, `Auth/Logout.bru` (updated) |

### Notes
- Frontend tidak tersentuh — masih pakai `Authorization: Bearer` + JSON response
- Cookie-based auth ditambahkan secara backward-compatible
- Old `X-Merchant-Id` header masih didukung sebagai fallback
- `adapter/iam/client.go` (merchant introspect) masih ada sebagai dead code
