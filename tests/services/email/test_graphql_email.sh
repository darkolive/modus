#!/bin/bash

echo "🧪 Testing Email Service via GraphQL"
echo "Recipient: darren@darkolive.co.uk"
echo ""

# GraphQL mutation to send OTP email
echo "🚀 Sending OTP email via GraphQL..."

curl -X POST http://localhost:8686/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { sendOTP(req: { channel: \"email\", recipient: \"darren@darkolive.co.uk\" }) { oTPID sent channel expiresAt message } }"
  }' | jq '.'

echo ""
echo "✅ GraphQL email test completed!"
