package hecateregister

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hypermodeinc/modus/sdk/go/pkg/dgraph"
)

// UserRegistrationRequest represents the request to register a new user
type UserRegistrationRequest struct {
	// Channel information from CharonOTP verification
	ChannelDID   string `json:"channelDID"`   // Unique identifier from OTP verification
	ChannelType  string `json:"channelType"`  // "email" or "phone"
	Recipient    string `json:"recipient"`    // email address or phone number
	
	// User profile information
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	DisplayName  string `json:"displayName,omitempty"`
	
	// Optional profile data
	Timezone     string `json:"timezone,omitempty"`
	Language     string `json:"language,omitempty"`
	
	// Registration metadata
	IPAddress    string `json:"ipAddress,omitempty"`
	UserAgent    string `json:"userAgent,omitempty"`
}

// UserRegistrationResponse represents the response after user registration
type UserRegistrationResponse struct {
	Success      bool      `json:"success"`
	UserID       string    `json:"userId"`
	Message      string    `json:"message"`
	
	// PII tokenization results
	PIITokens    map[string]string `json:"piiTokens,omitempty"`
	
	// Identity verification status
	IdentityCheckID string `json:"identityCheckId,omitempty"`
	
	// Audit information
	AuditEventID    string    `json:"auditEventId,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

// PIITokenizationRequest for internal PII handling
type PIITokenizationRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

// PIITokenizationResponse from internal/pii service
type PIITokenizationResponse struct {
	Tokens map[string]string `json:"tokens"`
	Status string           `json:"status"`
}

// AuditEvent for ISO compliance
type AuditEvent struct {
	EventType    string                 `json:"eventType"`
	UserID       string                 `json:"userId"`
	Timestamp    time.Time              `json:"timestamp"`
	IPAddress    string                 `json:"ipAddress,omitempty"`
	UserAgent    string                 `json:"userAgent,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// generateUserID creates a unique user identifier
func generateUserID() string {
	// Generate a unique user ID (could use UUID or other method)
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

// tokenizePII handles PII tokenization via internal/pii service
func tokenizePII(req PIITokenizationRequest) (*PIITokenizationResponse, error) {
	// TODO: Integrate with internal/pii service for ISO-compliant tokenization
	// For now, return placeholder tokens
	tokens := map[string]string{
		"firstName": fmt.Sprintf("tok_fn_%d", time.Now().UnixNano()),
		"lastName":  fmt.Sprintf("tok_ln_%d", time.Now().UnixNano()),
	}
	
	if req.Email != "" {
		tokens["email"] = fmt.Sprintf("tok_em_%d", time.Now().UnixNano())
	}
	
	if req.Phone != "" {
		tokens["phone"] = fmt.Sprintf("tok_ph_%d", time.Now().UnixNano())
	}
	
	return &PIITokenizationResponse{
		Tokens: tokens,
		Status: "success",
	}, nil
}

// triggerIdentityCheck initiates identity verification via JanusFace
func triggerIdentityCheck(userID string) (string, error) {
	// TODO: Integrate with JanusFace agent for identity enrollment
	// For now, return placeholder identity check ID
	identityCheckID := fmt.Sprintf("id_check_%s_%d", userID, time.Now().UnixNano())
	
	fmt.Printf("🔍 Identity check initiated: %s for user %s\n", identityCheckID, userID)
	return identityCheckID, nil
}

// emitAuditEvent creates an ISO-compliant audit trail entry
func emitAuditEvent(event AuditEvent) (string, error) {
	// TODO: Integrate with ISO audit-trail system
	// For now, log the event and return placeholder audit ID
	auditID := fmt.Sprintf("audit_%d", time.Now().UnixNano())
	
	eventJSON, _ := json.MarshalIndent(event, "", "  ")
	fmt.Printf("📋 Audit Event [%s]:\n%s\n", auditID, string(eventJSON))
	
	return auditID, nil
}

// createUserInDgraph stores the new user record in Dgraph
func createUserInDgraph(req UserRegistrationRequest, userID string) error {
	// Determine channel DID field based on channel type
	var channelDIDField, channelField, channelVerifiedField string
	var channelValue string
	
	if req.ChannelType == "email" {
		channelDIDField = "emailDID"
		channelField = "email"
		channelVerifiedField = "emailVerified"
		channelValue = req.Recipient
	} else if req.ChannelType == "phone" {
		channelDIDField = "phoneDID"
		channelField = "phone"
		channelVerifiedField = "phoneVerified"
		channelValue = req.Recipient
	} else {
		return fmt.Errorf("unsupported channel type: %s", req.ChannelType)
	}
	
	// Create DQL mutation for user creation
	nquads := fmt.Sprintf(`
		_:user <dgraph.type> "User" .
		_:user <%s> "%s" .
		_:user <%s> "%s" .
		_:user <%s> "true"^^<xs:boolean> .
		_:user <createdAt> "%s"^^<xs:dateTime> .
		_:user <status> "active" .
	`, channelDIDField, req.ChannelDID,
		channelField, channelValue,
		channelVerifiedField,
		time.Now().Format(time.RFC3339))
	
	// Add user profile if provided
	if req.FirstName != "" || req.LastName != "" {
		profileNquads := fmt.Sprintf(`
			_:profile <dgraph.type> "UserProfile" .
			_:profile <userId> "%s" .
			_:profile <firstName> "%s" .
			_:profile <lastName> "%s" .
			_:profile <displayName> "%s" .
			_:profile <timezone> "%s" .
			_:profile <language> "%s" .
			_:profile <updatedAt> "%s"^^<xs:dateTime> .
		`, userID, req.FirstName, req.LastName, 
			req.DisplayName, req.Timezone, req.Language,
			time.Now().Format(time.RFC3339))
		
		nquads += profileNquads
	}
	
	// Execute mutation using Dgraph SDK
	mutationObj := dgraph.NewMutation().WithSetNquads(nquads)
	result, err := dgraph.ExecuteMutations("dgraph", mutationObj)
	if err != nil {
		return fmt.Errorf("failed to create user in Dgraph: %v", err)
	}
	
	// Extract the created user UID
	if len(result.Uids) > 0 {
		if uid, exists := result.Uids["user"]; exists {
			fmt.Printf("✅ User created with UID: %s\n", uid)
		}
	}
	
	return nil
}

// RegisterUser is the main exported function to register a new user
func RegisterUser(ctx context.Context, req UserRegistrationRequest) (UserRegistrationResponse, error) {
	fmt.Printf("🌙 HecateRegister: Initiating user registration for %s\n", req.Recipient)
	
	// Generate unique user ID
	userID := generateUserID()
	
	// Step 1: PII Tokenization for ISO compliance
	piiReq := PIITokenizationRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
	
	if req.ChannelType == "email" {
		piiReq.Email = req.Recipient
	} else if req.ChannelType == "phone" {
		piiReq.Phone = req.Recipient
	}
	
	piiResp, err := tokenizePII(piiReq)
	if err != nil {
		return UserRegistrationResponse{
			Success: false,
			Message: "Failed to tokenize PII data",
		}, fmt.Errorf("PII tokenization failed: %v", err)
	}
	
	// Step 2: Create user record in Dgraph
	if err := createUserInDgraph(req, userID); err != nil {
		return UserRegistrationResponse{
			Success: false,
			Message: "Failed to create user account",
		}, fmt.Errorf("user creation failed: %v", err)
	}
	
	// Step 3: Trigger identity verification
	identityCheckID, err := triggerIdentityCheck(userID)
	if err != nil {
		fmt.Printf("⚠️ Identity check failed (non-critical): %v\n", err)
		// Don't fail registration if identity check fails
	}
	
	// Step 4: Emit audit event for ISO compliance
	auditEvent := AuditEvent{
		EventType: "UserRegistered",
		UserID:    userID,
		Timestamp: time.Now(),
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		Metadata: map[string]interface{}{
			"channelType":     req.ChannelType,
			"channelDID":      req.ChannelDID,
			"registrationSource": "HecateRegister",
			"piiTokenized":    true,
			"identityCheckID": identityCheckID,
		},
	}
	
	auditEventID, err := emitAuditEvent(auditEvent)
	if err != nil {
		fmt.Printf("⚠️ Audit event failed (non-critical): %v\n", err)
		// Don't fail registration if audit fails
	}
	
	// Return successful registration response
	return UserRegistrationResponse{
		Success:         true,
		UserID:          userID,
		Message:         "User registration completed successfully",
		PIITokens:       piiResp.Tokens,
		IdentityCheckID: identityCheckID,
		AuditEventID:    auditEventID,
		CreatedAt:       time.Now(),
	}, nil
}
