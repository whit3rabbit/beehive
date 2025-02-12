# Documentation

This folder contains all documentation related to the beehive project, including API specifications and database schema definitions. It is organized into two main sections:

- **API Documentation:**  
  Contains OpenAPI YAML files that define the RESTful endpoints for both the agent and the manager.

- **Database Schemas:**  
  Provides JSON Schema validation files for the MongoDB collections used by the manager, as well as the SQLite schema for local agent storage.

---

## Directory Structure

```plaintext
.
└── docs
    ├── api
    │   ├── agent.yml       # OpenAPI specification for the Agent API
    │   └── manager.yml     # OpenAPI specification for the Manager API
    └── db_schema
        ├── agent
        │   └── agent.sqlite   # SQLite schema for agent local storage
        └── manager
            ├── agents.json    # MongoDB JSON Schema for the agents collection
            ├── logs.json      # MongoDB JSON Schema for the logs collection
            ├── roles.json     # MongoDB JSON Schema for the roles collection
            └── tasks.json     # MongoDB JSON Schema for the tasks collection
```

---

## API Documentation

- **Agent API (`docs/api/agent.yml`):**  

  Describes endpoints that the agent exposes for:
  - Polling tasks
  - Updating task status
  - Sending heartbeat signals

- **Manager API (`docs/api/manager.yml`):**  

  Defines endpoints for the central manager, including:
  - Agent registration
  - Task creation, status checking, and cancellation
  - Role management for agents
  - Logging and auditing endpoints

---

## Database Schemas

### Manager (MongoDB)

The manager uses MongoDB with JSON Schema validation to enforce data integrity. The relevant schema files include:

- **`agents.json`**: Schema for the agents collection.
- **`logs.json`**: Schema for recording API request logs and task executions.
- **`roles.json`**: Schema for defining and managing agent roles.
- **`tasks.json`**: Schema for the tasks collection, which stores task details.

### Agent (SQLite)

For lightweight local storage on the agent:

- **`agent.sqlite`** contains the SQLite schema for:
  - **Tasks Table:** Stores tasks received by the agent.
  - **Logs Table:** Maintains local logs for task executions and API interactions.

---

## Updating Documentation

- **API Specs:**  
  When changes are made to the API endpoints or data structures, update the corresponding YAML files in the `docs/api` directory. Tools like [Swagger Editor](https://editor.swagger.io/) can help validate and preview your OpenAPI specs.

- **Database Schemas:**  
  Ensure that any modifications to the data model are reflected in the JSON schema files (for MongoDB) and the SQLite script (for the agent). This helps maintain consistency and data integrity across the system.
