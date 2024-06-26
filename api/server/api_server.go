package server

import (
	"context"
	"encoding/asn1"
	"encoding/base64"
	"math/big"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/skip-mev/platform-take-home/logging"
	"github.com/skip-mev/platform-take-home/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"

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

// TODO: log errors
func (s *APIServerImpl) CreateWallet(ctx context.Context, request *types.CreateWalletRequest) (*types.CreateWalletResponse, error) {
	logging.FromContext(ctx).Info("CreateWallet", zap.String("name", request.Name))

	// TODO: validate if key is already created
	// TODO: abstract all vault calls in the vault utils. promote vault utils to vault client proxy
	resp, err := s.vaultClient.Secrets.TransitCreateKey(ctx, request.GetName(),
		schema.TransitCreateKeyRequest{
			Type: "ecdsa-p256",
		}, vault.WithMountPath(vaultClient.VAULT_MOUNT_POINT))
	if err != nil {
		return nil, errors.Wrap(err, "error creating key")
	}

	wallet, err := getWallet(resp.Data)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing vault response")
	}
	// TODO: populate error
	return &types.CreateWalletResponse{Wallet: wallet}, nil
}

func (s *APIServerImpl) GetWallet(ctx context.Context, request *types.GetWalletRequest) (*types.GetWalletResponse, error) {
	logging.FromContext(ctx).Info("GetWallet", zap.String("name", request.Name))

	resp, err := s.vaultClient.Secrets.TransitReadKey(ctx, request.GetName(),
		vault.WithMountPath(vaultClient.VAULT_MOUNT_POINT))
	if err != nil {
		var resErr *vault.ResponseError
		if errors.As(err, &resErr) && resErr.StatusCode == http.StatusNotFound {
			return nil, status.Error(codes.NotFound, "wallet not found") // TODO: wrap error message for logging
		}
		return nil, errors.Wrap(err, "error getting key from vault")
	}

	wallet, err := getWallet(resp.Data)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing vault response")
	}

	// TODO: populate error
	return &types.GetWalletResponse{Wallet: wallet}, nil
}

func (s *APIServerImpl) GetWallets(ctx context.Context, request *types.EmptyRequest) (*types.GetWalletsResponse, error) {
	getWalletResponse := types.GetWalletsResponse{}
	resp, err := s.vaultClient.Secrets.TransitListKeys(ctx, vault.WithMountPath(vaultClient.VAULT_MOUNT_POINT))
	if err != nil {
		return nil, errors.Wrap(err, "error getting keys from vault")
	}
	for _, key := range resp.Data.Keys {
		resp, err := s.vaultClient.Secrets.TransitReadKey(ctx, key,
			vault.WithMountPath(vaultClient.VAULT_MOUNT_POINT))
		if err != nil {
			return nil, errors.Wrap(err, "error getting key from vault")
		}
		wallet, err := getWallet(resp.Data)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing vault response")
		}

		getWalletResponse.Wallets = append(getWalletResponse.Wallets, wallet)
	}

	return &getWalletResponse, nil
}

func (s *APIServerImpl) Sign(ctx context.Context, request *types.WalletSignatureRequest) (*types.WalletSignatureResponse, error) {
	resp, err := s.vaultClient.Secrets.TransitSign(ctx, request.GetWalletName(),
		schema.TransitSignRequest{
			Input:         base64.StdEncoding.EncodeToString(request.GetTxBytes()),
			HashAlgorithm: "sha2-256", // default value. not needed
		}, vault.WithMountPath(vaultClient.VAULT_MOUNT_POINT))
	if err != nil {
		return nil, errors.Wrap(err, "error creating key")
	}

	// "vault:v1:<base64-encoded-signature>
	vaultSig := resp.Data["signature"].(string)
	asn1Sig, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(vaultSig, "vault:v1:"))
	if err != nil {
		return nil, errors.Wrap(err, "error base64 decoding signature")
	}

	var ecdsaSig struct {
		R, S *big.Int
	}
	_, err = asn1.Unmarshal((asn1Sig), &ecdsaSig)
	if err != nil {
		return nil, errors.Wrap(err, "error asn1 decoding signature")
	}

	ecdsaSig.S = NormalizeS(ecdsaSig.S)

	return &types.WalletSignatureResponse{
		Signature: signatureRaw(ecdsaSig.R, ecdsaSig.S),
	}, nil
}
