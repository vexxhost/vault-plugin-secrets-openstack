package openstack

import (
	"context"
	"errors"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/applicationcredentials"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	SecretTokenType = "token"
)

func secretToken(b *backend) *framework.Secret {
	return &framework.Secret{
		Type: SecretTokenType,
		Fields: map[string]*framework.FieldSchema{
			"token": {
				Type:        framework.TypeString,
				Description: "Application credential token",
			},
		},
		Revoke: b.secretTokenRevoke,
	}
}

func (b *backend) secretTokenRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	cfg, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error reading access config: %w", err)
	}
	if cfg == nil {
		return nil, errors.New("access config not found")
	}

	identityClient, err := client(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating identity client: %w", err)
	}

	IDRaw, ok := req.Secret.InternalData["application_credential_id"]
	if !ok {
		return nil, fmt.Errorf("application_credential_id is missing on the lease")
	}
	id, ok := IDRaw.(string)
	if !ok {
		return nil, errors.New("unable to convert accessor_id")
	}

	if err := applicationcredentials.Delete(ctx, identityClient, cfg.UserID, id).ExtractErr(); err != nil {
		return nil, err
	}

	return nil, nil
}
