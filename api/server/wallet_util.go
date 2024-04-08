package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"

	"github.com/pkg/errors"
	"github.com/skip-mev/platform-take-home/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// construct wallet from vault transit key
func GetWallet(data map[string]interface{})(*types.Wallet, error){
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