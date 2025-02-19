print('Starting MongoDB initialization script...');

// Switch to admin database first
db = db.getSiblingDB('admin');

// MongoDB Docker uses MONGO_INITDB_ROOT_USERNAME and MONGO_INITDB_ROOT_PASSWORD for authentication
print('Authenticating as root user...');
db.auth(process.env.MONGO_INITDB_ROOT_USERNAME, process.env.MONGO_INITDB_ROOT_PASSWORD);

// Switch to the application database
print('Switching to application database...');
db = db.getSiblingDB(process.env.MONGO_INITDB_DATABASE);

// Create application user (for the app) with the provided credentials
print('Creating application user...');
db.createUser({
    user: process.env.MONGODB_USER,
    pwd: process.env.MONGODB_PASSWORD,
    roles: [
        {
            role: 'readWrite',
            db: process.env.MONGO_INITDB_DATABASE
        }
    ]
});

// Create necessary collections
print('Creating collections...');
db.createCollection('admins');
db.createCollection('agents');
db.createCollection('roles');
db.createCollection('tasks');
db.createCollection('logs');

// Create indexes
print('Creating indexes...');
db.admins.createIndex({ "username": 1 }, { unique: true });
db.agents.createIndex({ "uuid": 1 }, { unique: true });
db.agents.createIndex({ "api_key": 1 }, { unique: true });
db.agents.createIndex({ "last_seen": 1 });

// Insert default admin user using the pre-hashed password from environment variable
print('Inserting default admin user...');
db.admins.insertOne({
    username: process.env.ADMIN_DEFAULT_USERNAME,
    password: process.env.ADMIN_DEFAULT_PASSWORD_HASH,
    created_at: new Date(),
    updated_at: new Date()
});

// Insert multiple roles
print('Inserting roles...');
const roles = [
    {
        name: "worker",
        description: "Default worker role",
        applications: ["scanner", "executor"],
        defaultTasks: ["scan", "execute"],
        created_at: new Date()
    },
    {
        name: "admin",
        description: "Administrative role",
        applications: ["scanner", "executor", "manager"],
        defaultTasks: ["scan", "execute", "manage"],
        created_at: new Date()
    },
    {
        name: "viewer",
        description: "Read-only role",
        applications: ["viewer"],
        defaultTasks: ["view"],
        created_at: new Date()
    }
];
db.roles.insertMany(roles);

// Insert multiple agents
print('Inserting agents...');
const agents = [
    {
        uuid: "test-agent-001",
        hostname: "test-host-1",
        mac_hash: "ab12cd34ef56",
        nickname: "Test Agent 1",
        role: "worker",
        api_key: "test_api_key_hash_1",
        api_secret: "test_api_secret_hash_1",
        status: "active",
        last_seen: new Date(),
        created_at: new Date()
    },
    {
        uuid: "test-agent-002",
        hostname: "test-host-2",
        mac_hash: "cd34ef56gh78",
        nickname: "Test Agent 2",
        role: "admin",
        api_key: "test_api_key_hash_2",
        api_secret: "test_api_secret_hash_2",
        status: "inactive",
        last_seen: new Date(Date.now() - 86400000), // 1 day ago
        created_at: new Date(Date.now() - 604800000) // 1 week ago
    },
    {
        uuid: "test-agent-003",
        hostname: "test-host-3",
        mac_hash: "ef56gh78ij90",
        nickname: "Test Agent 3",
        role: "viewer",
        api_key: "test_api_key_hash_3",
        api_secret: "test_api_secret_hash_3",
        status: "disconnected",
        last_seen: new Date(Date.now() - 259200000), // 3 days ago
        created_at: new Date(Date.now() - 1209600000) // 2 weeks ago
    }
];
db.agents.insertMany(agents);

// Insert sample tasks
print('Inserting tasks...');
const tasks = [
    {
        agent_id: "test-agent-001",
        type: "scan",
        parameters: {
            target: "localhost",
            ports: "80,443,8080",
            timeout: 300
        },
        status: "completed",
        output: {
            logs: "Scan completed successfully\nPorts 80,443 open",
            error: ""
        },
        created_at: new Date(Date.now() - 3600000), // 1 hour ago
        updated_at: new Date(),
        timeout: 300,
        started_at: new Date(Date.now() - 3300000) // 55 minutes ago
    },
    {
        agent_id: "test-agent-002",
        type: "execute",
        parameters: {
            command: "systeminfo",
            shell: "powershell"
        },
        status: "running",
        output: {
            logs: "Executing system information command...",
            error: ""
        },
        created_at: new Date(Date.now() - 300000), // 5 minutes ago
        updated_at: new Date(),
        timeout: 600,
        started_at: new Date(Date.now() - 240000) // 4 minutes ago
    },
    {
        agent_id: "test-agent-001",
        type: "scan",
        parameters: {
            target: "192.168.1.1",
            ports: "22,3389",
            timeout: 300
        },
        status: "failed",
        output: {
            logs: "Scan initiated...",
            error: "Connection timeout after 300 seconds"
        },
        created_at: new Date(Date.now() - 86400000), // 1 day ago
        updated_at: new Date(Date.now() - 86100000), // 23.95 hours ago
        timeout: 300,
        started_at: new Date(Date.now() - 86400000) // 1 day ago
    }
];
db.tasks.insertMany(tasks);

// Insert sample logs
print('Inserting logs...');
const logs = [
    {
        timestamp: new Date(),
        endpoint: "/api/agent/register",
        agent_id: "test-agent-001",
        status: "success",
        details: "Agent registration successful"
    },
    {
        timestamp: new Date(Date.now() - 3600000), // 1 hour ago
        endpoint: "/api/task/create",
        agent_id: "test-agent-002",
        status: "success",
        details: "Task created: systeminfo execution"
    },
    {
        timestamp: new Date(Date.now() - 7200000), // 2 hours ago
        endpoint: "/api/agent/heartbeat",
        agent_id: "test-agent-003",
        status: "warning",
        details: "Agent heartbeat delayed"
    },
    {
        timestamp: new Date(Date.now() - 86400000), // 1 day ago
        endpoint: "/api/task/status",
        agent_id: "test-agent-001",
        status: "error",
        details: "Task execution timeout"
    }
];
db.logs.insertMany(logs);

// Print verification of setup
print('MongoDB Test Environment Setup Complete:');
print('----------------------------------------');
print('Root Admin User:', db.runCommand({ connectionStatus: 1 }).authInfo.authenticatedUsers[0].user);
print('Application Database:', db.getName());
print('Application User:', process.env.MONGODB_USER);
print('Collections Created:', JSON.stringify(db.getCollectionNames()));

// Use countDocuments() instead of count()
async function getCollectionCounts() {
    const adminCount = await db.admins.countDocuments();
    const rolesCount = await db.roles.countDocuments();
    const agentsCount = await db.agents.countDocuments();
    const tasksCount = await db.tasks.countDocuments();
    const logsCount = await db.logs.countDocuments();
    
    print('Admin Users Count:', adminCount);
    print('Roles Count:', rolesCount);
    print('Agents Count:', agentsCount);
    print('Tasks Count:', tasksCount);
    print('Logs Count:', logsCount);
}

getCollectionCounts();
print('----------------------------------------');

// Print sample data details
print('\nSample Data Summary:');
print('----------------------------------------');
print('Roles:', JSON.stringify(db.roles.distinct('name')));
print('Agent Statuses:', JSON.stringify(db.agents.distinct('status')));
print('Task Types:', JSON.stringify(db.tasks.distinct('type')));
print('Task Statuses:', JSON.stringify(db.tasks.distinct('status')));
print('Log Statuses:', JSON.stringify(db.logs.distinct('status')));
print('----------------------------------------');
