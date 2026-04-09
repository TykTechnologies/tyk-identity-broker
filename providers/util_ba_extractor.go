package providers

import (
	b64 "encoding/base64"
	"net/http"
	"strings"
)

func ExtractBAUsernameAndPasswordFromRequest(r *http.Request) (string, string) {
	uName := ""
	pw := ""
	authHeader := r.Header.Get("Authorization")
	splitFields := strings.Split(authHeader, " ")
	if len(splitFields) == 2 {
		upEnc, decErr := b64.StdEncoding.DecodeString(splitFields[1])
		if decErr == nil {
			// split out again
			splitUP := strings.Split(string(upEnc), ":")
			if len(splitUP) == 2 {
				uName = splitUP[0]
				pw = splitUP[1]
			}
		}
	}

	return uName, pw
}
