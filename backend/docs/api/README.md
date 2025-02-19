# Beehive Manager API Documentation

## Overview

The Beehive Manager API provides endpoints for managing agents, tasks, and administrative functions. The API uses JWT for admin authentication and API keys for agent authentication.

## Authentication

### Admin Authentication

- Uses JWT (JSON Web Tokens)
- Token must be included in the `Authorization` header as `Bearer <token>`
- Tokens expire after the configured duration (default 24 hours)

### Agent Authentication

- Uses API key and signature-based authentication
- Requires `X-API-Key` header for the API key
- Requires `X-Signature` header with HMAC-SHA256 signature of request body

## Base URL

```http
https://<your-server>:8080
```

## Endpoints

### Admin Routes

#### Login

```http
POST /admin/login
```

Request body:

```json
{
    "username": "string",
    "password": "string"
}
```

Response:

```json
{
    "token": "string",
    "username": "string"
}
```

#### List Roles

```http
GET /admin/roles
```

Response:

```json
[
    {
        "id": "string",
        "name": "string",
        "description": "string",
        "applications": ["string"],
        "default_tasks": ["string"],
        "created_at": "string"
    }
]
```

#### Create Role

```http
POST /admin/roles
```

Request body:

```json
{
    "name": "string",
    "description": "string",
    "applications": ["string"],
    "default_tasks": ["string"]
}
```

### Agent Routes

#### Register Agent

```http
POST /api/agent/register
```

Request body:

```json
{
    "uuid": "string",
    "hostname": "string",
    "mac_hash": "string",
    "nickname": "string",
    "role": "string"
}
```

Response:

```json
{
    "api_key": "string",
    "api_secret": "string",
    "status": "registered",
    "timestamp": "string"
}
```

#### Agent Heartbeat

```http
POST /api/agent/heartbeat
```

Request body:

```json
{
    "uuid": "string"
}
```

Response:

```json
{
    "status": "heartbeat_received",
    "timestamp": "string"
}
```

#### List Agent Tasks

```http
GET /api/agent/{agent_id}/tasks
```

Response:

```json
[
    {
        "id": "string",
        "type": "string",
        "parameters": {},
        "status": "string",
        "output": {},
        "created_at": "string",
        "updated_at": "string"
    }
]
```

### Task Management

#### Create Task

```http
POST /api/task/create
```

Request body:

```json
{
    "task": {
        "agent_id": "string",
        "type": "scan|execute|monitor",
        "parameters": {},
        "timeout": 0
    }
}
```

Response:

```json
{
    "task_id": "string",
    "status": "queued",
    "timestamp": "string"
}
```

#### Get Task Status

```http
GET /api/task/status/{task_id}
```

Response:

```json
{
    "id": "string",
    "agent_id": "string",
    "type": "string",
    "parameters": {},
    "status": "string",
    "output": {},
    "created_at": "string",
    "updated_at": "string",
    "timeout": 0,
    "started_at": "string"
}
```

#### Cancel Task

```http
POST /api/task/cancel/{task_id}
```

Response:

```json
{
    "task_id": "string",
    "status": "cancelled",
    "timestamp": "string"
}
```

## Status Codes

The API uses standard HTTP status codes:

- `200 OK`: Successful request
- `201 Created`: Resource successfully created
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication failed
- `403 Forbidden`: Permission denied
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

## Rate Limiting

- Admin routes are limited to 5 requests per minute
- Failed login attempts are limited to 5 attempts per 15 minutes
- After exceeding login attempts, account is blocked for 15 minutes

## Error Responses

All error responses follow this format:

```json
{
    "error": "string",
    "details": "string"
}
```
