#!/bin/bash

# Test Backend Phase 3-5: Config, Fund, Settlement APIs
# This script tests configuration, fund management, and debt settlement

BASE_URL="http://localhost:8080/api/v1"
FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "Testing Backend Phase 3-5 APIs"
echo "========================================="

# Clean up existing data
echo -e "\n${YELLOW}[SETUP] Cleaning up existing data...${NC}"
curl -s "$BASE_URL/matches" | jq -r '.[].id' 2>/dev/null | while read -r match_id; do
    curl -s -X DELETE "$BASE_URL/matches/$match_id" > /dev/null 2>&1
done
curl -s "$BASE_URL/users" | jq -r '.[].id' 2>/dev/null | while read -r user_id; do
    curl -s -X DELETE "$BASE_URL/users/$user_id" > /dev/null 2>&1
done

# ===================================
# PHASE 5: CONFIG API TESTS
# ===================================

echo -e "\n${YELLOW}=========================================${NC}"
echo -e "${YELLOW}PHASE 5: CONFIGURATION API${NC}"
echo -e "${YELLOW}=========================================${NC}"

# Test 1: Get all config
echo -e "\n${YELLOW}[TEST 1] Get all configuration...${NC}"
CONFIG_COUNT=$(curl -s "$BASE_URL/config" | jq 'length')
if [ "$CONFIG_COUNT" -eq 3 ]; then
    echo -e "${GREEN}✓ All config entries retrieved (3)${NC}"
else
    echo -e "${RED}✗ Config count incorrect (expected 3, got $CONFIG_COUNT)${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 2: Get specific config key
echo -e "\n${YELLOW}[TEST 2] Get debt_threshold config...${NC}"
DEBT_THRESHOLD=$(curl -s "$BASE_URL/config/debt_threshold" | jq -r '.value')
if [ "$DEBT_THRESHOLD" = "-6" ]; then
    echo -e "${GREEN}✓ Debt threshold retrieved correctly: $DEBT_THRESHOLD${NC}"
else
    echo -e "${RED}✗ Debt threshold incorrect (expected -6, got $DEBT_THRESHOLD)${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 3: Update config
echo -e "\n${YELLOW}[TEST 3] Update point_to_vnd config...${NC}"
RESPONSE=$(curl -s -X PUT "$BASE_URL/config/point_to_vnd" -H "Content-Type: application/json" -d '{"value":"25000"}')
NEW_VALUE=$(echo "$RESPONSE" | jq -r '.value')
if [ "$NEW_VALUE" = "25000" ]; then
    echo -e "${GREEN}✓ Config updated successfully${NC}"
else
    echo -e "${RED}✗ Config update failed${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 4: Validate config update (negative test)
echo -e "\n${YELLOW}[TEST 4] Test invalid debt_threshold (positive value)...${NC}"
RESPONSE=$(curl -s -X PUT "$BASE_URL/config/debt_threshold" -H "Content-Type: application/json" -d '{"value":"5"}')
ERROR_CODE=$(echo "$RESPONSE" | jq -r '.code')
if [ "$ERROR_CODE" = "VALIDATION_ERROR" ]; then
    echo -e "${GREEN}✓ Invalid config rejected correctly${NC}"
else
    echo -e "${RED}✗ Invalid config not rejected${NC}"
    FAILED=$((FAILED + 1))
fi

# Restore original value
curl -s -X PUT "$BASE_URL/config/point_to_vnd" -H "Content-Type: application/json" -d '{"value":"22000"}' > /dev/null

# ===================================
# PHASE 4: FUND API TESTS
# ===================================

echo -e "\n${YELLOW}=========================================${NC}"
echo -e "${YELLOW}PHASE 4: FUND MANAGEMENT API${NC}"
echo -e "${YELLOW}=========================================${NC}"

# Test 5: Get initial fund balance
echo -e "\n${YELLOW}[TEST 5] Get initial fund balance...${NC}"
BALANCE=$(curl -s "$BASE_URL/fund/balance" | jq -r '.balance')
if [ "$BALANCE" -eq 0 ]; then
    echo -e "${GREEN}✓ Initial balance is 0${NC}"
else
    echo -e "${RED}✗ Initial balance incorrect (expected 0, got $BALANCE)${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 6: Create deposit
echo -e "\n${YELLOW}[TEST 6] Create deposit (100,000 VND)...${NC}"
DEPOSIT=$(curl -s -X POST "$BASE_URL/fund/deposit" -H "Content-Type: application/json" -d '{
  "amount": 100000,
  "description": "Initial deposit"
}')
DEPOSIT_AMOUNT=$(echo "$DEPOSIT" | jq -r '.amount')
if [ "$DEPOSIT_AMOUNT" -eq 100000 ]; then
    echo -e "${GREEN}✓ Deposit created successfully${NC}"
else
    echo -e "${RED}✗ Deposit creation failed${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 7: Verify balance after deposit
echo -e "\n${YELLOW}[TEST 7] Verify balance after deposit...${NC}"
BALANCE=$(curl -s "$BASE_URL/fund/balance" | jq -r '.balance')
if [ "$BALANCE" -eq 100000 ]; then
    echo -e "${GREEN}✓ Balance updated correctly: 100,000 VND${NC}"
else
    echo -e "${RED}✗ Balance incorrect (expected 100000, got $BALANCE)${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 8: Create withdrawal
echo -e "\n${YELLOW}[TEST 8] Create withdrawal (30,000 VND)...${NC}"
WITHDRAWAL=$(curl -s -X POST "$BASE_URL/fund/withdrawal" -H "Content-Type: application/json" -d '{
  "amount": 30000,
  "description": "Equipment purchase"
}')
WITHDRAWAL_AMOUNT=$(echo "$WITHDRAWAL" | jq -r '.amount')
if [ "$WITHDRAWAL_AMOUNT" -eq 30000 ]; then
    echo -e "${GREEN}✓ Withdrawal created successfully${NC}"
else
    echo -e "${RED}✗ Withdrawal creation failed${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 9: Verify balance after withdrawal
echo -e "\n${YELLOW}[TEST 9] Verify balance after withdrawal...${NC}"
BALANCE=$(curl -s "$BASE_URL/fund/balance" | jq -r '.balance')
if [ "$BALANCE" -eq 70000 ]; then
    echo -e "${GREEN}✓ Balance updated correctly: 70,000 VND${NC}"
else
    echo -e "${RED}✗ Balance incorrect (expected 70000, got $BALANCE)${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 10: Test insufficient balance
echo -e "\n${YELLOW}[TEST 10] Test withdrawal with insufficient balance...${NC}"
RESPONSE=$(curl -s -X POST "$BASE_URL/fund/withdrawal" -H "Content-Type: application/json" -d '{
  "amount": 100000,
  "description": "Over withdrawal"
}')
ERROR_CODE=$(echo "$RESPONSE" | jq -r '.code')
if [ "$ERROR_CODE" = "INSUFFICIENT_BALANCE" ]; then
    echo -e "${GREEN}✓ Insufficient balance error returned correctly${NC}"
else
    echo -e "${RED}✗ Insufficient balance not handled (got: $ERROR_CODE)${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 11: Get fund transactions
echo -e "\n${YELLOW}[TEST 11] Get fund transactions...${NC}"
TRANSACTION_COUNT=$(curl -s "$BASE_URL/fund/transactions" | jq 'length')
if [ "$TRANSACTION_COUNT" -eq 2 ]; then
    echo -e "${GREEN}✓ All transactions retrieved (2)${NC}"
else
    echo -e "${RED}✗ Transaction count incorrect (expected 2, got $TRANSACTION_COUNT)${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 12: Get fund stats
echo -e "\n${YELLOW}[TEST 12] Get fund statistics...${NC}"
STATS=$(curl -s "$BASE_URL/fund/stats")
STAT_BALANCE=$(echo "$STATS" | jq -r '.current_balance')
STAT_TOTAL=$(echo "$STATS" | jq -r '.total_transactions')
if [ "$STAT_BALANCE" -eq 70000 ] && [ "$STAT_TOTAL" -eq 2 ]; then
    echo -e "${GREEN}✓ Fund stats correct (balance: 70,000, transactions: 2)${NC}"
else
    echo -e "${RED}✗ Fund stats incorrect${NC}"
    FAILED=$((FAILED + 1))
fi

# ===================================
# PHASE 3: SETTLEMENT API TESTS
# ===================================

echo -e "\n${YELLOW}=========================================${NC}"
echo -e "${YELLOW}PHASE 3: DEBT SETTLEMENT API${NC}"
echo -e "${YELLOW}=========================================${NC}"

# Create test users for settlement
echo -e "\n${YELLOW}[SETUP] Creating test users...${NC}"
PLAYER1=$(curl -s -X POST "$BASE_URL/users" -H "Content-Type: application/json" -d '{"name":"Player1"}' | jq -r '.id')
PLAYER2=$(curl -s -X POST "$BASE_URL/users" -H "Content-Type: application/json" -d '{"name":"Player2"}' | jq -r '.id')
PLAYER3=$(curl -s -X POST "$BASE_URL/users" -H "Content-Type: application/json" -d '{"name":"Player3"}' | jq -r '.id')
echo "Created 3 players for settlement testing"

# Test 13: Create matches to build debt
echo -e "\n${YELLOW}[TEST 13] Creating matches to build debt for Player1...${NC}"
for i in {1..7}; do
    curl -s -X POST "$BASE_URL/matches" -H "Content-Type: application/json" -d "{
        \"match_type\":\"1v1\",
        \"team1\":[\"$PLAYER1\"],
        \"team2\":[\"$PLAYER2\"],
        \"winner_team\":2
    }" > /dev/null
done
PLAYER1_SCORE=$(curl -s "$BASE_URL/users/$PLAYER1" | jq -r '.current_score')
echo "Player1 score after 7 losses: $PLAYER1_SCORE (should trigger settlement at -6)"

# Test 14: Verify auto-settlement triggered
echo -e "\n${YELLOW}[TEST 14] Verify auto-settlement was triggered...${NC}"
sleep 1 # Give time for settlement to process
SETTLEMENT_COUNT=$(curl -s "$BASE_URL/settlements" | jq 'length')
if [ "$SETTLEMENT_COUNT" -ge 1 ]; then
    echo -e "${GREEN}✓ Settlement auto-triggered (count: $SETTLEMENT_COUNT)${NC}"
else
    echo -e "${RED}✗ Settlement not triggered (count: $SETTLEMENT_COUNT)${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 15: Verify debtor score reset
echo -e "\n${YELLOW}[TEST 15] Verify Player1 score reset to 0...${NC}"
PLAYER1_SCORE=$(curl -s "$BASE_URL/users/$PLAYER1" | jq -r '.current_score')
if [ "$PLAYER1_SCORE" -eq 0 ]; then
    echo -e "${GREEN}✓ Debtor score reset correctly${NC}"
else
    echo -e "${RED}✗ Debtor score not reset (expected 0, got $PLAYER1_SCORE)${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 16: Verify fund balance increased
echo -e "\n${YELLOW}[TEST 16] Verify fund received settlement deposit...${NC}"
NEW_BALANCE=$(curl -s "$BASE_URL/fund/balance" | jq -r '.balance')
if [ "$NEW_BALANCE" -gt 70000 ]; then
    INCREASE=$((NEW_BALANCE - 70000))
    echo -e "${GREEN}✓ Fund balance increased by $INCREASE VND${NC}"
else
    echo -e "${RED}✗ Fund balance did not increase${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 17: Get settlement details
echo -e "\n${YELLOW}[TEST 17] Get settlement details...${NC}"
SETTLEMENT=$(curl -s "$BASE_URL/settlements" | jq '.[0]')
SETTLEMENT_MONEY=$(echo "$SETTLEMENT" | jq -r '.money_amount')
SETTLEMENT_FUND=$(echo "$SETTLEMENT" | jq -r '.fund_amount')
SETTLEMENT_WINNER=$(echo "$SETTLEMENT" | jq -r '.winner_distribution')
echo "Settlement: Total=$SETTLEMENT_MONEY, Fund=$SETTLEMENT_FUND, Winners=$SETTLEMENT_WINNER"
if [ "$SETTLEMENT_MONEY" -gt 0 ]; then
    echo -e "${GREEN}✓ Settlement details retrieved${NC}"
else
    echo -e "${RED}✗ Settlement details incorrect${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 18: Verify matches locked after settlement
echo -e "\n${YELLOW}[TEST 18] Verify related matches are locked...${NC}"
LOCKED_COUNT=$(curl -s "$BASE_URL/users/$PLAYER1/matches" | jq '[.[] | select(.is_locked == true)] | length')
if [ "$LOCKED_COUNT" -ge 1 ]; then
    echo -e "${GREEN}✓ Matches locked after settlement (count: $LOCKED_COUNT)${NC}"
else
    echo -e "${RED}✗ Matches not locked${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 19: Test manual settlement trigger (no debt case)
echo -e "\n${YELLOW}[TEST 19] Test manual settlement with no debt...${NC}"
RESPONSE=$(curl -s -X POST "$BASE_URL/settlements/trigger" -H "Content-Type: application/json" -d "{\"debtor_id\":\"$PLAYER2\"}")
ERROR_CODE=$(echo "$RESPONSE" | jq -r '.code')
if [ "$ERROR_CODE" = "NO_DEBT" ]; then
    echo -e "${GREEN}✓ Settlement correctly rejected for user with no debt${NC}"
else
    echo -e "${RED}✗ Settlement validation failed${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 20: Get user settlement history
echo -e "\n${YELLOW}[TEST 20] Get Player1 settlement history...${NC}"
HISTORY_COUNT=$(curl -s "$BASE_URL/users/$PLAYER1/settlements" | jq 'length')
if [ "$HISTORY_COUNT" -ge 1 ]; then
    echo -e "${GREEN}✓ User settlement history retrieved${NC}"
else
    echo -e "${RED}✗ Settlement history not found${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 21: Get settlement stats
echo -e "\n${YELLOW}[TEST 21] Get settlement statistics...${NC}"
STATS=$(curl -s "$BASE_URL/settlements/stats")
TOTAL_SETTLEMENTS=$(echo "$STATS" | jq -r '.total_settlements')
TODAY_SETTLEMENTS=$(echo "$STATS" | jq -r '.today_settlements')
echo "Total settlements: $TOTAL_SETTLEMENTS, Today: $TODAY_SETTLEMENTS"
if [ "$TOTAL_SETTLEMENTS" -ge 1 ]; then
    echo -e "${GREEN}✓ Settlement stats correct${NC}"
else
    echo -e "${RED}✗ Settlement stats incorrect${NC}"
    FAILED=$((FAILED + 1))
fi

# Summary
echo -e "\n========================================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All 21 tests passed! ✓${NC}"
    echo -e "${GREEN}Backend Phase 3-5 APIs fully functional${NC}"
else
    echo -e "${RED}$FAILED test(s) failed ✗${NC}"
fi
echo "========================================="

exit $FAILED
