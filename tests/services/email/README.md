# Email Service Tests

This directory contains integration tests for the DO Study email service using MailerSend.

## Files

- `email_test.go` - Comprehensive test suite for email service functionality
- `run_email_tests.sh` - Test runner script with benchmarks
- `README.md` - This documentation

## Test Coverage

### Integration Tests
- **SendOTPEmail with DefaultTemplate** - Tests OTP email sending using the configured default template
- **SendWelcomeEmail with DefaultTemplate** - Tests welcome email sending using the configured default template
- **Template Configuration** - Verifies that template IDs are properly configured

### Unit Tests
- **Basic Functionality** - Smoke tests for email service functions
- **Error Handling** - Tests proper error handling for invalid inputs

### Benchmarks
- **Performance Testing** - Benchmarks email sending performance

## Setup

1. **Set your MailerSend API Key:**
   ```bash
   export MAILERSEND_API_KEY=your_api_key_here
   ```

2. **Update test email address** in `email_test.go`:
   ```go
   testEmail := "your-test-email@domain.com"
   ```

## Running Tests

### Option 1: Use the test runner script
```bash
./run_email_tests.sh
```

### Option 2: Run tests manually
```bash
# Run all tests
go test -v ./...

# Run specific test
go test -v -run TestEmailServiceIntegration

# Run benchmarks
go test -bench=. -benchmem ./...
```

## Template Configuration

The tests verify that the email service uses the configured template IDs:

- **OTPTemplateID**: `neqvygm91v8l0p7w` (OTP emails)
- **WelcomeTemplateID**: `3234678` (Welcome emails)  
- **DefaultTemplateID**: `vywj2lpz701g7oqz` (Fallback template)

## Expected Behavior

- If no `MAILERSEND_API_KEY` is set, integration tests will be skipped
- Tests verify successful email sending with proper response structure
- Template fallback logic is tested to ensure proper configuration
- Error handling is verified for invalid inputs

## Troubleshooting

- **Tests skipped**: Set `MAILERSEND_API_KEY` environment variable
- **API errors**: Check your MailerSend API key and account status
- **Template errors**: Verify template IDs exist in your MailerSend account
- **Import errors**: Ensure you're running tests from the correct directory

## Notes

- Tests use real MailerSend API calls when API key is provided
- Update test email addresses to avoid sending to invalid addresses
- Template IDs should match your actual MailerSend templates
- Tests include both success and failure scenarios
