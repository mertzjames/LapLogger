#!/bin/bash

# LapLogger - Startup Script
echo "Starting LapLogger..."

# Start backend in background
echo "Starting Go backend on :8080..."
cd backend && go run . &
BACKEND_PID=$!

# Wait a bit for backend to start
sleep 3

# Start frontend
echo "Starting React frontend on :3000..."
cd ../frontend && npm start &
FRONTEND_PID=$!

echo "✅ LapLogger is running!"
echo "🔗 Frontend: http://localhost:3000"
echo "🔗 Backend API: http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop both servers"

# Wait for user to press Ctrl+C
trap 'kill $BACKEND_PID $FRONTEND_PID; exit' INT
wait
