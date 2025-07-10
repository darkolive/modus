# WebAuthn Registration Flow Test Documentation

## Overview

The `test_webauthn_registration.sh` script provides comprehensive end-to-end testing of the WebAuthn registration flow within the CerberusMFA authentication system.

## Test Flow Architecture

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│ CharonOTP   │───▶│ CerberusMFA  │───▶│ WebAuthn    │───▶│ Registration │
│ (OTP Verify)│    │ (User Check) │    │ (Challenge) │    │ (Complete)   │
└─────────────┘    └──────────────┘    └─────────────┘    └──────────────┘
```

## Test Steps

### Step 1: OTP Send
- **Endpoint**: `sendOTP`
- **Purpose**: Initiate OTP verification via email/phone
- **Input**: Channel (email/phone), recipient, userID
- **Output**: OTP ID, expiration time, delivery status

### Step 2: OTP Verification
- **Endpoint**: `verifyOTP`
- **Purpose**: Verify the OTP code and obtain channel DID
- **Input**: OTP ID, OTP code, userID
- **Output**: Channel DID, channel type, verification status

### Step 3: CerberusMFA Routing
- **Endpoint**: `cerberusMFA`
- **Purpose**: Check user existence and determine next action
- **Input**: Channel DID, channel type
- **Output**: User exists flag, action (signin/register), available methods

### Step 4: WebAuthn Registration Challenge
- **Endpoint**: `createWebAuthnRegistrationChallenge`
- **Purpose**: Generate WebAuthn registration challenge
- **Input**: User ID, username, display name
- **Output**: Challenge data, relying party info, credential parameters

### Step 5: WebAuthn Registration Verification
- **Endpoint**: `verifyWebAuthnRegistration`
- **Purpose**: Verify WebAuthn registration response
- **Input**: User ID, challenge, client data, attestation object
- **Output**: Success status, credential ID, user ID

## Usage

### Basic Usage
```bash
./test_webauthn_registration.sh
```

### Custom Configuration
```bash
# Custom email
./test_webauthn_registration.sh -e user@example.com

# Custom GraphQL endpoint
./test_webauthn_registration.sh -u http://localhost:8686/graphql

# Custom email and endpoint
./test_webauthn_registration.sh -e test@domain.com -u http://localhost:8686/graphql
```

### Command Line Options
- `-e, --email EMAIL`: Test email address
- `-p, --phone PHONE`: Test phone number  
- `-u, --endpoint URL`: GraphQL endpoint URL
- `-h, --help`: Show help message

## Prerequisites

### Server Requirements
1. **Modus Server Running**: The GraphQL server must be running
   ```bash
   cd /path/to/modus
   go run .
   ```

2. **Schema Deployed**: Ensure the latest schema is deployed
   ```bash
   cd /path/to/modus/db
   ./combine_schema.sh
   ./deploy_all_schemas.sh
   ```

### System Requirements
- `curl`: For making HTTP requests
- `grep`: For parsing JSON responses
- `bash`: Shell environment

## Expected Output

### Successful Test Run
```
🚀 Starting WebAuthn Registration Flow Test
==================================================
Test Configuration:
  📧 Email: test.user@example.com
  📱 Phone: +1234567890
  👤 User ID: test-user-1720564563
  🏷️  Username: testuser1720564563
  📝 Display Name: Test User
  🌐 GraphQL Endpoint: http://localhost:8686/graphql

🔧 Checking GraphQL endpoint availability...
✅ GraphQL endpoint is available

🔧 Step 1: Testing OTP Send (Email Channel)
✅ OTP sent successfully. OTP ID: otp_12345

🔧 Step 2: Testing OTP Verification (Simulated)
⚠️  OTP verification may have failed (expected for test OTP)
💡 In production, use the actual OTP received via email
💡 Using simulated Channel DID: did:channel:email:abc123...

🔧 Step 3: Testing CerberusMFA User Check
✅ CerberusMFA routing successful: New user registration required

🔧 Step 4: Testing WebAuthn Registration Challenge
✅ WebAuthn registration challenge created successfully
💡 Challenge: random_challenge_string

🔧 Step 5: Testing WebAuthn Registration Verification (Simulated)
⚠️  WebAuthn registration verification failed (expected with simulated data)
💡 In production, use actual attestation data from the browser

🎉 WebAuthn Registration Flow Test Completed!
```

## Test Limitations

### Simulated Components
1. **OTP Verification**: Uses test OTP code (123456) instead of real email/SMS
2. **WebAuthn Attestation**: Uses simulated attestation data instead of real authenticator response
3. **Channel DID**: Generated from email hash for testing

### Production Differences
1. **Real OTP Delivery**: Actual email/SMS delivery required
2. **Browser Integration**: Real WebAuthn API calls from browser
3. **Authenticator Interaction**: Physical authenticator (fingerprint, face, security key)

## Troubleshooting

### Common Issues

#### GraphQL Endpoint Not Available
```
❌ GraphQL endpoint is not available at http://localhost:8686/graphql
💡 Please ensure the Modus server is running with: go run .
```
**Solution**: Start the Modus server in the project directory

#### Schema Not Deployed
```
❌ Failed to send OTP
Response: {"errors":[{"message":"Unknown type OTPRequest"}]}
```
**Solution**: Deploy the latest schema
```bash
cd modus/db
./combine_schema.sh
./deploy_all_schemas.sh
```

#### Missing Dependencies
```
❌ curl is required but not installed
```
**Solution**: Install required tools
```bash
# macOS
brew install curl

# Ubuntu/Debian
sudo apt-get install curl
```

## Integration with Frontend

### JavaScript WebAuthn Integration
```javascript
// Step 1: Get registration challenge
const challengeResponse = await fetch('/graphql', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    query: `mutation CreateWebAuthnRegistrationChallenge($req: WebAuthnChallengeRequest!) {
      createWebAuthnRegistrationChallenge(req: $req) {
        challenge
        relyingParty { id name }
        user { id name displayName }
        pubKeyCredParams { type alg }
        authenticatorSelection {
          authenticatorAttachment
          requireResidentKey
          userVerification
        }
        timeout
        attestation
        excludeCredentials { type id transports }
      }
    }`,
    variables: { req: { userID, username, displayName } }
  })
});

// Step 2: Create credential with WebAuthn API
const credential = await navigator.credentials.create({
  publicKey: {
    challenge: base64ToArrayBuffer(challengeData.challenge),
    rp: challengeData.relyingParty,
    user: {
      id: stringToArrayBuffer(challengeData.user.id),
      name: challengeData.user.name,
      displayName: challengeData.user.displayName
    },
    pubKeyCredParams: challengeData.pubKeyCredParams,
    authenticatorSelection: challengeData.authenticatorSelection,
    timeout: challengeData.timeout,
    attestation: challengeData.attestation,
    excludeCredentials: challengeData.excludeCredentials
  }
});

// Step 3: Verify registration
const verificationResponse = await fetch('/graphql', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    query: `mutation VerifyWebAuthnRegistration($req: WebAuthnRegistrationRequest!) {
      verifyWebAuthnRegistration(req: $req) {
        success
        credentialID
        message
        userID
      }
    }`,
    variables: {
      req: {
        userID,
        challenge: challengeData.challenge,
        clientDataJSON: arrayBufferToBase64(credential.response.clientDataJSON),
        attestationObject: arrayBufferToBase64(credential.response.attestationObject)
      }
    }
  })
});
```

## Security Considerations

### Test Environment Only
- **Simulated Data**: Test uses fake attestation data
- **Test OTP**: Uses predictable OTP code (123456)
- **No Real Authentication**: Does not provide actual security

### Production Requirements
- **Real Authenticators**: Physical security keys, biometrics
- **HTTPS Required**: WebAuthn requires secure context
- **Origin Validation**: Proper relying party configuration
- **Challenge Uniqueness**: Cryptographically secure challenges

## Next Steps

1. **Frontend Integration**: Implement WebAuthn JavaScript in your frontend
2. **Real OTP Testing**: Configure email/SMS providers for actual OTP delivery
3. **Authenticator Testing**: Test with physical security keys and biometric authenticators
4. **Error Handling**: Add comprehensive error handling and user feedback
5. **Production Hardening**: Implement full W3C WebAuthn specification compliance

## Related Documentation

- [CerberusMFA Agent Documentation](./agents/auth/CerberusMFA/readme.md)
- [WebAuthn Service Implementation](./services/webauthn/)
- [Schema Documentation](./db/schema/)
- [GraphQL API Reference](./main.go)
