#!/bin/sh

# Trap SIGTERM and SIGINT
trap 'cleanup' SIGTERM SIGINT

# Cleanup function for graceful shutdown
cleanup() {
    echo "Received shutdown signal, performing cleanup..."
    
    # Send SIGTERM to the main process
    if [ -f /app/tmp/manager.pid ]; then
        kill -TERM $(cat /app/tmp/manager.pid) 2>/dev/null
    fi
    
    # Wait for the process to terminate
    wait
    
    echo "Cleanup completed"
    exit 0
}

# Function to wait for MongoDB
wait_for_mongodb() {
    echo "Waiting for MongoDB to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if wget --spider --quiet http://mongo:27017; then
            echo "MongoDB is up - executing command"
            return 0
        fi
        
        echo "MongoDB is unavailable (attempt $attempt/$max_attempts) - sleeping"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "Failed to connect to MongoDB after $max_attempts attempts"
    return 1
}

# Wait for MongoDB to be ready
if ! wait_for_mongodb; then
    echo "Could not connect to MongoDB - exiting"
    exit 1
fi

# Create PID file directory if it doesn't exist
mkdir -p /app/tmp

# Start the application and save its PID
echo "Starting the application..."
/app/bin/manager & 
echo $! > /app/tmp/manager.pid

# Wait for the application process
wait $!