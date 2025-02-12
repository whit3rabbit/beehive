

```
manager/
├── cmd/
│   └── manager/
│       └── main.go         // Entry point: initialize server, load config, start HTTP listener
├── config/
│   └── config.yaml         // Configuration file (MongoDB URI, admin credentials, etc.)
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── agent.go    // Handlers for /agent endpoints (register, heartbeat, tasks)
│   │   │   ├── task.go     // Handlers for /task endpoints (create, status, cancel)
│   │   │   └── role.go     // Handlers for /roles endpoints (list, create, retrieve)
│   │   └── routes.go       // Set up HTTP routes matching manager API docs
│   ├── admin/
│   │   └── auth.go         // Admin login and session management
│   ├── middleware/
│   │   └── auth.go         // API key, HMAC, and admin auth middleware for protecting endpoints
│   ├── models/
│   │   ├── agent.go        // Agent model (MongoDB schema mapping)
│   │   ├── task.go         // Task model
│   │   ├── role.go         // Role model
│   │   └── admin.go        // Admin account model
│   └── mongodb/
│       └── client.go       // MongoDB client initialization and helper functions
├── docs/
│   └── api/
│       └── manager.yml     // API documentation matching the manager API endpoints
├── go.mod                  // Module definition and dependencies
└── go.sum                  // Dependency checksums
```