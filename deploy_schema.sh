#!/bin/bash

# Deploy DQL schema to Hypermode Dgraph
echo "Deploying DQL schema to Dgraph..."

curl -X POST https://do-study-do-study.hypermode.host/dgraph/alter \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data-binary "@./db/schema.dql"

echo -e "\n\nVerifying schema deployment..."

# To verify the schema, we can query the schema using DQL
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{"query": "schema {}"}'
