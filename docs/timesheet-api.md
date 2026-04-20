# Timesheet API Documentation

Base URL: `http://localhost:8080/api/v1`

> **All endpoints require:**
> - `X-API-Key` header
> - `Authorization: Bearer <access_token>` header

> | Route | Allowed roles |
> |---|---|
> | `POST /api/v1/timesheets/estimate` | All authenticated roles |
> | `POST /api/v1/timesheets` | `consultant` |
> | `GET /api/v1/timesheets/me?year={year}&month={month}` | `consultant` |
> | `GET /api/v1/timesheets` | `super_admin`, `admin`, `manager` |
> | `GET /api/v1/timesheets/{id}` | `super_admin`, `admin`, `manager` |
> | `PATCH /api/v1/timesheets/{id}/review` | `super_admin`, `admin`, `manager` |

---

## Table of Contents

- [1. Calculate Pay Estimate](#1-calculate-pay-estimate)
- [2. Submit Timesheet](#2-submit-timesheet)
- [3. View My Timesheet](#3-view-my-timesheet)
- [4. List All Timesheets](#4-list-all-timesheets)
- [5. Get Timesheet by ID](#5-get-timesheet-by-id)
- [6. Review Timesheet](#6-review-timesheet)
- [Common Error Responses](#common-error-responses)
- [Status Codes Summary](#status-codes-summary)

---

## 1. Calculate Pay Estimate

```
POST /api/v1/timesheets/estimate
```

Pure calculation — no data stored. The caller's `fixed_monthly_salary` and `ot_rate` are fetched automatically from their employee record.

> **Pay scenarios — determined by comparing `support_hours` vs `whole_month_hours` (weekdays × 8h):**
>
> | Scenario | Condition | Formula |
> |---|---|---|
> | `Full` | Hours worked = whole month, no OT | `fixed_monthly_salary` |
> | `Over` | Hours worked ≥ whole month AND OT hours > 0 | `fixed_monthly_salary + (ot_hours × ot_rate)` |
> | `Partial` | Hours worked < whole month | `support_hours × hourly_rate` |
>
> `whole_month_hours` = actual weekdays (Mon–Fri) in the given month × 8.  
> `hourly_rate` = `fixed_monthly_salary / whole_month_hours`.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `year` | integer | ✅ | Calendar year (e.g. `2026`) |
| `month` | integer | ✅ | Month number `1–12` |
| `support_hours` | number | ✅ | Total regular working hours logged this month |
| `overtime_hours` | number | ✅ | Total overtime hours logged this month |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/timesheets/estimate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "year": 2026,
    "month": 2,
    "support_hours": 160.0,
    "overtime_hours": 8.0
  }'
```

#### ✅ 200 OK — Scenario: Over

```json
{
  "success": true,
  "data": {
    "year": 2026,
    "month": 2,
    "whole_month_hours": 160.0,
    "support_hours": 160.0,
    "overtime_hours": 8.0,
    "fixed_monthly_salary": 183333.33,
    "ot_rate": 500.0,
    "hourly_rate": 1145.83,
    "scenario": "Over",
    "estimated_pay": 187333.33,
    "currency": "INR"
  }
}
```

#### ✅ 200 OK — Scenario: Full

```json
{
  "success": true,
  "data": {
    "year": 2026,
    "month": 2,
    "whole_month_hours": 160.0,
    "support_hours": 160.0,
    "overtime_hours": 0.0,
    "fixed_monthly_salary": 183333.33,
    "ot_rate": 500.0,
    "hourly_rate": 1145.83,
    "scenario": "Full",
    "estimated_pay": 183333.33,
    "currency": "INR"
  }
}
```

#### ✅ 200 OK — Scenario: Partial

```json
{
  "success": true,
  "data": {
    "year": 2026,
    "month": 2,
    "whole_month_hours": 160.0,
    "support_hours": 8.0,
    "overtime_hours": 0.0,
    "fixed_monthly_salary": 183333.33,
    "ot_rate": 500.0,
    "hourly_rate": 1145.83,
    "scenario": "Partial",
    "estimated_pay": 9166.64,
    "currency": "INR"
  }
}
```

#### Pay Scenario Examples

> Assuming: `fixed_monthly_salary = 183,333.33`, `ot_rate = 500`, February 2026 (20 weekdays → `whole_month_hours = 160`)

| support_hours | overtime_hours | Scenario | estimated_pay | Formula used |
|---|---|---|---|---|
| `160.0` | `0.0` | `Full` | `183,333.33` | Fixed monthly salary |
| `160.0` | `8.0` | `Over` | `187,333.33` | `183,333.33 + (8 × 500)` |
| `120.0` | `0.0` | `Partial` | `137,499.60` | `120 × 1,145.83` |
| `8.0` | `0.0` | `Partial` | `9,166.64` | `8 × 1,145.83` |
| `80.0` | `5.0` | `Partial` | `91,666.40` | `80 × 1,145.83` (OT ignored when partial) |

---

## 2. Submit Timesheet

```
POST /api/v1/timesheets
```

Consultant submits their monthly timesheet. Submitting the same month again resets the status to `pending` and overwrites previous values.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `year` | integer | ✅ | Calendar year (e.g. `2026`) |
| `month` | integer | ✅ | Month number `1–12` |
| `support_hours` | number | ✅ | Total regular working hours |
| `overtime_hours` | number | ✅ | Total overtime hours |
| `notes` | string | ❌ | Optional notes to the reviewer |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/timesheets \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -H "Authorization: Bearer <consultant_access_token>" \
  -d '{
    "year": 2026,
    "month": 4,
    "support_hours": 152.0,
    "overtime_hours": 4.0,
    "notes": "Worked from client site for week 2"
  }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-...",
    "employee_id": "e5f6g7h8-...",
    "employee_code": "CON001",
    "first_name": "John",
    "last_name": "Doe",
    "year": 2026,
    "month": 4,
    "support_hours": 152.0,
    "overtime_hours": 4.0,
    "notes": "Worked from client site for week 2",
    "status": "pending",
    "submitted_at": "2026-04-20T10:30:00Z"
  }
}
```

#### ❌ 403 Forbidden — Non-consultant role

```json
{
  "success": false,
  "error": "Forbidden"
}
```

---

## 3. View My Timesheet

```
GET /api/v1/timesheets/me?year={year}&month={month}
```

Returns the consultant's submitted timesheet for the given month. Both `year` and `month` are required query parameters.

#### Query Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `year` | integer | ✅ | Calendar year (e.g. `2026`) |
| `month` | integer | ✅ | Month number `1–12` |

#### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/timesheets/me?year=2026&month=4" \
  -H "X-API-Key: your-api-key" \
  -H "Authorization: Bearer <consultant_access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-...",
    "employee_id": "e5f6g7h8-...",
    "employee_code": "CON001",
    "first_name": "John",
    "last_name": "Doe",
    "year": 2026,
    "month": 4,
    "support_hours": 152.0,
    "overtime_hours": 4.0,
    "notes": "Worked from client site for week 2",
    "status": "approved",
    "reviewer_id": "u9v0w1x2-...",
    "review_note": "All good.",
    "reviewed_at": "2026-04-21T09:00:00Z",
    "submitted_at": "2026-04-20T10:30:00Z"
  }
}
```

#### ❌ 400 Bad Request — Missing or invalid query params

```json
{
  "success": false,
  "error": "year and month query parameters are required (e.g. ?year=2026&month=4)"
}
```

#### ❌ 404 Not Found — No timesheet for that month

```json
{
  "success": false,
  "error": "timesheet not found for the given month"
}
```

---

## 4. List All Timesheets

```
GET /api/v1/timesheets
```

Returns all consultant timesheets. `admin` and `manager` see only their branch; `super_admin` sees all.

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/timesheets \
  -H "X-API-Key: your-api-key" \
  -H "Authorization: Bearer <admin_access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": [
    {
      "id": "a1b2c3d4-...",
      "employee_id": "e5f6g7h8-...",
      "employee_code": "CON001",
      "first_name": "John",
      "last_name": "Doe",
      "year": 2026,
      "month": 4,
      "support_hours": 152.0,
      "overtime_hours": 4.0,
      "notes": "Worked from client site for week 2",
      "status": "pending",
      "submitted_at": "2026-04-20T10:30:00Z"
    }
  ]
}
```

---

## 5. Get Timesheet by ID

```
GET /api/v1/timesheets/{id}
```

Returns a single timesheet. `admin` and `manager` are restricted to their branch.

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/timesheets/a1b2c3d4-... \
  -H "X-API-Key: your-api-key" \
  -H "Authorization: Bearer <admin_access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-...",
    "employee_id": "e5f6g7h8-...",
    "employee_code": "CON001",
    "first_name": "John",
    "last_name": "Doe",
    "year": 2026,
    "month": 4,
    "support_hours": 152.0,
    "overtime_hours": 4.0,
    "notes": "Worked from client site for week 2",
    "status": "pending",
    "submitted_at": "2026-04-20T10:30:00Z"
  }
}
```

#### ❌ 404 Not Found

```json
{
  "success": false,
  "error": "timesheet not found"
}
```

---

## 6. Review Timesheet

```
PATCH /api/v1/timesheets/{id}/review
```

Approve or reject a consultant's timesheet. Records the reviewer's identity automatically from the JWT.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `status` | string | ✅ | `approved` or `rejected` |
| `review_note` | string | ❌ | Optional feedback to the consultant |

#### cURL — Approve

```bash
curl -X PATCH http://localhost:8080/api/v1/timesheets/a1b2c3d4-.../review \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -H "Authorization: Bearer <admin_access_token>" \
  -d '{
    "status": "approved",
    "review_note": "Hours verified. Good work."
  }'
```

#### cURL — Reject

```bash
curl -X PATCH http://localhost:8080/api/v1/timesheets/a1b2c3d4-.../review \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -H "Authorization: Bearer <admin_access_token>" \
  -d '{
    "status": "rejected",
    "review_note": "Support hours exceed working days in month. Please resubmit."
  }'
```

#### ✅ 200 OK — Approved

```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-...",
    "employee_id": "e5f6g7h8-...",
    "employee_code": "CON001",
    "first_name": "John",
    "last_name": "Doe",
    "year": 2026,
    "month": 4,
    "support_hours": 152.0,
    "overtime_hours": 4.0,
    "notes": "Worked from client site for week 2",
    "status": "approved",
    "reviewer_id": "u9v0w1x2-...",
    "review_note": "Hours verified. Good work.",
    "reviewed_at": "2026-04-21T09:00:00Z",
    "submitted_at": "2026-04-20T10:30:00Z"
  }
}
```

#### ❌ 400 Bad Request — Invalid status value

```json
{
  "success": false,
  "error": "Key: 'ReviewRequest.Status' Error:Field validation for 'Status' failed on the 'oneof' tag"
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

### ❌ 403 Forbidden — Insufficient role

```json
{
  "success": false,
  "error": "Forbidden"
}
```

---

## Status Codes Summary

| Code | Meaning | When |
|---|---|---|
| `200` | OK | Request successful |
| `400` | Bad Request | Malformed JSON, validation error, or no employee record |
| `401` | Unauthorized | Missing or invalid API key / JWT |
| `403` | Forbidden | Role not permitted for this endpoint |
| `404` | Not Found | Timesheet not found or outside caller's branch |
| `500` | Internal Server Error | Database error or unexpected failure |
