#!/bin/bash

# Drop All Schemas Script
# Uses Dgraph's drop_all operation to completely clear all schema and data

echo "🗑️  Dropping ALL Schemas and Data from Hypermode Dgraph..."
echo "========================================================"
echo "⚠️  WARNING: This will delete ALL schema definitions AND data!"
echo "⚠️  This includes all predicates, types, and stored data!"
echo "⚠️  This action cannot be undone!"
echo ""

# Confirmation prompt
read -p "Are you sure you want to drop all schemas and data? (yes/no): " confirm
if [[ $confirm != "yes" ]]; then
    echo "❌ Drop operation cancelled."
    exit 0
fi

echo ""
echo "🔥 Proceeding with drop_all operation..."
echo ""

# Check if API_KEY environment variable is set
if [ -z "$API_KEY" ]; then
    echo "⚠️  API_KEY environment variable not set, using default from .env.dev.local"
    API_KEY="nZgKQjXX2XBRpt"
fi

# Perform drop_all operation
echo "🗑️  Executing drop_all operation..."
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    --data '{"drop_all": true}' \
    "https://do-study-do-study.hypermode.host/dgraph/alter")

# Extract HTTP status code (last line)
http_code=$(echo "$response" | tail -n1)
# Extract response body (all lines except last)
response_body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo "✅ Drop_all operation completed successfully!"
    if [ -n "$response_body" ] && [ "$response_body" != "{}" ]; then
        echo "📋 Response: $response_body"
    fi
else
    echo "❌ Drop_all operation failed (HTTP $http_code)"
    echo "📋 Response: $response_body"
    exit 1
fi

echo ""
echo "🎉 All schemas and data have been dropped!"
echo "📊 Database is now completely empty"
echo ""
echo "💡 Tip: You can now deploy fresh schemas using deploy_all_schemas.sh"
echo "🚀 Usage: API_KEY=nZgKQjXX2XBRpt ./deploy_all_schemas.sh"
