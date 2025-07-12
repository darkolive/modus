package charonotp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/hypermodeinc/modus/sdk/go/pkg/console"
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

// hashString creates a SHA-256 hash of the input string
func hashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// logAuditEvent creates an audit trail entry for OTP operations
// Follows ThemisLog agent patterns for ISO 27001 compliance
func logAuditEvent(category, action, objectType, objectId, performedBy, details string) {
	// Generate unique audit ID
	auditID := fmt.Sprintf("audit_%d", time.Now().UnixNano())
	
	// Create audit entry with proper timestamp and retention
	retentionDate := time.Now().AddDate(7, 0, 0) // 7 years retention for compliance
	
	// Escape quotes in details to prevent DQL syntax errors
	escapedDetails := strings.ReplaceAll(details, `"`, `\"`) 
	
	nquads := fmt.Sprintf(`_:audit <id> "%s" .
_:audit <category> "%s" .
_:audit <action> "%s" .
_:audit <objectType> "%s" .
_:audit <objectId> "%s" .
_:audit <performedBy> "%s" .
_:audit <timestamp> "%s"^^<xs:dateTime> .
_:audit <details> "%s" .
_:audit <severity> "INFO" .
_:audit <source> "CharonOTP" .
_:audit <retentionDate> "%s"^^<xs:dateTime> .
_:audit <dgraph.type> "AuditEntry" .`,
		auditID, category, action, objectType, objectId, performedBy,
		time.Now().Format(time.RFC3339), escapedDetails, retentionDate.Format(time.RFC3339))
	
	// Store audit entry asynchronously (fire-and-forget for performance)
	mutationObj := dgraph.NewMutation().WithSetNquads(nquads)
	_, err := dgraph.ExecuteMutations("dgraph", mutationObj)
	if err != nil {
		// Log audit failure but don't block main operation
		console.Warn(fmt.Sprintf("Audit logging failed: %s", err.Error()))
	}
}



// executeMutation executes a DQL mutation using Dgraph SDK
func executeMutation(nquads string) error {
	mutationObj := dgraph.NewMutation().WithSetNquads(nquads)
	_, err := dgraph.ExecuteMutations("dgraph", mutationObj)
	if err != nil {
		return fmt.Errorf("failed to execute mutation: %w", err)
	}
	return nil
}

// storeOTPInDgraph stores the OTP record in Dgraph using Modus SDK best practices
func storeOTPInDgraph(channel, recipient, otpCode string, expiresAt time.Time) (string, error) {
	// Use Modus SDK console for structured logging
	// Debug: console.Log(fmt.Sprintf("Starting OTP storage for channel: %s", channel))
	
	start := time.Now()
	
	// Hash sensitive data for privacy (ISO 27001 compliance)
	channelHash := hashString(recipient)
	otpHash := hashString(otpCode)
	
	// Generate temporary OTP ID for immediate response
	otpID := fmt.Sprintf("otp_%d", start.UnixNano())
	
	// Create N-Quads format using proven working pattern from memories
	nquads := fmt.Sprintf(`_:channelotp <channelHash> "%s" .
_:channelotp <channelType> "%s" .
_:channelotp <otpHash> "%s" .
_:channelotp <verified> "false"^^<xs:boolean> .
_:channelotp <expiresAt> "%s"^^<xs:dateTime> .
_:channelotp <used> "false"^^<xs:boolean> .
_:channelotp <dgraph.type> "ChannelOTP" .`,
		channelHash, channel, otpHash, expiresAt.Format(time.RFC3339))
	
	// Execute mutation using proven Modus SDK pattern
	mutationObj := dgraph.NewMutation().WithSetNquads(nquads)
	result, err := dgraph.ExecuteMutations("dgraph", mutationObj)
	
	if err != nil {
		// Log error with audit trail
		console.Error(fmt.Sprintf("OTP storage failed: %s (duration: %v)", err.Error(), time.Since(start)))
		
		// Create audit entry for failed OTP storage
		logAuditEvent("AUTHENTICATION", "OTP_STORAGE_FAILED", "ChannelOTP", otpID, "CharonOTP", 
			fmt.Sprintf(`{"channel":"%s","error":"%s","duration_ms":%d}`, 
				channel, err.Error(), time.Since(start).Milliseconds()))
		
		return "", fmt.Errorf("failed to store OTP: %w", err)
	}
	
	// Extract UID from response using proven pattern
	otpUID, exists := result.Uids["channelotp"]
	if !exists {
		console.Warn(fmt.Sprintf("No UID returned from Dgraph for channel: %s", channel))
		return otpID, nil // Return generated ID as fallback
	}
	
	// Log successful storage with audit trail
	// Debug: console.Log(fmt.Sprintf("OTP stored successfully: %s (duration: %v)", otpUID, time.Since(start)))
	
	// Create audit entry for successful OTP storage
	logAuditEvent("AUTHENTICATION", "OTP_GENERATED", "ChannelOTP", otpUID, "CharonOTP",
		fmt.Sprintf(`{"channel":"%s","expiresAt":"%s","duration_ms":%d}`,
			channel, expiresAt.Format(time.RFC3339), time.Since(start).Milliseconds()))
	
	return otpUID, nil
}

// sendOTPViaEmail sends OTP via email using the async email queue for instant response
func sendOTPViaEmail(recipient, otpCode string) error {
	// Use the ASYNC email service for non-blocking OTP emails
	response, err := email.SendOTPEmailAsync(
		recipient,
		otpCode,
	)
	
	if err != nil {
		return fmt.Errorf("failed to queue OTP email: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("email service error: %s", response.Error)
	}
	
	// Email is now queued for background processing - instant return!
	return nil
}

// sendOTPViaOtherChannels sends OTP via SMS, WhatsApp, or Telegram using IrisMessage
func sendOTPViaOtherChannels(channel string, recipient, _ string) error {
	// TODO: Implement IrisMessage integration for SMS, WhatsApp, Telegram
	// This is a placeholder until IrisMessage agent is implemented
	
	// Log the attempt for debugging
	// Debug: console.Log(fmt.Sprintf("Attempting to send OTP via %s to %s (code: %s...)", channel, recipient, otpCode[:2]))
	
	switch channel {
	case "sms":
		// TODO: Call IrisMessage SMS function
		return fmt.Errorf("SMS channel not yet implemented for %s - waiting for IrisMessage agent", recipient)
	case "whatsapp":
		// TODO: Call IrisMessage WhatsApp function
		return fmt.Errorf("WhatsApp channel not yet implemented for %s - waiting for IrisMessage agent", recipient)
	case "telegram":
		// TODO: Call IrisMessage Telegram function
		return fmt.Errorf("Telegram channel not yet implemented for %s - waiting for IrisMessage agent", recipient)
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
	
	// Generate OTP
	otpCode, err := generateOTP()
	if err != nil {
		return OTPResponse{}, fmt.Errorf("failed to generate OTP: %w", err)
	}
	
	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(expiryMins) * time.Minute)
	
	// Send OTP via appropriate channel FIRST (fast path)
	var sendErr error
	switch req.Channel {
	case "email":
		sendErr = sendOTPViaEmail(req.Recipient, otpCode)
	case "sms", "whatsapp", "telegram":
		sendErr = sendOTPViaOtherChannels(req.Channel, req.Recipient, otpCode)
	default:
		return OTPResponse{}, fmt.Errorf("unsupported channel: %s", req.Channel)
	}

	// Log send error but don't return early - allow OTP storage and graceful response
	if sendErr != nil {
		console.Error(fmt.Sprintf("Failed to send OTP via %s: %v", req.Channel, sendErr))
	}

	// Store OTP in Dgraph synchronously (WASM compatible)
	// Debug: console.Log("Starting synchronous OTP storage")
	storageStart := time.Now()
	otpID, storageErr := storeOTPInDgraph(req.Channel, req.Recipient, otpCode, expiresAt)
	if storageErr != nil {
		console.Error(fmt.Sprintf("OTP storage failed after %v: %v", time.Since(storageStart), storageErr))
		// Use fallback ID for response even if storage fails
		otpID = fmt.Sprintf("otp_%d", time.Now().UnixNano())
	} else {
		// Debug: console.Log(fmt.Sprintf("OTP storage completed in %v", time.Since(storageStart)))
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
	
	// Debug: console.Log(fmt.Sprintf("🔍 Verifying OTP: channel=%s, code=%s", req.Recipient, req.OTPCode))
	// Debug: console.Log(fmt.Sprintf("🔍 Hashes: channelHash=%s, otpHash=%s", channelHash, otpHash))
	
	// Query Dgraph to find matching OTP record using proper Modus SDK
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
	
	// Debug: console.Log(fmt.Sprintf("🔍 DQL Query: %s", query))
	
	// Execute query using Modus SDK
	queryObj := dgraph.NewQuery(query)
	result, err := dgraph.ExecuteQuery("dgraph", queryObj)
	if err != nil {
		console.Error(fmt.Sprintf("❌ Query execution failed: %v", err))
		return VerifyOTPResponse{
			Verified: false,
			Message:  "Failed to verify OTP: database error",
		}, fmt.Errorf("failed to query OTP: %w", err)
	}
	
	// Debug: console.Log(fmt.Sprintf("🔍 Query result JSON: %s", result.Json))
	
	// Parse query response directly from result.Json
	var response struct {
		OTPVerification []struct {
			UID         string    `json:"uid"`
			ExpiresAt   time.Time `json:"expiresAt"`
			UserID      string    `json:"userId"`
			ChannelType string    `json:"channelType"`
		} `json:"otp_verification"`
	}
	
	if result.Json == "" {
		// Debug: console.Log("🔍 Empty JSON response from Dgraph")
		return VerifyOTPResponse{
			Verified: false,
			Message:  "Invalid OTP code or OTP has already been used",
		}, nil
	}
	
	if err := json.Unmarshal([]byte(result.Json), &response); err != nil {
		console.Error(fmt.Sprintf("❌ JSON parsing failed: %v", err))
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

	// Use direct HTTP mutation to avoid v25 SDK compatibility issues
	err := executeMutation(nquads)
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

	// Debug: console.Log(fmt.Sprintf("🔍 Checking user existence with query: %s", query))
	
	// Execute query using Modus SDK
	queryObj := dgraph.NewQuery(query)
	result, err := dgraph.ExecuteQuery("dgraph", queryObj)
	if err != nil {
		console.Error(fmt.Sprintf("❌ User query execution failed: %v", err))
		return false, "", fmt.Errorf("failed to query user: %v", err)
	}

	// Debug: console.Log(fmt.Sprintf("🔍 User query result JSON: %s", result.Json))

	// Parse the response directly from result.Json
	var response struct {
		User []struct {
			UID    string `json:"uid"`
			Email  string `json:"email,omitempty"`
			Phone  string `json:"phone,omitempty"`
			Status string `json:"status"`
		} `json:"user"`
	}

	if result.Json == "" {
		// Debug: console.Log("🔍 Empty JSON response from user query - user does not exist")
		return false, "", nil
	}

	if err := json.Unmarshal([]byte(result.Json), &response); err != nil {
		console.Error(fmt.Sprintf("❌ User query JSON parsing failed: %v", err))
		return false, "", fmt.Errorf("failed to parse user query response: %v", err)
	}

	// Check if user exists
	if len(response.User) > 0 {
		// Debug: console.Log(fmt.Sprintf("✅ User found: UID=%s", response.User[0].UID))
		return true, response.User[0].UID, nil
	}

	// Debug: console.Log("🔍 No user found with this channelDID")

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