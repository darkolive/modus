#!/bin/bash

# Deploy Combined Schema to Hypermode Dgraph
# Uses the working combined.dql schema file with correct DQL syntax

echo "🚀 Deploying Combined Schema to Hypermode Dgraph..."
echo "=================================================="
echo "📄 Using auto-generated schema.dql with correct DQL syntax"
echo ""

# Check if API_KEY environment variable is set
if [ -z "$API_KEY" ]; then
    echo "⚠️  API_KEY environment variable not set, using default from .env.dev.local"
    API_KEY="nZgKQjXX2XBRpt"
fi

# Path to the auto-generated schema file
SCHEMA_FILE="$(dirname "$0")/schema/schema.dql"

# Check if schema file exists
if [ ! -f "$SCHEMA_FILE" ]; then
    echo "❌ Error: Schema file not found at $SCHEMA_FILE"
    echo "💡 Run ./combine_schema.sh to generate the schema.dql file"
    exit 1
fi

# Check if file is empty
if [ ! -s "$SCHEMA_FILE" ]; then
    echo "❌ Error: Schema file is empty"
    echo "💡 Run ./combine_schema.sh to generate the schema.dql file"
    exit 1
fi

echo "📄 Deploying auto-generated schema..."
echo "   📁 File: $SCHEMA_FILE"
echo "   📊 Size: $(wc -l < "$SCHEMA_FILE") lines"
echo ""

# Deploy the combined schema
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    --data-binary "@$SCHEMA_FILE" \
    "https://do-study-do-study.hypermode.host/dgraph/alter")

# Extract HTTP status code (last line)
http_code=$(echo "$response" | tail -n1)
# Extract response body (all lines except last)
response_body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo "✅ Auto-generated schema deployed successfully!"
    if [ -n "$response_body" ] && [ "$response_body" != "{}" ]; then
        echo "📋 Response: $response_body"
    fi
    echo ""
    echo "🎉 Schema deployment completed!"
    echo "================================================"
    echo "📊 Summary:"
    echo "   ✅ Auto-generated schema deployed with correct DQL syntax"
    echo "   🔧 Includes all predicates and type definitions"
    echo "   📈 Ready for use by authentication agents"
    echo ""
    echo "💡 Tip: Use ./check_schema.sh to verify the deployed schema"
else
    echo "❌ Auto-generated schema deployment failed (HTTP $http_code)"
    echo "📋 Response: $response_body"
    exit 1
fi
