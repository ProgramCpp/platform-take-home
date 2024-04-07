package vault

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/pkg/errors"
)

func NewClient(ctx context.Context, addr string) (*vault.Client, error) {
	client, err := vault.New(
		vault.WithAddress(addr),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating vault client")
	}
	
	secret, err := vault.Unwrap[schema.KvV2ReadResponse](ctx, client, os.Getenv("WRAPPED_SECRET_ID"))
	if err != nil {
		return nil, errors.Wrap(err, "error unerapping vault token")
	}

	resp, err := client.Auth.AppRoleLogin(
		ctx,
		schema.AppRoleLoginRequest{
			RoleId:   os.Getenv("ROLE_ID"),
			SecretId: secret.Auth.ClientToken,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "error with app role login")
	}

	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		return nil, errors.Wrap(err, "error initializing client token")
	}

	return client, nil
}
