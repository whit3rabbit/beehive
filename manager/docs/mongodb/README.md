# Agent Management System Database Documentation

## Overview

This document describes the MongoDB database schema for the Agent Management System. The system manages agents, tasks, roles, administrators, and logging functionality through several collections.

## Collections

### Admins Collection

Administrators of the system with authentication capabilities.

```json
{
    "_id": ObjectId,
    "username": String,  // Required
    "email": String,     // Optional
    "password": String,  // Required, stored as hash
    "created_at": DateTime,
    "updated_at": DateTime
}
```

Password policies can be configured with the following parameters:

- Minimum length
- Uppercase requirement
- Lowercase requirement
- Number requirement
- Special character requirement

### Agents Collection

Represents connected agents in the system.

```json
{
    "_id": ObjectId,
    "uuid": String,      // Required, unique identifier
    "hostname": String,  // Required
    "mac_hash": String,  // Required
    "nickname": String,
    "role": String,
    "api_key": String,   // Hidden from JSON responses
    "api_secret": String,// Hidden from JSON responses
    "status": String,    // "active", "inactive", "disconnected"
    "last_seen": DateTime,
    "created_at": DateTime
}
```

Agents maintain their connection through heartbeat mechanisms, updating their `last_seen` timestamp regularly.

### Roles Collection

Defines agent roles and their associated permissions.

```json
{
    "_id": ObjectId,
    "name": String,          // Required
    "description": String,
    "applications": [String],// List of allowed applications
    "default_tasks": [String],// List of default tasks
    "created_at": DateTime
}
```

### Tasks Collection

Records tasks assigned to agents. Tasks must be one of the following types: "command_shell", "file_operation", "ui_automation", or "browser_automation".

```json
{
    "_id": ObjectId,
    "agent_id": String,     // Required
    "type": String,         // Required: "command_shell", "file_operation", "ui_automation", "browser_automation"
    "parameters": Object,   // Required, task-specific parameters
    "status": String,       // Required: "queued", "running", "completed", "failed", "cancelled", "timeout"
    "output": {
        "logs": String,
        "error": String
    },
    "created_at": DateTime,
    "updated_at": DateTime,
    "timeout": Integer,
    "started_at": DateTime
}
```

### Logs Collection

System-wide logging for monitoring and auditing.

```json
{
    "_id": ObjectId,
    "timestamp": DateTime,
    "endpoint": String,
    "agent_id": String,    // Optional
    "status": String,
    "details": String      // Optional
}
```

## Indexes

Recommended indexes for optimal performance:

1. Agents Collection:
   - `uuid`: Unique index
   - `mac_hash`: Unique index
   - `status`: Index for status queries
   - `last_seen`: Index for heartbeat monitoring

2. Tasks Collection:
   - `agent_id`: Index for agent-specific queries
   - `status`: Index for status monitoring
   - `created_at`: Index for temporal queries

3. Logs Collection:
   - `timestamp`: Index for temporal queries
   - `agent_id`: Index for agent-specific logs
   - `endpoint`: Index for endpoint monitoring

## Data Types

- `ObjectId`: MongoDB's internal ID type
- `String`: UTF-8 encoded strings
- `DateTime`: ISO 8601 date and time
- `Integer`: 32-bit integer
- `Object`: Nested JSON document
- `[String]`: Array of strings

## API Authentication

The system supports three authentication methods:

1. **Bearer JWT** - Used for admin authentication
2. **API Key** - Used for agent authentication
3. **Signature** - Additional security layer for agent requests

## Best Practices

1. **Data Validation**
   - Implement schema validation for all collections
   - Validate required fields before insertion
   - Use appropriate data types for each field

2. **Security**
   - Store passwords using secure hashing algorithms
   - Keep API keys and secrets encrypted
   - Implement role-based access control

3. **Performance**
   - Use appropriate indexes for frequent queries
   - Implement data archival for old logs
   - Monitor collection sizes and implement cleanup strategies

4. **Maintenance**
   - Regular backup of all collections
   - Monitor index usage and optimize as needed
   - Implement data retention policies

## Usage Examples

### Agent Registration

```javascript
db.agents.insertOne({
    uuid: "unique-identifier",
    hostname: "agent-host-1",
    mac_hash: "hashed-mac-address",
    status: "active",
    created_at: new Date(),
    last_seen: new Date()
})
```

### Task Creation

```javascript
db.tasks.insertOne({
    agent_id: "agent-uuid",
    type: "command_shell",
    parameters: {
        command: "echo 'Hello World'"
    },
    status: "queued",
    created_at: new Date(),
    updated_at: new Date(),
    timeout: 3600
})
```

### Log Entry

```javascript
db.logs.insertOne({
    timestamp: new Date(),
    endpoint: "/api/v1/tasks",
    agent_id: "agent-uuid",
    status: "success",
    details: "Task created successfully"
})
```
