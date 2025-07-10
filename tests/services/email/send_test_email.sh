#!/bin/bash

# Simple test email sender using modus/services/email/email.go
# Sends a test email to verify the email service is working

set -e

# Configuration
RECIPIENT="darren@darkolive.co.uk"
PROJECT_ROOT="/Users/darrenknipe/Websites/DO Study"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}📧 Sending Test Email${NC}"
echo "Recipient: $RECIPIENT"
echo ""

# Create temporary Go program to send test email
cat > /tmp/send_test_email.go << 'EOF'
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Add the project root to Go path for imports
	projectRoot := os.Args[1]
	recipient := os.Args[2]
	
	// Change to project directory
	if err := os.Chdir(projectRoot); err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}
	
	// Import the email service (this will be done via go run)
	fmt.Printf("Sending test email to: %s\n", recipient)
	fmt.Printf("Using project root: %s\n", projectRoot)
}
EOF

# Create the actual email sending program in the modus directory
cat > /tmp/email_sender.go << EOF
package main

import (
	"fmt"
	"log"
	"os"
	
	"modus/services/email"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run email_sender.go <recipient>")
	}
	
	recipient := os.Args[1]
	
	fmt.Printf("🔧 Email service ready (using %s)\n", email.GetProviderInfo())
	
	fmt.Printf("📧 Sending test OTP email to: %s\n", recipient)
	
	// Send OTP email using default template
	otpCode := "123456" // Test OTP code
	response, err := email.SendOTPEmail(recipient, otpCode)
	if err != nil {
		log.Fatalf("❌ Failed to send OTP email: %v", err)
	}
	
	fmt.Printf("✅ OTP Email sent successfully!\n")
	fmt.Printf("   Message ID: %s\n", response.MessageID)
	fmt.Printf("   Success: %t\n", response.Success)
	fmt.Printf("   Message: %s\n", response.Message)
	
	fmt.Printf("\n📧 Sending test Welcome email to: %s\n", recipient)
	
	// Send Welcome email using default template
	userName := "Test User"
	response2, err := email.SendWelcomeEmail(recipient, userName)
	if err != nil {
		log.Fatalf("❌ Failed to send Welcome email: %v", err)
	}
	
	fmt.Printf("✅ Welcome Email sent successfully!\n")
	fmt.Printf("   Message ID: %s\n", response2.MessageID)
	fmt.Printf("   Success: %t\n", response2.Success)
	fmt.Printf("   Message: %s\n", response2.Message)
	
	fmt.Printf("\n🎉 All test emails sent successfully!\n")
	fmt.Printf("📬 Check your inbox at %s for the test emails\n", recipient)
}
EOF

echo -e "${BLUE}🚀 Running email test...${NC}"

# Run the email sender from modus directory
cd "$PROJECT_ROOT/modus"
if go run /tmp/email_sender.go "$RECIPIENT"; then
	echo ""
	echo -e "${GREEN}✅ Test email sending completed successfully!${NC}"
	echo -e "${BLUE}📬 Check your inbox at $RECIPIENT${NC}"
else
	echo ""
	echo -e "${RED}❌ Test email sending failed${NC}"
	exit 1
fi

# Cleanup
rm -f /tmp/send_test_email.go /tmp/email_sender.go

echo ""
echo -e "${GREEN}🎉 Email service test completed!${NC}"
