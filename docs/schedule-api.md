# Schedule API Documentation

Base URL: `http://localhost:8080/api/v1`

> **All endpoints require:**
> - `X-API-Key` header
> - `Authorization: Bearer <access_token>` header
>
> | Route prefix | Required role |
> |---|---|
> | `/api/v1/schedule/*` | `super_admin`, `admin`, `manager` |

> **Role-based filtering:**
> | Role | Returns |
> |---|---|
> | `super_admin` | Office timings from all branches |
> | `admin` | Office timings from their branch only |
> | `manager` | Office timings from their branch only |
>
> Branch filtering is applied automatically from the JWT `branch_id` claim.

---

## Table of Contents

- [Office Timings](#office-timings)
  - [1. List All Office Timings](#1-list-all-office-timings)
  - [2. Get Office Timing by ID](#2-get-office-timing-by-id)
  - [3. Create Office Timing](#3-create-office-timing)
  - [4. Update Office Timing](#4-update-office-timing)
  - [5. Delete Office Timing](#5-delete-office-timing)
  - [6. Activate Office Timing](#6-activate-office-timing)
- [Common Error Responses](#common-error-responses)
- [Status Codes Summary](#status-codes-summary)
- [Day of Week Reference](#day-of-week-reference)
- [Response Fields Reference](#response-fields-reference)

---

## Office Timings

### 1. List All Office Timings

```
GET /api/v1/schedule/office-timings
```

Returns a flat list of office timings (without day details). Use [Get by ID](#2-get-office-timing-by-id) to retrieve full day schedules.

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/schedule/office-timings \
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
      "branch_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Standard Week",
      "is_active": true,
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    },
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440002",
      "branch_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Night Shift",
      "is_active": true,
      "created_at": "2026-04-02T10:00:00Z",
      "updated_at": "2026-04-02T10:00:00Z"
    }
  ]
}
```

#### ✅ 200 OK — No records

```json
{
  "success": true,
  "data": null
}
```

---

### 2. Get Office Timing by ID

```
GET /api/v1/schedule/office-timings/{id}
```

Returns the office timing along with its full day-by-day schedule.

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/schedule/office-timings/aa0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "aa0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Standard Week",
    "is_active": true,
    "created_at": "2026-04-01T10:00:00Z",
    "updated_at": "2026-04-01T10:00:00Z",
    "days": [
      {
        "id": "bb0e8400-e29b-41d4-a716-446655440001",
        "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001",
        "day_of_week": 0,
        "is_working_day": false,
        "start_time": null,
        "end_time": null,
        "break_minutes": 0
      },
      {
        "id": "bb0e8400-e29b-41d4-a716-446655440002",
        "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001",
        "day_of_week": 1,
        "is_working_day": true,
        "start_time": "09:00:00",
        "end_time": "18:00:00",
        "break_minutes": 60
      },
      {
        "id": "bb0e8400-e29b-41d4-a716-446655440003",
        "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001",
        "day_of_week": 2,
        "is_working_day": true,
        "start_time": "09:00:00",
        "end_time": "18:00:00",
        "break_minutes": 60
      }
    ]
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
  "error": "office timing not found"
}
```

---

### 3. Create Office Timing

```
POST /api/v1/schedule/office-timings
```

Creates a new office timing for the caller's branch. Each day entry must be unique per `day_of_week` (0–6). Include all 7 days or only working days — your choice.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | ✅ | Timing label (e.g. `"Standard Week"`, `"Night Shift"`) |
| `days` | array | ✅ | Array of day schedule objects (at least 1) |

**Day object fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `day_of_week` | integer | ✅ | `0`=Sunday … `6`=Saturday |
| `is_working_day` | boolean | ✅ | Whether this is a working day |
| `start_time` | string | ❌ | Work start time in `"HH:MM:SS"` format |
| `end_time` | string | ❌ | Work end time in `"HH:MM:SS"` format |
| `break_minutes` | integer | ❌ | Break duration in minutes (default `0`) |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/schedule/office-timings \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "name": "Standard Week",
    "days": [
      { "day_of_week": 0, "is_working_day": false },
      { "day_of_week": 1, "is_working_day": true, "start_time": "09:00:00", "end_time": "18:00:00", "break_minutes": 60 },
      { "day_of_week": 2, "is_working_day": true, "start_time": "09:00:00", "end_time": "18:00:00", "break_minutes": 60 },
      { "day_of_week": 3, "is_working_day": true, "start_time": "09:00:00", "end_time": "18:00:00", "break_minutes": 60 },
      { "day_of_week": 4, "is_working_day": true, "start_time": "09:00:00", "end_time": "18:00:00", "break_minutes": 60 },
      { "day_of_week": 5, "is_working_day": true, "start_time": "09:00:00", "end_time": "18:00:00", "break_minutes": 60 },
      { "day_of_week": 6, "is_working_day": false }
    ]
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "id": "aa0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Standard Week",
    "is_active": true,
    "created_at": "2026-04-14T10:00:00Z",
    "updated_at": "2026-04-14T10:00:00Z",
    "days": [
      { "id": "bb0e8400-e29b-41d4-a716-446655440001", "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001", "day_of_week": 0, "is_working_day": false, "start_time": null, "end_time": null, "break_minutes": 0 },
      { "id": "bb0e8400-e29b-41d4-a716-446655440002", "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001", "day_of_week": 1, "is_working_day": true, "start_time": "09:00:00", "end_time": "18:00:00", "break_minutes": 60 },
      { "id": "bb0e8400-e29b-41d4-a716-446655440003", "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001", "day_of_week": 2, "is_working_day": true, "start_time": "09:00:00", "end_time": "18:00:00", "break_minutes": 60 },
      { "id": "bb0e8400-e29b-41d4-a716-446655440004", "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001", "day_of_week": 3, "is_working_day": true, "start_time": "09:00:00", "end_time": "18:00:00", "break_minutes": 60 },
      { "id": "bb0e8400-e29b-41d4-a716-446655440005", "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001", "day_of_week": 4, "is_working_day": true, "start_time": "09:00:00", "end_time": "18:00:00", "break_minutes": 60 },
      { "id": "bb0e8400-e29b-41d4-a716-446655440006", "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001", "day_of_week": 5, "is_working_day": true, "start_time": "09:00:00", "end_time": "18:00:00", "break_minutes": 60 },
      { "id": "bb0e8400-e29b-41d4-a716-446655440007", "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001", "day_of_week": 6, "is_working_day": false, "start_time": null, "end_time": null, "break_minutes": 0 }
    ]
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
  "error": "Key: 'CreateOfficeTimingRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"
}
```

---

### 4. Update Office Timing

```
PUT /api/v1/schedule/office-timings/{id}
```

Replaces the timing name and **all** its days atomically. Any existing days are deleted and replaced with the new set.

> `admin` and `manager` can only update timings that belong to their own branch.

#### Request Body

Same structure as [Create](#3-create-office-timing).

#### cURL

```bash
curl -X PUT http://localhost:8080/api/v1/schedule/office-timings/aa0e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "name": "Standard Week (Updated)",
    "days": [
      { "day_of_week": 0, "is_working_day": false },
      { "day_of_week": 1, "is_working_day": true, "start_time": "08:30:00", "end_time": "17:30:00", "break_minutes": 45 },
      { "day_of_week": 2, "is_working_day": true, "start_time": "08:30:00", "end_time": "17:30:00", "break_minutes": 45 },
      { "day_of_week": 3, "is_working_day": true, "start_time": "08:30:00", "end_time": "17:30:00", "break_minutes": 45 },
      { "day_of_week": 4, "is_working_day": true, "start_time": "08:30:00", "end_time": "17:30:00", "break_minutes": 45 },
      { "day_of_week": 5, "is_working_day": true, "start_time": "08:30:00", "end_time": "17:30:00", "break_minutes": 45 },
      { "day_of_week": 6, "is_working_day": false }
    ]
  }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "aa0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Standard Week (Updated)",
    "is_active": true,
    "created_at": "2026-04-14T10:00:00Z",
    "updated_at": "2026-04-14T11:00:00Z",
    "days": [
      { "id": "cc0e8400-e29b-41d4-a716-446655440001", "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001", "day_of_week": 0, "is_working_day": false, "start_time": null, "end_time": null, "break_minutes": 0 },
      { "id": "cc0e8400-e29b-41d4-a716-446655440002", "office_timing_id": "aa0e8400-e29b-41d4-a716-446655440001", "day_of_week": 1, "is_working_day": true, "start_time": "08:30:00", "end_time": "17:30:00", "break_minutes": 45 }
    ]
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
  "error": "failed to update office timing"
}
```

---

### 5. Delete Office Timing

```
DELETE /api/v1/schedule/office-timings/{id}
```

Deletes the office timing and all its days (cascade). If this timing is currently assigned as a branch's active timing, `branches.office_timing_id` will be set to `null` automatically (via `ON DELETE SET NULL`).

> `admin` and `manager` can only delete timings from their own branch.

#### cURL

```bash
curl -X DELETE http://localhost:8080/api/v1/schedule/office-timings/aa0e8400-e29b-41d4-a716-446655440001 \
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
  "error": "failed to delete office timing"
}
```

---

### 6. Activate Office Timing

```
PUT /api/v1/schedule/office-timings/{id}/activate
```

Sets the specified timing as the **active schedule** for its branch by updating `branches.office_timing_id`. A branch can only have one active timing at a time — calling this replaces any previously active one.

> `admin` and `manager` can only activate timings that belong to their own branch.

#### cURL

```bash
curl -X PUT http://localhost:8080/api/v1/schedule/office-timings/aa0e8400-e29b-41d4-a716-446655440001/activate \
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

#### ❌ 404 Not Found

```json
{
  "success": false,
  "error": "office timing not found"
}
```

#### ❌ 500 Internal Server Error

```json
{
  "success": false,
  "error": "failed to activate office timing"
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
| `201` | Created | Office timing created successfully |
| `400` | Bad Request | Malformed JSON body |
| `401` | Unauthorized | Missing/invalid API key or JWT |
| `403` | Forbidden | Role is not `super_admin`, `admin`, or `manager`; or timing belongs to another branch |
| `404` | Not Found | Office timing ID does not exist |
| `422` | Unprocessable Entity | Validation failed (e.g. missing required field) |
| `500` | Internal Server Error | Database error or unexpected failure |

---

## Day of Week Reference

| Value | Day |
|---|---|
| `0` | Sunday |
| `1` | Monday |
| `2` | Tuesday |
| `3` | Wednesday |
| `4` | Thursday |
| `5` | Friday |
| `6` | Saturday |

---

## Response Fields Reference

### OfficeTiming

| Field | Type | Description |
|---|---|---|
| `id` | string | Office timing UUID |
| `branch_id` | string | Branch this timing belongs to |
| `name` | string | Timing label |
| `is_active` | boolean | Whether this timing is enabled |
| `created_at` | string | Creation timestamp |
| `updated_at` | string | Last update timestamp |
| `days` | array | Day schedules (only included in Get by ID, Create, Update responses) |

### OfficeTimingDay

| Field | Type | Description |
|---|---|---|
| `id` | string | Day record UUID |
| `office_timing_id` | string | Parent timing UUID |
| `day_of_week` | integer | Day index: `0`=Sunday … `6`=Saturday |
| `is_working_day` | boolean | Whether employees work on this day |
| `start_time` | string \| null | Work start time `"HH:MM:SS"` (`null` if non-working day) |
| `end_time` | string \| null | Work end time `"HH:MM:SS"` (`null` if non-working day) |
| `break_minutes` | integer | Break duration in minutes |
