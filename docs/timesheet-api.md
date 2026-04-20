# Timesheet Estimate API Documentation

Base URL: `http://localhost:8080/api/v1`

> **All endpoints require:**
> - `X-API-Key` header
> - `Authorization: Bearer <access_token>` header
>
> | Route | Allowed roles |
> |---|---|
> | `POST /api/v1/timesheets/estimate` | All authenticated roles |

> The estimate is always calculated for the **currently logged-in user**.  
> Salary config (`fixed_monthly_salary`, `ot_rate`) is fetched automatically from the `employees` table using the JWT identity.

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

---

## Table of Contents

- [1. Calculate Pay Estimate](#1-calculate-pay-estimate)
- [Common Error Responses](#common-error-responses)
- [Status Codes Summary](#status-codes-summary)
- [Response Fields Reference](#response-fields-reference)

---

## 1. Calculate Pay Estimate

```
POST /api/v1/timesheets/estimate
```

Calculates the estimated pay for a given month based on hours submitted. No data is stored — this is a pure calculation endpoint. The caller's `fixed_monthly_salary` and `ot_rate` are fetched automatically from their employee record.

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
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "year": 2026,
    "month": 2,
    "support_hours": 160.0,
    "overtime_hours": 8.0
  }'
```

---

#### ✅ 200 OK — Scenario: Over (full month worked + OT)

> February 2026 has 20 weekdays → `whole_month_hours = 160.0`  
> `support_hours (160) ≥ whole_month_hours (160)` AND `overtime_hours (8) > 0` → **Over**

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

#### ✅ 200 OK — Scenario: Full (exactly full month, no OT)

> `support_hours (160) = whole_month_hours (160)` AND `overtime_hours = 0` → **Full**

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

#### ✅ 200 OK — Scenario: Partial (less than full month)

> `support_hours (8) < whole_month_hours (160)` → **Partial**  
> `estimated_pay = 8 × 1145.83 = 9166.64`

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

#### ❌ 400 Bad Request — No employee record linked to account

```json
{
  "success": false,
  "error": "employee record not found for your account"
}
```

#### ❌ 400 Bad Request — Validation error

```json
{
  "success": false,
  "error": "Key: 'EstimateRequest.Year' Error:Field validation for 'Year' failed on the 'required' tag"
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
| `200` | OK | Estimate computed successfully |
| `400` | Bad Request | Malformed JSON, validation error, or no employee record linked |
| `401` | Unauthorized | Missing or invalid API key / JWT |
| `500` | Internal Server Error | Database error or unexpected failure |

---

## Response Fields Reference

| Field | Type | Description |
|---|---|---|
| `year` | integer | Calendar year passed in the request |
| `month` | integer | Month number `1–12` passed in the request |
| `whole_month_hours` | number | Actual weekdays (Mon–Fri) in the month × 8 — the full-month threshold |
| `support_hours` | number | Regular working hours passed in the request |
| `overtime_hours` | number | Overtime hours passed in the request |
| `fixed_monthly_salary` | number | Caller's configured fixed monthly salary |
| `ot_rate` | number | Caller's configured per-OT-hour rate |
| `hourly_rate` | number | `fixed_monthly_salary / whole_month_hours` — derived at compute time |
| `scenario` | string | `Full` \| `Over` \| `Partial` — which pay formula was applied |
| `estimated_pay` | number | Computed estimated gross pay for the month |
| `currency` | string | 3-letter currency code from the employee record (e.g. `USD`, `INR`) |

---

## Pay Scenario Examples

> Assuming: `fixed_monthly_salary = 183,333.33`, `ot_rate = 500`, month = February 2026 (20 weekdays → `whole_month_hours = 160`)

| support_hours | overtime_hours | Scenario | estimated_pay | Formula used |
|---|---|---|---|---|
| `160.0` | `0.0` | `Full` | `183,333.33` | Fixed monthly salary |
| `160.0` | `8.0` | `Over` | `187,333.33` | `183,333.33 + (8 × 500)` |
| `120.0` | `0.0` | `Partial` | `137,499.60` | `120 × 1,145.83` |
| `8.0` | `0.0` | `Partial` | `9,166.64` | `8 × 1,145.83` |
| `80.0` | `5.0` | `Partial` | `91,666.40` | `80 × 1,145.83` (OT ignored when not full month) |
