# Admin API Documentation

Base URL: `http://localhost:8080/api/v1/admin`

> **All admin endpoints require:**
> - `X-API-Key` header
> - `Authorization: Bearer <access_token>` header
> - The authenticated user must have the **`super_admin`** role

---

## Table of Contents

- [Branches](#branches)
  - [1. List All Branches](#1-list-all-branches)
  - [2. Get Branch by ID](#2-get-branch-by-id)
  - [3. Create Branch](#3-create-branch)
  - [4. Update Branch](#4-update-branch)
  - [5. Delete Branch](#5-delete-branch)
- [Users](#users)
  - [1. List All Users](#1-list-all-users)
  - [2. Get User by ID](#2-get-user-by-id)
  - [3. Create User](#3-create-user)
  - [4. Update User](#4-update-user)
  - [5. Reset User Password](#5-reset-user-password)
  - [6. Delete User](#6-delete-user)
- [Menus (super_admin only)](#menus-super_admin-only)
  - [1. List All Menus (Flat)](#1-list-all-menus-flat)
  - [2. List All Menus (Tree)](#2-list-all-menus-tree)
  - [3. Get Menu by ID](#3-get-menu-by-id)
  - [4. Create Menu](#4-create-menu)
  - [5. Update Menu](#5-update-menu)
  - [6. Delete Menu](#6-delete-menu)
- [Role Permissions (super_admin only)](#role-permissions-super_admin-only)
  - [1. List All Role Permissions](#1-list-all-role-permissions)
  - [2. Get Role Permission by ID](#2-get-role-permission-by-id)
  - [3. Create Role Permission](#3-create-role-permission)
  - [4. Update Role Permission](#4-update-role-permission)
  - [5. Delete Role Permission](#5-delete-role-permission)
- [Common Error Responses](#common-error-responses)
- [Status Codes Summary](#status-codes-summary)
- [Available User Roles](#available-user-roles)
- [Available Resources](#available-resources)
- [Available User Statuses](#available-user-statuses)

---

## Branches

### 1. List All Branches

```
GET /api/v1/admin/branches
```

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/admin/branches \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Main Branch",
      "code": "BRANCH01",
      "address": "123 Main Street",
      "phone": "+1234567890",
      "email": "main@oleron.com",
      "logo_url": null,
      "is_active": true,
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    }
  ]
}
```

---

### 2. Get Branch by ID

```
GET /api/v1/admin/branches/{id}
```

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/admin/branches/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Main Branch",
    "code": "BRANCH01",
    "address": "123 Main Street",
    "phone": "+1234567890",
    "email": "main@oleron.com",
    "logo_url": null,
    "is_active": true,
    "created_at": "2026-04-01T10:00:00Z",
    "updated_at": "2026-04-01T10:00:00Z"
  }
}
```

#### ❌ 404 Not Found

```json
{
  "success": false,
  "error": "branch not found"
}
```

---

### 3. Create Branch

```
POST /api/v1/admin/branches
```

#### Request Body

| Field      | Type   | Required | Description                          |
|-----------|--------|----------|--------------------------------------|
| `name`     | string | ✅       | Branch display name                  |
| `code`     | string | ✅       | Unique branch code (e.g. `BRANCH02`) |
| `address`  | string | ❌       | Physical address                     |
| `phone`    | string | ❌       | Contact phone                        |
| `email`    | string | ❌       | Contact email                        |
| `logo_url` | string | ❌       | URL to branch logo                   |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/admin/branches \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "name": "North Branch",
    "code": "BRANCH02",
    "address": "456 North Avenue",
    "phone": "+0987654321",
    "email": "north@oleron.com"
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "North Branch",
    "code": "BRANCH02",
    "address": "456 North Avenue",
    "phone": "+0987654321",
    "email": "north@oleron.com",
    "is_active": true,
    "created_at": "2026-04-01T10:30:00Z",
    "updated_at": "2026-04-01T10:30:00Z"
  }
}
```

#### ❌ 400 Bad Request

```json
{
  "success": false,
  "error": "name and code are required"
}
```

---

### 4. Update Branch

```
PUT /api/v1/admin/branches/{id}
```

#### Request Body

| Field       | Type   | Required | Description               |
|------------|--------|----------|---------------------------|
| `name`      | string | ❌       | Updated name              |
| `code`      | string | ❌       | Updated code              |
| `address`   | string | ❌       | Updated address           |
| `phone`     | string | ❌       | Updated phone             |
| `email`     | string | ❌       | Updated email             |
| `logo_url`  | string | ❌       | Updated logo URL          |
| `is_active` | bool   | ❌       | Enable/disable the branch |

#### cURL

```bash
curl -X PUT http://localhost:8080/api/v1/admin/branches/660e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "name": "North Branch (Updated)",
    "is_active": false
  }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "North Branch (Updated)",
    "code": "BRANCH02",
    "address": "456 North Avenue",
    "phone": "+0987654321",
    "email": "north@oleron.com",
    "is_active": false,
    "created_at": "2026-04-01T10:30:00Z",
    "updated_at": "2026-04-01T10:45:00Z"
  }
}
```

---

### 5. Delete Branch

```
DELETE /api/v1/admin/branches/{id}
```

#### cURL

```bash
curl -X DELETE http://localhost:8080/api/v1/admin/branches/660e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "message": "branch deleted"
  }
}
```

> ⚠️ **Note:** Deleting a branch that still has users will fail due to the foreign key constraint (`ON DELETE RESTRICT`).

---

## Users

### 1. List All Users

```
GET /api/v1/admin/users
```

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/admin/users \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "branch_id": "550e8400-e29b-41d4-a716-446655440000",
      "first_name": "Super",
      "last_name": "Admin",
      "email": "admin@oleron.com",
      "phone": null,
      "role": "super_admin",
      "status": "active",
      "avatar_url": null,
      "last_login_at": "2026-04-01T10:00:00Z",
      "created_at": "2026-04-01T09:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    }
  ]
}
```

---

### 2. Get User by ID

```
GET /api/v1/admin/users/{id}
```

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/admin/users/770e8400-e29b-41d4-a716-446655440000 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "first_name": "Super",
    "last_name": "Admin",
    "email": "admin@oleron.com",
    "role": "super_admin",
    "status": "active",
    "created_at": "2026-04-01T09:00:00Z",
    "updated_at": "2026-04-01T10:00:00Z"
  }
}
```

#### ❌ 404 Not Found

```json
{
  "success": false,
  "error": "user not found"
}
```

---

### 3. Create User

```
POST /api/v1/admin/users
```

#### Request Body

| Field        | Type   | Required | Description                                                        |
|-------------|--------|----------|--------------------------------------------------------------------|
| `branch_id`  | string | ✅       | UUID of the branch to assign the user to                           |
| `first_name` | string | ✅       | User's first name                                                  |
| `last_name`  | string | ✅       | User's last name                                                   |
| `email`      | string | ✅       | Unique email address                                               |
| `password`   | string | ✅       | Password (min 6 chars, stored as bcrypt hash)                      |
| `role`       | string | ✅       | One of: `super_admin`, `admin`, `manager`, `employee`              |
| `phone`      | string | ❌       | Contact phone number                                               |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/admin/users \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@oleron.com",
    "password": "SecurePass123",
    "role": "employee",
    "phone": "+1122334455"
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "id": "880e8400-e29b-41d4-a716-446655440002",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@oleron.com",
    "phone": "+1122334455",
    "role": "employee",
    "status": "active",
    "created_at": "2026-04-01T11:00:00Z",
    "updated_at": "2026-04-01T11:00:00Z"
  }
}
```

#### ❌ 400 Bad Request

```json
{
  "success": false,
  "error": "first_name, last_name, email, password, branch_id, and role are required"
}
```

---

### 4. Update User

```
PUT /api/v1/admin/users/{id}
```

#### Request Body

| Field        | Type   | Required | Description                                           |
|-------------|--------|----------|-------------------------------------------------------|
| `branch_id`  | string | ❌       | Reassign to a different branch                        |
| `first_name` | string | ❌       | Updated first name                                    |
| `last_name`  | string | ❌       | Updated last name                                     |
| `email`      | string | ❌       | Updated email                                         |
| `phone`      | string | ❌       | Updated phone                                         |
| `role`       | string | ❌       | One of: `super_admin`, `admin`, `manager`, `employee` |
| `status`     | string | ❌       | One of: `active`, `inactive`, `suspended`, `pending`  |

#### cURL

```bash
curl -X PUT http://localhost:8080/api/v1/admin/users/880e8400-e29b-41d4-a716-446655440002 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "role": "manager",
    "status": "active"
  }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "880e8400-e29b-41d4-a716-446655440002",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@oleron.com",
    "phone": "+1122334455",
    "role": "manager",
    "status": "active",
    "created_at": "2026-04-01T11:00:00Z",
    "updated_at": "2026-04-01T11:15:00Z"
  }
}
```

---

### 5. Reset User Password

```
PATCH /api/v1/admin/users/{id}/password
```

#### Request Body

| Field      | Type   | Required | Description                |
|-----------|--------|----------|----------------------------|
| `password` | string | ✅       | New password (min 6 chars) |

#### cURL

```bash
curl -X PATCH http://localhost:8080/api/v1/admin/users/880e8400-e29b-41d4-a716-446655440002/password \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "password": "NewSecure456"
  }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "message": "password reset successful"
  }
}
```

#### ❌ 400 Bad Request

```json
{
  "success": false,
  "error": "password is required"
}
```

---

### 6. Delete User

```
DELETE /api/v1/admin/users/{id}
```

#### cURL

```bash
curl -X DELETE http://localhost:8080/api/v1/admin/users/880e8400-e29b-41d4-a716-446655440002 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "message": "user deleted"
  }
}
```

---

## Menus (super_admin only)

> Menu items use a **tree structure** via `parent_id`. Top-level menus have `parent_id: null`.
> The `resource` field links to `role_permissions` to control which menus each role can see.
> Icons are handled on the frontend based on the `label` value.

### 1. List All Menus (Flat)

Returns all menus as a flat list, regardless of hierarchy.

```
GET /api/v1/admin/menus
```

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/admin/menus \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

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
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    },
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440002",
      "parent_id": null,
      "label": "Employees",
      "path": null,
      "resource": "employee",
      "sort_order": 2,
      "is_active": true,
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    },
    {
      "id": "bb0e8400-e29b-41d4-a716-446655440001",
      "parent_id": "aa0e8400-e29b-41d4-a716-446655440002",
      "label": "Employee List",
      "path": "/employees",
      "resource": "employee",
      "sort_order": 1,
      "is_active": true,
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    }
  ]
}
```

---

### 2. List All Menus (Tree)

Returns all menus as a **nested tree** with children embedded inside their parents.

```
GET /api/v1/admin/menus/tree
```

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/admin/menus/tree \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": [
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440001",
      "parent_id": null,
      "label": "Dashboard",
      "path": "/dashboard",
      "sort_order": 1,
      "is_active": true,
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    },
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440002",
      "parent_id": null,
      "label": "Employees",
      "resource": "employee",
      "sort_order": 2,
      "is_active": true,
      "children": [
        {
          "id": "bb0e8400-e29b-41d4-a716-446655440001",
          "parent_id": "aa0e8400-e29b-41d4-a716-446655440002",
          "label": "Employee List",
          "path": "/employees",
          "resource": "employee",
          "sort_order": 1,
          "is_active": true,
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
          "created_at": "2026-04-01T10:00:00Z",
          "updated_at": "2026-04-01T10:00:00Z"
        }
      ],
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    }
  ]
}
```

---

### 3. Get Menu by ID

```
GET /api/v1/admin/menus/{id}
```

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/admin/menus/aa0e8400-e29b-41d4-a716-446655440002 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "aa0e8400-e29b-41d4-a716-446655440002",
    "parent_id": null,
    "label": "Employees",
    "resource": "employee",
    "sort_order": 2,
    "is_active": true,
    "created_at": "2026-04-01T10:00:00Z",
    "updated_at": "2026-04-01T10:00:00Z"
  }
}
```

#### ❌ 404 Not Found

```json
{
  "success": false,
  "error": "menu not found"
}
```

---

### 4. Create Menu

```
POST /api/v1/admin/menus
```

#### Request Body

| Field        | Type    | Required | Description                                                                              |
|-------------|---------|----------|------------------------------------------------------------------------------------------|
| `label`      | string  | ✅       | Display label (e.g. `"Attendance"`)                                                      |
| `parent_id`  | string  | ❌       | UUID of parent menu (`null` for top-level)                                               |
| `path`       | string  | ❌       | Route path (`null` for parent menus with children)                                       |
| `resource`   | string  | ❌       | One of: `employee`, `attendance`, `payroll`, `report`, `settings` |
| `sort_order` | int     | ❌       | Display order (default: `0`)                                                             |

#### cURL — Create a top-level menu

```bash
curl -X POST http://localhost:8080/api/v1/admin/menus \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "label": "Payroll",
    "path": "/payroll",
    "resource": "payroll",
    "sort_order": 4
  }'
```

#### cURL — Create a sub-menu

```bash
curl -X POST http://localhost:8080/api/v1/admin/menus \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "parent_id": "aa0e8400-e29b-41d4-a716-446655440002",
    "label": "Work Schedule",
    "path": "/employees/schedule",
    "resource": "attendance",
    "sort_order": 2
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "id": "cc0e8400-e29b-41d4-a716-446655440001",
    "parent_id": null,
    "label": "Payroll",
    "path": "/payroll",
    "resource": "payroll",
    "sort_order": 4,
    "is_active": true,
    "created_at": "2026-04-01T12:00:00Z",
    "updated_at": "2026-04-01T12:00:00Z"
  }
}
```

#### ❌ 400 Bad Request

```json
{
  "success": false,
  "error": "label is required"
}
```

---

### 5. Update Menu

```
PUT /api/v1/admin/menus/{id}
```

#### Request Body

| Field        | Type    | Required | Description                                        |
|-------------|---------|----------|----------------------------------------------------|
| `parent_id`  | string  | ❌       | Move to different parent (`null` for top-level)    |
| `label`      | string  | ❌       | Updated label                                      |
| `path`       | string  | ❌       | Updated route path                                 |
| `resource`   | string  | ❌       | Updated resource name                              |
| `sort_order` | int     | ❌       | Updated display order                              |
| `is_active`  | bool    | ❌       | Enable/disable the menu                            |

#### cURL

```bash
curl -X PUT http://localhost:8080/api/v1/admin/menus/cc0e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "label": "Payroll & Salary",
    "sort_order": 3,
    "is_active": true
  }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "cc0e8400-e29b-41d4-a716-446655440001",
    "parent_id": null,
    "label": "Payroll & Salary",
    "path": "/payroll",
    "resource": "payroll",
    "sort_order": 3,
    "is_active": true,
    "created_at": "2026-04-01T12:00:00Z",
    "updated_at": "2026-04-01T12:15:00Z"
  }
}
```

---

### 6. Delete Menu

```
DELETE /api/v1/admin/menus/{id}
```

> ⚠️ **Note:** Deleting a parent menu will **cascade delete** all its children.

#### cURL

```bash
curl -X DELETE http://localhost:8080/api/v1/admin/menus/cc0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "message": "menu deleted"
  }
}
```

---

## Role Permissions (super_admin only)

> `role_permissions` controls what each role can do for a given `resource`.
> Used by menu filtering and backend authorization rules.

### 1. List All Role Permissions

```
GET /api/v1/admin/role-permissions
```

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/admin/role-permissions \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": [
    {
      "id": "de0e8400-e29b-41d4-a716-446655440001",
      "role": "manager",
      "resource": "attendance",
      "can_view": true,
      "can_create": true,
      "can_edit": true,
      "can_delete": false,
      "created_at": "2026-04-01T10:00:00Z"
    },
    {
      "id": "de0e8400-e29b-41d4-a716-446655440002",
      "role": "employee",
      "resource": "attendance",
      "can_view": true,
      "can_create": false,
      "can_edit": false,
      "can_delete": false,
      "created_at": "2026-04-01T10:00:00Z"
    }
  ]
}
```

---

### 2. Get Role Permission by ID

```
GET /api/v1/admin/role-permissions/{id}
```

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/admin/role-permissions/de0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "de0e8400-e29b-41d4-a716-446655440001",
    "role": "manager",
    "resource": "attendance",
    "can_view": true,
    "can_create": true,
    "can_edit": true,
    "can_delete": false,
    "created_at": "2026-04-01T10:00:00Z"
  }
}
```

#### ❌ 404 Not Found

```json
{
  "success": false,
  "error": "role permission not found"
}
```

---

### 3. Create Role Permission

```
POST /api/v1/admin/role-permissions
```

#### Request Body

| Field        | Type   | Required | Description                                                             |
|-------------|--------|----------|-------------------------------------------------------------------------|
| `role`       | string | ✅       | One of: `super_admin`, `admin`, `manager`, `employee`                   |
| `resource`   | string | ✅       | One of: `employee`, `attendance`, `payroll`, `report`, `settings`       |
| `can_view`   | bool   | ❌       | View permission (default: `false`)                                      |
| `can_create` | bool   | ❌       | Create permission (default: `false`)                                    |
| `can_edit`   | bool   | ❌       | Edit permission (default: `false`)                                      |
| `can_delete` | bool   | ❌       | Delete permission (default: `false`)                                    |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/admin/role-permissions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "role": "manager",
    "resource": "payroll",
    "can_view": true,
    "can_create": false,
    "can_edit": false,
    "can_delete": false
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "id": "de0e8400-e29b-41d4-a716-446655440003",
    "role": "manager",
    "resource": "payroll",
    "can_view": true,
    "can_create": false,
    "can_edit": false,
    "can_delete": false,
    "created_at": "2026-04-01T12:00:00Z"
  }
}
```

#### ❌ 400 Bad Request

```json
{
  "success": false,
  "error": "role and resource are required"
}
```

---

### 4. Update Role Permission

```
PUT /api/v1/admin/role-permissions/{id}
```

#### Request Body

| Field        | Type   | Required | Description              |
|-------------|--------|----------|--------------------------|
| `role`       | string | ❌       | Updated role             |
| `resource`   | string | ❌       | Updated resource key     |
| `can_view`   | bool   | ❌       | Updated view permission  |
| `can_create` | bool   | ❌       | Updated create permission|
| `can_edit`   | bool   | ❌       | Updated edit permission  |
| `can_delete` | bool   | ❌       | Updated delete permission|

#### cURL

```bash
curl -X PUT http://localhost:8080/api/v1/admin/role-permissions/de0e8400-e29b-41d4-a716-446655440003 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "can_edit": true
  }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "de0e8400-e29b-41d4-a716-446655440003",
    "role": "manager",
    "resource": "payroll",
    "can_view": true,
    "can_create": false,
    "can_edit": true,
    "can_delete": false,
    "created_at": "2026-04-01T12:00:00Z"
  }
}
```

#### ❌ 404 Not Found

```json
{
  "success": false,
  "error": "role permission not found"
}
```

---

### 5. Delete Role Permission

```
DELETE /api/v1/admin/role-permissions/{id}
```

#### cURL

```bash
curl -X DELETE http://localhost:8080/api/v1/admin/role-permissions/de0e8400-e29b-41d4-a716-446655440003 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "message": "role permission deleted"
  }
}
```

---

## Common Error Responses

### ❌ 401 Unauthorized — Missing/Invalid API Key

```json
{
  "success": false,
  "error": "missing API key"
}
```

### ❌ 401 Unauthorized — Missing/Invalid JWT

```json
{
  "success": false,
  "error": "Unauthorized"
}
```

### ❌ 403 Forbidden — Non-super_admin Access

```json
{
  "success": false,
  "error": "insufficient permissions"
}
```

### ❌ 500 Internal Server Error

```json
{
  "success": false,
  "error": "failed to create branch: ERROR: duplicate key value violates unique constraint..."
}
```

---

## Status Codes Summary

| Code  | Meaning               | When                                                                   |
|-------|-----------------------|------------------------------------------------------------------------|
| `200` | OK                    | Successful read, update, or delete                                     |
| `201` | Created               | Successfully created a branch, user, menu, or role permission          |
| `400` | Bad Request           | Invalid JSON, missing required fields, or invalid enum value           |
| `401` | Unauthorized          | Missing/invalid API key or JWT token                                   |
| `403` | Forbidden             | Authenticated user is not a `super_admin`                              |
| `404` | Not Found             | Branch, user, menu, or role permission with the given ID doesn't exist |
| `500` | Internal Server Error | Database constraint violation or unexpected server error               |

---

## Available User Roles

| Role          | Description                                               |
|--------------|-----------------------------------------------------------|
| `super_admin` | Full system access across all branches                    |
| `admin`       | Branch-level full access — acts as branch manager         |
| `manager`     | Manages employees; views and approves attendance & payroll|
| `employee`    | Punches in/out via mobile; views own attendance & salary  |

---

## Available Resources

| Resource     | Description                              |
|-------------|------------------------------------------|
| `employee`   | Employee records and profiles            |
| `attendance` | Punch-in/out records and working hours   |
| `payroll`    | Salary calculation and payment records   |
| `report`     | Attendance and payroll reports           |
| `settings`   | System and branch configuration          |

---

## Available User Statuses

| Status      | Description               |
|------------|---------------------------|
| `active`    | User can log in           |
| `inactive`  | Account disabled          |
| `suspended` | Temporarily suspended     |
| `pending`   | Awaiting activation       |
