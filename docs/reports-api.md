# Reports API Documentation

Base URL: `http://localhost:8080/api/v1`

> **All report endpoints require:**
> - `X-API-Key` header
> - `Authorization: Bearer <access_token>` header
>
> | Route prefix | Required role |
> |---|---|
> | `/api/v1/reports/*` | `super_admin`, `admin`, `manager` |

---

## Table of Contents

- [Attendance Report](#attendance-report)
  - [1. Get Attendance Report](#1-get-attendance-report)
- [Common Error Responses](#common-error-responses)
- [Status Codes Summary](#status-codes-summary)
- [Attendance Status Values](#attendance-status-values)

---

## Attendance Report

> **Role-based filtering:**
> | Role | Returns |
> |---|---|
> | `super_admin` | All attendance records across all branches |
> | `admin` | Attendance records from their branch only |
> | `manager` | Attendance records from their branch only |
>
> Branch filtering is applied automatically from the JWT `branch_id` claim — no extra param needed.

### 1. Get Attendance Report

```
GET /api/v1/reports/attendance
```

#### Query Parameters

| Param       | Type   | Required | Description                                                               |
|------------|--------|----------|---------------------------------------------------------------------------|
| `date_from` | string | ❌       | Start date filter in `YYYY-MM-DD` format                                  |
| `date_to`   | string | ❌       | End date filter in `YYYY-MM-DD` format                                    |
| `user_id`   | string | ❌       | Filter by a specific employee's user UUID                                 |
| `status`    | string | ❌       | One of: `present`, `absent`, `half_day`, `late`, `on_leave`               |

#### cURL — All attendance for current branch

```bash
curl -X GET http://localhost:8080/api/v1/reports/attendance \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### cURL — Filter by date range

```bash
curl -X GET "http://localhost:8080/api/v1/reports/attendance?date_from=2026-04-01&date_to=2026-04-14" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### cURL — Filter by employee + status

```bash
curl -X GET "http://localhost:8080/api/v1/reports/attendance?user_id=ff0e8400-e29b-41d4-a716-446655440001&status=present" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": [
    {
      "id": "aa1e8400-e29b-41d4-a716-446655440001",
      "user_id": "ff0e8400-e29b-41d4-a716-446655440001",
      "work_date": "2026-04-14T00:00:00Z",
      "punch_in": "2026-04-14T09:02:00Z",
      "punch_out": "2026-04-14T17:30:00Z",
      "work_hours": 8.47,
      "status": "present",
      "notes": null,
      "first_name": "Jane",
      "last_name": "Smith",
      "email": "jane.smith@oleron.com",
      "employee_code": "EMP001",
      "branch_id": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2026-04-14T09:02:00Z",
      "updated_at": "2026-04-14T17:30:00Z"
    },
    {
      "id": "aa1e8400-e29b-41d4-a716-446655440002",
      "user_id": "ff0e8400-e29b-41d4-a716-446655440002",
      "work_date": "2026-04-14T00:00:00Z",
      "punch_in": "2026-04-14T09:45:00Z",
      "punch_out": null,
      "work_hours": null,
      "status": "late",
      "notes": "Traffic delay",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@oleron.com",
      "employee_code": "EMP002",
      "branch_id": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2026-04-14T09:45:00Z",
      "updated_at": "2026-04-14T09:45:00Z"
    }
  ]
}
```

> ℹ️ `punch_out` and `work_hours` are `null` if the employee has not yet punched out.

#### ✅ 200 OK — No records found

```json
{
  "success": true,
  "data": null
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
  "error": "failed to fetch attendance report"
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

### ❌ 500 Internal Server Error

```json
{
  "success": false,
  "error": "failed to fetch attendance report"
}
```

---

## Status Codes Summary

| Code  | Meaning               | When                                                        |
|-------|-----------------------|-------------------------------------------------------------|
| `200` | OK                    | Report fetched successfully (may return empty data)         |
| `401` | Unauthorized          | Missing/invalid API key or JWT token                        |
| `403` | Forbidden             | Role is not `super_admin`, `admin`, or `manager`            |
| `500` | Internal Server Error | Database error or unexpected server error                   |

---

## Attendance Status Values

| Status      | Description                                      |
|------------|--------------------------------------------------|
| `present`   | Employee punched in on time                      |
| `absent`    | No punch-in recorded for the day                 |
| `half_day`  | Employee worked less than the expected half shift |
| `late`      | Employee punched in after the expected start time |
| `on_leave`  | Employee was on approved leave                   |

---

## Response Fields Reference

| Field           | Type      | Description                                          |
|----------------|-----------|------------------------------------------------------|
| `id`            | string    | Attendance record UUID                               |
| `user_id`       | string    | Employee's user account UUID                         |
| `work_date`     | string    | The date of the attendance record                    |
| `punch_in`      | string    | Punch-in timestamp (`null` if not yet punched in)    |
| `punch_out`     | string    | Punch-out timestamp (`null` if not yet punched out)  |
| `work_hours`    | number    | Total hours worked, computed on punch-out            |
| `status`        | string    | Attendance status for the day                        |
| `notes`         | string    | Optional note on the record                          |
| `first_name`    | string    | Employee first name (joined from users)              |
| `last_name`     | string    | Employee last name (joined from users)               |
| `email`         | string    | Employee email (joined from users)                   |
| `employee_code` | string    | Employee punch code (joined from employees)          |
| `branch_id`     | string    | Branch UUID (joined from users)                      |
| `created_at`    | string    | Record creation timestamp                            |
| `updated_at`    | string    | Record last updated timestamp                        |
