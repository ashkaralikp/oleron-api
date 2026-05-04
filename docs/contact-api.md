# Contact API

Public website contact-form submission endpoint.

## Endpoint

`POST /api/v1/contact-submissions`

- No JWT required
- No API key required
- Rate limited per client IP: `5 requests / minute`

## Request Body

```json
{
  "name": "Jane Doe",
  "company": "Acme Ltd",
  "email": "jane@example.com",
  "phone": "+1 555 0100",
  "category": "sales",
  "message": "I'd like to know more about your services."
}
```

## Validation

- `name`: required, max 150 chars
- `email`: required, valid email, max 255 chars
- `message`: required, max 5000 chars
- `company`: optional, max 150 chars
- `phone`: optional, max 50 chars
- `category`: optional, max 100 chars

## Success Response

Status: `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "0f9e9d98-40d0-4f52-bc84-f438243f3f1f",
    "name": "Jane Doe",
    "company": "Acme Ltd",
    "email": "jane@example.com",
    "phone": "+1 555 0100",
    "category": "sales",
    "message": "I'd like to know more about your services.",
    "status": "new",
    "ip_address": "203.0.113.10",
    "user_agent": "Mozilla/5.0",
    "created_at": "2026-05-04T11:55:00Z",
    "updated_at": "2026-05-04T11:55:00Z"
  }
}
```

## Error Responses

`400 Bad Request`

```json
{
  "success": false,
  "error": "invalid request body"
}
```

`422 Unprocessable Entity`

```json
{
  "success": false,
  "error": "Key: 'CreateSubmissionRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
}
```

`429 Too Many Requests`

```json
{
  "success": false,
  "error": "rate limit exceeded"
}
```
