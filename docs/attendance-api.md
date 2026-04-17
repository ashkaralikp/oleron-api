# Attendance API Documentation

Base URL: `http://localhost:8080/api/v1`

> **All endpoints require:**
> - `X-API-Key` header
> - `Authorization: Bearer <access_token>` header
>
> | Route prefix | Required role |
> |---|---|
> | `/api/v1/attendance/*` | All authenticated roles (`super_admin`, `admin`, `manager`, `employee`) |

---

## Table of Contents

- [Punch In / Punch Out](#punch-in--punch-out)
  - [1. Punch](#1-punch)
- [Today's Status](#todays-status)
  - [2. Get Today's Attendance](#2-get-todays-attendance)
- [Common Error Responses](#common-error-responses)
- [Status Codes Summary](#status-codes-summary)
- [Attendance Status Values](#attendance-status-values)
- [Response Fields Reference](#response-fields-reference)

---

## Punch In / Punch Out

### 1. Punch

```
POST /api/v1/attendance/punch
```

**Smart endpoint** — no body required. The server decides the action based on the caller's record for today:

| State | Action taken |
|---|---|
| No record for today | **Punch in** — creates a record, sets `punch_in = now` |
| Record exists, no `punch_out` | **Punch out** — sets `punch_out = now`, computes `work_hours` and final status |
| Record exists, `punch_out` already set | Returns `409 Conflict` |

> **Status is automatically calculated** from the employee's assigned office timing and branch calendar:
> - Compared against `start_time` with a 15-minute grace period on punch-in
> - Compared against `end_time` and expected hours on punch-out
> - If today is a branch holiday → `422` error, punch blocked

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/attendance/punch \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 201 Created — Punch In

```json
{
  "success": true,
  "data": {
    "action": "punch_in",
    "id": "at0e8400-e29b-41d4-a716-446655440001",
    "user_id": "ff0e8400-e29b-41d4-a716-446655440001",
    "work_date": "2026-04-16T00:00:00Z",
    "punch_in": "2026-04-16T09:03:00Z",
    "punch_out": null,
    "work_hours": null,
    "status": "present",
    "created_at": "2026-04-16T09:03:00Z",
    "updated_at": "2026-04-16T09:03:00Z"
  }
}
```

> `status` is `"present"` if punched in within the 15-minute grace window, `"late_in"` otherwise.

#### ✅ 200 OK — Punch Out

```json
{
  "success": true,
  "data": {
    "action": "punch_out",
    "id": "at0e8400-e29b-41d4-a716-446655440001",
    "user_id": "ff0e8400-e29b-41d4-a716-446655440001",
    "work_date": "2026-04-16T00:00:00Z",
    "punch_in": "2026-04-16T09:03:00Z",
    "punch_out": "2026-04-16T18:01:00Z",
    "work_hours": 8.97,
    "status": "present",
    "created_at": "2026-04-16T09:03:00Z",
    "updated_at": "2026-04-16T18:01:00Z"
  }
}
```

#### ✅ 200 OK — Punch Out (late in + early out)

```json
{
  "success": true,
  "data": {
    "action": "punch_out",
    "id": "at0e8400-e29b-41d4-a716-446655440002",
    "user_id": "ff0e8400-e29b-41d4-a716-446655440002",
    "work_date": "2026-04-16T00:00:00Z",
    "punch_in": "2026-04-16T09:45:00Z",
    "punch_out": "2026-04-16T16:30:00Z",
    "work_hours": 6.75,
    "status": "late_in_early_out",
    "created_at": "2026-04-16T09:45:00Z",
    "updated_at": "2026-04-16T16:30:00Z"
  }
}
```

#### ❌ 409 Conflict — Already Punched Out

```json
{
  "success": false,
  "error": "already punched out for today"
}
```

#### ❌ 422 Unprocessable Entity — Public Holiday

```json
{
  "success": false,
  "error": "today is a public holiday"
}
```

#### ❌ 422 Unprocessable Entity — Non-working Day

```json
{
  "success": false,
  "error": "today is not a working day"
}
```

#### ❌ 500 Internal Server Error

```json
{
  "success": false,
  "error": "failed to process punch"
}
```

---

## Today's Status

### 2. Get Today's Attendance

```
GET /api/v1/attendance/today
```

Returns the calling user's attendance record for today. If no punch has been made yet, returns `{ "punched_in": false, "punched_out": false }`.

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/attendance/today \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK — Punched in, not yet out

```json
{
  "success": true,
  "data": {
    "punched_in": true,
    "punched_out": false,
    "id": "at0e8400-e29b-41d4-a716-446655440001",
    "user_id": "ff0e8400-e29b-41d4-a716-446655440001",
    "work_date": "2026-04-16T00:00:00Z",
    "punch_in": "2026-04-16T09:03:00Z",
    "punch_out": null,
    "work_hours": null,
    "status": "present"
  }
}
```

#### ✅ 200 OK — Fully punched out

```json
{
  "success": true,
  "data": {
    "punched_in": true,
    "punched_out": true,
    "id": "at0e8400-e29b-41d4-a716-446655440001",
    "user_id": "ff0e8400-e29b-41d4-a716-446655440001",
    "work_date": "2026-04-16T00:00:00Z",
    "punch_in": "2026-04-16T09:03:00Z",
    "punch_out": "2026-04-16T18:01:00Z",
    "work_hours": 8.97,
    "status": "present"
  }
}
```

#### ✅ 200 OK — Not yet punched in

```json
{
  "success": true,
  "data": {
    "punched_in": false,
    "punched_out": false
  }
}
```

#### ❌ 500 Internal Server Error

```json
{
  "success": false,
  "error": "failed to fetch today's attendance"
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

---

## Status Codes Summary

| Code | Meaning | When |
|---|---|---|
| `200` | OK | Punch out succeeded or today's record fetched |
| `201` | Created | Punch in succeeded (new record created) |
| `401` | Unauthorized | Missing/invalid API key or JWT |
| `409` | Conflict | Employee already punched out for today |
| `422` | Unprocessable Entity | Today is a public holiday or non-working day |
| `500` | Internal Server Error | Database error or unexpected failure |

---

## Attendance Status Values

Status is set automatically on punch-in and updated on punch-out based on the employee's office timing.

| Status | Set on | Condition |
|---|---|---|
| `present` | Punch-in | Punched in within 15-minute grace period of `start_time` |
| `late_in` | Punch-in | Punched in more than 15 minutes after `start_time` |
| `early_out` | Punch-out | Punched out before `end_time` (was `present`) |
| `late_in_early_out` | Punch-out | Punched out before `end_time` (was `late_in`) |
| `half_day` | Punch-out | `work_hours` < half of expected hours for the day |
| `absent` | — | Default for records created manually; no punch-in made |
| `on_leave` | — | Set manually by admin/manager via attendance management |

> **Grace period:** 15 minutes. Punch-in at `09:00` start time is `present` up to `09:15`, `late_in` after.

> **Status priority on punch-out:** `half_day` is checked first (overrides everything), then `early_out` / `late_in_early_out`.

---

## Response Fields Reference

### Punch Response

| Field | Type | Description |
|---|---|---|
| `action` | string | `"punch_in"` or `"punch_out"` |
| `id` | string | Attendance record UUID |
| `user_id` | string | The user who punched |
| `work_date` | string | Date of the attendance record |
| `punch_in` | string \| null | Punch-in timestamp |
| `punch_out` | string \| null | Punch-out timestamp (`null` until punched out) |
| `work_hours` | number \| null | Total hours worked, computed on punch-out |
| `status` | string | Current attendance status |
| `created_at` | string | Record creation timestamp |
| `updated_at` | string | Last update timestamp |

### Today Response

| Field | Type | Description |
|---|---|---|
| `punched_in` | boolean | Whether the user has punched in today |
| `punched_out` | boolean | Whether the user has punched out today |
| `id` | string \| omitted | Attendance record UUID (omitted if not punched in) |
| `user_id` | string \| omitted | User UUID (omitted if not punched in) |
| `work_date` | string \| omitted | Date (omitted if not punched in) |
| `punch_in` | string \| omitted | Punch-in timestamp |
| `punch_out` | string \| omitted | Punch-out timestamp |
| `work_hours` | number \| omitted | Hours worked so far |
| `status` | string \| omitted | Current status |
| `notes` | string \| omitted | Optional notes on the record |
