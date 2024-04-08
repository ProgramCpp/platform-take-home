package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"
	"github.com/skip-mev/platform-take-home/logging"
	"github.com/skip-mev/platform-take-home/types"
	"go.uber.org/zap"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"

	sdk "github.com/cosmos/cosmos-sdk/types"

	vaultClient "github.com/skip-mev/platform-take-home/vault"
)

type APIServerImpl struct {
	types.UnimplementedAPIServer

	logger      *zap.Logger
	vaultClient *vault.Client
}

var _ types.APIServer = &APIServerImpl{}

func NewDefaultAPIServer(logger *zap.Logger, vaultClient *vault.Client) *APIServerImpl {
	return &APIServerImpl{
		logger:      logger,
		vaultClient: vaultClient,
	}
}

func (s *APIServerImpl) CreateWallet(ctx context.Context, request *types.CreateWalletRequest) (*types.CreateWalletResponse, error) {
	logging.FromContext(ctx).Info("CreateWallet", zap.String("name", request.Name))

	resp, err := s.vaultClient.Secrets.TransitCreateKey(ctx, request.Name,
		schema.TransitCreateKeyRequest{
			Type: "ecdsa-p256",
		}, vault.WithMountPath(vaultClient.VAULT_MOUNT_POINT))
	if err != nil {
		return nil, errors.Wrap(err, "error creating key")
	}

	publicKeyPEM := resp.Data["keys"].(map[string]interface{})["1"].(map[string]interface{})["public_key"].(string)

	block, _ := pem.Decode([]byte(publicKeyPEM))
	if err != nil {
		return nil, errors.Wrap(err, "error decoding public key PEM")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
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

	res := types.CreateWalletResponse{
		Wallet: &types.Wallet{
			Name: request.Name,
			Pubkey: publicKeyCompressed,
			AddressBytes: publicKeySha[:], 
			Address: address,
		},
	}
	fmt.Printf("%+v", res)
	return &res, nil
}

func (s *APIServerImpl) GetWallet(ctx context.Context, request *types.GetWalletRequest) (*types.GetWalletResponse, error) {
	// TODO: implement this
	logging.FromContext(ctx).Info("GetWallet", zap.String("name", request.Name))

	return &types.GetWalletResponse{
		Wallet: &types.Wallet{},
	}, nil
}

func (s *APIServerImpl) GetWallets(ctx context.Context, request *types.EmptyRequest) (*types.GetWalletsResponse, error) {
	// TODO: implement this
	return &types.GetWalletsResponse{
		Wallets: nil,
	}, nil
}

func (s *APIServerImpl) Sign(ctx context.Context, request *types.WalletSignatureRequest) (*types.WalletSignatureResponse, error) {
	// TODO: implement this
	return &types.WalletSignatureResponse{
		Signature: nil,
	}, nil
}
