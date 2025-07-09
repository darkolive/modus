#!/bin/bash

# List All Users - Query all user-related data
echo "👥 Querying All Users and Related Data..."
echo "========================================"

# DQL query to get all User types with all fields
echo "🔍 Fetching all User records..."
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ users(func: type(User)) { uid email emailVerified phone phoneVerified createdAt lastLoginAt status } }"
  }' | jq '.'

echo -e "\n\n🔍 Fetching all UserProfile records..."
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ profiles(func: type(UserProfile)) { uid userId firstName lastName displayName avatar bio timezone language updatedAt } }"
  }' | jq '.'

echo -e "\n\n🔍 Fetching all UserPreferences records..."
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ preferences(func: type(UserPreferences)) { uid userId emailNotifications smsNotifications theme updatedAt } }"
  }' | jq '.'

echo -e "\n\n📊 Summary Query - Users with linked profiles and preferences..."
curl -X POST https://do-study-do-study.hypermode.host/dgraph/query \
  --header "Authorization: Bearer nZgKQjXX2XBRpt" \
  --header "Content-Type: application/json" \
  --data '{
    "query": "{ users(func: type(User)) { uid email emailVerified phone phoneVerified createdAt lastLoginAt status ~userId @filter(type(UserProfile)) { uid firstName lastName displayName avatar bio timezone language updatedAt } ~userId @filter(type(UserPreferences)) { uid emailNotifications smsNotifications theme updatedAt } } }"
  }' | jq '.'

echo -e "\n\n✅ User data query completed!"
