package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Test types matching main.go GraphQL types
type SessionRequest struct {
	UserID     string `json:"userId"`
	ChannelDID string `json:"channelDID"`
	Action     string `json:"action"`
}

type SessionResponse struct {
	Success     bool   `json:"success"`
	SessionID   string `json:"sessionId"`
	AccessToken string `json:"accessToken"`
	ExpiresAt   int64  `json:"expiresAt"`
	Message     string `json:"message"`
	UserID      string `json:"userId"`
}

type ValidationRequest struct {
	Token string `json:"token"`
}

type ValidationResponse struct {
	Valid     bool   `json:"valid"`
	UserID    string `json:"userId,omitempty"`
	ExpiresAt int64  `json:"expiresAt,omitempty"`
	Message   string `json:"message,omitempty"`
}

type RefreshRequest struct {
	Token string `json:"token"`
}

type RefreshResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
	Message   string `json:"message,omitempty"`
}

type RevocationRequest struct {
	Token  string `json:"token"`
	Reason string `json:"reason,omitempty"`
}

type RevocationResponse struct {
	Revoked   bool   `json:"revoked"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// GraphQL request/response structures
type GraphQLRequest struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

const baseURL = "http://localhost:8080/graphql"

func main() {
	fmt.Println("🚀 Starting ChronosSession Integration Tests")
	fmt.Println(strings.Repeat("=", 50))

	// Test data
	testUserID := "test-user-123"
	testChannelDID := "test-channel-did-456"
	testAction := "signin"

	var sessionToken string
	var refreshedToken string

	// Test 1: Session Creation (Issue)
	fmt.Println("\n📝 Test 1: Session Creation (Issue)")
	sessionResp, err := testCreateSession(testUserID, testChannelDID, testAction)
	if err != nil {
		fmt.Printf("❌ Session creation failed: %v\n", err)
		return
	}
	
	if sessionResp.Success {
		sessionToken = sessionResp.AccessToken
		fmt.Printf("✅ Session created successfully\n")
		fmt.Printf("   Token: %s...\n", sessionToken[:20])
		fmt.Printf("   UserID: %s\n", sessionResp.UserID)
		fmt.Printf("   ExpiresAt: %s\n", time.Unix(sessionResp.ExpiresAt, 0).Format(time.RFC3339))
	} else {
		fmt.Printf("❌ Session creation failed: %s\n", sessionResp.Message)
		return
	}

	// Test 2: Session Validation
	fmt.Println("\n🔍 Test 2: Session Validation")
	validationResp, err := testValidateSession(sessionToken)
	if err != nil {
		fmt.Printf("❌ Session validation failed: %v\n", err)
		return
	}
	
	if validationResp.Valid {
		fmt.Printf("✅ Session validation successful\n")
		fmt.Printf("   Valid: %t\n", validationResp.Valid)
		fmt.Printf("   UserID: %s\n", validationResp.UserID)
		fmt.Printf("   ExpiresAt: %s\n", time.Unix(validationResp.ExpiresAt, 0).Format(time.RFC3339))
	} else {
		fmt.Printf("❌ Session validation failed: %s\n", validationResp.Message)
	}

	// Test 3: Session Refresh
	fmt.Println("\n🔄 Test 3: Session Refresh")
	refreshResp, err := testRefreshSession(sessionToken)
	if err != nil {
		fmt.Printf("❌ Session refresh failed: %v\n", err)
	} else {
		refreshedToken = refreshResp.Token
		fmt.Printf("✅ Session refresh successful\n")
		fmt.Printf("   New Token: %s...\n", refreshedToken[:20])
		fmt.Printf("   ExpiresAt: %s\n", time.Unix(refreshResp.ExpiresAt, 0).Format(time.RFC3339))
		fmt.Printf("   Message: %s\n", refreshResp.Message)
	}

	// Test 4: Validate Refreshed Session
	if refreshedToken != "" {
		fmt.Println("\n🔍 Test 4: Validate Refreshed Session")
		validationResp, err := testValidateSession(refreshedToken)
		if err != nil {
			fmt.Printf("❌ Refreshed session validation failed: %v\n", err)
		} else if validationResp.Valid {
			fmt.Printf("✅ Refreshed session validation successful\n")
			fmt.Printf("   Valid: %t\n", validationResp.Valid)
			fmt.Printf("   UserID: %s\n", validationResp.UserID)
		} else {
			fmt.Printf("❌ Refreshed session validation failed: %s\n", validationResp.Message)
		}
	}

	// Test 5: Session Revocation
	fmt.Println("\n🚫 Test 5: Session Revocation")
	tokenToRevoke := refreshedToken
	if tokenToRevoke == "" {
		tokenToRevoke = sessionToken
	}
	
	revocationResp, err := testRevokeSession(tokenToRevoke, "Integration test cleanup")
	if err != nil {
		fmt.Printf("❌ Session revocation failed: %v\n", err)
	} else if revocationResp.Revoked {
		fmt.Printf("✅ Session revocation successful\n")
		fmt.Printf("   Revoked: %t\n", revocationResp.Revoked)
		fmt.Printf("   Message: %s\n", revocationResp.Message)
		fmt.Printf("   Timestamp: %s\n", revocationResp.Timestamp)
	} else {
		fmt.Printf("❌ Session revocation failed: %s\n", revocationResp.Message)
	}

	// Test 6: Validate Revoked Session
	fmt.Println("\n🔍 Test 6: Validate Revoked Session")
	validationResp, err = testValidateSession(tokenToRevoke)
	if err != nil {
		fmt.Printf("❌ Revoked session validation test failed: %v\n", err)
	} else if !validationResp.Valid {
		fmt.Printf("✅ Revoked session correctly invalid\n")
		fmt.Printf("   Valid: %t\n", validationResp.Valid)
		fmt.Printf("   Message: %s\n", validationResp.Message)
	} else {
		fmt.Printf("❌ Revoked session still shows as valid!\n")
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🎯 ChronosSession Integration Tests Complete")
}

func testCreateSession(userID, channelDID, action string) (*SessionResponse, error) {
	query := `
		mutation CreateSession($req: SessionRequest!) {
			createSession(req: $req) {
				success
				sessionId
				accessToken
				expiresAt
				message
				userId
			}
		}
	`
	
	variables := map[string]interface{}{
		"req": SessionRequest{
			UserID:     userID,
			ChannelDID: channelDID,
			Action:     action,
		},
	}

	var response struct {
		CreateSession SessionResponse `json:"createSession"`
	}

	err := makeGraphQLRequest(query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.CreateSession, nil
}

func testValidateSession(token string) (*ValidationResponse, error) {
	query := `
		query ValidateSession($req: ValidationRequest!) {
			validateSession(req: $req) {
				valid
				userId
				expiresAt
				message
			}
		}
	`
	
	variables := map[string]interface{}{
		"req": ValidationRequest{
			Token: token,
		},
	}

	var response struct {
		ValidateSession ValidationResponse `json:"validateSession"`
	}

	err := makeGraphQLRequest(query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.ValidateSession, nil
}

func testRefreshSession(token string) (*RefreshResponse, error) {
	query := `
		mutation RefreshSession($req: RefreshRequest!) {
			refreshSession(req: $req) {
				token
				expiresAt
				message
			}
		}
	`
	
	variables := map[string]interface{}{
		"req": RefreshRequest{
			Token: token,
		},
	}

	var response struct {
		RefreshSession RefreshResponse `json:"refreshSession"`
	}

	err := makeGraphQLRequest(query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.RefreshSession, nil
}

func testRevokeSession(token, reason string) (*RevocationResponse, error) {
	query := `
		mutation RevokeSession($req: RevocationRequest!) {
			revokeSession(req: $req) {
				revoked
				message
				timestamp
			}
		}
	`
	
	variables := map[string]interface{}{
		"req": RevocationRequest{
			Token:  token,
			Reason: reason,
		},
	}

	var response struct {
		RevokeSession RevocationResponse `json:"revokeSession"`
	}

	err := makeGraphQLRequest(query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.RevokeSession, nil
}

func makeGraphQLRequest(query string, variables interface{}, response interface{}) error {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var gqlResp GraphQLResponse
	err = json.Unmarshal(body, &gqlResp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %v", gqlResp.Errors)
	}

	// Marshal data back to JSON and unmarshal into the expected response type
	dataBytes, err := json.Marshal(gqlResp.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	err = json.Unmarshal(dataBytes, response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data into response: %v", err)
	}

	return nil
}
