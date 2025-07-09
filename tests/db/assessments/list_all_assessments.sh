#!/bin/bash

# List All Assessments - Query all assessment-related data
echo "📝 Querying All Assessment Data..."
echo "=================================="

echo "⚠️  Assessment schema not yet defined."
echo "   Create schema in db/schema/assessments/assessments.dql first"
echo ""
echo "Example query structure:"
echo "curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \\"
echo "  --header \"Authorization: Bearer nZgKQjXX2XBRpt\" \\"
echo "  --header \"Content-Type: application/json\" \\"
echo "  --data '{"
echo "    \"query\": \"{ assessments(func: type(Assessment)) { uid title type questions createdAt } }\""
echo "  }' | jq '.'"

echo -e "\n✅ Assessment query template ready!"
