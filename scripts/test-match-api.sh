#!/bin/bash

# Test Match API endpoints
# This script tests the match recording functionality

BASE_URL="http://localhost:8080/api/v1"
FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "Testing Match API Endpoints"
echo "========================================="

# Clean up: Delete all existing matches and users
echo -e "\n${YELLOW}[SETUP] Cleaning up existing data...${NC}"
# Delete matches first (to avoid foreign key issues)
curl -s "$BASE_URL/matches" | jq -r '.[].id' | while read -r match_id; do
    curl -s -X DELETE "$BASE_URL/matches/$match_id" > /dev/null 2>&1
done
# Then delete users
curl -s "$BASE_URL/users" | jq -r '.[].id' | while read -r user_id; do
    curl -s -X DELETE "$BASE_URL/users/$user_id" > /dev/null 2>&1
done

# Step 1: Create test users
echo -e "\n${YELLOW}[STEP 1] Creating test users...${NC}"
USER1=$(curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice"}' | jq -r '.id')
echo "Created User 1 (Alice): $USER1"

USER2=$(curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob"}' | jq -r '.id')
echo "Created User 2 (Bob): $USER2"

USER3=$(curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie"}' | jq -r '.id')
echo "Created User 3 (Charlie): $USER3"

USER4=$(curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"David"}' | jq -r '.id')
echo "Created User 4 (David): $USER4"

# Verify initial scores are 0
echo -e "\n${YELLOW}[VERIFY] Checking initial scores...${NC}"
RESPONSE=$(curl -s "$BASE_URL/users")
ALICE_SCORE=$(echo "$RESPONSE" | jq -r ".[] | select(.name==\"Alice\") | .current_score")
BOB_SCORE=$(echo "$RESPONSE" | jq -r ".[] | select(.name==\"Bob\") | .current_score")
echo "Alice score: $ALICE_SCORE, Bob score: $BOB_SCORE"
if [ "$ALICE_SCORE" -eq 0 ] && [ "$BOB_SCORE" -eq 0 ]; then
    echo -e "${GREEN}✓ Initial scores correct${NC}"
else
    echo -e "${RED}✗ Initial scores incorrect${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 2: Create a 1v1 match (Alice vs Bob, Alice wins)
echo -e "\n${YELLOW}[STEP 2] Creating 1v1 match (Alice wins vs Bob)...${NC}"
MATCH1=$(curl -s -X POST "$BASE_URL/matches" -H "Content-Type: application/json" -d "{\"match_type\":\"1v1\",\"team1\":[\"$USER1\"],\"team2\":[\"$USER2\"],\"winner_team\":1}" | jq -r '.id')

if [ -n "$MATCH1" ] && [ "$MATCH1" != "null" ]; then
    echo -e "${GREEN}✓ Created 1v1 match: $MATCH1${NC}"
else
    echo -e "${RED}✗ Failed to create 1v1 match${NC}"
    FAILED=$((FAILED + 1))
fi

# Verify scores after 1v1 match
echo -e "\n${YELLOW}[VERIFY] Checking scores after 1v1...${NC}"
RESPONSE=$(curl -s "$BASE_URL/users")
ALICE_SCORE=$(echo "$RESPONSE" | jq -r ".[] | select(.name==\"Alice\") | .current_score")
BOB_SCORE=$(echo "$RESPONSE" | jq -r ".[] | select(.name==\"Bob\") | .current_score")
echo "Alice score: $ALICE_SCORE (expected: 1), Bob score: $BOB_SCORE (expected: -1)"
if [ "$ALICE_SCORE" -eq 1 ] && [ "$BOB_SCORE" -eq -1 ]; then
    echo -e "${GREEN}✓ Scores updated correctly after 1v1${NC}"
else
    echo -e "${RED}✗ Scores incorrect after 1v1${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 3: Create a 2v2 match (Alice+Charlie vs Bob+David, Team2 wins)
echo -e "\n${YELLOW}[STEP 3] Creating 2v2 match (Team2 wins: Bob+David vs Alice+Charlie)...${NC}"
MATCH2=$(curl -s -X POST "$BASE_URL/matches" -H "Content-Type: application/json" -d "{\"match_type\":\"2v2\",\"team1\":[\"$USER1\",\"$USER3\"],\"team2\":[\"$USER2\",\"$USER4\"],\"winner_team\":2}" | jq -r '.id')

if [ -n "$MATCH2" ] && [ "$MATCH2" != "null" ]; then
    echo -e "${GREEN}✓ Created 2v2 match: $MATCH2${NC}"
else
    echo -e "${RED}✗ Failed to create 2v2 match${NC}"
    FAILED=$((FAILED + 1))
fi

# Verify scores after 2v2 match
echo -e "\n${YELLOW}[VERIFY] Checking scores after 2v2...${NC}"
RESPONSE=$(curl -s "$BASE_URL/users")
ALICE_SCORE=$(echo "$RESPONSE" | jq -r ".[] | select(.name==\"Alice\") | .current_score")
BOB_SCORE=$(echo "$RESPONSE" | jq -r ".[] | select(.name==\"Bob\") | .current_score")
CHARLIE_SCORE=$(echo "$RESPONSE" | jq -r ".[] | select(.name==\"Charlie\") | .current_score")
DAVID_SCORE=$(echo "$RESPONSE" | jq -r ".[] | select(.name==\"David\") | .current_score")
echo "Alice: $ALICE_SCORE (expected: 0), Bob: $BOB_SCORE (expected: 0)"
echo "Charlie: $CHARLIE_SCORE (expected: -1), David: $DAVID_SCORE (expected: 1)"
if [ "$ALICE_SCORE" -eq 0 ] && [ "$BOB_SCORE" -eq 0 ] && [ "$CHARLIE_SCORE" -eq -1 ] && [ "$DAVID_SCORE" -eq 1 ]; then
    echo -e "${GREEN}✓ Scores updated correctly after 2v2${NC}"
else
    echo -e "${RED}✗ Scores incorrect after 2v2${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 4: Test invalid match type
echo -e "\n${YELLOW}[STEP 4] Testing invalid match type (3v3)...${NC}"
RESPONSE=$(curl -s -X POST "$BASE_URL/matches" -H "Content-Type: application/json" -d "{\"match_type\":\"3v3\",\"team1\":[\"$USER1\"],\"team2\":[\"$USER2\"],\"winner_team\":1}")
ERROR_CODE=$(echo "$RESPONSE" | jq -r '.code')
if [ "$ERROR_CODE" = "INVALID_MATCH_TYPE" ]; then
    echo -e "${GREEN}✓ Invalid match type rejected correctly${NC}"
else
    echo -e "${RED}✗ Invalid match type not handled properly${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 5: Test invalid team size for 1v1 (2 players in team1)
echo -e "\n${YELLOW}[STEP 5] Testing invalid team size for 1v1...${NC}"
RESPONSE=$(curl -s -X POST "$BASE_URL/matches" -H "Content-Type: application/json" -d "{\"match_type\":\"1v1\",\"team1\":[\"$USER1\",\"$USER3\"],\"team2\":[\"$USER2\"],\"winner_team\":1}")
ERROR_CODE=$(echo "$RESPONSE" | jq -r '.code')
if [ "$ERROR_CODE" = "INVALID_TEAM_SIZE" ]; then
    echo -e "${GREEN}✓ Invalid team size rejected correctly${NC}"
else
    echo -e "${RED}✗ Invalid team size not handled properly${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 6: Test duplicate player in same match
echo -e "\n${YELLOW}[STEP 6] Testing duplicate player in match...${NC}"
RESPONSE=$(curl -s -X POST "$BASE_URL/matches" -H "Content-Type: application/json" -d "{\"match_type\":\"1v1\",\"team1\":[\"$USER1\"],\"team2\":[\"$USER1\"],\"winner_team\":1}")
ERROR_CODE=$(echo "$RESPONSE" | jq -r '.code')
if [ "$ERROR_CODE" = "DUPLICATE_PLAYER" ]; then
    echo -e "${GREEN}✓ Duplicate player rejected correctly${NC}"
else
    echo -e "${RED}✗ Duplicate player not handled properly${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 7: Test invalid winner team
echo -e "\n${YELLOW}[STEP 7] Testing invalid winner team (3)...${NC}"
RESPONSE=$(curl -s -X POST "$BASE_URL/matches" -H "Content-Type: application/json" -d "{\"match_type\":\"1v1\",\"team1\":[\"$USER1\"],\"team2\":[\"$USER2\"],\"winner_team\":3}")
ERROR_CODE=$(echo "$RESPONSE" | jq -r '.code')
if [ "$ERROR_CODE" = "INVALID_WINNER_TEAM" ]; then
    echo -e "${GREEN}✓ Invalid winner team rejected correctly${NC}"
else
    echo -e "${RED}✗ Invalid winner team not handled properly${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 8: Get all matches
echo -e "\n${YELLOW}[STEP 8] Getting all matches...${NC}"
RESPONSE=$(curl -s "$BASE_URL/matches")
MATCH_COUNT=$(echo "$RESPONSE" | jq 'length')
echo "Found $MATCH_COUNT matches"
if [ "$MATCH_COUNT" -eq 2 ]; then
    echo -e "${GREEN}✓ Correct number of matches${NC}"
else
    echo -e "${RED}✗ Incorrect match count (expected 2, got $MATCH_COUNT)${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 9: Get match by ID
echo -e "\n${YELLOW}[STEP 9] Getting match by ID...${NC}"
RESPONSE=$(curl -s "$BASE_URL/matches/$MATCH1")
MATCH_TYPE=$(echo "$RESPONSE" | jq -r '.match_type')
WINNER_TEAM=$(echo "$RESPONSE" | jq -r '.winner_team')
if [ "$MATCH_TYPE" = "1v1" ] && [ "$WINNER_TEAM" -eq 1 ]; then
    echo -e "${GREEN}✓ Match details retrieved correctly${NC}"
else
    echo -e "${RED}✗ Match details incorrect${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 10: Get recent matches
echo -e "\n${YELLOW}[STEP 10] Getting recent matches (limit 1)...${NC}"
RESPONSE=$(curl -s "$BASE_URL/matches/recent?limit=1")
MATCH_COUNT=$(echo "$RESPONSE" | jq 'length')
if [ "$MATCH_COUNT" -eq 1 ]; then
    echo -e "${GREEN}✓ Recent matches retrieved correctly${NC}"
else
    echo -e "${RED}✗ Recent matches count incorrect${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 11: Get matches by user
echo -e "\n${YELLOW}[STEP 11] Getting matches for Alice...${NC}"
RESPONSE=$(curl -s "$BASE_URL/users/$USER1/matches")
MATCH_COUNT=$(echo "$RESPONSE" | jq 'length')
if [ "$MATCH_COUNT" -eq 2 ]; then
    echo -e "${GREEN}✓ User matches retrieved correctly (Alice in 2 matches)${NC}"
else
    echo -e "${RED}✗ User matches count incorrect (expected 2, got $MATCH_COUNT)${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 12: Get match stats
echo -e "\n${YELLOW}[STEP 12] Getting match statistics...${NC}"
RESPONSE=$(curl -s "$BASE_URL/matches/stats")
TOTAL_MATCHES=$(echo "$RESPONSE" | jq -r '.total_matches')
TODAY_MATCHES=$(echo "$RESPONSE" | jq -r '.today_matches')
echo "Total matches: $TOTAL_MATCHES, Today matches: $TODAY_MATCHES"
if [ "$TOTAL_MATCHES" -eq 2 ] && [ "$TODAY_MATCHES" -eq 2 ]; then
    echo -e "${GREEN}✓ Match statistics correct${NC}"
else
    echo -e "${RED}✗ Match statistics incorrect${NC}"
    FAILED=$((FAILED + 1))
fi

# Step 13: Delete a match and verify scores are reverted
echo -e "\n${YELLOW}[STEP 13] Deleting match and verifying score reversion...${NC}"
RESPONSE=$(curl -s -X DELETE "$BASE_URL/matches/$MATCH1")
MESSAGE=$(echo "$RESPONSE" | jq -r '.message')
if [ "$MESSAGE" = "Match deleted successfully" ]; then
    echo -e "${GREEN}✓ Match deleted successfully${NC}"
else
    echo -e "${RED}✗ Match deletion failed${NC}"
    FAILED=$((FAILED + 1))
fi

# Verify scores after deletion
echo -e "\n${YELLOW}[VERIFY] Checking scores after match deletion...${NC}"
RESPONSE=$(curl -s "$BASE_URL/users")
ALICE_SCORE=$(echo "$RESPONSE" | jq -r ".[] | select(.name==\"Alice\") | .current_score")
BOB_SCORE=$(echo "$RESPONSE" | jq -r ".[] | select(.name==\"Bob\") | .current_score")
echo "Alice: $ALICE_SCORE (expected: -1), Bob: $BOB_SCORE (expected: 1)"
if [ "$ALICE_SCORE" -eq -1 ] && [ "$BOB_SCORE" -eq 1 ]; then
    echo -e "${GREEN}✓ Scores reverted correctly after match deletion${NC}"
else
    echo -e "${RED}✗ Scores not reverted correctly${NC}"
    FAILED=$((FAILED + 1))
fi

# Summary
echo -e "\n========================================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
else
    echo -e "${RED}$FAILED test(s) failed ✗${NC}"
fi
echo "========================================="

exit $FAILED
