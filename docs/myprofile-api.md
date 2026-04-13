# My Profile API Documentation

Base URL: `http://localhost:8080/api/v1`

> All my profile endpoints require the `X-API-Key` header.
> All my profile endpoints require a valid JWT access token in the `Authorization` header.

---

## 1. Update My Profile

Updates the authenticated user's own profile details.

- This endpoint updates the current user from the JWT `sub` claim.
- Allowed fields are `first_name`, `last_name`, `email`, `phone`, and `avatar_url`.
- `phone` and `avatar_url` can be cleared by sending an empty string.
- Password changes are handled by a separate endpoint.

### Endpoint

```
PUT /api/v1/profile/me
```

### Headers

| Header          | Required | Description                        |
|----------------|----------|------------------------------------|
| `X-API-Key`     | ✅       | Mobile app API key                 |
| `Authorization` | ✅       | `Bearer <access_token>`            |
| `Content-Type`  | ✅       | `application/json`                 |

### Request Body

| Field         | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `first_name` | string | ❌       | Updated first name |
| `last_name`  | string | ❌       | Updated last name |
| `email`      | string | ❌       | Updated email |
| `phone`      | string | ❌       | Updated phone number, or `""` to clear |
| `avatar_url` | string | ❌       | Updated avatar URL, or `""` to clear |

### cURL

```bash
curl -X PUT http://localhost:8080/api/v1/profile/me \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "first_name": "Super",
    "last_name": "Admin",
    "phone": "+919999999999",
    "avatar_url": "https://cdn.example.com/avatars/admin.png"
  }'
```

### Responses

#### ✅ 200 OK — Profile Updated

```json
{
  "success": true,
  "data": {
    "id": "b0bb4f78-4b8c-4c48-9f2f-47ff0a87c001",
    "branch_id": "a0aa4f78-4b8c-4c48-9f2f-47ff0a87c000",
    "first_name": "Super",
    "last_name": "Admin",
    "email": "admin@oleron.com",
    "phone": "+919999999999",
    "role": "super_admin",
    "status": "active",
    "avatar_url": "https://cdn.example.com/avatars/admin.png",
    "last_login_at": "2026-04-06T09:30:00Z",
    "created_at": "2026-04-01T09:00:00Z",
    "updated_at": "2026-04-06T10:15:00Z"
  }
}
```

#### ❌ 400 Bad Request — Invalid Body

```json
{
  "success": false,
  "error": "invalid request body"
}
```

#### ❌ 400 Bad Request — Invalid Field Value

```json
{
  "success": false,
  "error": "email cannot be empty"
}
```

#### ❌ 401 Unauthorized — Missing API Key

```json
{
  "success": false,
  "error": "missing API key"
}
```

#### ❌ 401 Unauthorized — Missing or Invalid Token

```json
{
  "success": false,
  "error": "missing or invalid authorization header"
}
```

#### ❌ 403 Forbidden — User Not Found in Context

```json
{
  "success": false,
  "error": "unable to determine user"
}
```

#### ❌ 404 Not Found — User Not Found

```json
{
  "success": false,
  "error": "user not found"
}
```

#### ❌ 500 Internal Server Error — Update Failed

```json
{
  "success": false,
  "error": "failed to update profile: duplicate key value violates unique constraint \"users_email_key\""
}
```

---

## 2. Change My Password

Changes the authenticated user's password.

- This endpoint updates the current user from the JWT `sub` claim.
- The user must provide the current password in `old_password`.
- The new password must be provided in `new_password`.

### Endpoint

```
PATCH /api/v1/profile/me/password
```

### Headers

| Header          | Required | Description                        |
|----------------|----------|------------------------------------|
| `X-API-Key`     | ✅       | Mobile app API key                 |
| `Authorization` | ✅       | `Bearer <access_token>`            |
| `Content-Type`  | ✅       | `application/json`                 |

### Request Body

| Field          | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `old_password` | string | ✅       | User's current password |
| `new_password` | string | ✅       | User's new password |

### cURL

```bash
curl -X PATCH http://localhost:8080/api/v1/profile/me/password \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "old_password": "Admin@123",
    "new_password": "NewSecurePassword@123"
  }'
```

### Responses

#### ✅ 200 OK — Password Updated

```json
{
  "success": true,
  "data": {
    "message": "password updated successfully"
  }
}
```

#### ❌ 400 Bad Request — Invalid Body

```json
{
  "success": false,
  "error": "invalid request body"
}
```

#### ❌ 400 Bad Request — Missing or Invalid Password Input

```json
{
  "success": false,
  "error": "old_password is incorrect"
}
```

Possible `400` errors:

- `old_password is required`
- `new_password is required`
- `old_password cannot be empty`
- `new_password cannot be empty`
- `old_password is incorrect`

#### ❌ 401 Unauthorized — Missing API Key

```json
{
  "success": false,
  "error": "missing API key"
}
```

#### ❌ 401 Unauthorized — Missing or Invalid Token

```json
{
  "success": false,
  "error": "missing or invalid authorization header"
}
```

#### ❌ 403 Forbidden — User Not Found in Context

```json
{
  "success": false,
  "error": "unable to determine user"
}
```

#### ❌ 404 Not Found — User Not Found

```json
{
  "success": false,
  "error": "user not found"
}
```

#### ❌ 500 Internal Server Error — Update Failed

```json
{
  "success": false,
  "error": "failed to change password: failed to hash password"
}
```

---

## 3. Get My Menus

Returns the navigation menu tree for the currently authenticated user.

- Menus and submenus are filtered by `role_permissions`.
- Only active menus are returned.
- The `permissions` object is populated from the current user's `role_permissions`.

### Endpoint

```
GET /api/v1/menus/me
```

### Headers

| Header          | Required | Description                        |
|----------------|----------|------------------------------------|
| `X-API-Key`     | ✅       | Mobile app API key                 |
| `Authorization` | ✅       | `Bearer <access_token>`            |

### cURL

```bash
curl -X GET http://localhost:8080/api/v1/menus/me \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

### Responses

#### ✅ 200 OK — Menus Retrieved

```json
{
  "success": true,
  "data": [
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440001",
      "parent_id": null,
      "label": "Dashboard",
      "path": "/dashboard",
      "resource": null,
      "sort_order": 1,
      "is_active": true,
      "permissions": {
        "can_view": true,
        "can_create": false,
        "can_edit": false,
        "can_delete": false
      },
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z",
      "children": []
    },
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440002",
      "parent_id": null,
      "label": "Employees",
      "path": null,
      "resource": "employee",
      "sort_order": 2,
      "is_active": true,
      "permissions": {
        "can_view": true,
        "can_create": false,
        "can_edit": false,
        "can_delete": false
      },
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z",
      "children": [
        {
          "id": "bb0e8400-e29b-41d4-a716-446655440001",
          "parent_id": "aa0e8400-e29b-41d4-a716-446655440002",
          "label": "Employee List",
          "path": "/employees",
          "resource": "employee",
          "sort_order": 1,
          "is_active": true,
          "permissions": {
            "can_view": true,
            "can_create": false,
            "can_edit": false,
            "can_delete": false
          },
          "created_at": "2026-04-01T10:00:00Z",
          "updated_at": "2026-04-01T10:00:00Z"
        },
        {
          "id": "bb0e8400-e29b-41d4-a716-446655440002",
          "parent_id": "aa0e8400-e29b-41d4-a716-446655440002",
          "label": "Work Schedule",
          "path": "/employees/schedule",
          "resource": "attendance",
          "sort_order": 2,
          "is_active": true,
          "permissions": {
            "can_view": true,
            "can_create": false,
            "can_edit": false,
            "can_delete": false
          },
          "created_at": "2026-04-01T10:00:00Z",
          "updated_at": "2026-04-01T10:00:00Z"
        }
      ]
    }
  ]
}
```

#### ❌ 401 Unauthorized — Missing API Key

```json
{
  "success": false,
  "error": "missing API key"
}
```

#### ❌ 401 Unauthorized — Missing or Invalid Token

```json
{
  "success": false,
  "error": "missing or invalid authorization header"
}
```

#### ❌ 403 Forbidden — Role Not Found in Context

```json
{
  "success": false,
  "error": "unable to determine user role"
}
```

#### ❌ 500 Internal Server Error — Query Failed

```json
{
  "success": false,
  "error": "failed to fetch menus"
}
```

---

## Menu Response Notes

| Field         | Type     | Description |
|--------------|----------|-------------|
| `id`          | string   | Menu ID (UUID) |
| `parent_id`   | string   | Parent menu ID, or `null` for top-level |
| `label`       | string   | Menu label shown in the app |
| `path`        | string   | Route path for clickable menus |
| `resource`    | string   | Permission resource key |
| `sort_order`  | int      | Display order within the level |
| `is_active`   | bool     | Menu active flag |
| `permissions` | object   | CRUD permissions for the current user from `role_permissions` |
| `children`    | array    | Nested submenu items |
| `created_at`  | string   | Creation timestamp |
| `updated_at`  | string   | Last update timestamp |

### `permissions` Object

| Field         | Type   | Description |
|--------------|--------|-------------|
| `can_view`    | bool   | Whether the user can view the resource |
| `can_create`  | bool   | Whether the user can create records |
| `can_edit`    | bool   | Whether the user can edit records |
| `can_delete`  | bool   | Whether the user can delete records |

---

## Status Codes Summary

| Code  | Meaning                 | When |
|-------|-------------------------|------|
| `200` | OK                      | Profile updated or menus fetched successfully |
| `400` | Bad Request             | Invalid JSON body or invalid profile field value |
| `401` | Unauthorized            | Missing API key or invalid JWT |
| `403` | Forbidden               | User or role not available in request context |
| `404` | Not Found               | Authenticated user record not found |
| `500` | Internal Server Error   | Database or server error |

---

## Notes

- This module is intended for self-service endpoints for the authenticated user.
- The menu response is always a tree, not a flat list.
