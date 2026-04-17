# Recruitment API Documentation

Base URL: `http://localhost:8080/api/v1`

> **Protected endpoints require:**
> - `X-API-Key` header
> - `Authorization: Bearer <access_token>` header
>
> | Route prefix | Required role |
> |---|---|
> | `/api/v1/recruitment/*` | `super_admin`, `admin`, `manager` |

> **One public endpoint (no JWT required):**
> | Endpoint | Auth |
> |---|---|
> | `POST /api/v1/recruitment/vacancies/{id}/apply` | `X-API-Key` only — for candidates |

> **Role-based filtering:**
> | Role | Access |
> |---|---|
> | `super_admin` | All vacancies across all branches |
> | `admin` | Vacancies for their branch only |
> | `manager` | Vacancies for their branch only |

> **Hire flow:**
> ```
> vacancy (open) → application (applied) → (shortlisted) → interview scheduled → (hired) → employee created
> ```

---

## Table of Contents

- [Vacancies](#vacancies)
  - [1. List All Vacancies](#1-list-all-vacancies)
  - [2. Get Vacancy by ID](#2-get-vacancy-by-id)
  - [3. Create Vacancy](#3-create-vacancy)
  - [4. Update Vacancy](#4-update-vacancy)
  - [5. Update Vacancy Status](#5-update-vacancy-status)
  - [6. Delete Vacancy](#6-delete-vacancy)
- [Applications](#applications)
  - [7. Apply for a Vacancy (Public)](#7-apply-for-a-vacancy-public)
  - [8. List Applications for a Vacancy](#8-list-applications-for-a-vacancy)
  - [9. Get Application by ID](#9-get-application-by-id)
  - [10. Update Application Status](#10-update-application-status)
  - [11. Delete Application](#11-delete-application)
- [Interviews](#interviews)
  - [12. Schedule Interview](#12-schedule-interview)
  - [13. Update Interview](#13-update-interview)
  - [14. Delete Interview](#14-delete-interview)
- [Hire](#hire)
  - [15. Hire Applicant](#15-hire-applicant)
- [Common Error Responses](#common-error-responses)
- [Status Codes Summary](#status-codes-summary)
- [Reference Tables](#reference-tables)
- [Response Fields Reference](#response-fields-reference)

---

## Vacancies

### 1. List All Vacancies

```
GET /api/v1/recruitment/vacancies
```

Returns all vacancies for the branch (summary only, no applications). `super_admin` sees all branches. Ordered by most recent first.

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/recruitment/vacancies \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": [
    {
      "id": "vv0e8400-e29b-41d4-a716-446655440001",
      "branch_id": "550e8400-e29b-41d4-a716-446655440000",
      "created_by": "880e8400-e29b-41d4-a716-446655440002",
      "title": "Senior Go Developer",
      "department": "Engineering",
      "description": "We are looking for an experienced Go developer...",
      "requirements": "5+ years of Go experience, PostgreSQL knowledge",
      "positions": 2,
      "status": "open",
      "deadline": "2026-05-31T00:00:00Z",
      "created_at": "2026-04-17T09:00:00Z",
      "updated_at": "2026-04-17T09:00:00Z",
      "application_count": 14
    }
  ]
}
```

#### ✅ 200 OK — No vacancies

```json
{
  "success": true,
  "data": null
}
```

---

### 2. Get Vacancy by ID

```
GET /api/v1/recruitment/vacancies/{id}
```

Returns a single vacancy with its current application count.

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/recruitment/vacancies/vv0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "vv0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "created_by": "880e8400-e29b-41d4-a716-446655440002",
    "title": "Senior Go Developer",
    "department": "Engineering",
    "description": "We are looking for an experienced Go developer...",
    "requirements": "5+ years of Go experience, PostgreSQL knowledge",
    "positions": 2,
    "status": "open",
    "deadline": "2026-05-31T00:00:00Z",
    "created_at": "2026-04-17T09:00:00Z",
    "updated_at": "2026-04-17T09:00:00Z",
    "application_count": 14
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
  "error": "vacancy not found"
}
```

---

### 3. Create Vacancy

```
POST /api/v1/recruitment/vacancies
```

Creates a new job vacancy for the caller's branch. Created with status `draft` — publish it with [Update Vacancy Status](#5-update-vacancy-status).

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | ✅ | Job title (max 150 characters) |
| `department` | string | ❌ | Department name (e.g. `"Engineering"`) |
| `description` | string | ❌ | Full job description / JD |
| `requirements` | string | ❌ | Skills and qualifications |
| `positions` | integer | ❌ | Number of openings (default: `1`) |
| `deadline` | string | ❌ | Application deadline `YYYY-MM-DD` |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/recruitment/vacancies \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "title": "Senior Go Developer",
    "department": "Engineering",
    "description": "We are looking for an experienced Go developer to join our team.",
    "requirements": "5+ years of Go experience, PostgreSQL knowledge, REST API design",
    "positions": 2,
    "deadline": "2026-05-31"
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "id": "vv0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "created_by": "880e8400-e29b-41d4-a716-446655440002",
    "title": "Senior Go Developer",
    "department": "Engineering",
    "description": "We are looking for an experienced Go developer to join our team.",
    "requirements": "5+ years of Go experience, PostgreSQL knowledge, REST API design",
    "positions": 2,
    "status": "draft",
    "deadline": "2026-05-31T00:00:00Z",
    "created_at": "2026-04-17T09:00:00Z",
    "updated_at": "2026-04-17T09:00:00Z"
  }
}
```

#### ❌ 422 Unprocessable Entity

```json
{
  "success": false,
  "error": "Key: 'CreateVacancyRequest.Title' Error:Field validation for 'Title' failed on the 'required' tag"
}
```

---

### 4. Update Vacancy

```
PUT /api/v1/recruitment/vacancies/{id}
```

Updates vacancy fields. All fields are optional — only provided fields are changed.

> `admin` and `manager` can only update vacancies from their own branch.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | ❌ | Job title |
| `department` | string | ❌ | Department name |
| `description` | string | ❌ | Full job description |
| `requirements` | string | ❌ | Skills and qualifications |
| `positions` | integer | ❌ | Number of openings (min: `1`) |
| `deadline` | string | ❌ | Application deadline `YYYY-MM-DD` |

#### cURL

```bash
curl -X PUT http://localhost:8080/api/v1/recruitment/vacancies/vv0e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "positions": 3,
    "deadline": "2026-06-15"
  }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "vv0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "created_by": "880e8400-e29b-41d4-a716-446655440002",
    "title": "Senior Go Developer",
    "department": "Engineering",
    "description": "We are looking for an experienced Go developer to join our team.",
    "requirements": "5+ years of Go experience, PostgreSQL knowledge, REST API design",
    "positions": 3,
    "status": "open",
    "deadline": "2026-06-15T00:00:00Z",
    "created_at": "2026-04-17T09:00:00Z",
    "updated_at": "2026-04-17T10:00:00Z"
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
  "error": "vacancy not found"
}
```

---

### 5. Update Vacancy Status

```
PATCH /api/v1/recruitment/vacancies/{id}/status
```

Moves a vacancy through its lifecycle. There is no enforced order — you can transition freely between statuses.

> `admin` and `manager` can only update vacancies from their own branch.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `status` | string | ✅ | One of: `draft`, `open`, `closed`, `cancelled` |

#### cURL — Publish

```bash
curl -X PATCH http://localhost:8080/api/v1/recruitment/vacancies/vv0e8400-e29b-41d4-a716-446655440001/status \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{ "status": "open" }'
```

#### cURL — Close

```bash
curl -X PATCH http://localhost:8080/api/v1/recruitment/vacancies/vv0e8400-e29b-41d4-a716-446655440001/status \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{ "status": "closed" }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "vv0e8400-e29b-41d4-a716-446655440001",
    "branch_id": "550e8400-e29b-41d4-a716-446655440000",
    "created_by": "880e8400-e29b-41d4-a716-446655440002",
    "title": "Senior Go Developer",
    "department": "Engineering",
    "description": "We are looking for an experienced Go developer to join our team.",
    "requirements": "5+ years of Go experience, PostgreSQL knowledge, REST API design",
    "positions": 3,
    "status": "open",
    "deadline": "2026-06-15T00:00:00Z",
    "created_at": "2026-04-17T09:00:00Z",
    "updated_at": "2026-04-17T10:30:00Z"
  }
}
```

#### ❌ 422 Unprocessable Entity

```json
{
  "success": false,
  "error": "Key: 'UpdateVacancyStatusRequest.Status' Error:Field validation for 'Status' failed on the 'oneof' tag"
}
```

---

### 6. Delete Vacancy

```
DELETE /api/v1/recruitment/vacancies/{id}
```

Deletes a vacancy and all its applications. **Only `draft` vacancies can be deleted.**

> `admin` and `manager` can only delete vacancies from their own branch.

#### cURL

```bash
curl -X DELETE http://localhost:8080/api/v1/recruitment/vacancies/vv0e8400-e29b-41d4-a716-446655440001 \
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
  "error": "only draft vacancies can be deleted"
}
```

---

## Applications

### 7. Apply for a Vacancy (Public)

```
POST /api/v1/recruitment/vacancies/{id}/apply
```

> **This endpoint does not require a JWT.** Only the `X-API-Key` header is needed. Intended for use from a public-facing careers page or candidate portal.

Submits an application for the specified vacancy. The vacancy must have status `open`.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `first_name` | string | ✅ | Applicant's first name |
| `last_name` | string | ✅ | Applicant's last name |
| `email` | string | ✅ | Applicant's email address |
| `phone` | string | ❌ | Contact phone number |
| `cv_url` | string | ❌ | Link to uploaded CV / resume |
| `cover_letter` | string | ❌ | Cover letter text |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/recruitment/vacancies/vv0e8400-e29b-41d4-a716-446655440001/apply \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -d '{
    "first_name": "Jane",
    "last_name": "Smith",
    "email": "jane.smith@example.com",
    "phone": "+1234567890",
    "cv_url": "https://storage.example.com/cvs/jane-smith.pdf",
    "cover_letter": "I am excited to apply for this position..."
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "id": "ap0e8400-e29b-41d4-a716-446655440001",
    "vacancy_id": "vv0e8400-e29b-41d4-a716-446655440001",
    "first_name": "Jane",
    "last_name": "Smith",
    "email": "jane.smith@example.com",
    "phone": "+1234567890",
    "cv_url": "https://storage.example.com/cvs/jane-smith.pdf",
    "cover_letter": "I am excited to apply for this position...",
    "status": "applied",
    "notes": null,
    "applied_at": "2026-04-17T11:00:00Z",
    "updated_at": "2026-04-17T11:00:00Z"
  }
}
```

#### ❌ 404 Not Found

```json
{
  "success": false,
  "error": "vacancy not found"
}
```

#### ❌ 422 Unprocessable Entity — Vacancy not open

```json
{
  "success": false,
  "error": "vacancy is not open for applications"
}
```

---

### 8. List Applications for a Vacancy

```
GET /api/v1/recruitment/vacancies/{id}/applications
```

Returns all applications for a vacancy. Optionally filter by status.

#### Query Parameters

| Parameter | Type | Description |
|---|---|---|
| `status` | string | Filter by status: `applied`, `shortlisted`, `rejected`, `interview_scheduled`, `hired`, `withdrawn` |

#### cURL

```bash
curl -X GET "http://localhost:8080/api/v1/recruitment/vacancies/vv0e8400-e29b-41d4-a716-446655440001/applications?status=shortlisted" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": [
    {
      "id": "ap0e8400-e29b-41d4-a716-446655440001",
      "vacancy_id": "vv0e8400-e29b-41d4-a716-446655440001",
      "first_name": "Jane",
      "last_name": "Smith",
      "email": "jane.smith@example.com",
      "phone": "+1234567890",
      "cv_url": "https://storage.example.com/cvs/jane-smith.pdf",
      "cover_letter": "I am excited to apply for this position...",
      "status": "shortlisted",
      "notes": "Strong background in Go, good culture fit",
      "applied_at": "2026-04-17T11:00:00Z",
      "updated_at": "2026-04-18T09:00:00Z"
    }
  ]
}
```

#### ✅ 200 OK — No results

```json
{
  "success": true,
  "data": null
}
```

---

### 9. Get Application by ID

```
GET /api/v1/recruitment/applications/{id}
```

Returns a single application with all its interview sessions.

#### cURL

```bash
curl -X GET http://localhost:8080/api/v1/recruitment/applications/ap0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "ap0e8400-e29b-41d4-a716-446655440001",
    "vacancy_id": "vv0e8400-e29b-41d4-a716-446655440001",
    "first_name": "Jane",
    "last_name": "Smith",
    "email": "jane.smith@example.com",
    "phone": "+1234567890",
    "cv_url": "https://storage.example.com/cvs/jane-smith.pdf",
    "cover_letter": "I am excited to apply for this position...",
    "status": "interview_scheduled",
    "notes": "Strong background in Go",
    "applied_at": "2026-04-17T11:00:00Z",
    "updated_at": "2026-04-18T10:00:00Z",
    "interviews": [
      {
        "id": "iv0e8400-e29b-41d4-a716-446655440001",
        "application_id": "ap0e8400-e29b-41d4-a716-446655440001",
        "interviewer_id": "880e8400-e29b-41d4-a716-446655440002",
        "scheduled_at": "2026-04-20T10:00:00Z",
        "type": "video",
        "location": "https://meet.example.com/interview-room-1",
        "outcome": "pending",
        "feedback": null,
        "created_at": "2026-04-18T10:00:00Z",
        "updated_at": "2026-04-18T10:00:00Z"
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
  "error": "application not found"
}
```

---

### 10. Update Application Status

```
PATCH /api/v1/recruitment/applications/{id}/status
```

Moves an application through the pipeline. Optionally attach or update reviewer notes.

> `admin` and `manager` can only update applications for vacancies in their own branch.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `status` | string | ✅ | One of: `shortlisted`, `rejected`, `interview_scheduled`, `hired`, `withdrawn` |
| `notes` | string | ❌ | Reviewer notes — if provided, replaces existing notes |

#### cURL — Shortlist

```bash
curl -X PATCH http://localhost:8080/api/v1/recruitment/applications/ap0e8400-e29b-41d4-a716-446655440001/status \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{ "status": "shortlisted", "notes": "Strong Go background, good culture fit" }'
```

#### cURL — Reject

```bash
curl -X PATCH http://localhost:8080/api/v1/recruitment/applications/ap0e8400-e29b-41d4-a716-446655440001/status \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{ "status": "rejected", "notes": "Does not meet minimum experience requirement" }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "ap0e8400-e29b-41d4-a716-446655440001",
    "vacancy_id": "vv0e8400-e29b-41d4-a716-446655440001",
    "first_name": "Jane",
    "last_name": "Smith",
    "email": "jane.smith@example.com",
    "phone": "+1234567890",
    "cv_url": "https://storage.example.com/cvs/jane-smith.pdf",
    "cover_letter": "I am excited to apply for this position...",
    "status": "shortlisted",
    "notes": "Strong Go background, good culture fit",
    "applied_at": "2026-04-17T11:00:00Z",
    "updated_at": "2026-04-18T09:00:00Z"
  }
}
```

---

### 11. Delete Application

```
DELETE /api/v1/recruitment/applications/{id}
```

Permanently deletes an application and all its interview records.

#### cURL

```bash
curl -X DELETE http://localhost:8080/api/v1/recruitment/applications/ap0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true
}
```

---

## Interviews

### 12. Schedule Interview

```
POST /api/v1/recruitment/applications/{id}/interviews
```

Schedules an interview session for a shortlisted candidate. Multiple interviews can be scheduled per application (e.g. technical round followed by HR round).

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `interviewer_id` | string | ✅ | UUID of the user (branch staff) conducting the interview |
| `scheduled_at` | string | ✅ | Interview date and time in RFC3339 format |
| `type` | string | ✅ | One of: `phone`, `video`, `in_person` |
| `location` | string | ❌ | Room name, address, or video link |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/recruitment/applications/ap0e8400-e29b-41d4-a716-446655440001/interviews \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "interviewer_id": "880e8400-e29b-41d4-a716-446655440002",
    "scheduled_at": "2026-04-20T10:00:00Z",
    "type": "video",
    "location": "https://meet.example.com/interview-room-1"
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "id": "iv0e8400-e29b-41d4-a716-446655440001",
    "application_id": "ap0e8400-e29b-41d4-a716-446655440001",
    "interviewer_id": "880e8400-e29b-41d4-a716-446655440002",
    "scheduled_at": "2026-04-20T10:00:00Z",
    "type": "video",
    "location": "https://meet.example.com/interview-room-1",
    "outcome": "pending",
    "feedback": null,
    "created_at": "2026-04-18T10:00:00Z",
    "updated_at": "2026-04-18T10:00:00Z"
  }
}
```

#### ❌ 422 Unprocessable Entity — Invalid date format

```json
{
  "success": false,
  "error": "invalid scheduled_at: use RFC3339 format (e.g. 2026-04-17T10:00:00Z)"
}
```

---

### 13. Update Interview

```
PUT /api/v1/recruitment/interviews/{id}
```

Updates interview details. All fields are optional — use to reschedule, change the type/location, or record the outcome after the session.

> `admin` and `manager` can only update interviews for their own branch's applications.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `scheduled_at` | string | ❌ | New date/time in RFC3339 format (reschedule) |
| `type` | string | ❌ | One of: `phone`, `video`, `in_person` |
| `location` | string | ❌ | Room name, address, or video link |
| `outcome` | string | ❌ | One of: `pending`, `passed`, `failed`, `no_show` |
| `feedback` | string | ❌ | Interviewer notes after the session |

#### cURL — Record outcome

```bash
curl -X PUT http://localhost:8080/api/v1/recruitment/interviews/iv0e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "outcome": "passed",
    "feedback": "Excellent problem-solving skills, clear communicator. Recommend proceeding to offer."
  }'
```

#### ✅ 200 OK

```json
{
  "success": true,
  "data": {
    "id": "iv0e8400-e29b-41d4-a716-446655440001",
    "application_id": "ap0e8400-e29b-41d4-a716-446655440001",
    "interviewer_id": "880e8400-e29b-41d4-a716-446655440002",
    "scheduled_at": "2026-04-20T10:00:00Z",
    "type": "video",
    "location": "https://meet.example.com/interview-room-1",
    "outcome": "passed",
    "feedback": "Excellent problem-solving skills, clear communicator. Recommend proceeding to offer.",
    "created_at": "2026-04-18T10:00:00Z",
    "updated_at": "2026-04-20T11:30:00Z"
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
  "error": "interview not found"
}
```

---

### 14. Delete Interview

```
DELETE /api/v1/recruitment/interviews/{id}
```

Deletes an interview session.

#### cURL

```bash
curl -X DELETE http://localhost:8080/api/v1/recruitment/interviews/iv0e8400-e29b-41d4-a716-446655440001 \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>"
```

#### ✅ 200 OK

```json
{
  "success": true
}
```

---

## Hire

### 15. Hire Applicant

```
POST /api/v1/recruitment/applications/{id}/hire
```

Converts a successful applicant into an employee. This action atomically:
1. Creates a `users` record with role `employee`
2. Creates an `employees` record linked to the user
3. Marks the application status as `hired`

> The applicant's name, email, and phone are carried over from the application. The `temp_password` you provide should be shared with the new hire — they can change it via the profile API.

> `admin` and `manager` can only hire from applications in their own branch.

#### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `employee_code` | string | ✅ | Unique punch code for the employee (e.g. `"EMP042"`) |
| `joining_date` | string | ✅ | Start date `YYYY-MM-DD` |
| `temp_password` | string | ✅ | Temporary password (min 8 characters) — must be shared with the hire |
| `hourly_rate` | number | ❌ | Hourly pay rate (default: `0`) |
| `currency` | string | ❌ | 3-letter currency code (default: `USD`) |
| `designation` | string | ❌ | Job title / role label (e.g. `"Software Engineer"`) |
| `employment_type` | string | ❌ | One of: `full_time`, `part_time`, `contract` (default: `full_time`) |

#### cURL

```bash
curl -X POST http://localhost:8080/api/v1/recruitment/applications/ap0e8400-e29b-41d4-a716-446655440001/hire \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-mobile-app-api-key" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "employee_code": "EMP042",
    "joining_date": "2026-05-01",
    "temp_password": "Welcome@2026",
    "hourly_rate": 35.00,
    "currency": "USD",
    "designation": "Senior Go Developer",
    "employment_type": "full_time"
  }'
```

#### ✅ 201 Created

```json
{
  "success": true,
  "data": {
    "user_id": "ff0e8400-e29b-41d4-a716-446655440099",
    "employee_id": "ee0e8400-e29b-41d4-a716-446655440099",
    "email": "jane.smith@example.com",
    "message": "candidate successfully hired and employee account created"
  }
}
```

#### ❌ 404 Not Found

```json
{
  "success": false,
  "error": "application not found"
}
```

#### ❌ 422 Unprocessable Entity — Email already registered

```json
{
  "success": false,
  "error": "failed to create user account: email may already be registered"
}
```

#### ❌ 422 Unprocessable Entity — Duplicate employee code

```json
{
  "success": false,
  "error": "failed to create employee record: employee code may already exist"
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
| `201` | Created | Vacancy, application, interview, or employee created |
| `400` | Bad Request | Malformed JSON body |
| `401` | Unauthorized | Missing/invalid API key or JWT |
| `403` | Forbidden | Role lacks access or resource belongs to another branch |
| `404` | Not Found | Resource ID does not exist |
| `422` | Unprocessable Entity | Validation failure, vacancy not open, or conflict (duplicate code/email) |
| `500` | Internal Server Error | Unexpected database or server error |

---

## Reference Tables

### Vacancy Status

| Status | Description | Accepts applications |
|---|---|---|
| `draft` | Created, not yet published | ❌ |
| `open` | Actively accepting applications | ✅ |
| `closed` | Position filled or stopped | ❌ |
| `cancelled` | Cancelled before filling | ❌ |

### Application Status

| Status | Description |
|---|---|
| `applied` | Submitted by candidate, pending review |
| `shortlisted` | CV reviewed and selected for next step |
| `rejected` | Not proceeding |
| `interview_scheduled` | Interview session booked |
| `hired` | Converted to employee |
| `withdrawn` | Candidate withdrew their application |

### Interview Type

| Value | Description |
|---|---|
| `phone` | Phone call |
| `video` | Video call (Zoom, Meet, Teams, etc.) |
| `in_person` | On-site interview |

### Interview Outcome

| Value | Description |
|---|---|
| `pending` | Interview not yet held |
| `passed` | Candidate progressed |
| `failed` | Did not pass |
| `no_show` | Candidate did not attend |

---

## Response Fields Reference

### Vacancy

| Field | Type | Description |
|---|---|---|
| `id` | string | Vacancy UUID |
| `branch_id` | string | Branch this vacancy belongs to |
| `created_by` | string | UUID of the user who created the vacancy |
| `title` | string | Job title |
| `department` | string \| null | Department name |
| `description` | string \| null | Full job description |
| `requirements` | string \| null | Skills and qualifications |
| `positions` | integer | Number of openings |
| `status` | string | `draft`, `open`, `closed`, or `cancelled` |
| `deadline` | string \| null | Application deadline timestamp |
| `created_at` | string | Creation timestamp |
| `updated_at` | string | Last update timestamp |
| `application_count` | integer | Total applications received (included in list and get responses) |

### Application

| Field | Type | Description |
|---|---|---|
| `id` | string | Application UUID |
| `vacancy_id` | string | Parent vacancy UUID |
| `first_name` | string | Applicant first name |
| `last_name` | string | Applicant last name |
| `email` | string | Applicant email |
| `phone` | string \| null | Applicant phone |
| `cv_url` | string \| null | Link to CV / resume |
| `cover_letter` | string \| null | Cover letter text |
| `status` | string | Current pipeline stage |
| `notes` | string \| null | Reviewer notes |
| `applied_at` | string | Submission timestamp |
| `updated_at` | string | Last status update timestamp |
| `interviews` | array | Interview sessions (only in Get by ID response) |

### Interview

| Field | Type | Description |
|---|---|---|
| `id` | string | Interview UUID |
| `application_id` | string | Parent application UUID |
| `interviewer_id` | string | User UUID of the interviewer |
| `scheduled_at` | string | Interview date and time (RFC3339) |
| `type` | string | `phone`, `video`, or `in_person` |
| `location` | string \| null | Room, address, or video link |
| `outcome` | string | `pending`, `passed`, `failed`, or `no_show` |
| `feedback` | string \| null | Interviewer notes after the session |
| `created_at` | string | Creation timestamp |
| `updated_at` | string | Last update timestamp |

### HireResult

| Field | Type | Description |
|---|---|---|
| `user_id` | string | UUID of the newly created user account |
| `employee_id` | string | UUID of the newly created employee record |
| `email` | string | Email address of the new employee |
| `message` | string | Confirmation message |
