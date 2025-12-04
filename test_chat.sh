#!/bin/bash

# Script untuk test fitur chat AI
# Usage: ./test_chat.sh

set -e

BASE_URL="http://localhost:8080/api/v1"
EMAIL="test@example.com"
PASSWORD="password123"

echo "🧪 Testing Zodiac AI Chat Feature"
echo "=================================="
echo ""

# 1. Register (skip if already exists)
echo "1️⃣  Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"testuser\",
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"zodiac_sign\": \"Aries\",
    \"birth_date\": \"1990-04-15\"
  }" || echo '{"success":false}')

if echo "$REGISTER_RESPONSE" | grep -q '"success":true'; then
  echo "✅ User registered successfully"
else
  echo "⚠️  User might already exist, continuing..."
fi
echo ""

# 2. Login
echo "2️⃣  Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
  echo "❌ Failed to get access token"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Logged in successfully"
echo "Token: ${ACCESS_TOKEN:0:20}..."
echo ""

# 3. Create Chat Session
echo "3️⃣  Creating chat session..."
SESSION_RESPONSE=$(curl -s -X POST "$BASE_URL/chat/sessions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "title": "Test Chat Session"
  }')

SESSION_ID=$(echo "$SESSION_RESPONSE" | grep -o '"_id":"[^"]*' | cut -d'"' -f4)

if [ -z "$SESSION_ID" ]; then
  echo "❌ Failed to create session"
  echo "Response: $SESSION_RESPONSE"
  exit 1
fi

echo "✅ Session created successfully"
echo "Session ID: $SESSION_ID"
echo ""

# 4. Send Message to AI
echo "4️⃣  Sending message to AI..."
MESSAGE_RESPONSE=$(curl -s -X POST "$BASE_URL/chat/sessions/$SESSION_ID/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "message": "Halo! Apa kabar hari ini?"
  }')

echo "Response:"
echo "$MESSAGE_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$MESSAGE_RESPONSE"
echo ""

# Check if successful
if echo "$MESSAGE_RESPONSE" | grep -q '"success":true'; then
  echo "✅ AI Chat is working!"
  
  # Extract AI response
  AI_RESPONSE=$(echo "$MESSAGE_RESPONSE" | grep -o '"content":"[^"]*","sender":"ai"' | head -1 | cut -d'"' -f4)
  if [ ! -z "$AI_RESPONSE" ]; then
    echo ""
    echo "🤖 AI Response: $AI_RESPONSE"
  fi
else
  echo "❌ AI Chat failed"
  echo "Check server logs for details"
fi

echo ""
echo "=================================="
echo "Test completed!"
