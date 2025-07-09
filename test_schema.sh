#!/bin/bash

# Test if the EmailOtp type is available in the schema
echo "Testing if EmailOtp type is available in the schema..."
echo ""

response=$(curl -s -X POST https://do-study-do-study.hypermode.host:443/graphql \
  -H "Authorization: Bearer nZgKQjXX2XBRpt" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { __type(name: \"EmailOtp\") { name fields { name type { name kind } } } }"}')

echo "Response:"
echo "$response" | jq . 2>/dev/null || echo "$response"
echo ""

# Also test the schema endpoint
echo "Testing schema endpoint..."
schema_response=$(curl -s -X POST https://do-study-do-study.hypermode.host:443/graphql \
  -H "Authorization: Bearer nZgKQjXX2XBRpt" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { __schema { types { name } } }"}')

echo "Available types:"
echo "$schema_response" | jq '.data.__schema.types[].name' 2>/dev/null || echo "$schema_response"
