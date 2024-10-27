package jwe

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"github.com/go-jose/go-jose/v3"

	"github.com/markbates/goth/providers/openidConnect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateMockPrivateKey() (*tls.Certificate, error) {
	// Generate a new RSA private key for testing
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	cert := &tls.Certificate{
		PrivateKey: privKey,
		// Normally you would also populate the Certificate fields, but it's not needed for this test
	}
	return cert, nil
}

func createJWE(payload []byte, recipient *rsa.PublicKey) (string, error) {

	encrypter, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{
			Algorithm: jose.RSA_OAEP_256,
			Key:       recipient,
		},
		(&jose.EncrypterOptions{}).WithType("JWT"))
	if err != nil {
		return "", err
	}
	jwe, err := encrypter.Encrypt(payload)
	if err != nil {
		return "", err
	}
	return jwe.CompactSerialize()
}

// Test case for Handler.Decrypt
func TestHandler_Decrypt(t *testing.T) {
	// Generate a mock private key
	mockCert, err := generateMockPrivateKey()
	assert.NoError(t, err)

	// Create a valid JWE token for testing
	jweString, err := createJWE([]byte("test token"), mockCert.PrivateKey.(*rsa.PrivateKey).Public().(*rsa.PublicKey))
	assert.NoError(t, err)

	tests := []struct {
		name         string
		handler      *Handler
		token        string
		expected     string
		expectError  bool
		errorMessage string
	}{
		{
			name: "Disabled Handler",
			handler: &Handler{
				Enabled: false,
			},
			token:       jweString,
			expected:    jweString,
			expectError: false,
		},
		{
			name: "Key Not Loaded",
			handler: &Handler{
				Enabled: true,
				Key:     nil,
			},
			token:        jweString,
			expected:     "",
			expectError:  true,
			errorMessage: "JWE Private Key not loaded",
		},
		{
			name: "Successful Decryption",
			handler: &Handler{
				Enabled: true,
				Key:     mockCert,
			},
			token:       jweString,
			expected:    "test token",
			expectError: false,
		},
		{
			name: "Invalid Token",
			handler: &Handler{
				Enabled: true,
				Key:     mockCert,
			},
			token:        "invalid-token",
			expected:     "",
			expectError:  true,
			errorMessage: "error parsing JWE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := tt.handler.Decrypt(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, decrypted)
			}
		})
	}
}

func TestDecryptIDToken(t *testing.T) {
	mockCert, err := generateMockPrivateKey()
	assert.NoError(t, err)

	// Create a valid JWE token for testing
	jweString, err := createJWE([]byte("test token"), mockCert.PrivateKey.(*rsa.PrivateKey).Public().(*rsa.PublicKey))
	assert.NoError(t, err)

	// Setup a valid JWE handler
	jweHandler := &Handler{
		Enabled: true,
		Key:     mockCert,
	}

	tests := []struct {
		name          string
		jwtSession    *openidConnect.Session
		expectError   bool
		expectedToken string
	}{
		{
			name: "Successful Decryption",
			jwtSession: &openidConnect.Session{
				IDToken: jweString,
			},
			expectError:   false,
			expectedToken: "test token",
		},
		{
			name: "Invalid Token",
			jwtSession: &openidConnect.Session{
				IDToken: "invalid-token",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DecryptIDToken(jweHandler, tt.jwtSession)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, tt.jwtSession.IDToken)
			}
		})
	}
}
