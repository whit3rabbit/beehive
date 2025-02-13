db.auth('admin', 'test_password')

db = db.getSiblingDB('manager_test_db')

db.createUser({
  user: 'test_user',
  pwd: 'test_password',
  roles: [
    {
      role: 'readWrite',
      db: 'manager_test_db'
    }
  ]
})
