#!/bin/bash

# RBAC Discovery Script for Dgraph v25
# Investigates current RBAC configuration and capabilities

echo "🔍 Discovering Dgraph v25 RBAC Configuration..."
echo "=============================================="

# Get API key
if [ -z "$API_KEY" ]; then
    if [ -f ".env.dev.local" ]; then
        API_KEY=$(grep "^API_KEY=" ".env.dev.local" | cut -d'=' -f2 | tr -d '"')
        echo "⚠️  Using API_KEY from .env.dev.local"
    else
        echo "❌ Error: API_KEY environment variable not set"
        exit 1
    fi
fi

API_ENDPOINT="https://do-study-do-study.hypermode.host"

echo ""
echo "🔍 Phase 1: Check Admin Endpoints"
echo "================================"

# Check if admin endpoints are available
echo "📡 Testing admin endpoint availability..."
admin_response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"query": "{ health }"}' \
    "$API_ENDPOINT/admin")

admin_code=$(echo "$admin_response" | tail -n1)
admin_body=$(echo "$admin_response" | sed '$d')

if [ "$admin_code" = "200" ]; then
    echo "✅ Admin endpoint accessible"
    echo "📋 Response: $admin_body"
else
    echo "❌ Admin endpoint not accessible (HTTP $admin_code)"
    echo "📋 Response: $admin_body"
fi

echo ""
echo "🔍 Phase 2: Check Current Users/Groups"
echo "====================================="

# Try to query current users (if RBAC is enabled)
echo "👥 Checking for existing users..."
users_response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"query": "{ queryUser { name } }"}' \
    "$API_ENDPOINT/admin")

users_code=$(echo "$users_response" | tail -n1)
users_body=$(echo "$users_response" | sed '$d')

echo "📋 Users query response (HTTP $users_code): $users_body"

# Try to query current groups
echo "👥 Checking for existing groups..."
groups_response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"query": "{ queryGroup { name } }"}' \
    "$API_ENDPOINT/admin")

groups_code=$(echo "$groups_response" | tail -n1)
groups_body=$(echo "$groups_response" | sed '$d')

echo "📋 Groups query response (HTTP $groups_code): $groups_body"

echo ""
echo "🔍 Phase 3: Test Schema Permissions"
echo "=================================="

# Test current schema access level
echo "🔐 Testing current API key permissions..."
schema_test=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"query": "schema { }"}' \
    "$API_ENDPOINT/graphql")

schema_code=$(echo "$schema_test" | tail -n1)
schema_body=$(echo "$schema_test" | sed '$d')

echo "📋 Schema access (HTTP $schema_code): $schema_body"

echo ""
echo "🔍 Phase 4: Check Available RBAC Operations"
echo "=========================================="

# Check what RBAC operations are available
echo "🛠️  Testing RBAC mutation availability..."
rbac_test=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"query": "mutation { __schema { mutationType { fields { name } } } }"}' \
    "$API_ENDPOINT/admin")

rbac_code=$(echo "$rbac_test" | tail -n1)
rbac_body=$(echo "$rbac_test" | sed '$d')

echo "📋 Available mutations (HTTP $rbac_code): $rbac_body"

echo ""
echo "📊 RBAC Discovery Summary"
echo "========================"
echo "🔹 Admin endpoint: $([ "$admin_code" = "200" ] && echo "✅ Available" || echo "❌ Not available")"
echo "🔹 User management: $([ "$users_code" = "200" ] && echo "✅ Available" || echo "❌ Not available")"
echo "🔹 Group management: $([ "$groups_code" = "200" ] && echo "✅ Available" || echo "❌ Not available")"
echo "🔹 Schema access: $([ "$schema_code" = "200" ] && echo "✅ Available" || echo "❌ Limited")"
echo ""
echo "💡 Next steps will depend on these discovery results..."
