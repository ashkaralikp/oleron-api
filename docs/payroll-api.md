# Payroll API Documentation

Base URL: `http://localhost:8080/api/v1`

> **All endpoints require:**
> - `X-API-Key` header
> - `Authorization: Bearer <access_token>` header
>
> | Route prefix | Required role |
> |---|---|
> | `/api/v1/payroll/*` | `super_admin`, `admin`, `manager` |

> **Role-based filtering:**
> | Role | Access |
> |---|---|
> | `super_admin` | All payroll runs across all branches |
> | `admin` | Payroll runs for their branch only |
> | `manager` | Payroll runs for their branch only |

> **How payroll is calculated:**
> 1. All active employees in the branch are fetched
> 2. Attendance records for the period are aggregated per employee
> 3. Expected working days are derived from the branch's `office_timings`, minus holidays in `branch_calendar`
> 4. `gross_pay = total_hours × hourly_rate`
> 5. `net_pay = gross_pay − deductions` (deductions default to `0`)

---

## Table of Contents

- [Payroll Runs](#payroll-runs)
  - [1. List All Payroll Runs](#1-list-all-payroll-runs)
  - [2. Get Payroll Run by ID](#2-get-payroll-run-by-id)
  - [3. Generate Payroll](#3-generate-payroll)
  - [4. Update Status](#4-update-status)
  - [5. Delete Payroll Run](#5-delete-payroll-run)
- [Common Error Responses](#common-error-responses)
- [Status Codes Summary](#status-codes-summary)
- [Payroll Status Reference](#payroll-status-reference)
- [Response Fields Reference](#response-fields-reference)

---

## Payroll Runs

### 1. List All Payroll Runs

```
GET /api/v1/payroll
```

Returns all payroll runs for the branch (summary only, no items). Ordered by most recent first.

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/payroll \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": [
    {
      "id": "pr0e8400-e29b-41d4-a716-446655440001",
      "branch_id": "550e8400-e29b-41d4-a716-446655440000",
      "period_from": "2026-04-01T00:00:00Z",
      "period_to": "2026-04-30T00:00:00Z",
      "generated_by": "880e8400-e29b-41d4-a716-446655440002",
      "status": "draft",
      "total_amount": 12500.00,
      "currency": "USD",
      "notes": null,
      "created_at": "2026-04-17T09:00:00Z",
      "updated_at": "2026-04-17T09:00:00Z"
    }
  ]
}
```

#### ✅ 200 OK — No runs

```json
{
  "success": true,
  "data": null
}
```

---

### 2. Get Payroll Run by ID

```
GET /api/v1/payroll/{id}
```

Returns the payroll run with all employee-level items.

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/payroll/pr0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "pr0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "period_from": "2026-04-01T00:00:00Z",
    "period_to": "2026-04-30T00:00:00Z",
    "generated_by": "880e8400-e29b-41d4-a716-446655440002",
    "status": "draft",
    "total_amount": 12500.00,
    "currency": "USD",
    "notes": "April 2026 payroll",
    "created_at": "2026-04-17T09:00:00Z",
    "updated_at": "2026-04-17T09:00:00Z",
    "items": [
      {
        "id": "pi0e8400-e29b-41d4-a716-446655440001",
        "payroll_run_id": "pr0e8400-e29b-41d4-a716-446655440001",
        "employee_id": "ee0e8400-e29b-41d4-a716-446655440001",
        "user_id": "ff0e8400-e29b-41d4-a716-446655440001",
        "first_name": "Jane",
        "last_name": "Smith",
        "email": "jane.smith@oleron.com",
        "employee_code": "EMP001",
        "working_days": 22,
        "present_days": 20,
        "absent_days": 1,
        "leave_days": 1,
        "total_hours": 176.50,
        "hourly_rate": 25.00,
        "currency": "USD",
        "gross_pay": 4412.50,
        "deductions": 0.00,
        "net_pay": 4412.50,
        "created_at": "2026-04-17T09:00:00Z"
      },
      {
        "id": "pi0e8400-e29b-41d4-a716-446655440002",
        "payroll_run_id": "pr0e8400-e29b-41d4-a716-446655440001",
        "employee_id": "ee0e8400-e29b-41d4-a716-446655440002",
        "user_id": "ff0e8400-e29b-41d4-a716-446655440002",
        "first_name": "John",
        "last_name": "Doe",
        "email": "john.doe@oleron.com",
        "employee_code": "EMP002",
        "working_days": 22,
        "present_days": 22,
        "absent_days": 0,
        "leave_days": 0,
        "total_hours": 185.00,
        "hourly_rate": 30.00,
        "currency": "USD",
        "gross_pay": 5550.00,
        "deductions": 0.00,
        "net_pay": 5550.00,
        "created_at": "2026-04-17T09:00:00Z"
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
  "error": "payroll run not found"
}
```

---

### 3. Generate Payroll

```
POST /api/v1/payroll/generate
```

Generates a new payroll run for the caller's branch covering the specified period. Fetches all active employees, aggregates their attendance, and computes pay automatically.

> - A duplicate run for the same `branch_id + period_from + period_to` will fail.
> - The run is created with status `draft`. Review before approving.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `period_from` | string | ✅ | Pay period start date `YYYY-MM-DD` |
| `period_to` | string | ✅ | Pay period end date `YYYY-MM-DD` |
| `currency` | string | ❌ | 3-letter currency code (default: `USD`) |
| `notes` | string | ❌ | Optional label e.g. `"April 2026 payroll"` |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/payroll/generate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "period_from": "2026-04-01",
    "period_to": "2026-04-30",
    "currency": "USD",
    "notes": "April 2026 payroll"
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "id": "pr0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "period_from": "2026-04-01T00:00:00Z",
    "period_to": "2026-04-30T00:00:00Z",
    "generated_by": "880e8400-e29b-41d4-a716-446655440002",
    "status": "draft",
    "total_amount": 9962.50,
    "currency": "USD",
    "notes": "April 2026 payroll",
    "created_at": "2026-04-17T09:00:00Z",
    "updated_at": "2026-04-17T09:00:00Z",
    "items": [...]
  }
}
```

#### ❌ 422 Unprocessable Entity — No employees

```json
{
  "success": false,
  "error": "no active employees found in this branch"
}
```

#### ❌ 422 Unprocessable Entity — Validation failed

```json
{
  "success": false,
  "error": "Key: 'GeneratePayrollRequest.PeriodFrom' Error:Field validation for 'PeriodFrom' failed on the 'required' tag"
}
```

#### ❌ 500 Internal Server Error — Duplicate period

```json
{
  "success": false,
  "error": "failed to generate payroll"
}
```

> ⚠️ A payroll run for the same branch and date range can only be created once. Delete the existing draft first if you need to regenerate.

---

### 4. Update Status

```
PATCH /api/v1/payroll/{id}/status
```

Moves a payroll run through its lifecycle. Status transitions are enforced:

```
draft → approved → paid
```

> `admin` and `manager` can only update runs from their own branch.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `status` | string | ✅ | One of: `approved`, `paid` |

#### cURL — Approve

```bash
curl -X PATCH http://localhost:8080/api/v1/payroll/pr0e8400-e29b-41d4-a716-446655440001/status \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{ "status": "approved" }'
```

#### cURL — Mark as Paid

```bash
curl -X PATCH http://localhost:8080/api/v1/payroll/pr0e8400-e29b-41d4-a716-446655440001/status \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{ "status": "paid" }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "pr0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "period_from": "2026-04-01T00:00:00Z",
    "period_to": "2026-04-30T00:00:00Z",
    "generated_by": "880e8400-e29b-41d4-a716-446655440002",
    "status": "approved",
    "total_amount": 9962.50,
    "currency": "USD",
    "notes": "April 2026 payroll",
    "created_at": "2026-04-17T09:00:00Z",
    "updated_at": "2026-04-17T09:30:00Z"
  }
}
```

#### ❌ 422 Unprocessable Entity — Invalid transition

```json
{
  "success": false,
  "error": "only draft runs can be approved"
}
```

```json
{
  "success": false,
  "error": "only approved runs can be marked as paid"
}
```

```json
{
  "success": false,
  "error": "payroll run is already marked as paid"
}
```

---

### 5. Delete Payroll Run

```
DELETE /api/v1/payroll/{id}
```

Deletes a payroll run and all its items. **Only `draft` runs can be deleted.** Approved or paid runs are permanent records.

#### cURL

```bash
curl -X DELETE http://localhost:8080/api/v1/payroll/pr0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true
}
```

#### ❌ 422 Unprocessable Entity

```json
{
  "success": false,
  "error": "only draft payroll runs can be deleted"
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
| `200` | OK | Run fetched or status updated |
| `201` | Created | Payroll generated successfully |
| `400` | Bad Request | Malformed JSON body |
| `401` | Unauthorized | Missing/invalid API key or JWT |
| `403` | Forbidden | Role lacks access or run belongs to another branch |
| `404` | Not Found | Payroll run ID does not exist |
| `422` | Unprocessable Entity | Validation failure, no employees, or invalid status transition |
| `500` | Internal Server Error | Database error (e.g. duplicate period) |

---

## Payroll Status Reference

| Status | Description | Can edit | Can delete |
|---|---|---|---|
| `draft` | Generated, pending review | ✅ | ✅ |
| `approved` | Reviewed and approved | ❌ | ❌ |
| `paid` | Payment processed | ❌ | ❌ |

---

## Response Fields Reference

### PayrollRun

| Field | Type | Description |
|---|---|---|
| `id` | string | Payroll run UUID |
| `branch_id` | string | Branch this run belongs to |
| `period_from` | string | Pay period start date |
| `period_to` | string | Pay period end date |
| `generated_by` | string | UUID of the user who generated the run |
| `status` | string | `draft`, `approved`, or `paid` |
| `total_amount` | number | Sum of all employee `net_pay` values |
| `currency` | string | 3-letter currency code |
| `notes` | string \| null | Optional label |
| `items` | array | Employee pay items (only in Get by ID and Generate responses) |
| `created_at` | string | Generation timestamp |
| `updated_at` | string | Last status update timestamp |

### PayrollItem

| Field | Type | Description |
|---|---|---|
| `id` | string | Item UUID |
| `payroll_run_id` | string | Parent run UUID |
| `employee_id` | string | Employee record UUID |
| `user_id` | string | User account UUID |
| `first_name` | string | Employee first name |
| `last_name` | string | Employee last name |
| `email` | string | Employee email |
| `employee_code` | string | Punch code |
| `working_days` | integer | Expected working days in the period (from office timing, minus holidays) |
| `present_days` | integer | Days with a valid punch-in |
| `absent_days` | integer | Days with no punch-in |
| `leave_days` | integer | Days marked `on_leave` |
| `total_hours` | number | Sum of `work_hours` across all attendance records in the period |
| `hourly_rate` | number | Rate at time of payroll generation |
| `currency` | string | Employee's currency |
| `gross_pay` | number | `total_hours × hourly_rate` |
| `deductions` | number | Deductions applied (default `0`) |
| `net_pay` | number | `gross_pay − deductions` |
| `created_at` | string | Item creation timestamp |
