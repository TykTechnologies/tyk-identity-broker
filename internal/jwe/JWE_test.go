package jwe

import (
	"crypto/rsa"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/markbates/goth/providers/openidConnect"

	"github.com/stretchr/testify/assert"
)

// Test case for Handler.Decrypt
func TestHandler_Decrypt(t *testing.T) {
	// Generate a mock private key
	mockCert, err := GenerateMockPrivateKey()
	assert.NoError(t, err)

	// Create a valid JWE token for testing
	jweString, err := CreateJWE([]byte("test token"), mockCert.PrivateKey.(*rsa.PrivateKey).Public().(*rsa.PublicKey))
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

// Test that every supported algorithm combination decrypts and unsupported ones are rejected
func TestHandler_Decrypt_Algorithms(t *testing.T) {
	mockCert, err := GenerateMockPrivateKey()
	assert.NoError(t, err)

	handler := &Handler{
		Enabled: true,
		Key:     mockCert,
	}
	publicKey := mockCert.PrivateKey.(*rsa.PrivateKey).Public()

	for _, keyAlg := range keyAlgorithms {
		for _, enc := range contentEncryptions {
			t.Run(string(keyAlg)+"/"+string(enc), func(t *testing.T) {
				encrypter, err := jose.NewEncrypter(enc, jose.Recipient{Algorithm: keyAlg, Key: publicKey}, nil)
				assert.NoError(t, err)
				object, err := encrypter.Encrypt([]byte("test token"))
				assert.NoError(t, err)
				token, err := object.CompactSerialize()
				assert.NoError(t, err)

				decrypted, err := handler.Decrypt(token)
				assert.NoError(t, err)
				assert.Equal(t, "test token", decrypted)
			})
		}
	}

	t.Run("Unsupported key algorithm", func(t *testing.T) {
		encrypter, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{Algorithm: jose.A128KW, Key: make([]byte, 16)}, nil)
		assert.NoError(t, err)
		object, err := encrypter.Encrypt([]byte("test token"))
		assert.NoError(t, err)
		token, err := object.CompactSerialize()
		assert.NoError(t, err)

		_, err = handler.Decrypt(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing JWE")
	})

	t.Run("Unsupported content encryption", func(t *testing.T) {
		token, err := CreateJWE([]byte("test token"), publicKey.(*rsa.PublicKey))
		assert.NoError(t, err)
		parts := strings.Split(token, ".")
		parts[0] = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RSA-OAEP-256","enc":"A999GCM"}`))

		_, err = handler.Decrypt(strings.Join(parts, "."))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing JWE")
	})
}

func TestDecryptIDToken(t *testing.T) {
	mockCert, err := GenerateMockPrivateKey()
	assert.NoError(t, err)

	// Create a valid JWE token for testing
	jweString, err := CreateJWE([]byte("test token"), mockCert.PrivateKey.(*rsa.PrivateKey).Public().(*rsa.PublicKey))
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
		jweHandler    *Handler
	}{
		{
			name: "Successful Decryption",
			jwtSession: &openidConnect.Session{
				IDToken: jweString,
			},
			expectError:   false,
			expectedToken: "test token",
			jweHandler:    jweHandler,
		},
		{
			name: "Invalid Token",
			jwtSession: &openidConnect.Session{
				IDToken: "invalid-token",
			},
			expectError: true,
			jweHandler:  jweHandler,
		},
		{
			name: "Failed Decryption due Key not loaded",
			jwtSession: &openidConnect.Session{
				IDToken: jweString,
			},
			expectError:   true,
			expectedToken: "test token",
			jweHandler: &Handler{
				Enabled: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DecryptIDToken(tt.jweHandler, tt.jwtSession)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, tt.jwtSession.IDToken)
			}
		})
	}
}
