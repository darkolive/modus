#!/bin/bash

# Deploy GraphQL schema to Hypermode Dgraph
echo "Deploying GraphQL schema to Dgraph..."

# Read the GraphQL schema file
schema_content=$(cat ./db/schema.graphql)

# Deploy the GraphQL schema using the admin endpoint
curl -X POST https://do-study-do-study.hypermode.host/admin/schema \
  -H "Authorization: Bearer nZgKQjXX2XBRpt" \
  -H "Content-Type: application/json" \
  -d "{\"schema\": $(echo "$schema_content" | jq -Rs .)}"

echo -e "\n\nVerifying GraphQL schema deployment..."

# Test if EmailOtp type is now available
curl -s -X POST https://do-study-do-study.hypermode.host:443/graphql \
  -H "Authorization: Bearer nZgKQjXX2XBRpt" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { __type(name: \"EmailOtp\") { name fields { name type { name kind } } } }"}' | jq .

echo -e "\n\nTesting addEmailOtp mutation availability..."

# Test if addEmailOtp mutation is available
curl -s -X POST https://do-study-do-study.hypermode.host:443/graphql \
  -H "Authorization: Bearer nZgKQjXX2XBRpt" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { __schema { mutationType { fields { name } } } }"}' | jq '.data.__schema.mutationType.fields[] | select(.name | contains("EmailOtp"))'
