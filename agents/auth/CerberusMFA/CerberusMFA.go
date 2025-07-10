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
		// New user - proceed to registration flow
		log.Printf("🆕 New user detected for channel %s", req.ChannelDID)

		return &CerberusMFAResponse{
			UserExists:       false,
			Action:          "register",
			UserID:          "",
			AvailableMethods: []string{"passwordless"},
			NextStep:        "Complete user registration and setup Passwordless authentication",
			Message:         "Welcome! Let's create your account and set up secure authentication.",
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
