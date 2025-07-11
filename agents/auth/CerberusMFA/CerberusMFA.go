package cerberusmfa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hypermodeinc/modus/sdk/go/pkg/dgraph"
	"modus/services/webauthn"
)

// CerberusMFARequest represents the input for MFA flow decision
type CerberusMFARequest struct {
	ChannelDID  string `json:"channelDID"`  // From CharonOTP verification
	ChannelType string `json:"channelType"` // email, phone, etc.
}

// CerberusMFAResponse represents the MFA flow decision response
type CerberusMFAResponse struct {
	UserExists       bool     `json:"userExists"`
	Action          string   `json:"action"`          // "signin" or "register"
	UserID          string   `json:"userId,omitempty"`
	AvailableMethods []string `json:"availableMethods"`
	NextStep        string   `json:"nextStep"`
	Message         string   `json:"message"`
}

// UserChannelsResult represents the database query result for user channels
type UserChannelsResult struct {
	UserChannels []struct {
		UID         string    `json:"uid"`
		UserID      string    `json:"userId"`
		ChannelType string    `json:"channelType"`
		ChannelHash string    `json:"channelHash"`
		Verified    bool      `json:"verified"`
		Primary     bool      `json:"primary"`
		CreatedAt   time.Time `json:"createdAt"`
		LastUsedAt  time.Time `json:"lastUsedAt"`
	} `json:"userChannels"`
}

// CerberusMFA is the main function that determines authentication flow
func CerberusMFA(req CerberusMFARequest) (*CerberusMFAResponse, error) {
	log.Printf("🐕 CerberusMFA: Checking user existence for channel %s (%s)", req.ChannelDID, req.ChannelType)

	// Check if user exists by channel hash
	userExists, userID, err := checkUserByChannel(req.ChannelDID, req.ChannelType)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %v", err)
	}

	if userExists {
		// Existing user - proceed to sign-in flow
		log.Printf("✅ Existing user found: %s", userID)
		
		// Update last used timestamp for the channel
		if err := updateChannelLastUsed(req.ChannelDID, req.ChannelType); err != nil {
			log.Printf("⚠️ Failed to update channel last used: %v", err)
		}

		return &CerberusMFAResponse{
			UserExists:       true,
			Action:          "signin",
			UserID:          userID,
			AvailableMethods: []string{"webauthn", "passwordless"},
			NextStep:        "Choose authentication method: WebAuthn (biometric/hardware) or Passwordless DID",
			Message:         "Welcome back! Please complete authentication.",
		}, nil
	} else {
		// New user - create user account first
		log.Printf("🆕 New user detected for channel %s", req.ChannelDID)
		
		// Create the new user with PENDING status and 'registered' role
		newUserID, err := CreateNewUser(req.ChannelDID, req.ChannelType)
		if err != nil {
			log.Printf("❌ Failed to create new user: %v", err)
			return &CerberusMFAResponse{
				UserExists:       false,
				Action:          "error",
				UserID:          "",
				AvailableMethods: []string{},
				NextStep:        "Registration failed",
				Message:         "Failed to create user account. Please try again.",
			}, nil
		}
		
		log.Printf("✅ Created new user: %s", newUserID)

		return &CerberusMFAResponse{
			UserExists:       true, // Now the user exists after creation
			Action:          "register",
			UserID:          newUserID,
			AvailableMethods: []string{"webauthn", "passwordless"},
			NextStep:        "Complete authentication setup: Choose WebAuthn (biometric/hardware) or Passwordless",
			Message:         "Welcome! Your account has been created. Please set up secure authentication.",
		}, nil
	}
}

// checkUserByChannel checks if a user exists by channel hash
func checkUserByChannel(channelDID, channelType string) (bool, string, error) {
	// Create DQL query to check if user exists by channel hash
	query := fmt.Sprintf(`{
		userChannels(func: eq(channelHash, "%s")) @filter(eq(channelType, "%s")) {
			uid
			userId
			channelType
			channelHash
			verified
			primary
			createdAt
			lastUsedAt
		}
	}`, channelDID, channelType)

	// Execute the query
	resp, err := dgraph.ExecuteQuery("dgraph", dgraph.NewQuery(query))
	if err != nil {
		return false, "", fmt.Errorf("failed to query user channels: %v", err)
	}

	// Parse the response
	var result UserChannelsResult
	if err := json.Unmarshal([]byte(resp.Json), &result); err != nil {
		return false, "", fmt.Errorf("failed to parse user channels response: %v", err)
	}

	// Check if user exists
	if len(result.UserChannels) > 0 {
		channel := result.UserChannels[0]
		if channel.Verified {
			return true, channel.UserID, nil
		} else {
			// Channel exists but not verified - treat as new user for security
			log.Printf("⚠️ Found unverified channel for %s - treating as new user", channelDID)
			return false, "", nil
		}
	}

	return false, "", nil
}

// updateChannelLastUsed updates the lastUsedAt timestamp for a channel
func updateChannelLastUsed(channelDID, channelType string) error {
	// Find the channel UID first
	query := fmt.Sprintf(`{
		channel(func: eq(channelHash, "%s")) @filter(eq(channelType, "%s")) {
			uid
		}
	}`, channelDID, channelType)

	resp, err := dgraph.ExecuteQuery("dgraph", dgraph.NewQuery(query))
	if err != nil {
		return fmt.Errorf("failed to find channel for update: %v", err)
	}

	var result struct {
		Channel []struct {
			UID string `json:"uid"`
		} `json:"channel"`
	}

	if err := json.Unmarshal([]byte(resp.Json), &result); err != nil {
		return fmt.Errorf("failed to parse channel query: %v", err)
	}

	if len(result.Channel) > 0 {
		// Update the specific channel using N-Quads format
		updateNquads := fmt.Sprintf(`<%s> <lastUsedAt> "%s" .`, result.Channel[0].UID, time.Now().Format(time.RFC3339))

		// Create mutation object with N-Quads
		mutationObj := dgraph.NewMutation().WithSetNquads(updateNquads)
		_, err := dgraph.ExecuteMutations("dgraph", mutationObj)
		if err != nil {
			return fmt.Errorf("failed to update channel lastUsedAt: %v", err)
		}

		log.Printf("✅ Updated lastUsedAt for channel %s", channelDID)
	}

	return nil
}

// Helper function to create a new user channel entry (for registration flow)
func CreateUserChannel(userID, channelDID, channelType string, verified, primary bool) error {
	nquads := fmt.Sprintf(`_:channel <dgraph.type> "UserChannels" .
_:channel <userId> "%s" .
_:channel <channelType> "%s" .
_:channel <channelHash> "%s" .
_:channel <verified> "%t" .
_:channel <primary> "%t" .
_:channel <createdAt> "%s" .
_:channel <lastUsedAt> "%s" .`,
		userID, channelType, channelDID, verified, primary,
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339))

	mutationObj := dgraph.NewMutation().WithSetNquads(nquads)
	_, err := dgraph.ExecuteMutations("dgraph", mutationObj)
	if err != nil {
		return fmt.Errorf("failed to create user channel: %v", err)
	}

	log.Printf("✅ Created user channel: %s -> %s", userID, channelType)
	return nil
}

// CreateNewUser creates a new user with PENDING status and assigns 'registered' role
func CreateNewUser(channelDID, channelType string) (string, error) {
	log.Printf("🆕 Creating new user for channel: %s (%s)", channelDID, channelType)
	
	// Generate a unique user ID using timestamp and channel hash
	userID := fmt.Sprintf("user_%d_%s", time.Now().Unix(), channelDID[len(channelDID)-8:])
	
	// First, get the 'registered' role UID
	roleQuery := `query {
		registeredRole(func: eq(name, "registered")) {
			uid
			name
		}
	}`
	
	roleResult, err := dgraph.ExecuteQuery("dgraph", dgraph.NewQuery(roleQuery))
	if err != nil {
		return "", fmt.Errorf("failed to query registered role: %v", err)
	}
	
	var roleData struct {
		RegisteredRole []struct {
			UID  string `json:"uid"`
			Name string `json:"name"`
		} `json:"registeredRole"`
	}
	
	err = json.Unmarshal([]byte(roleResult.Json), &roleData)
	if err != nil {
		return "", fmt.Errorf("failed to parse role query: %v", err)
	}
	
	var roleUID string
	if len(roleData.RegisteredRole) > 0 {
		roleUID = roleData.RegisteredRole[0].UID
		log.Printf("✅ Found registered role: %s", roleUID)
	} else {
		// If no registered role exists, continue without it but log warning
		log.Printf("⚠️  Warning: 'registered' role not found, creating user without role")
	}
	
	// Create the user with all required fields
	currentTime := time.Now().Format(time.RFC3339)
	nquads := fmt.Sprintf(`_:user <dgraph.type> "User" .
_:user <status> "PENDING" .
_:user <did> "%s" .
_:user <createdAt> "%s" .
_:user <updatedAt> "%s" .`,
		userID, currentTime, currentTime)
	
	// Add role assignment if role exists
	if roleUID != "" {
		nquads += fmt.Sprintf(`
_:user <roles> <%s> .`, roleUID)
	}
	
	// Execute user creation mutation
	mutationObj := dgraph.NewMutation().WithSetNquads(nquads)
	result, err := dgraph.ExecuteMutations("dgraph", mutationObj)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %v", err)
	}
	
	// Extract the created user UID
	var newUserUID string
	if result.Uids != nil {
		if uid, exists := result.Uids["user"]; exists {
			newUserUID = uid
		}
	}
	
	if newUserUID == "" {
		return "", fmt.Errorf("failed to get created user UID")
	}
	
	log.Printf("✅ Created new user: %s (UID: %s)", userID, newUserUID)
	
	// Create the user channel association
	err = CreateUserChannel(userID, channelDID, channelType, true, true)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to create user channel: %v", err)
	}
	
	return userID, nil
}

// WebAuthn Integration Functions

// InitiateWebAuthnRegistration creates a WebAuthn registration challenge
func InitiateWebAuthnRegistration(userID, username, displayName string) (*webauthn.ChallengeResponse, error) {
	log.Printf("🔐 CerberusMFA: Initiating WebAuthn registration for user %s", userID)
	
	ctx := context.Background()
	webauthnService := webauthn.NewWebAuthnService()
	
	req := webauthn.ChallengeRequest{
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
	}
	
	response, err := webauthnService.CreateRegistrationChallenge(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn registration challenge: %v", err)
	}
	
	log.Printf("✅ WebAuthn registration challenge created for user %s", userID)
	return &response, nil
}

// VerifyWebAuthnRegistration verifies a WebAuthn registration response
func VerifyWebAuthnRegistration(req webauthn.RegistrationRequest) (*webauthn.RegistrationResponse, error) {
	log.Printf("🔐 CerberusMFA: Verifying WebAuthn registration for user %s", req.UserID)
	
	ctx := context.Background()
	webauthnService := webauthn.NewWebAuthnService()
	
	response, err := webauthnService.VerifyRegistration(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify WebAuthn registration: %v", err)
	}
	
	if response.Success {
		log.Printf("✅ WebAuthn registration verified for user %s", req.UserID)
	} else {
		log.Printf("❌ WebAuthn registration failed for user %s: %s", req.UserID, response.Message)
	}
	
	return &response, nil
}

// InitiateWebAuthnAuthentication creates a WebAuthn authentication challenge
func InitiateWebAuthnAuthentication(userID string) (*webauthn.AssertionChallengeResponse, error) {
	log.Printf("🔐 CerberusMFA: Initiating WebAuthn authentication for user %s", userID)
	
	webauthnService := webauthn.NewWebAuthnService()
	
	req := webauthn.AssertionChallengeRequest{
		UserID: userID,
	}
	
	response, err := webauthnService.CreateAuthenticationChallenge(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn authentication challenge: %v", err)
	}
	
	log.Printf("✅ WebAuthn authentication challenge created for user %s", userID)
	return &response, nil
}

// VerifyWebAuthnAuthentication verifies a WebAuthn authentication response
func VerifyWebAuthnAuthentication(req webauthn.AuthenticationRequest) (*webauthn.AuthenticationResponse, error) {
	log.Printf("🔐 CerberusMFA: Verifying WebAuthn authentication for user %s", req.UserID)
	
	webauthnService := webauthn.NewWebAuthnService()
	
	response, err := webauthnService.VerifyAuthentication(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify WebAuthn authentication: %v", err)
	}
	
	if response.Success {
		log.Printf("✅ WebAuthn authentication verified for user %s", req.UserID)
	} else {
		log.Printf("❌ WebAuthn authentication failed for user %s: %s", req.UserID, response.Message)
	}
	
	return &response, nil
}
