package main

import (
	"context"
	"fmt"
	"time"

	charonotp "modus/agents/auth/CharonOTP"
	cerberusmfa "modus/agents/auth/CerberusMFA"
	chronossession "modus/agents/sessions/ChronosSession"
	"modus/services/webauthn"
)

// Valid OTP channel values (for reference)
// Supported channels: "email", "sms", "whatsapp", "telegram"

// OTPRequest represents the request to generate and send OTP
type OTPRequest struct {
	Channel   string `json:"channel"`
	Recipient string `json:"recipient"`
}

// OTPResponse represents the response from OTP generation and sending
type OTPResponse struct {
	OTPID     string    `json:"oTPID"`
	Sent      bool      `json:"sent"`
	Channel   string    `json:"channel"`
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
	Verified   bool   `json:"verified"`
	Message    string `json:"message"`
	UserID     string `json:"userId,omitempty"`
	Action     string `json:"action,omitempty"`     // "signin" or "register"
	ChannelDID string `json:"channelDID,omitempty"` // Unique identifier for the channel
}

// CerberusMFARequest represents the request for MFA flow decision
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

// WebAuthn Types for GraphQL

// WebAuthnChallengeRequest represents a request for WebAuthn challenge
type WebAuthnChallengeRequest struct {
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

// WebAuthnChallengeResponse represents a WebAuthn challenge response
type WebAuthnChallengeResponse struct {
	Challenge              string                    `json:"challenge"`
	RelyingParty          RelyingPartyInfo          `json:"rp"`
	User                  UserInfo                  `json:"user"`
	PubKeyCredParams      []PubKeyCredParam         `json:"pubKeyCredParams"`
	AuthenticatorSelection AuthenticatorSelection   `json:"authenticatorSelection"`
	Timeout               int                       `json:"timeout"`
	Attestation           string                    `json:"attestation"`
	ExcludeCredentials    []PublicKeyCredDescriptor `json:"excludeCredentials,omitempty"`
}

type RelyingPartyInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type PubKeyCredParam struct {
	Type string `json:"type"`
	Alg  int    `json:"alg"`
}

type AuthenticatorSelection struct {
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"`
	RequireResidentKey      bool   `json:"requireResidentKey"`
	UserVerification        string `json:"userVerification"`
}

type PublicKeyCredDescriptor struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	Transports []string `json:"transports,omitempty"`
}

// WebAuthnRegistrationRequest represents a WebAuthn registration request
type WebAuthnRegistrationRequest struct {
	UserID            string `json:"userId"`
	Challenge         string `json:"challenge"`
	ClientDataJSON    string `json:"clientDataJSON"`
	AttestationObject string `json:"attestationObject"`
}

// WebAuthnRegistrationResponse represents a WebAuthn registration response
type WebAuthnRegistrationResponse struct {
	Success      bool   `json:"success"`
	CredentialID string `json:"credentialId"`
	Message      string `json:"message"`
	UserID       string `json:"userId"`
}

// WebAuthnAuthRequest represents a WebAuthn authentication request
type WebAuthnAuthRequest struct {
	UserID            string `json:"userId"`
	Challenge         string `json:"challenge"`
	ClientDataJSON    string `json:"clientDataJSON"`
	AuthenticatorData string `json:"authenticatorData"`
	Signature         string `json:"signature"`
	UserHandle        string `json:"userHandle,omitempty"`
}

// WebAuthnAuthResponse represents a WebAuthn authentication response
type WebAuthnAuthResponse struct {
	Success   bool   `json:"success"`
	UserID    string `json:"userId"`
	Message   string `json:"message"`
	SessionID string `json:"sessionId,omitempty"`
}

// WebAuthnAssertionChallengeRequest represents a request for assertion challenge
type WebAuthnAssertionChallengeRequest struct {
	UserID string `json:"userId,omitempty"`
}

// WebAuthnAssertionChallengeResponse represents an assertion challenge response
type WebAuthnAssertionChallengeResponse struct {
	Challenge        string                    `json:"challenge"`
	Timeout          int                       `json:"timeout"`
	RelyingPartyID   string                    `json:"rpId"`
	AllowCredentials []PublicKeyCredDescriptor `json:"allowCredentials,omitempty"`
	UserVerification string                    `json:"userVerification"`
}

// Session Management Types

// SessionRequest represents a request to create a session after successful authentication
type SessionRequest struct {
	UserID     string `json:"userId"`
	ChannelDID string `json:"channelDID"`
	Action     string `json:"action"` // "signin" or "register"
}

// SessionResponse represents the response containing session information
type SessionResponse struct {
	Success     bool   `json:"success"`
	SessionID   string `json:"sessionId"`
	AccessToken string `json:"accessToken"`
	ExpiresAt   int64  `json:"expiresAt"`
	Message     string `json:"message"`
	UserID      string `json:"userId"`
}

// ValidateSessionRequest represents a request to validate an existing session
type ValidateSessionRequest struct {
	SessionID   string `json:"sessionId,omitempty"`
	AccessToken string `json:"accessToken,omitempty"`
}

// ValidateSessionResponse represents the response from session validation
type ValidateSessionResponse struct {
	Valid     bool   `json:"valid"`
	UserID    string `json:"userId"`
	Message   string `json:"message"`
	ExpiresAt int64  `json:"expiresAt,omitempty"`
}

// ChronosSession-compatible types for session lifecycle management

// ValidationRequest for ChronosSession token validation
type ValidationRequest struct {
	Token string `json:"token"`
}

// ValidationResponse for ChronosSession token validation results
type ValidationResponse struct {
	Valid     bool   `json:"valid"`
	UserID    string `json:"userId,omitempty"`
	ExpiresAt int64  `json:"expiresAt,omitempty"`
	Message   string `json:"message,omitempty"`
}

// RefreshRequest for extending an existing session
type RefreshRequest struct {
	Token string `json:"token"`
}

// RefreshResponse for session refresh results
type RefreshResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
	Message   string `json:"message,omitempty"`
}

// RevocationRequest for revoking a session
type RevocationRequest struct {
	Token  string `json:"token"`
	Reason string `json:"reason,omitempty"`
}

// RevocationResponse for session revocation results
type RevocationResponse struct {
	Revoked   bool   `json:"revoked"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}



// Convert main package verify types to charonotp package types
func convertToCharonVerifyRequest(req VerifyOTPRequest) charonotp.VerifyOTPRequest {
	return charonotp.VerifyOTPRequest{
		OTPCode:   req.OTPCode,
		Recipient: req.Recipient,
	}
}

func convertFromCharonVerifyResponse(resp charonotp.VerifyOTPResponse) VerifyOTPResponse {
	return VerifyOTPResponse{
		Verified:   resp.Verified,
		Message:    resp.Message,
		UserID:     resp.UserID,
		Action:     resp.Action,
		ChannelDID: resp.ChannelDID,
	}
}

// SendOTP is the exported wrapper function for Modus GraphQL
func SendOTP(req OTPRequest) (OTPResponse, error) {
	// Convert to charonotp types
	charonReq := charonotp.OTPRequest{
		Channel:   req.Channel,
		Recipient: req.Recipient,
	}
	
	// Call the charonotp agent to send OTP
	resp, err := charonotp.SendOTP(context.Background(), charonReq)
	if err != nil {
		return OTPResponse{}, err
	}
	
	// Convert response back to main types
	return OTPResponse{
		OTPID:     resp.OTPID,
		Sent:      resp.Sent,
		Channel:   resp.Channel,
		ExpiresAt: resp.ExpiresAt,
		Message:   resp.Message,
	}, nil
}

// VerifyOTP is the exported wrapper function for Modus
func VerifyOTP(req VerifyOTPRequest) (VerifyOTPResponse, error) {
	// Convert to charonotp types
	charonReq := convertToCharonVerifyRequest(req)

	// Call the charonotp agent
	resp, err := charonotp.VerifyOTP(charonReq)
	if err != nil {
		return VerifyOTPResponse{}, err
	}

	// Convert response back to main types
	return convertFromCharonVerifyResponse(resp), nil
}

// Convert main package types to cerberusmfa package types
func convertToCerberusMFARequest(req CerberusMFARequest) cerberusmfa.CerberusMFARequest {
	return cerberusmfa.CerberusMFARequest{
		ChannelDID:  req.ChannelDID,
		ChannelType: req.ChannelType,
	}
}

// Convert cerberusmfa package types to main package types
func convertFromCerberusMFAResponse(resp *cerberusmfa.CerberusMFAResponse) CerberusMFAResponse {
	return CerberusMFAResponse{
		UserExists:       resp.UserExists,
		Action:          resp.Action,
		UserID:          resp.UserID,
		AvailableMethods: resp.AvailableMethods,
		NextStep:        resp.NextStep,
		Message:         resp.Message,
	}
}

// CheckUserAndRoute determines if user should signin or register after OTP verification
func CheckUserAndRoute(req CerberusMFARequest) (CerberusMFAResponse, error) {
	// Convert to cerberusmfa types
	cerberusReq := convertToCerberusMFARequest(req)

	// Call the cerberusmfa agent
	resp, err := cerberusmfa.CerberusMFA(cerberusReq)
	if err != nil {
		return CerberusMFAResponse{}, err
	}

	// Convert response back to main types
	return convertFromCerberusMFAResponse(resp), nil
}

// SigninUser handles existing user signin flow (placeholder for WebAuthn/Passwordless integration)
func SigninUser(req CerberusMFARequest) (CerberusMFAResponse, error) {
	// First check if user exists and get routing info
	routeResp, err := CheckUserAndRoute(req)
	if err != nil {
		return CerberusMFAResponse{}, err
	}

	// If user doesn't exist, return error
	if !routeResp.UserExists {
		return CerberusMFAResponse{
			UserExists: false,
			Action:     "register",
			Message:    "User not found. Please register first.",
		}, nil
	}

	// For existing users, prepare signin response
	return CerberusMFAResponse{
		UserExists:       true,
		Action:          "signin",
		UserID:          routeResp.UserID,
		AvailableMethods: routeResp.AvailableMethods,
		NextStep:        "webauthn_or_passwordless",
		Message:         "User verified. Proceed with WebAuthn or Passwordless signin.",
	}, nil
}

// RegisterUser handles new user registration flow (placeholder for HecateRegister integration)
func RegisterUser(req CerberusMFARequest) (CerberusMFAResponse, error) {
	// First check if user already exists
	routeResp, err := CheckUserAndRoute(req)
	if err != nil {
		return CerberusMFAResponse{}, err
	}

	// If user already exists, return error
	if routeResp.UserExists {
		return CerberusMFAResponse{
			UserExists: true,
			Action:     "signin",
			Message:    "User already exists. Please signin instead.",
		}, nil
	}

	// For new users, prepare registration response
	return CerberusMFAResponse{
		UserExists:       false,
		Action:          "register",
		AvailableMethods: []string{"profile_creation", "identity_verification"},
		NextStep:        "user_profile_creation",
		Message:         "New user detected. Proceed with registration and profile creation.",
	}, nil
}

// WebAuthn Integration Functions

// CreateWebAuthnRegistrationChallenge creates a WebAuthn registration challenge
func CreateWebAuthnRegistrationChallenge(req WebAuthnChallengeRequest) (WebAuthnChallengeResponse, error) {
	// Call CerberusMFA integration function
	response, err := cerberusmfa.InitiateWebAuthnRegistration(req.UserID, req.Username, req.DisplayName)
	if err != nil {
		return WebAuthnChallengeResponse{}, err
	}

	// Convert response
	return convertFromWebAuthnChallengeResponse(*response), nil
}

// VerifyWebAuthnRegistration verifies a WebAuthn registration
func VerifyWebAuthnRegistration(req WebAuthnRegistrationRequest) (WebAuthnRegistrationResponse, error) {
	// Convert to service types
	serviceReq := webauthn.RegistrationRequest{
		UserID:            req.UserID,
		Challenge:         req.Challenge,
		ClientDataJSON:    req.ClientDataJSON,
		AttestationObject: req.AttestationObject,
	}

	// Call CerberusMFA integration function
	response, err := cerberusmfa.VerifyWebAuthnRegistration(serviceReq)
	if err != nil {
		return WebAuthnRegistrationResponse{}, err
	}

	// Convert response
	return convertFromWebAuthnRegistrationResponse(*response), nil
}

// CreateWebAuthnAuthenticationChallenge creates a WebAuthn authentication challenge
func CreateWebAuthnAuthenticationChallenge(req WebAuthnAssertionChallengeRequest) (WebAuthnAssertionChallengeResponse, error) {
	// Call CerberusMFA integration function
	response, err := cerberusmfa.InitiateWebAuthnAuthentication(req.UserID)
	if err != nil {
		return WebAuthnAssertionChallengeResponse{}, err
	}

	// Convert response
	return convertFromWebAuthnAssertionResponse(*response), nil
}

// VerifyWebAuthnAuthentication verifies a WebAuthn authentication
func VerifyWebAuthnAuthentication(req WebAuthnAuthRequest) (WebAuthnAuthResponse, error) {
	// Convert to service types
	serviceReq := webauthn.AuthenticationRequest{
		UserID:            req.UserID,
		Challenge:         req.Challenge,
		ClientDataJSON:    req.ClientDataJSON,
		AuthenticatorData: req.AuthenticatorData,
		Signature:         req.Signature,
		UserHandle:        req.UserHandle,
	}

	// Call CerberusMFA integration function
	response, err := cerberusmfa.VerifyWebAuthnAuthentication(serviceReq)
	if err != nil {
		return WebAuthnAuthResponse{}, err
	}

	// Convert response
	return convertFromWebAuthnAuthResponse(*response), nil
}

// Conversion Functions for WebAuthn

func convertFromWebAuthnChallengeResponse(resp webauthn.ChallengeResponse) WebAuthnChallengeResponse {
	return WebAuthnChallengeResponse{
		Challenge: resp.Challenge,
		RelyingParty: RelyingPartyInfo{
			ID:   resp.RelyingParty.ID,
			Name: resp.RelyingParty.Name,
		},
		User: UserInfo{
			ID:          resp.User.ID,
			Name:        resp.User.Name,
			DisplayName: resp.User.DisplayName,
		},
		PubKeyCredParams:      convertPubKeyCredParams(resp.PubKeyCredParams),
		AuthenticatorSelection: convertAuthenticatorSelection(resp.AuthenticatorSelection),
		Timeout:               resp.Timeout,
		Attestation:           resp.Attestation,
		ExcludeCredentials:    convertPublicKeyCredDescriptors(resp.ExcludeCredentials),
	}
}

func convertFromWebAuthnRegistrationResponse(resp webauthn.RegistrationResponse) WebAuthnRegistrationResponse {
	return WebAuthnRegistrationResponse{
		Success:      resp.Success,
		CredentialID: resp.CredentialID,
		Message:      resp.Message,
		UserID:       resp.UserID,
	}
}

func convertFromWebAuthnAssertionResponse(resp webauthn.AssertionChallengeResponse) WebAuthnAssertionChallengeResponse {
	return WebAuthnAssertionChallengeResponse{
		Challenge:        resp.Challenge,
		Timeout:          resp.Timeout,
		RelyingPartyID:   resp.RelyingPartyID,
		AllowCredentials: convertPublicKeyCredDescriptors(resp.AllowCredentials),
		UserVerification: resp.UserVerification,
	}
}

func convertFromWebAuthnAuthResponse(resp webauthn.AuthenticationResponse) WebAuthnAuthResponse {
	return WebAuthnAuthResponse{
		Success:   resp.Success,
		UserID:    resp.UserID,
		Message:   resp.Message,
		SessionID: resp.SessionID,
	}
}

// Helper conversion functions
func convertPubKeyCredParams(params []webauthn.PubKeyCredParam) []PubKeyCredParam {
	result := make([]PubKeyCredParam, len(params))
	for i, p := range params {
		result[i] = PubKeyCredParam{
			Type: p.Type,
			Alg:  p.Alg,
		}
	}
	return result
}

func convertAuthenticatorSelection(sel webauthn.AuthenticatorSelection) AuthenticatorSelection {
	return AuthenticatorSelection{
		AuthenticatorAttachment: sel.AuthenticatorAttachment,
		RequireResidentKey:      sel.RequireResidentKey,
		UserVerification:        sel.UserVerification,
	}
}

func convertPublicKeyCredDescriptors(descs []webauthn.PublicKeyCredDescriptor) []PublicKeyCredDescriptor {
	result := make([]PublicKeyCredDescriptor, len(descs))
	for i, d := range descs {
		result[i] = PublicKeyCredDescriptor{
			Type:       d.Type,
			ID:         d.ID,
			Transports: d.Transports,
		}
	}
	return result
}

func main() {
	// This function is required by Modus but can be empty
	// All functionality is exposed through exported functions
}

// TestSimpleFunction is a basic test function to check GraphQL discovery
func TestSimpleFunction(input string) (string, error) {
	return fmt.Sprintf("Hello, %s! GraphQL is working.", input), nil
}

// Session Management Functions

// CreateSession creates a secure session after successful OTP verification and authentication
func CreateSession(req SessionRequest) (SessionResponse, error) {
	ctx := context.Background()
	
	// Initialize ChronosSession agent
	chronos, err := chronossession.Initialize()
	if err != nil {
		return SessionResponse{}, fmt.Errorf("failed to initialize ChronosSession: %v", err)
	}
	
	// Create session request for ChronosSession agent
	sessionReq := &chronossession.SessionRequest{
		UserID:     req.UserID,
		DeviceInfo: fmt.Sprintf("ChannelDID: %s, Action: %s", req.ChannelDID, req.Action),
	}
	
	// Create session using ChronosSession agent
	sessionResp, err := chronos.IssueSession(ctx, sessionReq)
	if err != nil {
		return SessionResponse{}, fmt.Errorf("failed to create session: %v", err)
	}
	
	return SessionResponse{
		Success:     true,
		SessionID:   sessionResp.Token, // Use token as sessionID
		AccessToken: sessionResp.Token,
		ExpiresAt:   sessionResp.ExpiresAt.Unix(),
		Message:     sessionResp.Message,
		UserID:      sessionResp.UserID,
	}, nil
}

// ValidateSession validates an existing session token using ChronosSession
func ValidateSession(req ValidationRequest) (ValidationResponse, error) {
	ctx := context.Background()
	
	// Initialize ChronosSession agent
	chronos, err := chronossession.Initialize()
	if err != nil {
		return ValidationResponse{}, fmt.Errorf("failed to initialize ChronosSession: %v", err)
	}
	
	// Create validation request for ChronosSession agent
	validationReq := &chronossession.ValidationRequest{
		Token: req.Token,
	}
	
	// Validate session using ChronosSession agent
	validationResp, err := chronos.ValidateSession(ctx, validationReq)
	if err != nil {
		return ValidationResponse{}, fmt.Errorf("failed to validate session: %v", err)
	}
	
	return ValidationResponse{
		Valid:     validationResp.Valid,
		UserID:    validationResp.UserID,
		ExpiresAt: validationResp.ExpiresAt.Unix(),
		Message:   validationResp.Message,
	}, nil
}

// RefreshSession extends an existing session using ChronosSession
func RefreshSession(req RefreshRequest) (RefreshResponse, error) {
	ctx := context.Background()
	
	// Initialize ChronosSession agent
	chronos, err := chronossession.Initialize()
	if err != nil {
		return RefreshResponse{}, fmt.Errorf("failed to initialize ChronosSession: %v", err)
	}
	
	// Create refresh request for ChronosSession agent
	refreshReq := &chronossession.RefreshRequest{
		Token: req.Token,
	}
	
	// Refresh session using ChronosSession agent
	refreshResp, err := chronos.RefreshSession(ctx, refreshReq)
	if err != nil {
		return RefreshResponse{}, fmt.Errorf("failed to refresh session: %v", err)
	}
	
	return RefreshResponse{
		Token:     refreshResp.Token,
		ExpiresAt: refreshResp.ExpiresAt.Unix(),
		Message:   refreshResp.Message,
	}, nil
}

// RevokeSession revokes an existing session using ChronosSession
func RevokeSession(req RevocationRequest) (RevocationResponse, error) {
	ctx := context.Background()
	
	// Initialize ChronosSession agent
	chronos, err := chronossession.Initialize()
	if err != nil {
		return RevocationResponse{}, fmt.Errorf("failed to initialize ChronosSession: %v", err)
	}
	
	// Create revocation request for ChronosSession agent
	revocationReq := &chronossession.RevocationRequest{
		Token:  req.Token,
		Reason: req.Reason,
	}
	
	// Revoke session using ChronosSession agent
	revocationResp, err := chronos.RevokeSession(ctx, revocationReq)
	if err != nil {
		return RevocationResponse{}, fmt.Errorf("failed to revoke session: %v", err)
	}
	
	return RevocationResponse{
		Revoked:   revocationResp.Revoked,
		Message:   revocationResp.Message,
		Timestamp: revocationResp.Timestamp,
	}, nil
}
