package jwe

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// using the keycloak one
var publicKeyPEM = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0Sk9ameCA5XL7Mjo17rL
A1NyZC5/QTa19qe4g8SHdwjNW0446HYIcNDlQekuQ3OX4avxcwkEABaTAp+tgPVj
+zWENcR8AaCemRaNPwkDxZep5g2Q8rRNeMCwGW4k53f6RNZN0lEnUpn/Qjg91Az3
PozcwBsm88rqLx8Z3G9NQPhrNQZshpXt38PU6d3fgTnZNPcdBS6bC0QLPk1vjVba
O30V029oY8xxtBkIRfa0k9fJkG0dLK1UkMiQR+t2yc5IBsBuTq59jE2af/Crzelj
zd+rhYD8/lhMG/nZ4nKuAEDquf7rP+74i1wj4VHB8ZXtPVKiWBAfxnlSZOlAf2zS
EwIDAQAB
-----END PUBLIC KEY-----
`

func getPublicKey() *rsa.PublicKey {
	// Parse the public key
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		fmt.Println("Failed to parse PEM block containing the public key")
		return nil
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Printf("Failed to parse public key: %v\n", err)
		return nil
	}

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		fmt.Println("Public key is not an RSA public key")
		return nil
	}
	return publicKey
}
