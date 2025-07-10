#!/bin/bash

# Minimal Email Service Test Script
# Tests modus/services/email/email.go by sending real emails

set -e

# Configuration
RECIPIENT="darren@darkolive.co.uk"
MODUS_DIR="/Users/darrenknipe/Websites/DO Study/modus"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🧪 Testing Email Service${NC}"
echo "Recipient: $RECIPIENT"
echo ""

cd "$MODUS_DIR"

# Test 1: SendOTPEmail
echo -e "${BLUE}📧 Testing SendOTPEmail...${NC}"
cat > /tmp/test_otp.go << EOF
package main

import (
	"fmt"
	"modus/services/email"
)

func main() {
	response, err := email.SendOTPEmail("$RECIPIENT", "123456")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	if response.Success {
		fmt.Printf("SUCCESS: OTP email sent - MessageID: %s\n", response.MessageID)
	} else {
		fmt.Printf("FAILED: %s\n", response.Error)
	}
}
EOF

if go run /tmp/test_otp.go; then
	echo -e "${GREEN}✅ SendOTPEmail test passed${NC}"
	OTP_SUCCESS=true
else
	echo -e "${RED}❌ SendOTPEmail test failed${NC}"
	OTP_SUCCESS=false
fi
rm -f /tmp/test_otp.go
echo ""

# Test 2: SendWelcomeEmail
echo -e "${BLUE}📧 Testing SendWelcomeEmail...${NC}"
cat > /tmp/test_welcome.go << EOF
package main

import (
	"fmt"
	"modus/services/email"
)

func main() {
	response, err := email.SendWelcomeEmail("$RECIPIENT", "Test User")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	if response.Success {
		fmt.Printf("SUCCESS: Welcome email sent - MessageID: %s\n", response.MessageID)
	} else {
		fmt.Printf("FAILED: %s\n", response.Error)
	}
}
EOF

if go run /tmp/test_welcome.go; then
	echo -e "${GREEN}✅ SendWelcomeEmail test passed${NC}"
	WELCOME_SUCCESS=true
else
	echo -e "${RED}❌ SendWelcomeEmail test failed${NC}"
	WELCOME_SUCCESS=false
fi
rm -f /tmp/test_welcome.go
echo ""

# Summary
echo -e "${BLUE}📊 Test Summary${NC}"
if [ "$OTP_SUCCESS" = true ] && [ "$WELCOME_SUCCESS" = true ]; then
	echo -e "${GREEN}🎉 All email service tests passed!${NC}"
	echo "Check your inbox at $RECIPIENT for test emails"
	exit 0
else
	echo -e "${RED}❌ Some email service tests failed${NC}"
	exit 1
fi
