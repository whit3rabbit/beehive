// Create application user and database
db.auth(process.env.MONGO_INITDB_ROOT_USERNAME, process.env.MONGO_INITDB_ROOT_PASSWORD)

db = db.getSiblingDB(process.env.MONGODB_DATABASE)

db.createUser({
    user: process.env.MONGODB_USER,
    pwd: process.env.MONGODB_PASSWORD,
    roles: [
        {
            role: "readWrite",
            db: process.env.MONGODB_DATABASE
        }
    ]
})

// Create initial collections
db.createCollection('admins')
db.createCollection('agents')
db.createCollection('roles')
db.createCollection('tasks')
db.createCollection('logs')

// Create indexes
db.admins.createIndex({ "username": 1 }, { unique: true })
db.agents.createIndex({ "uuid": 1 }, { unique: true })
db.agents.createIndex({ "api_key": 1 }, { unique: true })
db.agents.createIndex({ "last_seen": 1 })
db.tasks.createIndex({ "agent_id": 1 })
db.tasks.createIndex({ "status": 1 })
db.tasks.createIndex({ "created_at": 1 })
db.logs.createIndex({ "timestamp": 1 }, { expireAfterSeconds: 2592000 }) // 30 days TTL