package vault_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublicKey(t *testing.T) {
	// ECDSA token from vault
	// https://pkg.go.dev/crypto/x509#example-ParsePKIXPublicKey
	pubPEM := `
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEyb1H7FYoR71Cy2k/AOVt+rYfFZZd
+t8UN0Ont/frXNQxZMEADGTwD6ZUsk7kKwtsXKewBd5YYIhV4CpU66jXUA==
-----END PUBLIC KEY-----`
	block, _ := pem.Decode([]byte(pubPEM))
	assert.NotNil(t, block)

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	assert.NoError(t, err)

	publicKey := pub.(*ecdsa.PublicKey)
	publicKeyCompressed := elliptic.MarshalCompressed(elliptic.P256(), publicKey.X, publicKey.Y)
	assert.NotEmpty(t, publicKeyCompressed)
}
