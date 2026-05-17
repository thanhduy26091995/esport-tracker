#!/bin/bash

# FC25 Esport Score Tracker - User Management API Test Script

echo "🧪 Testing User Management API"
echo "================================"
echo ""

BASE_URL="http://localhost:8080/api/v1"

# Check if server is running
echo "1️⃣  Checking server status..."
HEALTH=$(curl -s http://localhost:8080/health)
if echo "$HEALTH" | grep -q "ok"; then
  echo "✅ Server is running"
else
  echo "❌ Server is not running. Please start the backend first:"
  echo "   cd backend && go run cmd/server/main.go"
  exit 1
fi
echo ""

# Test GET all users (should be empty initially)
echo "2️⃣  Getting all users (should be empty)..."
curl -s $BASE_URL/users | jq '.'
echo ""

# Test CREATE user
echo "3️⃣ Creating user 'Ronaldo'..."
RESPONSE=$(curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Ronaldo"}')
echo "$RESPONSE" | jq '.'
USER1_ID=$(echo "$RESPONSE" | jq -r '.id')
echo "Created user ID: $USER1_ID"
echo ""

# Test CREATE another user
echo "4️⃣  Creating user 'Messi'..."
RESPONSE=$(curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Messi"}')
echo "$RESPONSE" | jq '.'
USER2_ID=$(echo "$RESPONSE" | jq -r '.id')
echo "Created user ID: $USER2_ID"
echo ""

# Test CREATE third user
echo "5️⃣  Creating user 'Neymar'..."
curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Neymar"}' | jq '.'
echo ""

# Test duplicate name (should fail)
echo "6️⃣  Testing duplicate name (should fail)..."
curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Ronaldo"}' | jq '.'
echo ""

# Test GET all users (should have 3 users)
echo "7️⃣  Getting all users..."
curl -s $BASE_URL/users | jq '.'
echo ""

# Test GET user by ID
echo "8️⃣  Getting user by ID..."
curl -s $BASE_URL/users/$USER1_ID | jq '.'
echo ""

# Test UPDATE user
echo "9️⃣  Updating user 'Ronaldo' to 'CR7'..."
curl -s -X PUT $BASE_URL/users/$USER1_ID \
  -H "Content-Type: application/json" \
  -d '{"name":"CR7"}' | jq '.'
echo ""

# Test GET leaderboard
echo "🔟 Getting leaderboard (top 10)..."
curl -s "$BASE_URL/users/leaderboard?limit=10" | jq '.'
echo ""

# Test DELETE user
echo "1️⃣1️⃣  Deleting user 'Messi'..."
curl -s -X DELETE $BASE_URL/users/$USER2_ID | jq '.'
echo ""

# Test GET all users after delete
echo "1️⃣2️⃣  Getting all users after delete..."
curl -s $BASE_URL/users | jq '.'
echo ""

echo "✅ All tests completed!"
echo ""
echo "📊 Summary:"
echo "  - Created 3 users"
echo "  - Updated 1 user"
echo "  - Deleted 1 user"
echo "  - Final count: 2 users"
