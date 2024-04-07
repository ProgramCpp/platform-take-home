package server

import (
	"context"

	"github.com/pkg/errors"
	"github.com/skip-mev/platform-take-home/logging"
	"github.com/skip-mev/platform-take-home/types"
	"go.uber.org/zap"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

type APIServerImpl struct {
	types.UnimplementedAPIServer

	logger      *zap.Logger
	vaultClient *vault.Client
}

var _ types.APIServer = &APIServerImpl{}

const VAULT_MOUNT_POINT = "private_keys"

func NewDefaultAPIServer(logger *zap.Logger, vaultClient *vault.Client) *APIServerImpl {
	return &APIServerImpl{
		logger: logger,
		vaultClient: vaultClient,
	}
}

func (s *APIServerImpl) CreateWallet(ctx context.Context, request *types.CreateWalletRequest) (*types.CreateWalletResponse, error) {
	// TODO: implement this
	logging.FromContext(ctx).Info("CreateWallet", zap.String("name", request.Name))

	_, err := s.vaultClient.Secrets.TransitCreateKey(ctx, request.Name, schema.TransitCreateKeyRequest{
		Type: "ecdsa-p256",
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating key")
	}

	//response.Data

	return &types.CreateWalletResponse{
		Wallet: &types.Wallet{},
	}, nil
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
