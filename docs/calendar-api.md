# Branch Calendar API Documentation

Base URL: `http://localhost:8080/api/v1`

> **All endpoints require:**
> - `X-API-Key` header
> - `Authorization: Bearer <access_token>` header
>
> | Route prefix | Required role |
> |---|---|
> | `/api/v1/calendar/*` | `super_admin`, `admin`, `manager` |

> **Role-based filtering:**
> | Role | Returns |
> |---|---|
> | `super_admin` | Calendar entries from all branches |
> | `admin` | Entries from their branch only |
> | `manager` | Entries from their branch only |
>
> Branch filtering is applied automatically from the JWT `branch_id` claim.

> **Purpose:** The branch calendar stores **per-date overrides** to the weekly office schedule.
> It is used during payroll generation to determine:
> - Which working days were public holidays (no deduction for absence)
> - Which non-working days were makeup working days (expect attendance)

---

## Table of Contents

- [Branch Calendar](#branch-calendar)
  - [1. List Calendar Entries](#1-list-calendar-entries)
  - [2. Get Entry by ID](#2-get-entry-by-id)
  - [3. Create Entry](#3-create-entry)
  - [4. Update Entry](#4-update-entry)
  - [5. Delete Entry](#5-delete-entry)
- [Common Error Responses](#common-error-responses)
- [Status Codes Summary](#status-codes-summary)
- [Entry Type Reference](#entry-type-reference)
- [Response Fields Reference](#response-fields-reference)

---

## Branch Calendar

### 1. List Calendar Entries

```
GET /api/v1/calendar/branch-calendar
```

Returns all calendar entries for the branch. Supports optional filtering by date range and type.

#### Query Parameters

| Param  | Type   | Required | Description                                      |
|--------|--------|----------|--------------------------------------------------|
| `from` | string | ❌       | Start date filter in `YYYY-MM-DD` format         |
| `to`   | string | ❌       | End date filter in `YYYY-MM-DD` format           |
| `type` | string | ❌       | One of: `holiday`, `working_day`                 |

#### cURL — All entries

```bash
curl -X GET http://localhost:8080/api/v1/calendar/branch-calendar \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### cURL — Holidays for a date range

```bash
curl -X GET "http://localhost:8080/api/v1/calendar/branch-calendar?from=2026-01-01&to=2026-12-31&type=holiday" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": [
    {
      "id": "ca0e8400-e29b-41d4-a716-446655440001",
      "branch_id": "550e8400-e29b-41d4-a716-446655440000",
      "date": "2026-01-01T00:00:00Z",
      "type": "holiday",
      "name": "New Year's Day",
      "created_at": "2026-04-01T10:00:00Z"
    },
    {
      "id": "ca0e8400-e29b-41d4-a716-446655440002",
      "branch_id": "550e8400-e29b-41d4-a716-446655440000",
      "date": "2026-04-18T00:00:00Z",
      "type": "working_day",
      "name": "Makeup Saturday",
      "created_at": "2026-04-01T10:00:00Z"
    }
  ]
}
```

#### ✅ 200 OK — No entries

```json
{
  "success": true,
  "data": null
}
```

---

### 2. Get Entry by ID

```
GET /api/v1/calendar/branch-calendar/{id}
```

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/calendar/branch-calendar/ca0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "ca0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "date": "2026-01-01T00:00:00Z",
    "type": "holiday",
    "name": "New Year's Day",
    "created_at": "2026-04-01T10:00:00Z"
  }
}
```

#### ❌ 403 Forbidden

```json
{
  "success": false,
  "error": "insufficient permissions"
}
```

#### ❌ 404 Not Found

```json
{
  "success": false,
  "error": "calendar entry not found"
}
```

---

### 3. Create Entry

```
POST /api/v1/calendar/branch-calendar
```

Adds a holiday or working-day override for a specific date. Each branch can have at most one entry per date (`UNIQUE(branch_id, date)`).

#### Request Body

| Field  | Type   | Required | Description                                                       |
|--------|--------|----------|-------------------------------------------------------------------|
| `date` | string | ✅       | Date in `YYYY-MM-DD` format                                       |
| `type` | string | ✅       | One of: `holiday`, `working_day`                                  |
| `name` | string | ❌       | Label for the entry (e.g. `"Christmas Day"`, `"Makeup Saturday"`) |

#### cURL — Add a public holiday

```bash
curl -X POST http://localhost:8080/api/v1/calendar/branch-calendar \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "date": "2026-12-25",
    "type": "holiday",
    "name": "Christmas Day"
  }'
```

#### cURL — Add a makeup working day

```bash
curl -X POST http://localhost:8080/api/v1/calendar/branch-calendar \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "date": "2026-04-18",
    "type": "working_day",
    "name": "Makeup Saturday"
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "id": "ca0e8400-e29b-41d4-a716-446655440003",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "date": "2026-12-25T00:00:00Z",
    "type": "holiday",
    "name": "Christmas Day",
    "created_at": "2026-04-16T09:00:00Z"
  }
}
```

#### ❌ 400 Bad Request

```json
{
  "success": false,
  "error": "invalid request body"
}
```

#### ❌ 422 Unprocessable Entity

```json
{
  "success": false,
  "error": "Key: 'CreateCalendarEntryRequest.Type' Error:Field validation for 'Type' failed on the 'oneof' tag"
}
```

#### ❌ 500 Internal Server Error — Duplicate date

```json
{
  "success": false,
  "error": "failed to create calendar entry"
}
```

> ⚠️ Each branch can only have **one entry per date**. Attempting to create a second entry for the same date will fail with a unique constraint error.

---

### 4. Update Entry

```
PUT /api/v1/calendar/branch-calendar/{id}
```

Updates the `type` and/or `name` of an existing entry. The `date` cannot be changed — delete and recreate if needed.

> `admin` and `manager` can only update entries from their own branch.

#### Request Body

| Field  | Type   | Required | Description                      |
|--------|--------|----------|----------------------------------|
| `type` | string | ✅       | One of: `holiday`, `working_day` |
| `name` | string | ❌       | Updated label                    |

#### cURL

```bash
curl -X PUT http://localhost:8080/api/v1/calendar/branch-calendar/ca0e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "type": "holiday",
    "name": "New Year's Day (National)"
  }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "ca0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "date": "2026-01-01T00:00:00Z",
    "type": "holiday",
    "name": "New Year's Day (National)",
    "created_at": "2026-04-01T10:00:00Z"
  }
}
```

#### ❌ 403 Forbidden

```json
{
  "success": false,
  "error": "insufficient permissions"
}
```

#### ❌ 500 Internal Server Error

```json
{
  "success": false,
  "error": "failed to update calendar entry"
}
```

---

### 5. Delete Entry

```
DELETE /api/v1/calendar/branch-calendar/{id}
```

Removes a calendar entry. Once deleted, that date reverts to the standard weekly schedule defined in `office_timings`.

> `admin` and `manager` can only delete entries from their own branch.

#### cURL

```bash
curl -X DELETE http://localhost:8080/api/v1/calendar/branch-calendar/ca0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true
}
```

#### ❌ 403 Forbidden

```json
{
  "success": false,
  "error": "insufficient permissions"
}
```

#### ❌ 500 Internal Server Error

```json
{
  "success": false,
  "error": "failed to delete calendar entry"
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

### ❌ 403 Forbidden — Insufficient Role

```json
{
  "success": false,
  "error": "insufficient permissions"
}
```

---

## Status Codes Summary

| Code | Meaning | When |
|---|---|---|
| `200` | OK | Request succeeded |
| `201` | Created | Calendar entry created successfully |
| `400` | Bad Request | Malformed JSON body |
| `401` | Unauthorized | Missing/invalid API key or JWT |
| `403` | Forbidden | Role is not `super_admin`, `admin`, or `manager`; or entry belongs to another branch |
| `404` | Not Found | Entry ID does not exist |
| `422` | Unprocessable Entity | Validation failed (e.g. invalid `type` value) |
| `500` | Internal Server Error | Database error (including duplicate date constraint) |

---

## Entry Type Reference

| Type | Description | Effect on payroll |
|---|---|---|
| `holiday` | Non-working day override | No absence deduction for employees who didn't punch in |
| `working_day` | Working day override (e.g. makeup Saturday) | Attendance expected; absence counted even if weekly schedule says off |

---

## Response Fields Reference

| Field | Type | Description |
|---|---|---|
| `id` | string | Calendar entry UUID |
| `branch_id` | string | Branch this entry belongs to |
| `date` | string | The specific date (ISO 8601 timestamp) |
| `type` | string | `holiday` or `working_day` |
| `name` | string \| null | Optional label for the entry |
| `created_at` | string | Creation timestamp |
