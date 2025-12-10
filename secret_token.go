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
	// Get the roleset name from the lease
	rolesetRaw, ok := req.Secret.InternalData["roleset"]
	if !ok {
		return nil, errors.New("roleset is missing on the lease")
	}
	rolesetName, ok := rolesetRaw.(string)
	if !ok {
		return nil, errors.New("unable to convert roleset name")
	}

	// Load the roleset
	role, err := b.Role(ctx, req.Storage, rolesetName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving roleset: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("roleset %q not found", rolesetName)
	}

	cfg, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error reading access config: %w", err)
	}
	if cfg == nil {
		return nil, errors.New("access config not found")
	}

	identityClient, err := client(ctx, cfg, role)
	if err != nil {
		return nil, fmt.Errorf("error creating identity client: %w", err)
	}

	IDRaw, ok := req.Secret.InternalData["application_credential_id"]
	if !ok {
		return nil, fmt.Errorf("application_credential_id is missing on the lease")
	}
	id, ok := IDRaw.(string)
	if !ok {
		return nil, errors.New("unable to convert application_credential_id")
	}

	if err := applicationcredentials.Delete(ctx, identityClient, cfg.UserID, id).ExtractErr(); err != nil {
		return nil, err
	}

	return nil, nil
}
