// Get environment variables or use defaults
const dbUser = process.env.MONGODB_USER || 'manager';
const dbPassword = process.env.MONGODB_PASSWORD;
const dbName = process.env.MONGODB_DATABASE || 'manager_db';

if (!dbPassword) {
    throw new Error('MONGODB_PASSWORD environment variable is required');
}

// Create application database
db = db.getSiblingDB(dbName);

// Create application user
db.createUser({
    user: dbUser,
    pwd: dbPassword,
    roles: [
        {
            role: 'readWrite',
            db: dbName
        }
    ]
});

// Create collections
db.createCollection('admins');
db.createCollection('agents');
db.createCollection('roles');
db.createCollection('tasks');
db.createCollection('logs');

// Create indexes
db.admins.createIndex({ "username": 1 }, { unique: true });
db.agents.createIndex({ "uuid": 1 }, { unique: true });
db.agents.createIndex({ "api_key": 1 }, { unique: true });
db.agents.createIndex({ "last_seen": 1 });
db.tasks.createIndex({ "agent_id": 1 });
db.tasks.createIndex({ "status": 1 });
db.tasks.createIndex({ "created_at": 1 });
db.logs.createIndex({ "timestamp": 1 }, { expireAfterSeconds: 2592000 }); // 30 days TTL

print('MongoDB initialization completed successfully');