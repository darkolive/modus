#!/bin/bash

# Deploy GraphQL Schema with @auth directives
# This deploys to the /admin/schema endpoint for GraphQL access control

echo "🚀 Deploying GraphQL Schema with @auth directives..."
echo "=================================================="

# Load environment variables
if [ -f ".env.dev.local" ]; then
    export $(grep -v '^#' .env.dev.local | xargs)
    echo "📋 Loaded environment from .env.dev.local"
elif [ -z "$API_KEY" ]; then
    echo "❌ Error: API_KEY not found in environment or .env.dev.local"
    echo "💡 Usage: API_KEY=your_key $0"
    exit 1
fi

# Configuration
GRAPHQL_SCHEMA_FILE="./schema/graphql_auth_schema.graphql"
DGRAPH_ENDPOINT="https://do-study-do-study.hypermode.host"
ADMIN_ENDPOINT="$DGRAPH_ENDPOINT/admin/schema"

# Check if GraphQL schema file exists
if [ ! -f "$GRAPHQL_SCHEMA_FILE" ]; then
    echo "❌ Error: GraphQL schema file not found: $GRAPHQL_SCHEMA_FILE"
    exit 1
fi

echo "📄 Deploying GraphQL schema with @auth directives..."
echo "   📁 File: $GRAPHQL_SCHEMA_FILE"
echo "   🌐 Endpoint: $ADMIN_ENDPOINT"
echo "   📊 Size: $(wc -l < "$GRAPHQL_SCHEMA_FILE") lines"
echo ""

# Deploy GraphQL schema to /admin/schema endpoint
response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
    -X POST \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/graphql" \
    --data-binary "@$GRAPHQL_SCHEMA_FILE" \
    "$ADMIN_ENDPOINT")

# Parse response
http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
response_body=$(echo "$response" | sed '/HTTP_STATUS:/d')

echo "📋 Response Status: $http_status"
echo "📋 Response Body: $response_body"
echo ""

if [ "$http_status" = "200" ]; then
    echo "✅ GraphQL schema with @auth directives deployed successfully!"
    echo ""
    echo "🎉 Access Control Features Enabled:"
    echo "   🔐 JWT-based authentication required"
    echo "   👥 Role-based access control (RBAC)"
    echo "   🛡️  User-level data isolation"
    echo "   📊 Granular CRUD permissions per type"
    echo ""
    echo "💡 Next steps:"
    echo "   1. Configure JWT tokens in your application"
    echo "   2. Test GraphQL queries with proper Authorization headers"
    echo "   3. Verify access control rules are working as expected"
else
    echo "❌ GraphQL schema deployment failed!"
    echo "📋 HTTP Status: $http_status"
    echo "📋 Response: $response_body"
    exit 1
fi

echo ""
echo "🔧 Technical Notes:"
echo "   • GraphQL schema deployed to /admin/schema (not /dgraph/alter)"
echo "   • @auth directives provide query/mutation-level access control"
echo "   • JWT tokens must include \$USER and \$ROLE claims"
echo "   • DQL schema remains unchanged and provides database structure"
echo "   • Both schemas work together: DQL for structure, GraphQL for access control"
