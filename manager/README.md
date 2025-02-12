# Manager Service Directory Structure

```
manager/
├── cmd/
│   └── manager/
│       └── main.go         // Entry point: initialize server, load config, start HTTP listener
├── config/
│   └── config.yaml         // Configuration file (MongoDB URI, admin credentials, etc.)
├── api/
│   └── handlers/
│       ├── agent.go        // Handlers for /agent endpoints (register, heartbeat, tasks)
│       ├── task.go         // Handlers for /task endpoints (create, status, cancel)
│       └── role.go         // Handlers for /roles endpoints (list, create, retrieve)
├── admin/
│   └── auth.go             // Admin login and session management
├── middleware/
│   └── auth.go             // API key, HMAC, and admin auth middleware
├── models/
│   ├── agent.go            // Agent model (MongoDB schema mapping)
│   ├── task.go             // Task model
│   ├── role.go             // Role model
│   ├── admin.go            // Admin account model
│   └── log.go              // Logging model
├── mongodb/
│   └── client.go           // MongoDB client initialization and helpers
├── docs/
│   └── api/
│       └── manager.yml     // API documentation
├── .env.sample            // Environment variables template
├── go.mod                 // Module definition and dependencies
└── go.sum                 // Dependency checksums
```
