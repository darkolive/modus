package hermesmailer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hypermodeinc/modus/sdk/go/pkg/http"
)

type HermesMailer struct {
	// No client needed - we'll use Modus HTTP connections
}

func NewHermesMailer(apiKey string) *HermesMailer {
	// API key is handled by modus.json connection
	return &HermesMailer{}
}

type SendTemplateRequest struct {
	FromName   string            `json:"fromName"`
	FromEmail  string            `json:"fromEmail"`
	ToName     string            `json:"toName"`
	ToEmail    string            `json:"toEmail"`
	Subject    string            `json:"subject"`
	TemplateID string            `json:"templateId"`
	Variables  map[string]string `json:"variables"`
	Tags       []string          `json:"tags,omitempty"`
}

type SendTemplateResponse struct {
	MessageID string `json:"messageId"`
}

func (h *HermesMailer) Send(ctx context.Context, req *SendTemplateRequest) (*SendTemplateResponse, error) {
	// Build MailerSend API request payload (matching their exact format)
	payload := map[string]interface{}{
		"from": map[string]string{
			"email": req.FromEmail,
		},
		"to": []map[string]string{
			{
				"email": req.ToEmail,
			},
		},
		"subject": req.Subject,
		"template_id": req.TemplateID,
	}
	
	// Add personalization variables if provided
	if len(req.Variables) > 0 {
		payload["personalization"] = []map[string]interface{}{
			{
				"email": req.ToEmail,
				"data":  req.Variables,
			},
		}
	}
	
	// Add tags if provided
	if len(req.Tags) > 0 {
		payload["tags"] = req.Tags
	}
	
	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	// Make HTTP request using Modus connection
	request := http.NewRequest("https://api.mailersend.com/v1/email/", &http.RequestOptions{
		Method: "POST",
		Body: payloadBytes,
	})
	
	resp, err := http.Fetch(request)
	if err != nil {
		return nil, fmt.Errorf("MailerSend API Error: %w", err)
	}
	
	if !resp.Ok() {
		responseText := resp.Text()
		return nil, fmt.Errorf("MailerSend API returned error: %d %s - %s", resp.Status, resp.StatusText, responseText)
	}
	
	// Handle response parsing - MailerSend may return empty body on success
	messageID := ""
	
	// Try to parse response body if it exists
	if len(resp.Body) > 0 {
		var response map[string]interface{}
		if err := json.Unmarshal(resp.Body, &response); err == nil {
			// Check for API errors in response
			if errors, ok := response["errors"]; ok {
				return nil, fmt.Errorf("MailerSend API errors: %v", errors)
			}
			
			// Extract message ID from response
			if data, ok := response["data"].(map[string]interface{}); ok {
				if id, ok := data["message_id"].(string); ok {
					messageID = id
				}
			}
		}
	}
	
	// If no message ID from body, generate a placeholder since email was sent successfully
	if messageID == "" {
		messageID = fmt.Sprintf("sent-%d", resp.Status)
	}
	
	return &SendTemplateResponse{MessageID: messageID}, nil
}