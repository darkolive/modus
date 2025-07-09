#!/bin/bash

# Master Database Test Script
# Runs all database queries by category

echo "🧪 Running All Database Tests..."
echo "================================"

# Base directory for tests
TEST_DIR="$(dirname "$0")"

# Categories to test
CATEGORIES=("auth" "users" "courses" "centres" "assessments")

# Run tests for each category
for category in "${CATEGORIES[@]}"; do
    echo ""
    echo "📂 Testing $category data..."
    echo "$(printf '=%.0s' {1..40})"
    
    if [ -d "$TEST_DIR/$category" ]; then
        cd "$TEST_DIR/$category"
        
        # Run all .sh files in the category directory
        for script in *.sh; do
            if [ -f "$script" ] && [ "$script" != "run_all_tests.sh" ]; then
                echo ""
                echo "🔧 Running $script..."
                chmod +x "$script"
                ./"$script"
            fi
        done
        
        cd - > /dev/null
    else
        echo "   ❌ Directory $TEST_DIR/$category not found"
    fi
done

echo ""
echo "🎉 All database tests completed!"
echo ""
echo "📋 Test Summary:"
echo "   - Auth: ChannelOTP, AuthSession queries"
echo "   - Users: User, UserProfile, UserPreferences queries"
echo "   - Courses: (schema pending)"
echo "   - Centres: (schema pending)"
echo "   - Assessments: (schema pending)"
