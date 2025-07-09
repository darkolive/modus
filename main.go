package main

import (
	"context"
	"time"

	charonotp "modus/agents/auth/CharonOTP"
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
	OTPID     string    `json:"otpId"`
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

// Convert main package types to charonotp package types
func convertToCharonOTPRequest(req OTPRequest) charonotp.OTPRequest {
	return charonotp.OTPRequest{
		Channel:   req.Channel, // Now using string instead of enum
		Recipient: req.Recipient,
		UserID:    "", // Empty userID since not required for this agent
	}
}

func convertFromCharonOTPResponse(resp charonotp.OTPResponse) OTPResponse {
	return OTPResponse{
		OTPID:     resp.OTPID,
		Sent:      resp.Sent,
		Channel:   string(resp.Channel),
		ExpiresAt: resp.ExpiresAt,
		Message:   resp.Message,
	}
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

// SendOTP is the exported wrapper function for Modus
func SendOTP(req OTPRequest) (OTPResponse, error) {
	// Create context for internal use
	ctx := context.Background()
	charonReq := convertToCharonOTPRequest(req)
	charonResp, err := charonotp.SendOTP(ctx, charonReq)
	if err != nil {
		return OTPResponse{}, err
	}
	return convertFromCharonOTPResponse(charonResp), nil
}

// VerifyOTP is the exported wrapper function for Modus
func VerifyOTP(req VerifyOTPRequest) (VerifyOTPResponse, error) {
	// Convert to internal type and call the implementation
	charonReq := convertToCharonVerifyRequest(req)
	resp, err := charonotp.VerifyOTP(charonReq)
	if err != nil {
		return VerifyOTPResponse{}, err
	}
	
	// Convert response back to main package type
	return convertFromCharonVerifyResponse(resp), nil
}

func main() {
	// Explorer will use the exported functions
	// No actual implementation needed here
}
