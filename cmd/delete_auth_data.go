package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/hypermodeinc/modus/sdk/go/pkg/dgraph"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run delete_auth_data.go <connection_name>")
		fmt.Println("Example: go run delete_auth_data.go dgraph")
		os.Exit(1)
	}

	connectionName := os.Args[1]
	fmt.Println("🗑️  Deleting All Authentication Data (GraphQL Approach)...")
	fmt.Println("=========================================================")

	// Delete all ChannelOTP records using GraphQL delete mutation
	fmt.Println("🔍 Deleting all ChannelOTP records...")
	otpDeleteMutation := `
		mutation {
			deleteChannelOTP(filter: {}) {
				msg
				numUids
				channelOTP {
					uid
					channelType
					createdAt
				}
			}
		}
	`

	// Use the latest v25-compatible ExecuteQuery method
	otpDelResp, err := dgraph.ExecuteQuery(connectionName, dgraph.NewQuery(otpDeleteMutation))
	if err != nil {
		log.Fatalf("Failed to delete ChannelOTP records: %v", err)
	}

	// Parse OTP deletion response
	var otpDelResult struct {
		DeleteChannelOTP struct {
			Msg       string `json:"msg"`
			NumUids   int    `json:"numUids"`
			ChannelOTP []struct {
				UID         string `json:"uid"`
				ChannelType string `json:"channelType"`
				CreatedAt   string `json:"createdAt"`
			} `json:"channelOTP"`
		} `json:"deleteChannelOTP"`
	}

	if err := json.Unmarshal([]byte(otpDelResp.Json), &otpDelResult); err != nil {
		log.Fatalf("Failed to parse ChannelOTP deletion response: %v", err)
	}

	fmt.Printf("✅ ChannelOTP Deletion: %s\n", otpDelResult.DeleteChannelOTP.Msg)
	fmt.Printf("📊 Deleted %d ChannelOTP records\n", otpDelResult.DeleteChannelOTP.NumUids)
	if len(otpDelResult.DeleteChannelOTP.ChannelOTP) > 0 {
		fmt.Println("🗑️  Deleted records:")
		for _, otp := range otpDelResult.DeleteChannelOTP.ChannelOTP {
			fmt.Printf("   - UID: %s, Type: %s, Created: %s\n", otp.UID, otp.ChannelType, otp.CreatedAt)
		}
	}

	// Delete all AuthSession records using GraphQL delete mutation
	fmt.Println("\n🔍 Deleting all AuthSession records...")
	authDeleteMutation := `
		mutation {
			deleteAuthSession(filter: {}) {
				msg
				numUids
				authSession {
					uid
					userId
					createdAt
				}
			}
		}
	`

	// Use the latest v25-compatible ExecuteQuery method
	authDelResp, err := dgraph.ExecuteQuery(connectionName, dgraph.NewQuery(authDeleteMutation))
	if err != nil {
		log.Fatalf("Failed to delete AuthSession records: %v", err)
	}

	// Parse AuthSession deletion response
	var authDelResult struct {
		DeleteAuthSession struct {
			Msg         string `json:"msg"`
			NumUids     int    `json:"numUids"`
			AuthSession []struct {
				UID       string `json:"uid"`
				UserID    string `json:"userId"`
				CreatedAt string `json:"createdAt"`
			} `json:"authSession"`
		} `json:"deleteAuthSession"`
	}

	if err := json.Unmarshal([]byte(authDelResp.Json), &authDelResult); err != nil {
		log.Fatalf("Failed to parse AuthSession deletion response: %v", err)
	}

	fmt.Printf("✅ AuthSession Deletion: %s\n", authDelResult.DeleteAuthSession.Msg)
	fmt.Printf("📊 Deleted %d AuthSession records\n", authDelResult.DeleteAuthSession.NumUids)
	if len(authDelResult.DeleteAuthSession.AuthSession) > 0 {
		fmt.Println("🗑️  Deleted records:")
		for _, session := range authDelResult.DeleteAuthSession.AuthSession {
			fmt.Printf("   - UID: %s, UserID: %s, Created: %s\n", session.UID, session.UserID, session.CreatedAt)
		}
	}

	// Verification queries
	fmt.Println("\n🔍 Verifying deletion...")

	// Verify deletion with queries using v25-compatible methods
	otpVerifyQuery := `
		query {
			queryChannelOTP {
				uid
				channelType
				createdAt
			}
		}
	`

	// Use the latest v25-compatible ExecuteQuery method
	otpVerifyResp, err := dgraph.ExecuteQuery(connectionName, dgraph.NewQuery(otpVerifyQuery))
	if err != nil {
		log.Fatalf("Failed to verify ChannelOTP deletion: %v", err)
	}

	var otpVerifyResult struct {
		QueryChannelOTP []struct {
			UID         string `json:"uid"`
			ChannelType string `json:"channelType"`
			CreatedAt   string `json:"createdAt"`
		} `json:"queryChannelOTP"`
	}

	if err := json.Unmarshal([]byte(otpVerifyResp.Json), &otpVerifyResult); err != nil {
		log.Fatalf("Failed to parse ChannelOTP verification response: %v", err)
	}

	fmt.Printf("📊 Remaining ChannelOTP records: %d\n", len(otpVerifyResult.QueryChannelOTP))

	// Verify AuthSession deletion
	authVerifyQuery := `
		query {
			queryAuthSession {
				uid
				userId
				createdAt
			}
		}
	`

	// Use the latest v25-compatible ExecuteQuery method
	authVerifyResp, err := dgraph.ExecuteQuery(connectionName, dgraph.NewQuery(authVerifyQuery))
	if err != nil {
		log.Fatalf("Failed to verify AuthSession deletion: %v", err)
	}

	var authVerifyResult struct {
		QueryAuthSession []struct {
			UID       string `json:"uid"`
			UserID    string `json:"userId"`
			CreatedAt string `json:"createdAt"`
		} `json:"queryAuthSession"`
	}

	if err := json.Unmarshal([]byte(authVerifyResp.Json), &authVerifyResult); err != nil {
		log.Fatalf("Failed to parse AuthSession verification response: %v", err)
	}

	fmt.Printf("📊 Remaining AuthSession records: %d\n", len(authVerifyResult.QueryAuthSession))

	// Summary
	totalDeleted := otpDelResult.DeleteChannelOTP.NumUids + authDelResult.DeleteAuthSession.NumUids
	totalRemaining := len(otpVerifyResult.QueryChannelOTP) + len(authVerifyResult.QueryAuthSession)

	fmt.Printf("\n📈 Summary:\n")
	fmt.Printf("   🗑️  Total records deleted: %d\n", totalDeleted)
	fmt.Printf("   📊 Total records remaining: %d\n", totalRemaining)

	if totalRemaining == 0 {
		fmt.Println("\n✅ All authentication data successfully deleted!")
	} else {
		fmt.Println("\n⚠️  Some records may still remain - check verification output above")
	}
}
