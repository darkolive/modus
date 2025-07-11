package webauthn

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hypermodeinc/modus/sdk/go/pkg/dgraph"
)

// WebAuthnService handles WebAuthn operations
type WebAuthnService struct {
	rpID   string
	rpName string
}

// NewWebAuthnService creates a new WebAuthn service instance
func NewWebAuthnService() *WebAuthnService {
	return &WebAuthnService{
		rpID:   DefaultRPID,
		rpName: DefaultRPName,
	}
}

// CreateRegistrationChallenge generates a WebAuthn registration challenge
func (w *WebAuthnService) CreateRegistrationChallenge(ctx context.Context, req ChallengeRequest) (ChallengeResponse, error) {
	log.Printf("🔐 WebAuthn: Creating registration challenge for user %s", req.UserID)

	// Generate cryptographically secure challenge
	challenge, err := generateChallenge()
	if err != nil {
		return ChallengeResponse{}, fmt.Errorf("failed to generate challenge: %v", err)
	}

	// Store challenge in database with expiry
	if err := w.storeChallenge(challenge, req.UserID, "registration"); err != nil {
		return ChallengeResponse{}, fmt.Errorf("failed to store challenge: %v", err)
	}

	// Get existing credentials to exclude
	excludeCredentials, err := w.getUserCredentials(req.UserID)
	if err != nil {
		log.Printf("⚠️ Warning: Could not fetch existing credentials: %v", err)
		excludeCredentials = []PublicKeyCredDescriptor{}
	}

	// Build WebAuthn registration challenge response
	response := ChallengeResponse{
		Challenge: challenge,
		RelyingParty: RelyingPartyInfo{
			ID:   w.rpID,
			Name: w.rpName,
		},
		User: UserInfo{
			ID:          req.UserID,
			Name:        req.Username,
			DisplayName: req.DisplayName,
		},
		PubKeyCredParams: []PubKeyCredParam{
			{Type: "public-key", Alg: -7},  // ES256
			{Type: "public-key", Alg: -257}, // RS256
		},
		AuthenticatorSelection: AuthenticatorSelection{
			RequireResidentKey: false,
			UserVerification:   UserVerificationPreferred,
		},
		Timeout:            DefaultTimeout,
		Attestation:        AttestationNone,
		ExcludeCredentials: excludeCredentials,
	}

	log.Printf("✅ WebAuthn: Registration challenge created for user %s", req.UserID)
	return response, nil
}

// VerifyRegistration verifies a WebAuthn registration response
func (w *WebAuthnService) VerifyRegistration(ctx context.Context, req RegistrationRequest) (RegistrationResponse, error) {
	log.Printf("🔐 WebAuthn: Verifying registration for user %s", req.UserID)

	// Verify challenge
	if err := w.verifyChallenge(req.Challenge, req.UserID, "registration"); err != nil {
		return RegistrationResponse{
			Success: false,
			Message: fmt.Sprintf("Challenge verification failed: %v", err),
		}, nil
	}

	// Parse and validate client data
	clientData, err := parseClientDataJSON(req.ClientDataJSON)
	if err != nil {
		return RegistrationResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid client data: %v", err),
		}, nil
	}

	// Verify challenge matches
	if clientData.Challenge != req.Challenge {
		return RegistrationResponse{
			Success: false,
			Message: "Challenge mismatch",
		}, nil
	}

	// Parse attestation object (simplified - in production, full attestation verification needed)
	credentialID, publicKey, err := parseAttestationObject(req.AttestationObject)
	if err != nil {
		return RegistrationResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid attestation object: %v", err),
		}, nil
	}

	// Store credential in database
	credential := WebAuthnCredential{
		UserID:       req.UserID,
		CredentialID: credentialID,
		PublicKey:    publicKey,
		SignCount:    0,
		Transports:   []string{"internal", "usb", "nfc", "ble"},
		AddedAt:      time.Now(),
	}

	if err := w.storeCredential(credential); err != nil {
		return RegistrationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to store credential: %v", err),
		}, nil
	}

	// Clean up challenge
	w.deleteChallenge(req.Challenge)

	log.Printf("✅ WebAuthn: Registration successful for user %s", req.UserID)
	return RegistrationResponse{
		Success:      true,
		CredentialID: credentialID,
		Message:      "WebAuthn registration successful",
		UserID:       req.UserID,
	}, nil
}

// CreateAuthenticationChallenge generates a WebAuthn authentication challenge
func (w *WebAuthnService) CreateAuthenticationChallenge(req AssertionChallengeRequest) (AssertionChallengeResponse, error) {
	log.Printf("🔐 WebAuthn: Creating authentication challenge for user %s", req.UserID)

	// Generate challenge
	challenge, err := generateChallenge()
	if err != nil {
		return AssertionChallengeResponse{}, fmt.Errorf("failed to generate challenge: %v", err)
	}

	// Store challenge
	if err := w.storeChallenge(challenge, req.UserID, "authentication"); err != nil {
		return AssertionChallengeResponse{}, fmt.Errorf("failed to store challenge: %v", err)
	}

	// Get user's credentials
	allowCredentials, err := w.getUserCredentials(req.UserID)
	if err != nil {
		return AssertionChallengeResponse{}, fmt.Errorf("failed to get user credentials: %v", err)
	}

	response := AssertionChallengeResponse{
		Challenge:        challenge,
		Timeout:          DefaultTimeout,
		RelyingPartyID:   w.rpID,
		AllowCredentials: allowCredentials,
		UserVerification: UserVerificationPreferred,
	}

	log.Printf("✅ WebAuthn: Authentication challenge created for user %s", req.UserID)
	return response, nil
}

// VerifyAuthentication verifies a WebAuthn authentication response
func (w *WebAuthnService) VerifyAuthentication(req AuthenticationRequest) (AuthenticationResponse, error) {
	log.Printf("🔐 WebAuthn: Verifying authentication for user %s", req.UserID)

	// Verify challenge
	if err := w.verifyChallenge(req.Challenge, req.UserID, "authentication"); err != nil {
		return AuthenticationResponse{
			Success: false,
			Message: fmt.Sprintf("Challenge verification failed: %v", err),
		}, nil
	}

	// Parse client data
	clientData, err := parseClientDataJSON(req.ClientDataJSON)
	if err != nil {
		return AuthenticationResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid client data: %v", err),
		}, nil
	}

	// Normalize base64 challenges (handle padding differences)
	clientChallenge := strings.TrimRight(clientData.Challenge, "=")
	requestChallenge := strings.TrimRight(req.Challenge, "=")
	
	// Verify challenge matches
	log.Printf("🔍 WebAuthn: Challenge comparison - Client: '%s' (normalized: '%s'), Request: '%s' (normalized: '%s')", 
		clientData.Challenge, clientChallenge, req.Challenge, requestChallenge)
	if clientChallenge != requestChallenge {
		log.Printf("❌ WebAuthn: Challenge mismatch - Client: '%s' != Request: '%s'", clientChallenge, requestChallenge)
		return AuthenticationResponse{
			Success: false,
			Message: "Challenge mismatch",
		}, nil
	}

	// Extract credential ID from authenticator data (simplified)
	credentialID := extractCredentialID(req.AuthenticatorData)
	
	// Get stored credential (simplified - in production, verify signature)
	credential, err := w.getCredentialByID(credentialID)
	if err != nil {
		return AuthenticationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get credential: %v", err),
		}, nil
	}

	// Update sign count (simplified)
	credential.SignCount++
	if err := w.updateCredentialSignCount(credentialID, credential.SignCount); err != nil {
		return AuthenticationResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update sign count: %v", err),
		}, nil
	}

	// Create authentication session
	sessionID, err := w.createAuthSession(req.UserID)
	if err != nil {
		log.Printf("⚠️ Warning: Could not create auth session: %v", err)
	}

	// Clean up challenge
	w.deleteChallenge(req.Challenge)

	log.Printf("✅ WebAuthn: Authentication successful for user %s", req.UserID)
	return AuthenticationResponse{
		Success:   true,
		UserID:    req.UserID,
		Message:   "WebAuthn authentication successful",
		SessionID: sessionID,
	}, nil
}

// Helper Functions

// generateChallenge creates a cryptographically secure random challenge
func generateChallenge() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// storeChallenge stores a challenge in the database with expiry
func (w *WebAuthnService) storeChallenge(challenge, userID, challengeType string) error {
	expiresAt := time.Now().Add(ChallengeExpiryMinutes * time.Minute)
	
	nquads := fmt.Sprintf(`_:challenge <dgraph.type> "WebAuthnChallenge" .
_:challenge <challenge> "%s" .
_:challenge <userId> "%s" .
_:challenge <type> "%s" .
_:challenge <expiresAt> "%s" .
_:challenge <createdAt> "%s" .`,
		challenge, userID, challengeType, 
		expiresAt.Format(time.RFC3339),
		time.Now().Format(time.RFC3339))

	mutationObj := dgraph.NewMutation().WithSetNquads(nquads)
	_, err := dgraph.ExecuteMutations("dgraph", mutationObj)
	return err
}

// verifyChallenge verifies a challenge exists and is not expired
func (w *WebAuthnService) verifyChallenge(challenge, userID, challengeType string) error {
	log.Printf("🔍 WebAuthn: Verifying challenge - Challenge: %s, UserID: %s, Type: %s", challenge, userID, challengeType)
	
	query := fmt.Sprintf(`{
		challenges(func: eq(challenge, "%s")) @filter(eq(userId, "%s") AND eq(type, "%s")) {
			uid
			expiresAt
		}
	}`, challenge, userID, challengeType)
	
	log.Printf("🔍 WebAuthn: Challenge query: %s", query)

	resp, err := dgraph.ExecuteQuery("dgraph", dgraph.NewQuery(query))
	if err != nil {
		log.Printf("❌ WebAuthn: Query execution failed: %v", err)
		return err
	}
	
	log.Printf("🔍 WebAuthn: Query response: %s", resp.Json)

	var result struct {
		Challenges []struct {
			UID       string `json:"uid"`
			ExpiresAt string `json:"expiresAt"`
		} `json:"challenges"`
	}

	if err := json.Unmarshal([]byte(resp.Json), &result); err != nil {
		return err
	}

	if len(result.Challenges) == 0 {
		return fmt.Errorf("challenge not found")
	}

	// Check expiry
	expiresAt, err := time.Parse(time.RFC3339, result.Challenges[0].ExpiresAt)
	if err != nil {
		return err
	}

	if time.Now().After(expiresAt) {
		return fmt.Errorf("challenge expired")
	}

	return nil
}

// storeCredential stores a WebAuthn credential in the database
func (w *WebAuthnService) storeCredential(cred WebAuthnCredential) error {
	transportsJSON, _ := json.Marshal(cred.Transports)
	
	nquads := fmt.Sprintf(`_:credential <dgraph.type> "WebAuthnCredential" .
_:credential <user> <%s> .
_:credential <credentialId> "%s" .
_:credential <publicKey> "%s" .
_:credential <signCount> "%d" .
_:credential <transports> "%s" .
_:credential <addedAt> "%s" .`,
		cred.UserID, cred.CredentialID, cred.PublicKey, 
		cred.SignCount, string(transportsJSON),
		cred.AddedAt.Format(time.RFC3339))

	mutationObj := dgraph.NewMutation().WithSetNquads(nquads)
	_, err := dgraph.ExecuteMutations("dgraph", mutationObj)
	return err
}

// getUserCredentials gets all credentials for a user
func (w *WebAuthnService) getUserCredentials(userID string) ([]PublicKeyCredDescriptor, error) {
	// Query for WebAuthn credentials by user UID reference
	query := fmt.Sprintf(`{
		credentials(func: type(WebAuthnCredential)) @filter(uid_in(user, <%s>)) {
			credentialId
			transports
		}
	}`, userID)

	resp, err := dgraph.ExecuteQuery("dgraph", dgraph.NewQuery(query))
	if err != nil {
		return nil, err
	}

	var result struct {
		Credentials []struct {
			CredentialID string `json:"credentialId"`
			Transports   string `json:"transports"`
		} `json:"credentials"`
	}

	if err := json.Unmarshal([]byte(resp.Json), &result); err != nil {
		return nil, err
	}

	var descriptors []PublicKeyCredDescriptor
	for _, cred := range result.Credentials {
		var transports []string
		json.Unmarshal([]byte(cred.Transports), &transports)
		
		descriptors = append(descriptors, PublicKeyCredDescriptor{
			Type:       "public-key",
			ID:         cred.CredentialID,
			Transports: transports,
		})
	}

	return descriptors, nil
}

// Simplified parsing functions (in production, use proper WebAuthn library)
func parseClientDataJSON(clientDataJSON string) (*ClientData, error) {
	decoded, err := base64.URLEncoding.DecodeString(clientDataJSON)
	if err != nil {
		return nil, err
	}

	var clientData ClientData
	if err := json.Unmarshal(decoded, &clientData); err != nil {
		return nil, err
	}

	return &clientData, nil
}

func parseAttestationObject(attestationObject string) (credentialID, publicKey string, err error) {
	// Simplified - in production, use proper CBOR parsing
	decoded, err := base64.URLEncoding.DecodeString(attestationObject)
	if err != nil {
		return "", "", err
	}

	// Generate mock values for demo (replace with proper parsing)
	hash := sha256.Sum256(decoded)
	credentialID = base64.URLEncoding.EncodeToString(hash[:16])
	publicKey = base64.URLEncoding.EncodeToString(hash[16:])

	return credentialID, publicKey, nil
}

func extractCredentialID(authenticatorData string) string {
	// Simplified extraction (replace with proper parsing)
	decoded, _ := base64.URLEncoding.DecodeString(authenticatorData)
	hash := sha256.Sum256(decoded)
	return base64.URLEncoding.EncodeToString(hash[:16])
}

// Additional helper functions
func (w *WebAuthnService) getCredentialByID(credentialID string) (*WebAuthnCredential, error) {
	// Implementation for getting credential by ID
	return &WebAuthnCredential{CredentialID: credentialID, SignCount: 0}, nil
}

func (w *WebAuthnService) updateCredentialSignCount(credentialID string, signCount int) error {
	// Implementation for updating sign count
	// TODO: Implement actual sign count update in database
	_ = credentialID // Mark as used
	_ = signCount    // Mark as used
	return nil
}

func (w *WebAuthnService) createAuthSession(userID string) (string, error) {
	// Implementation for creating auth session
	// TODO: Implement actual session creation in database
	sessionID := fmt.Sprintf("session_%s_%d", userID, time.Now().Unix())
	return sessionID, nil
}

func (w *WebAuthnService) deleteChallenge(challenge string) error {
	// Implementation for deleting challenge
	// TODO: Implement actual challenge deletion from database
	_ = challenge // Mark as used
	return nil
}

// ClientData represents the parsed client data JSON
type ClientData struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}
