#!/bin/bash

# List All Courses - Query all course-related data
echo "📚 Querying All Course Data..."
echo "=============================="

echo "⚠️  Course schema not yet defined."
echo "   Create schema in db/schema/courses/courses.dql first"
echo ""
echo "Example query structure:"
echo "curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \\"
echo "  --header \"Authorization: Bearer nZgKQjXX2XBRpt\" \\"
echo "  --header \"Content-Type: application/json\" \\"
echo "  --data '{"
echo "    \"query\": \"{ courses(func: type(Course)) { uid title description createdAt } }\""
echo "  }' | jq '.'"

echo -e "\n✅ Course query template ready!"
