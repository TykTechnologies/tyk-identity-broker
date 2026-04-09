package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ReadNamesFromClaims(t *testing.T) {
	var testMatrix = []struct {
		rawData          map[string]interface{}
		ForeNameClaim    string
		SurNameClaim     string
		ForeNameExpected string
		SurNameExpected  string
	}{
		// read default claim
		{
			rawData: map[string]interface{}{
				DefaultForeNameClaim: "jhon",
				DefaultSurNameClaim:  "doe",
			},
			ForeNameClaim:    "",
			SurNameClaim:     "",
			ForeNameExpected: "jhon",
			SurNameExpected:  "doe",
		},
		// read custom claim
		{
			rawData: map[string]interface{}{
				"custom-forename-claim": "jhon",
				"custom-surname-claim":  "doe",
			},
			ForeNameClaim:    "custom-forename-claim",
			SurNameClaim:     "custom-surname-claim",
			ForeNameExpected: "jhon",
			SurNameExpected:  "doe",
		},
		// read custom claims that doesnt comes from idp...(bad mapping)
		{
			rawData:          map[string]interface{}{},
			ForeNameClaim:    "custom-forename-claim",
			SurNameClaim:     "custom-surname-claim",
			ForeNameExpected: "",
			SurNameExpected:  "",
		},
	}

	for _, ts := range testMatrix {
		forename, surname := ReadNamesFromClaims(ts.ForeNameClaim, ts.SurNameClaim, ts.rawData)
		assert.Equal(t, ts.ForeNameExpected, forename)
		assert.Equal(t, ts.SurNameExpected, surname)
	}
}

func Test_ReadEmailFromClaims(t *testing.T) {
	var testMatrix = []struct {
		rawData       map[string]interface{}
		emailClaim    string
		emailExpected string
	}{
		// read default claim
		{
			rawData: map[string]interface{}{
				DefaultEmailClaim: "jhon@doe.com",
			},
			emailExpected: "jhon@doe.com",
		},
		// read custom claim
		{
			rawData: map[string]interface{}{
				"custom-email-claim": "jhon@doe.com",
			},
			emailClaim:    "custom-email-claim",
			emailExpected: "jhon@doe.com",
		},
		// read custom claims that doesnt comes from idp...(bad mapping)
		{
			rawData:       map[string]interface{}{},
			emailClaim:    "custom-email-claim",
			emailExpected: "",
		},
		// WIF
		{
			rawData: map[string]interface{}{
				"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/": "",
				WIFUniqueName: "jhon@doe.com",
			},
			emailClaim:    "",
			emailExpected: "jhon@doe.com",
		},
	}

	for _, ts := range testMatrix {
		email := ReadEmailFromClaims(ts.emailClaim, ts.rawData)
		assert.Equal(t, ts.emailExpected, email)
	}
}
