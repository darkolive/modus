package charonotp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/hypermodeinc/modus/sdk/go/pkg/dgraph"
	"modus/services/email"
)

// OTPRequest represents the request to generate and send OTP
type OTPRequest struct {
	Channel     string `json:"channel"`     // "email", "sms", "whatsapp", "telegram"
	Recipient   string `json:"recipient"`   // email, phone number, etc.
	UserID      string `json:"userId,omitempty"`
}

// OTPResponse represents the response after OTP generation
type OTPResponse struct {
	OTPID     string    `json:"otpId"`
	Sent      bool      `json:"sent"`
	Verified  bool      `json:"verified"`
	Channel   string `json:"channel"`
	ExpiresAt time.Time `json:"expiresAt"`
	Message   string    `json:"message,omitempty"`
}

// VerifyOTPRequest represents the request to verify an OTP
type VerifyOTPRequest struct {
	OTPCode   string `json:"otpCode"`
	Recipient string `json:"recipient"`
}

// VerifyOTPResponse represents the response after OTP verification
type VerifyOTPResponse struct {
	Verified  bool   `json:"verified"`
	Message   string `json:"message"`
	UserID    string `json:"userId,omitempty"`
	Action    string `json:"action,omitempty"` // "signin" or "register"
	ChannelDID string `json:"channelDID,omitempty"` // Unique identifier for the channel
}

// ChannelOTPRecord represents the OTP stored in Dgraph (matches ChannelOTP schema)
type ChannelOTPRecord struct {
	UID         string    `json:"uid,omitempty"`
	ChannelHash string    `json:"channelHash"`    // Hashed email/phone for privacy
	ChannelType string    `json:"channelType"`    // "email", "sms", "whatsapp", etc.
	OTPHash     string    `json:"otpHash"`        // Hashed OTP code for security
	Verified    bool      `json:"verified"`       // Whether OTP has been verified
	ExpiresAt   time.Time `json:"expiresAt"`      // When OTP expires
	CreatedAt   time.Time `json:"createdAt"`      // When OTP was created
	UserID      string    `json:"userId,omitempty"` // Optional user link
	Purpose     string    `json:"purpose"`        // "signin", "signup", etc.
	Used        bool      `json:"used"`           // Whether OTP consumed
}

// GenerateOTP generates a 6-digit numerical OTP
func generateOTP() (string, error) {
	max := big.NewInt(999999)
	min := big.NewInt(100000)
	
	n, err := rand.Int(rand.Reader, max.Sub(max, min).Add(max, big.NewInt(1)))
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%06d", n.Add(n, min).Int64()), nil
}

// hashString creates a SHA256 hash of the input string
func hashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}



// storeOTPInDgraph stores the OTP record in Dgraph using DQL mutation
func storeOTPInDgraph(_ context.Context, otpCode, channel, recipient, userID, purpose string, expiresAt time.Time) (string, error) {
	createdAt := time.Now()
	
	// Hash sensitive data for privacy
	channelHash := hashString(recipient)
	otpHash := hashString(otpCode)
	
	// Build N-Quads for ChannelOTP (pure N-Quads format, no mutation wrapper)
	nquads := fmt.Sprintf(`_:channelotp <channelHash> "%s" .
_:channelotp <channelType> "%s" .
_:channelotp <otpHash> "%s" .
_:channelotp <verified> "false"^^<xs:boolean> .
_:channelotp <expiresAt> "%s"^^<xs:dateTime> .
_:channelotp <createdAt> "%s"^^<xs:dateTime> .
_:channelotp <userId> "%s" .
_:channelotp <purpose> "%s" .
_:channelotp <used> "false"^^<xs:boolean> .
_:channelotp <dgraph.type> "ChannelOTP" .`,
		channelHash, channel, otpHash,
		expiresAt.Format(time.RFC3339),
		createdAt.Format(time.RFC3339),
		userID, purpose,
	)
	
	// Create Dgraph mutation with proper N-Quads format
	mutationObj := dgraph.NewMutation().WithSetNquads(nquads)
	
	// Execute DQL mutation using Dgraph SDK
	result, err := dgraph.ExecuteMutations("dgraph", mutationObj)
	if err != nil {
		return "", fmt.Errorf("failed to store OTP in Dgraph: %w", err)
	}
	
	// Get UID directly from response.Uids map
	otpUID, exists := result.Uids["channelotp"]
	if !exists {
		return "", fmt.Errorf("failed to get OTP UID from Dgraph response")
	}

	return otpUID, nil
}

// sendOTPViaEmail sends OTP via email using the email service
func sendOTPViaEmail(recipient, otpCode string) error {
	// Use the email service for sending OTP emails
	response, err := email.SendOTPEmail(
		recipient,
		otpCode,
	)
	
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("email service error: %s", response.Error)
	}
	
	return nil
}

// sendOTPViaOtherChannels sends OTP via SMS, WhatsApp, or Telegram using IrisMessage
func sendOTPViaOtherChannels(_ context.Context, channel string, _ string, _ string, _ string) error {
	// TODO: Implement IrisMessage integration for SMS, WhatsApp, Telegram
	// This is a placeholder until IrisMessage agent is implemented
	
	// Format message for different channels
	// message := fmt.Sprintf("Your OTP code for %s is: %s. This code expires in 5 minutes.", purpose, otpCode)
	
	switch channel {
	case "sms":
		// TODO: Call IrisMessage SMS function
		return fmt.Errorf("SMS channel not yet implemented - waiting for IrisMessage agent")
	case "whatsapp":
		// TODO: Call IrisMessage WhatsApp function
		return fmt.Errorf("whatsApp channel not yet implemented - waiting for IrisMessage agent")
	case "telegram":
		// TODO: Call IrisMessage Telegram function
		return fmt.Errorf("telegram channel not yet implemented - waiting for IrisMessage agent")
	default:
		return fmt.Errorf("unsupported channel: %s", channel)
	}
}

// SendOTP is the main exported function to generate and send OTP
func SendOTP(ctx context.Context, req OTPRequest) (OTPResponse, error) {
	// Validate request
	if req.Channel == "" {
		return OTPResponse{}, fmt.Errorf("channel is required")
	}
	if req.Recipient == "" {
		return OTPResponse{}, fmt.Errorf("recipient is required")
	}
	
	// Set hardcoded default values
	expiryMins := 5 // Fixed 5 minutes expiry
	purpose := "authentication" // Fixed purpose
	
	// Generate OTP
	otpCode, err := generateOTP()
	if err != nil {
		return OTPResponse{}, fmt.Errorf("failed to generate OTP: %w", err)
	}
	
	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(expiryMins) * time.Minute)
	
	// Store OTP in Dgraph
	otpID, err := storeOTPInDgraph(ctx, otpCode, string(req.Channel), req.Recipient, req.UserID, purpose, expiresAt)
	if err != nil {
		return OTPResponse{}, fmt.Errorf("failed to store OTP: %w", err)
	}
	
	// Send OTP via selected channel
	var sendErr error
	switch req.Channel {
	case "email":
		sendErr = sendOTPViaEmail(req.Recipient, otpCode)
	case "sms", "whatsapp", "telegram":
		sendErr = sendOTPViaOtherChannels(ctx, req.Channel, req.Recipient, otpCode, purpose)
	default:
		sendErr = fmt.Errorf("unsupported channel: %s", req.Channel)
	}
	
	response := OTPResponse{
		OTPID:     otpID,
		Sent:      sendErr == nil,
		Verified:  false, // OTP not verified yet
		Channel:   req.Channel,
		ExpiresAt: expiresAt,
	}
	
	if sendErr != nil {
		response.Message = fmt.Sprintf("OTP generated but failed to send: %v", sendErr)
	} else {
		response.Message = fmt.Sprintf("OTP sent successfully via %s", req.Channel)
	}
	
	return response, nil
}



// VerifyOTP verifies an OTP code against the database
// Frontend should store channel value and pass it with the OTP code
func VerifyOTP(req VerifyOTPRequest) (VerifyOTPResponse, error) {
	ctx := context.Background()
	
	// Hash the provided channel and OTP for database comparison
	channelHash := hashString(req.Recipient)
	otpHash := hashString(req.OTPCode)
	
	// Query Dgraph to find matching OTP record
	query := fmt.Sprintf(`{
		otp_verification(func: eq(channelHash, "%s")) @filter(eq(otpHash, "%s") AND eq(verified, false) AND eq(used, false)) {
			uid
			channelHash
			otpHash
			verified
			used
			expiresAt
			createdAt
			userId
			purpose
			channelType
		}
	}`, channelHash, otpHash)
	
	// Execute query using Dgraph SDK
	queryObj := dgraph.NewQuery(query)
	result, err := dgraph.ExecuteQuery("dgraph", queryObj)
	if err != nil {
		return VerifyOTPResponse{
			Verified: false,
			Message:  "Failed to verify OTP: database error",
		}, fmt.Errorf("failed to query OTP: %w", err)
	}
	
	// Parse query response
	var response struct {
		OTPVerification []struct {
			UID         string    `json:"uid"`
			ExpiresAt   time.Time `json:"expiresAt"`
			UserID      string    `json:"userId"`
			ChannelType string    `json:"channelType"`
		} `json:"otp_verification"`
	}
	
	if err := json.Unmarshal([]byte(result.Json), &response); err != nil {
		return VerifyOTPResponse{
			Verified: false,
			Message:  "Failed to parse verification response",
		}, fmt.Errorf("failed to parse query response: %w", err)
	}
	
	// Check if OTP was found
	if len(response.OTPVerification) == 0 {
		return VerifyOTPResponse{
			Verified: false,
			Message:  "Invalid OTP code or OTP has already been used",
		}, nil
	}
	
	otpRecord := response.OTPVerification[0]
	
	// Check if OTP has expired
	if time.Now().After(otpRecord.ExpiresAt) {
		return VerifyOTPResponse{
			Verified: false,
			Message:  "OTP has expired",
		}, nil
	}
	
	// Mark OTP as verified and used
	if err := markOTPAsVerifiedAndUsed(ctx, otpRecord.UID); err != nil {
		return VerifyOTPResponse{
			Verified: false,
			Message:  "Failed to update OTP status",
		}, fmt.Errorf("failed to mark OTP as used: %w", err)
	}
	
	// Determine channel type from the OTP record
	channelType := otpRecord.ChannelType
	
	// Perform post-OTP verification to check if user exists
	action, userID, channelDID, err := PostOTPVerification(channelType, req.Recipient)
	if err != nil {
		return VerifyOTPResponse{
			Verified: false,
			Message:  "Failed to determine next action",
		}, fmt.Errorf("post-OTP verification failed: %w", err)
	}
	
	// Return successful verification with routing information
	return VerifyOTPResponse{
		Verified:   true,
		Message:    "OTP verified successfully",
		UserID:     userID,
		Action:     action,     // "signin" or "register"
		ChannelDID: channelDID, // Unique identifier for the channel
	}, nil
}

// markOTPAsVerifiedAndUsed marks an OTP as both verified and used in Dgraph
func markOTPAsVerifiedAndUsed(_ context.Context, otpUID string) error {
	// Create DQL mutation to mark OTP as verified and used
	nquads := fmt.Sprintf(`
		<%s> <verified> "true"^^<xs:boolean> .
		<%s> <used> "true"^^<xs:boolean> .
	`, otpUID, otpUID)

	// Use the latest v25-compatible ExecuteMutations method
	mutationObj := dgraph.NewMutation().WithSetNquads(nquads)
	_, err := dgraph.ExecuteMutations("dgraph", mutationObj)
	if err != nil {
		return fmt.Errorf("failed to mark OTP as verified and used: %v", err)
	}

	fmt.Printf("✅ OTP %s marked as verified and used\n", otpUID)
	return nil
}

// generateChannelDID creates a unique DID for a channel (email/phone)
func generateChannelDID(channel, recipient string) string {
	// Create a unique identifier based on channel type and recipient
	// This ensures each email/phone has a unique DID across the system
	return hashString(fmt.Sprintf("%s:%s", channel, recipient))
}

// checkUserExists checks if a user exists by channel DID
func checkUserExists(channelDID, channelType string) (bool, string, error) {
	// Create DQL query to check if user exists by channel DID
	var query string
	switch channelType {
	case "email":
		query = fmt.Sprintf(`{
			user(func: eq(emailDID, "%s")) {
				uid
				email
				emailDID
				status
			}
		}`, channelDID)
	case "phone":
		query = fmt.Sprintf(`{
			user(func: eq(phoneDID, "%s")) {
				uid
				phone
				phoneDID
				status
			}
		}`, channelDID)
	default:
		return false, "", fmt.Errorf("unsupported channel type: %s", channelType)
	}

	// Use the latest v25-compatible ExecuteQuery method
	resp, err := dgraph.ExecuteQuery("dgraph", dgraph.NewQuery(query))
	if err != nil {
		return false, "", fmt.Errorf("failed to query user: %v", err)
	}

	// Parse the response
	var result struct {
		User []struct {
			UID    string `json:"uid"`
			Email  string `json:"email,omitempty"`
			Phone  string `json:"phone,omitempty"`
			Status string `json:"status"`
		} `json:"user"`
	}

	if err := json.Unmarshal([]byte(resp.Json), &result); err != nil {
		return false, "", fmt.Errorf("failed to parse user query response: %v", err)
	}

	// Check if user exists
	if len(result.User) > 0 {
		return true, result.User[0].UID, nil
	}

	return false, "", nil
}

// PostOTPVerification handles the logic after OTP is successfully verified
// Checks if user exists and returns appropriate action (signin/register)
func PostOTPVerification(channel, recipient string) (string, string, string, error) {
	// Generate channel DID for unique identification
	channelDID := generateChannelDID(channel, recipient)

	// Check if user exists by channel DID
	userExists, userID, err := checkUserExists(channelDID, channel)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to check user existence: %v", err)
	}

	if userExists {
		// User exists - route to signin
		return "signin", userID, channelDID, nil
	} else {
		// User doesn't exist - route to register
		return "register", "", channelDID, nil
	}
}