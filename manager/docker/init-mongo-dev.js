print('Starting MongoDB initialization...');

// Enhanced readiness check with multiple attempts
let retries = 5;
while (retries-- > 0) {
    try {
        // Basic command to check server status
        db.adminCommand({ ping: 1 });
        break;
    } catch (error) {
        print(`Waiting for MongoDB to be ready... (${retries} retries left)`);
        sleep(3000);
    }
}

// Create admin user with full privileges
try {
    // First check if admin user exists
    const adminDB = db.getSiblingDB('admin');
    const adminUser = adminDB.getUser('admin');
    
    if (!adminUser) {
        adminDB.createUser({
            user: 'admin',
            pwd: 'admin_password',
            roles: [
                { role: 'root', db: 'admin' },  // Full superuser role
                { role: 'userAdminAnyDatabase', db: 'admin' },
                { role: 'dbAdminAnyDatabase', db: 'admin' }
            ]
        });
        print('Root admin user created successfully');
    } else {
        print('Admin user already exists, skipping creation');
    }
} catch (error) {
    throw new Error(`FATAL: Admin user check/creation failed: ${error}`);
}

// Secure authentication with retries
let authSuccess = false;
for (let i = 0; i < 5; i++) {
    try {
        if (db.getSiblingDB('admin').auth('admin', 'admin_password')) {
            authSuccess = true;
            break;
        }
    } catch (error) {
        print(`Authentication attempt ${i+1}/5 failed: ${error}`);
        sleep(2000);
    }
}

if (!authSuccess) {
    throw new Error('FATAL: Authentication failed after multiple attempts');
}

// Database setup with transaction support
try {
    const managerDB = db.getSiblingDB('manager_db');
    
    // User creation with transaction
    const session = managerDB.getMongo().startSession();
    session.startTransaction();
    
    try {
        // Check if manager user exists
        const managerUser = managerDB.getUser('manager');
        if (!managerUser) {
            managerDB.createUser({
                user: 'manager',
                pwd: 'manager_password',
                roles: [
                    { role: 'readWrite', db: 'manager_db' },
                    { role: 'dbAdmin', db: 'manager_db' }
                ]
            });
        } else {
            print('Manager user already exists, skipping creation');
        }
        
        // Collection creation in transaction
        ['admins', 'agents', 'roles', 'tasks', 'logs'].forEach(collection => {
            if (!managerDB.getCollection(collection).exists()) {
                managerDB.createCollection(collection);
            }
        });
        
        session.commitTransaction();
        print('Manager database initialized successfully');
    } catch (error) {
        session.abortTransaction();
        throw error;
    }
} catch (error) {
    print(`Manager setup error (non-fatal): ${error}`);
    print('Continuing with initialization...');
}

// Index creation with optimized options
const indexes = [
    { collection: 'admins', keys: { username: 1 }, options: { unique: true } },
    { 
        collection: 'agents', 
        keys: { uuid: 1 }, 
        options: { 
            unique: true,
            partialFilterExpression: { status: { $eq: "active" } }
        }
    },
    { collection: 'tasks', keys: { created_at: -1 }, options: { expireAfterSeconds: 2592000 } }
];

indexes.forEach(({ collection, keys, options }) => {
    try {
        db.getSiblingDB('manager_db')[collection].createIndex(keys, options || {});
    } catch (error) {
        if (error.codeName !== 'IndexKeySpecsConflict') {
            print(`WARNING: Index creation error: ${error}`);
        }
    }
});

// Data seeding with bulk operations
try {
    const adminsBulk = db.getSiblingDB('manager_db').admins.initializeUnorderedBulkOp();
    adminsBulk.find({ username: 'admin' }).upsert().updateOne({
        $setOnInsert: {
            username: 'admin',
            password: '$2a$10$IiPwfeWr7rGFZpXWR6wONuV7CUHbBJ4RZWqMWvwxF3f8qZvzZv3Ei',
            email: 'admin@example.com',
            created_at: new Date(),
            updated_at: new Date()
        }
    });
    adminsBulk.execute();
    
    const rolesBulk = db.getSiblingDB('manager_db').roles.initializeUnorderedBulkOp();
    rolesBulk.find({ name: 'admin' }).upsert().updateOne({
        $setOnInsert: {
            name: 'admin',
            description: 'Administrator role with full access',
            applications: ['all'],
            default_tasks: ['scan', 'execute', 'monitor'],
            created_at: new Date()
        }
    });
    rolesBulk.execute();
} catch (error) {
    print(`WARNING: Data seeding error: ${error}`);
}

print('\nMongoDB initialization completed successfully!');
print('=============================================');
print('Development Environment Credentials:');
print('1. Admin UI:');
print('   - Username: admin');
print('   - Password: DevAdmin123!@#');
print('2. MongoDB:');
print('   - Root User: admin/admin_password');
print('   - App User: manager/manager_password');
print('=============================================');