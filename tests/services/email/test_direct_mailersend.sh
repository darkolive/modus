#!/bin/bash

echo "🧪 Testing Direct MailerSend API Call (Outside Modus)"
echo "This will help determine if the issue is with Modus or MailerSend"
echo ""

# Read API key from environment
if [ -z "$API_KEY" ]; then
    echo "❌ API_KEY environment variable not set"
    echo "Please set API_KEY with your MailerSend API key"
    exit 1
fi

echo "🔑 Using API Key: ${API_KEY:0:10}..."
echo ""

# Direct curl request to MailerSend API
echo "🚀 Making direct HTTP POST to MailerSend API..."

curl -X POST "https://api.mailersend.com/v1/email" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "from": {
      "email": "test@trial-3vz9dle7zzjg2k8e.mlsender.net",
      "name": "DO Study Platform"
    },
    "to": [
      {
        "email": "darren@darkolive.co.uk",
        "name": "User"
      }
    ],
    "subject": "Direct API Test",
    "template_id": "neqvygm91v8l0p7w",
    "personalization": [
      {
        "email": "darren@darkolive.co.uk",
        "data": {
          "otp_code": "123456",
          "purpose": "direct API test",
          "expires": "5 minutes"
        }
      }
    ]
  }' \
  -w "\n\nHTTP Status: %{http_code}\nResponse Time: %{time_total}s\n" \
  -v

echo ""
echo "✅ Direct MailerSend API test completed!"
