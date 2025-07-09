#!/bin/bash

# List All OTPs - Query all authentication-related data
echo "🔐 Querying All Authentication Data..."
echo "====================================="

# DQL query to get all ChannelOTP types with all fields
echo "🔍 Fetching all ChannelOTP records..."
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ otps(func: type(ChannelOTP)) { uid channelHash channelType otpHash verified expiresAt createdAt userId purpose used } }"
  }' | jq '.'

echo -e "\n\n🔍 Fetching all AuthSession records..."
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ sessions(func: type(AuthSession)) { uid userId token expiresAt createdAt ipAddress userAgent active } }"
  }' | jq '.'

echo -e "\n\n📊 Active OTPs (not used and not expired)..."
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ active_otps(func: type(ChannelOTP)) @filter(eq(used, false) AND gt(expiresAt, \"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'\")) { uid channelType verified expiresAt createdAt userId purpose } }"
  }' | jq '.'

echo -e "\n\n📊 Verified OTPs..."
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ verified_otps(func: type(ChannelOTP)) @filter(eq(verified, true)) { uid channelType verified expiresAt createdAt userId purpose used } }"
  }' | jq '.'

echo -e "\n\n✅ Authentication data query completed!"
