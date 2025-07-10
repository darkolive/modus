# Email Service

This service provides email functionality for the DO Study LMS platform.

## Architecture

The email service abstracts email provider integrations, allowing agents to send emails without being tightly coupled to specific email services.

## Files

- `mailersend.go` - MailerSend integration service

## Usage

Agents should use this service instead of calling email providers directly:

```go
import "modus/services/email"

// Send OTP email
response, err := email.SendOTPEmail(
    "user@example.com",
    "123456",
    "template-id"
)

// Send custom email
response, err := email.SendEmail(email.EmailRequest{
    To: "user@example.com",
    From: "noreply@darkolive.co.uk",
    Subject: "Your Subject",
    TemplateID: "template-id",
    Variables: map[string]string{
        "variable_name": "value",
    },
})
```

## Configuration

The service uses the MailerSend connection configured in `modus.json`.

## Benefits

1. **Separation of Concerns**: Agents focus on business logic, service handles email integration
2. **Reusability**: Multiple agents can use the same email service
3. **Maintainability**: Email provider changes only affect the service layer
4. **Testing**: Easier to mock and test email functionality
