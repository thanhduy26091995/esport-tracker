#!/bin/bash

echo "=== Restarting Backend Server ==="

# Kill old processes
echo "Killing old processes..."
lsof -ti :8080 | xargs kill -9 2>/dev/null
pkill -9 -f "go run cmd/server" 2>/dev/null
pkill -9 main 2>/dev/null
sleep 2

# Start fresh server
echo "Starting server..."
cd /Users/duyb/Documents/Growth/esport/backend
go run cmd/server/main.go > nohup.out 2>&1 &
sleep 6

# Check if server started
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ Server started successfully"
    
    # Run tests
    echo ""
    echo "=== Running Phase 3-5 Tests ==="
    /Users/duyb/Documents/Growth/esport/test-backend-phase3-5.sh
else
    echo "❌ Server failed to start"
    echo "Last 20 lines of log:"
    tail -20 nohup.out
    exit 1
fi
