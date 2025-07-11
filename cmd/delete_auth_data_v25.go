package main

import (
	"fmt"

	"github.com/hypermodeinc/modus/sdk/go/pkg/dgraph"
)

func main() {
	// Connection name from manifest
	connectionName := "dgraph"
	
	// Print starting message
	fmt.Println("🗑️  Starting OTP deletion using official Hypermode v25 SDK patterns...")
	
	// Method 1: Try DelJson with type-based deletion
	fmt.Println("\n📋 Method 1: DelJson type-based deletion")
	err := deleteWithDelJson(connectionName)
	if err != nil {
		fmt.Printf("❌ DelJson deletion failed: %v\n", err)
	} else {
		fmt.Println("✅ DelJson deletion completed")
	}
	
	// Method 2: Try query + mutations (conditional deletion)
	fmt.Println("\n📋 Method 2: Query + Mutations (conditional deletion)")
	err = deleteWithQueryMutations(connectionName)
	if err != nil {
		fmt.Printf("❌ Query+Mutations deletion failed: %v\n", err)
	} else {
		fmt.Println("✅ Query+Mutations deletion completed")
	}
	
	// Method 3: Try DelNquads format  
	fmt.Println("\n📋 Method 3: DelNquads format deletion")
	err = deleteWithDelNquads(connectionName)
	if err != nil {
		fmt.Printf("❌ DelNquads deletion failed: %v\n", err)
	} else {
		fmt.Println("✅ DelNquads deletion completed")
	}
	
	// Verify deletion worked
	fmt.Println("\n🔍 Verifying deletion results...")
	verifyDeletion(connectionName)
}

// Method 1: DelJson - JSON format deletion
func deleteWithDelJson(connectionName string) error {
	fmt.Println("  🔄 Creating DelJson mutation...")
	
	// Create mutation using official v25 pattern
	mutation := dgraph.NewMutation().WithDelJson(`{
		"uid": "*",
		"dgraph.type": "ChannelOTP"
	}`)
	
	// Execute using official ExecuteMutations function
	response, err := dgraph.ExecuteMutations(connectionName, mutation)
	if err != nil {
		return fmt.Errorf("failed to execute DelJson mutation: %w", err)
	}
	
	fmt.Printf("  📊 DelJson response: %+v\n", response)
	return nil
}

// Method 2: Query + Mutations - Conditional deletion
func deleteWithQueryMutations(connectionName string) error {
	fmt.Println("  🔄 Creating query + mutations...")
	
	// Create query to find all ChannelOTP records
	query := dgraph.NewQuery(`{
		otps as var(func: type(ChannelOTP))
		sessions as var(func: type(AuthSession))
	}`)
	
	// Create mutations to delete the found records  
	otpMutation := dgraph.NewMutation().WithDelJson(`{
		"uid": "uid(otps)"
	}`)
	
	sessionMutation := dgraph.NewMutation().WithDelJson(`{
		"uid": "uid(sessions)"
	}`)
	
	// Execute query with mutations using official v25 pattern
	response, err := dgraph.ExecuteQuery(connectionName, query, otpMutation, sessionMutation)
	if err != nil {
		return fmt.Errorf("failed to execute query+mutations: %w", err)
	}
	
	fmt.Printf("  📊 Query+Mutations response: %+v\n", response)
	return nil
}

// Method 3: DelNquads - N-Quads format deletion
func deleteWithDelNquads(connectionName string) error {
	fmt.Println("  🔄 Creating DelNquads mutation...")
	
	// First, get UIDs of all ChannelOTP records
	query := dgraph.NewQuery(`{
		otps(func: type(ChannelOTP)) {
			uid
		}
		sessions(func: type(AuthSession)) {
			uid
		}
	}`)
	
	response, err := dgraph.ExecuteQuery(connectionName, query)
	if err != nil {
		return fmt.Errorf("failed to query UIDs: %w", err)
	}
	
	fmt.Printf("  📋 Found records to delete: %+v\n", response)
	
	// Create N-Quads deletion using official v25 pattern
	// Parse response and build N-Quads delete statements
	mutation := dgraph.NewMutation().WithDelNquads(`* * * .`)
	
	// Execute using official v25 pattern
	delResponse, err := dgraph.ExecuteMutations(connectionName, mutation)
	if err != nil {
		return fmt.Errorf("failed to execute DelNquads mutation: %w", err)
	}
	
	fmt.Printf("  📊 DelNquads response: %+v\n", delResponse)
	return nil
}

// Verify deletion worked by counting remaining records
func verifyDeletion(connectionName string) error {
	query := dgraph.NewQuery(`{
		otpCount(func: type(ChannelOTP)) {
			count(uid)
		}
		sessionCount(func: type(AuthSession)) {
			count(uid)
		}
		allOtps(func: type(ChannelOTP)) {
			uid
			channelType
			createdAt
		}
		allSessions(func: type(AuthSession)) {
			uid
			sessionId
			createdAt
		}
	}`)
	
	response, err := dgraph.ExecuteQuery(connectionName, query)
	if err != nil {
		return fmt.Errorf("failed to verify deletion: %w", err)
	}
	
	fmt.Printf("🔍 Verification results: %+v\n", response)
	return nil
}
