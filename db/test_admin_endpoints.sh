#!/bin/bash

# Test Admin Endpoints for ACL/RBAC Capabilities
# Based on Hypermode documentation: https://docs.hypermode.com/dgraph/enterprise/access-control-lists

echo "🔍 Testing Hypermode Dgraph Admin Endpoints for ACL/RBAC..."
echo "============================================================"

# Load environment variables
if [ -f ".env.dev.local" ]; then
    export $(grep -v '^#' .env.dev.local | xargs)
    echo "📋 Loaded API_KEY from .env.dev.local"
elif [ -z "$API_KEY" ]; then
    echo "❌ Error: API_KEY not found"
    exit 1
fi

BASE_URL="https://do-study-do-study.hypermode.host"

echo "🌐 Base URL: $BASE_URL"
echo "🔑 API Key: ${API_KEY:0:8}..."
echo ""

# Test endpoints based on Hypermode documentation
endpoints=(
    "/admin"
    "/admin/graphql" 
    "/admin/schema"
    "/graphql/admin"
    "/dgraph/admin"
)

# Test GraphQL introspection queries
queries=(
    '{"query":"query { __schema { queryType { name } } }"}'
    '{"query":"query { getGQLSchema { schema } }"}'
    '{"query":"query { __type(name: \"Query\") { fields { name } } }"}'
    '{"query":"query { queryUser { name } }"}'
    '{"query":"query { queryGroup { name } }"}'
)

echo "📋 Testing Admin Endpoints..."
echo "=============================="

for endpoint in "${endpoints[@]}"; do
    echo "🔗 Testing: $BASE_URL$endpoint"
    
    response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $API_KEY" \
        -H "Content-Type: application/json" \
        -d '{"query":"query { __schema { queryType { name } } }"}' \
        "$BASE_URL$endpoint")
    
    http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
    response_body=$(echo "$response" | sed '/HTTP_STATUS:/d')
    
    if [ "$http_status" = "200" ]; then
        echo "   ✅ Status: $http_status - AVAILABLE"
        if [ -n "$response_body" ]; then
            echo "   📋 Response: $response_body"
        fi
        echo ""
        
        # If endpoint is available, test ACL-specific queries
        echo "   🔍 Testing ACL queries on working endpoint..."
        
        acl_queries=(
            '{"query":"query { queryUser { name } }"}'
            '{"query":"query { queryGroup { name } }"}'
            '{"query":"mutation { addUser(input: {name: \"test\", password: \"test\"}) { user { name } } }"}'
        )
        
        for query in "${acl_queries[@]}"; do
            echo "      🧪 Query: $(echo $query | jq -r '.query' | head -c 50)..."
            
            acl_response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
                -X POST \
                -H "Authorization: Bearer $API_KEY" \
                -H "Content-Type: application/json" \
                -d "$query" \
                "$BASE_URL$endpoint")
            
            acl_status=$(echo "$acl_response" | grep "HTTP_STATUS:" | cut -d: -f2)
            acl_body=$(echo "$acl_response" | sed '/HTTP_STATUS:/d')
            
            echo "         📊 Status: $acl_status"
            if [ -n "$acl_body" ]; then
                echo "         📋 Response: $(echo $acl_body | head -c 100)..."
            fi
        done
        echo ""
        
    elif [ "$http_status" = "404" ]; then
        echo "   ❌ Status: $http_status - NOT FOUND"
    elif [ "$http_status" = "403" ]; then
        echo "   🔒 Status: $http_status - FORBIDDEN (endpoint exists but access denied)"
    else
        echo "   ⚠️  Status: $http_status - OTHER"
        if [ -n "$response_body" ]; then
            echo "   📋 Response: $response_body"
        fi
    fi
    echo ""
done

echo "📋 Testing GraphQL Schema Deployment..."
echo "======================================"

# Test schema deployment endpoints
schema_endpoints=(
    "/admin/schema"
    "/admin/graphql"
    "/graphql/admin"
)

for endpoint in "${schema_endpoints[@]}"; do
    echo "🔗 Testing schema deployment: $BASE_URL$endpoint"
    
    # Test with a simple GraphQL schema
    test_schema='type TestType { id: ID! name: String }'
    
    response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $API_KEY" \
        -H "Content-Type: application/graphql" \
        --data "$test_schema" \
        "$BASE_URL$endpoint")
    
    http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
    response_body=$(echo "$response" | sed '/HTTP_STATUS:/d')
    
    echo "   📊 Status: $http_status"
    if [ -n "$response_body" ]; then
        echo "   📋 Response: $(echo $response_body | head -c 200)..."
    fi
    echo ""
done

echo "📋 Summary & Recommendations"
echo "============================"
echo "Based on the test results above:"
echo ""
echo "✅ If any endpoint returned 200: ACL/RBAC may be available"
echo "🔒 If endpoints returned 403: ACL exists but requires different authentication"
echo "❌ If all endpoints returned 404: ACL/RBAC not exposed by Hypermode"
echo ""
echo "📚 Documentation Reference:"
echo "   https://docs.hypermode.com/dgraph/enterprise/access-control-lists"
echo ""
echo "💡 Next Steps:"
echo "   1. Review test results above"
echo "   2. If ACL is available, implement user/group management"
echo "   3. If not available, continue with application-level access control"
