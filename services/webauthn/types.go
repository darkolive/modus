package webauthn

import "time"

// WebAuthn Challenge Types
type ChallengeRequest struct {
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

type ChallengeResponse struct {
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

// WebAuthn Registration Types
type RegistrationRequest struct {
	UserID                string                     `json:"userId"`
	Challenge             string                     `json:"challenge"`
	ClientDataJSON        string                     `json:"clientDataJSON"`
	AttestationObject     string                     `json:"attestationObject"`
	AuthenticatorResponse AuthenticatorAttestationResponse `json:"response"`
}

type AuthenticatorAttestationResponse struct {
	ClientDataJSON    string `json:"clientDataJSON"`
	AttestationObject string `json:"attestationObject"`
}

type RegistrationResponse struct {
	Success      bool   `json:"success"`
	CredentialID string `json:"credentialId"`
	Message      string `json:"message"`
	UserID       string `json:"userId"`
}

// WebAuthn Authentication Types
type AuthenticationRequest struct {
	UserID                string                        `json:"userId"`
	Challenge             string                        `json:"challenge"`
	ClientDataJSON        string                        `json:"clientDataJSON"`
	AuthenticatorData     string                        `json:"authenticatorData"`
	Signature             string                        `json:"signature"`
	UserHandle            string                        `json:"userHandle,omitempty"`
	AuthenticatorResponse AuthenticatorAssertionResponse `json:"response"`
}

type AuthenticatorAssertionResponse struct {
	ClientDataJSON    string `json:"clientDataJSON"`
	AuthenticatorData string `json:"authenticatorData"`
	Signature         string `json:"signature"`
	UserHandle        string `json:"userHandle,omitempty"`
}

type AuthenticationResponse struct {
	Success   bool   `json:"success"`
	UserID    string `json:"userId"`
	Message   string `json:"message"`
	SessionID string `json:"sessionId,omitempty"`
}

// WebAuthn Assertion Challenge Types
type AssertionChallengeRequest struct {
	UserID string `json:"userId,omitempty"`
}

type AssertionChallengeResponse struct {
	Challenge        string                    `json:"challenge"`
	Timeout          int                       `json:"timeout"`
	RelyingPartyID   string                    `json:"rpId"`
	AllowCredentials []PublicKeyCredDescriptor `json:"allowCredentials,omitempty"`
	UserVerification string                    `json:"userVerification"`
}

// Internal Types
type StoredChallenge struct {
	Challenge string    `json:"challenge"`
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	Type      string    `json:"type"` // "registration" or "authentication"
}

type WebAuthnCredential struct {
	UID          string    `json:"uid,omitempty"`
	UserID       string    `json:"userId"`
	CredentialID string    `json:"credentialId"`
	PublicKey    string    `json:"publicKey"`
	SignCount    int       `json:"signCount"`
	Transports   []string  `json:"transports"`
	AddedAt      time.Time `json:"addedAt"`
}

// Error Types
type WebAuthnError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e WebAuthnError) Error() string {
	return e.Message
}

// Constants
const (
	// Challenge expiry time (5 minutes)
	ChallengeExpiryMinutes = 5
	
	// Relying Party Information
	DefaultRPID   = "do-study.hypermode.host"
	DefaultRPName = "DO Study LMS"
	
	// Timeout (60 seconds)
	DefaultTimeout = 60000
	
	// User Verification
	UserVerificationRequired    = "required"
	UserVerificationPreferred   = "preferred"
	UserVerificationDiscouraged = "discouraged"
	
	// Attestation
	AttestationNone   = "none"
	AttestationDirect = "direct"
	
	// Authenticator Attachment
	AttachmentPlatform     = "platform"
	AttachmentCrossPlatform = "cross-platform"
	
	// Error Codes
	ErrorInvalidChallenge    = "INVALID_CHALLENGE"
	ErrorExpiredChallenge    = "EXPIRED_CHALLENGE"
	ErrorInvalidCredential   = "INVALID_CREDENTIAL"
	ErrorUserNotFound        = "USER_NOT_FOUND"
	ErrorRegistrationFailed  = "REGISTRATION_FAILED"
	ErrorAuthenticationFailed = "AUTHENTICATION_FAILED"
)
