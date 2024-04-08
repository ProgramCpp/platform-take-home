package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"math/big"

	"github.com/pkg/errors"
	"github.com/skip-mev/platform-take-home/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// construct wallet from vault transit key
func getWallet(data map[string]interface{}) (*types.Wallet, error) {
	publicKeyPEM := data["keys"].(map[string]interface{})["1"].(map[string]interface{})["public_key"].(string)
	publicKeyBlock, _ := pem.Decode([]byte(publicKeyPEM))
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing public key")
	}

	ecdsaPublicKey := publicKey.(*ecdsa.PublicKey)
	publicKeyCompressed := elliptic.MarshalCompressed(elliptic.P256(), ecdsaPublicKey.X, ecdsaPublicKey.Y)
	publicKeySha := sha256.Sum256(publicKeyCompressed)
	address, err := sdk.Bech32ifyAddressBytes("cosmos", publicKeySha[:])
	if err != nil {
		return nil, errors.Wrap(err, "error creating bech32 address")
	}

	return &types.Wallet{
		Name:         data["name"].(string),
		Pubkey:       publicKeyCompressed,
		AddressBytes: publicKeySha[:],
		Address:      address,
	}, nil
}

// signatureRaw will serialize signature to R || S.
// R, S are padded to 32 bytes respectively.
// code roughly copied from secp256k1_nocgo.go
func signatureRaw(r, s *big.Int) []byte {
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	sigBytes := make([]byte, 64)
	// 0 pad the byte arrays from the left if they aren't big enough.
	copy(sigBytes[32-len(rBytes):32], rBytes)
	copy(sigBytes[64-len(sBytes):64], sBytes)
	return sigBytes
}

// TODO: move crypto utils to a separate common package
var p256Order = elliptic.P256().Params().N
var p256HalfOrder = new(big.Int).Rsh(p256Order, 1)

func IsSNormalized(sigS *big.Int) bool {
	return sigS.Cmp(p256HalfOrder) != 1
}

// NormalizeS will invert the s value if not already in the lower half
// of curve order value
func NormalizeS(sigS *big.Int) *big.Int {
	if IsSNormalized(sigS) {
		return sigS
	}

	return new(big.Int).Sub(p256Order, sigS)
}