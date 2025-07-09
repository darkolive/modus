#!/bin/bash

# List All Centres - Query all centre-related data
echo "🏢 Querying All Centre Data..."
echo "=============================="

echo "⚠️  Centre schema not yet defined."
echo "   Create schema in db/schema/centres/centres.dql first"
echo ""
echo "Example query structure:"
echo "curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \\"
echo "  --header \"Authorization: Bearer nZgKQjXX2XBRpt\" \\"
echo "  --header \"Content-Type: application/json\" \\"
echo "  --data '{"
echo "    \"query\": \"{ centres(func: type(Centre)) { uid name location contact createdAt } }\""
echo "  }' | jq '.'"

echo -e "\n✅ Centre query template ready!"
