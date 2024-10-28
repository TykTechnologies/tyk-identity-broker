package tothic

import (
	"crypto/rsa"
	"os"
	"reflect"
	"testing"

	"github.com/TykTechnologies/tyk-identity-broker/internal/jwe"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/openidConnect"
	"github.com/stretchr/testify/assert"
)

func TestKeyFromEnv(t *testing.T) {
	expected := "SECRET"

	t.Run("with new variable", func(t *testing.T) {
		_ = os.Setenv("TYK_IB_SESSION_SECRET", expected)
		_ = os.Setenv("SESSION_SECRET", "")
		assertDeepEqual(t, expected, KeyFromEnv())
	})

	t.Run("with deprecated", func(t *testing.T) {
		_ = os.Setenv("TYK_IB_SESSION_SECRET", "")
		_ = os.Setenv("SESSION_SECRET", expected)
		assertDeepEqual(t, expected, KeyFromEnv())
	})

	t.Run("with both", func(t *testing.T) {
		_ = os.Setenv("SESSION_SECRET", "SOME_OTHER_SECRET")
		_ = os.Setenv("TYK_IB_SESSION_SECRET", expected)
		assertDeepEqual(t, expected, KeyFromEnv())
	})
}

func TestProcessJWTSession_Success(t *testing.T) {

	mockCert, err := jwe.GenerateMockPrivateKey()
	assert.NoError(t, err)

	IDTokenContents := "test-id-token"

	// Create a valid JWE token for testing
	jweString, err := jwe.CreateJWE([]byte(IDTokenContents), mockCert.PrivateKey.(*rsa.PrivateKey).Public().(*rsa.PublicKey))
	assert.NoError(t, err)

	tcs := []struct {
		name         string
		sess         goth.Session
		jweHandler   jwe.Handler
		expectedSess goth.Session
		errExpected  bool
	}{
		{
			name: "no id token encryption",
			sess: &openidConnect.Session{
				IDToken: "anything",
			},
			expectedSess: &openidConnect.Session{
				IDToken: "anything",
			},
			jweHandler:  jwe.Handler{Enabled: false},
			errExpected: false,
		},
		{
			name: "failed decryption, no key present",
			sess: &openidConnect.Session{
				IDToken: "any-encrypted-val",
			},
			expectedSess: &openidConnect.Session{
				IDToken: "any-encrypted-val",
			},
			jweHandler:  jwe.Handler{Enabled: true},
			errExpected: true,
		},
		{
			name: "successful decryption",
			sess: &openidConnect.Session{
				IDToken: jweString,
			},
			jweHandler: jwe.Handler{
				Enabled: true,
				Key:     mockCert,
			},
			expectedSess: &openidConnect.Session{
				IDToken: IDTokenContents,
			},
			errExpected: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function
			sessResult, err := prepareJWTSession(tc.sess, &tc.jweHandler)
			// Assert results
			didErr := err != nil
			assert.Equal(t, tc.errExpected, didErr)
			if !tc.errExpected {
				assert.Equal(t, tc.expectedSess, sessResult)
			}
		})
	}

}

func assertDeepEqual(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, actual %v", expected, actual)
	}
}
