#!/bin/bash

# Delete All OTPs - Clean up all authentication-related data
echo "🗑️  Deleting All Authentication Data..."
echo "====================================="

# First, let's see what we're about to delete
echo "🔍 Current OTP records before deletion:"
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ otps(func: type(ChannelOTP)) { uid channelHash channelType otpHash verified expiresAt createdAt userId purpose used } }"
  }' | jq '.data.otps | length' 2>/dev/null

# Use the Go SDK-based deletion utility for proper Modus integration
echo "Using Modus SDK for proper deletion..."
cd "$(dirname "$0")/../../.." # Navigate to modus root directory
go run cmd/delete_auth_data.go dgraph

echo -e "\n🔍 Verifying deletion - checking remaining records..."

# Verify deletion by counting remaining records
echo "📊 Remaining ChannelOTP records:"
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ otps(func: type(ChannelOTP)) { uid } }"
  }' | jq '.data.otps | length' 2>/dev/null

echo "📊 Remaining AuthSession records:"
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ sessions(func: type(AuthSession)) { uid } }"
  }' | jq '.data.sessions | length' 2>/dev/null

echo -e "\n✅ Authentication data cleanup completed!"
echo "💡 Use ./list_all_otps.sh to verify the cleanup was successful"
