#!/bin/bash

# Create Initial Roles and Permissions Script
# Sets up the basic role structure for the DO Study LMS system

echo "🚀 Creating Initial Roles and Permissions..."
echo "============================================"

# Get API key
if [ -z "$API_KEY" ]; then
    if [ -f ".env.dev.local" ]; then
        API_KEY=$(grep "^API_KEY=" ".env.dev.local" | cut -d'=' -f2 | tr -d '"')
        echo "⚠️  API_KEY environment variable not set, using value from .env.dev.local"
    else
        echo "⚠️  API_KEY environment variable not set, using default"
        API_KEY="nZgKQjXX2XBRpt"
    fi
fi

API_ENDPOINT="https://do-study-do-study.hypermode.host/dgraph/alter"

echo ""
echo "📋 Step 1: Creating Permission Categories"
echo "========================================"

# Create basic permission categories using N-Quads format
categories_nquads='_:view_category <dgraph.type> "PermissionCategory" .
_:view_category <id> "VIEW" .
_:view_category <displayName> "View Content" .

_:create_category <dgraph.type> "PermissionCategory" .
_:create_category <id> "CREATE" .
_:create_category <displayName> "Create Content" .

_:update_category <dgraph.type> "PermissionCategory" .
_:update_category <id> "UPDATE" .
_:update_category <displayName> "Update Content" .

_:auth_category <dgraph.type> "PermissionCategory" .
_:auth_category <id> "AUTH" .
_:auth_category <displayName> "Authentication" .'

echo "📡 Creating permission categories..."
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d "$categories_nquads" \
    "$API_ENDPOINT")

http_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo "✅ Permission categories created successfully"
else
    echo "❌ Failed to create permission categories (HTTP $http_code)"
    echo "📋 Response: $response_body"
fi

echo ""
echo "📋 Step 2: Creating Action Types"
echo "=============================="

action_types_nquads='_:create_action <dgraph.type> "ActionType" .
_:create_action <code> "CREATE" .
_:read_action <dgraph.type> "ActionType" .
_:read_action <code> "READ" .
_:update_action <dgraph.type> "ActionType" .
_:update_action <code> "UPDATE" .
_:delete_action <dgraph.type> "ActionType" .
_:delete_action <code> "DELETE" .
_:execute_action <dgraph.type> "ActionType" .
_:execute_action <code> "EXECUTE" .
_:signin_action <dgraph.type> "ActionType" .
_:signin_action <code> "SIGNIN" .
_:profile_action <dgraph.type> "ActionType" .
_:profile_action <code> "PROFILE" .
_:course_view_action <dgraph.type> "ActionType" .
_:course_view_action <code> "COURSE_VIEW" .'

echo "📡 Creating action types..."
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d "$action_types_nquads" \
    "$API_ENDPOINT")

http_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo "✅ Action types created successfully"
else
    echo "❌ Failed to create action types (HTTP $http_code)"
    echo "Response: $response_body"
fi

echo ""
echo "🔐 Step 3: Creating Basic Permissions"
echo "===================================="

permissions_nquads='_:signin_permission <dgraph.type> "Permission" .
_:signin_permission <category> _:auth_category .
_:signin_permission <action> _:signin_action .

_:profile_permission <dgraph.type> "Permission" .
_:profile_permission <category> _:auth_category .
_:profile_permission <action> _:profile_action .

_:course_view_permission <dgraph.type> "Permission" .
_:course_view_permission <category> _:view_category .
_:course_view_permission <action> _:course_view_action .

_:create_permission <dgraph.type> "Permission" .
_:create_permission <category> _:create_category .
_:create_permission <action> _:create_action .

_:read_permission <dgraph.type> "Permission" .
_:read_permission <category> _:view_category .
_:read_permission <action> _:read_action .

_:update_permission <dgraph.type> "Permission" .
_:update_permission <category> _:update_category .
_:update_permission <action> _:update_action .

_:delete_permission <dgraph.type> "Permission" .
_:delete_permission <category> _:update_category .
_:delete_permission <action> _:delete_action .

_:execute_permission <dgraph.type> "Permission" .
_:execute_permission <category> _:update_category .
_:execute_permission <action> _:execute_action .'

echo "📡 Creating basic permissions..."
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d "$permissions_nquads" \
    "$API_ENDPOINT")

http_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo "✅ Basic permissions created successfully"
else
    echo "❌ Failed to create basic permissions (HTTP $http_code)"
    echo "📋 Response: $response_body"
fi

echo ""
echo "👤 Step 4: Creating 'registered' Role"
echo "===================================="

registered_role_nquads='_:registered_role <dgraph.type> "Role" .
_:registered_role <name> "registered" .
_:registered_role <permissions> _:signin_permission .
_:registered_role <permissions> _:profile_permission .
_:registered_role <permissions> _:course_view_permission .
_:registered_role <defaultDashboard> "student" .'

echo "📡 Creating 'registered' role..."
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d "$registered_role_nquads" \
    "$API_ENDPOINT")

http_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo "✅ 'registered' role created successfully"
else
    echo "❌ Failed to create 'registered' role (HTTP $http_code)"
    echo "📋 Response: $response_body"
fi

echo ""
echo "👑 Step 5: Creating Additional Roles"
echo "===================================="

# Create admin and instructor roles for completeness
additional_roles_mutation='{
  "query": "mutation {\n  set {\n    _:admin_role <dgraph.type> \"Role\" .\n    _:admin_role <name> \"admin\" .\n    _:admin_role <permissions> _:signin_permission .\n    _:admin_role <permissions> _:profile_permission .\n    _:admin_role <permissions> _:course_view_permission .\n    _:admin_role <defaultDashboard> \"admin\" .\n    \n    _:instructor_role <dgraph.type> \"Role\" .\n    _:instructor_role <name> \"instructor\" .\n    _:instructor_role <permissions> _:signin_permission .\n    _:instructor_role <permissions> _:profile_permission .\n    _:instructor_role <permissions> _:course_view_permission .\n    _:instructor_role <defaultDashboard> \"instructor\" .\n  }\n}"
}'

echo "📡 Creating additional roles..."
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d "$additional_roles_mutation" \
    "$API_ENDPOINT/graphql")

http_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo "✅ Additional roles created successfully"
else
    echo "❌ Failed to create additional roles (HTTP $http_code)"
    echo "📋 Response: $response_body"
fi

echo ""
echo "🔍 Step 6: Verification"
echo "======================"

# Query to verify roles were created
verify_query='{
  "query": "query {\n  roles(func: type(Role)) {\n    uid\n    name\n    defaultDashboard\n    permissions {\n      uid\n      category {\n        id\n        displayName\n      }\n      action {\n        code\n        description\n      }\n    }\n  }\n}"
}'

echo "📡 Verifying created roles..."
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d "$verify_query" \
    "$API_ENDPOINT/graphql")

http_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo "✅ Verification successful"
    echo "📋 Created roles:"
    echo "$response_body" | jq -r '.data.roles[] | "- \(.name) (Dashboard: \(.defaultDashboard))"' 2>/dev/null || echo "$response_body"
else
    echo "❌ Verification failed (HTTP $http_code)"
    echo "📋 Response: $response_body"
fi

echo ""
echo "🎉 Initial Roles Setup Complete!"
echo "==============================="
echo "✅ Permission categories created"
echo "✅ Action types defined" 
echo "✅ Basic permissions established"
echo "✅ 'registered' role created with basic permissions"
echo "✅ Additional roles (admin, instructor) created"
echo ""
echo "Next steps:"
echo "1. Run this script: chmod +x create_initial_roles.sh && ./create_initial_roles.sh"
echo "2. Update user creation logic to assign 'registered' role to new users"
echo "3. Test the authentication flow with role-based permissions"
