# Auth API Documentation

Base URL: `http://localhost:8080/api/v1`

> All auth endpoints require the `X-API-Key` header.

---

## 1. Login

Authenticate a user with email and password. Returns JWT access & refresh tokens.

### Endpoint

```
POST /api/v1/auth/login
```

### Headers

| Header       | Required | Description              |
|-------------|----------|--------------------------|
| `X-API-Key` | ✅       | Mobile app API key       |
| `Content-Type` | ✅    | `application/json`       |

### Request Body

| Field      | Type   | Required | Description           |
|-----------|--------|----------|-----------------------|
| `email`    | string | ✅       | User's email address  |
| `password` | string | ✅       | User's password       |

### cURL

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -d '{
    "email": "admin@oleron.com",
    "password": "Admin@123"
  }'
```

### Responses

#### ✅ 200 OK — Login Successful

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "first_name": "Super",
      "last_name": "Admin",
      "email": "admin@oleron.com",
      "role": "super_admin",
      "branch_id": "660e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

#### ❌ 400 Bad Request — Invalid JSON Body

```json
{
  "success": false,
  "error": "invalid request body"
}
```

#### ❌ 401 Unauthorized — Wrong Credentials

```json
{
  "success": false,
  "error": "invalid email or password"
}
```

#### ❌ 401 Unauthorized — Account Inactive

```json
{
  "success": false,
  "error": "account is not active"
}
```

#### ❌ 401 Unauthorized — Missing API Key

```json
{
  "success": false,
  "error": "missing API key"
}
```

#### ❌ 401 Unauthorized — Invalid API Key

```json
{
  "success": false,
  "error": "invalid API key"
}
```

---

## 2. Refresh Token

Exchange a valid refresh token for a new access token + refresh token pair.

### Endpoint

```
POST /api/v1/auth/refresh
```

### Headers

| Header       | Required | Description              |
|-------------|----------|--------------------------|
| `X-API-Key` | ✅       | Mobile app API key       |
| `Content-Type` | ✅    | `application/json`       |

### Request Body

| Field           | Type   | Required | Description                        |
|----------------|--------|----------|------------------------------------|
| `refresh_token` | string | ✅       | Refresh token from login response  |

### cURL

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### Responses

#### ✅ 200 OK — Token Refreshed

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600
  }
}
```

#### ❌ 400 Bad Request — Invalid JSON Body

```json
{
  "success": false,
  "error": "invalid request body"
}
```

#### ❌ 401 Unauthorized — Invalid/Expired Refresh Token

```json
{
  "success": false,
  "error": "invalid refresh token"
}
```

#### ❌ 401 Unauthorized — Missing API Key

```json
{
  "success": false,
  "error": "missing API key"
}
```

---

## Status Codes Summary

| Code  | Meaning                | When                                          |
|-------|------------------------|-----------------------------------------------|
| `200` | OK                     | Login or refresh successful                   |
| `400` | Bad Request            | Malformed JSON, missing required fields        |
| `401` | Unauthorized           | Wrong credentials, inactive account, bad API key, expired token |
| `429` | Too Many Requests      | Rate limit exceeded (if rate limiter enabled)  |
| `500` | Internal Server Error  | Token generation failure or server error       |

---

## Token Details

| Token          | Lifetime  | Usage                                     |
|---------------|-----------|-------------------------------------------|
| `access_token`  | **1 hour**  | Pass as `Authorization: Bearer <token>` on protected routes |
| `refresh_token` | **7 days**  | Use with `/auth/refresh` to get a new token pair |

### Using the Access Token

After login, include the access token in all protected API requests:

```bash
curl -X GET http://localhost:8080/api/v1/patients \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## JWT Claims (Access Token)

| Claim       | Type   | Description                |
|------------|--------|----------------------------|
| `sub`       | string | User ID (UUID)             |
| `role`      | string | User role (e.g. `super_admin`, `doctor`) |
| `branch_id` | string | Branch ID (UUID)           |
| `exp`       | int    | Expiration timestamp       |
| `iat`       | int    | Issued at timestamp        |
| `type`      | string | `"access"`                 |

## JWT Claims (Refresh Token)

| Claim  | Type   | Description            |
|--------|--------|------------------------|
| `sub`  | string | User ID (UUID)         |
| `exp`  | int    | Expiration timestamp   |
| `iat`  | int    | Issued at timestamp    |
| `type` | string | `"refresh"`            |
