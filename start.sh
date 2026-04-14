#!/bin/bash

# FC25 Esport Score Tracker - Development Setup Script

set -e

echo "🚀 FC25 Esport Score Tracker - Setup"
echo "===================================="

# Check if PostgreSQL is running
echo "📊 Checking PostgreSQL..."
if ! docker ps | grep -q esport-postgres; then
  echo "Starting PostgreSQL container..."
  docker run --name esport-postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=esport_tracker \
    -p 5432:5432 \
    -d postgres:14
  echo "✅ PostgreSQL started"
  echo "⏳ Waiting for PostgreSQL to be ready..."
  sleep 3
else
  echo "✅ PostgreSQL already running"
fi

# Start backend
echo ""
echo "🔧 Starting backend server..."
cd backend
go run cmd/server/main.go &
BACKEND_PID=$!
cd ..

# Wait for backend to start
echo "⏳ Waiting for backend to be ready..."
sleep 3

# Start frontend
echo ""
echo "🎨 Starting frontend server..."
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "✅ All services started!"
echo ""
echo "📝 Access the application:"
echo "   Frontend: http://localhost:5173"
echo "   Backend:  http://localhost:8080"
echo "   Health:   http://localhost:8080/health"
echo ""
echo "Press Ctrl+C to stop all services"
echo ""

# Trap Ctrl+C and kill both processes
trap "kill $BACKEND_PID $FRONTEND_PID; exit" INT

# Wait for either process to exit
wait
