package openstack

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathCreateCreds(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the credential",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathTokenRead,
		},
	}
}

func (b *backend) pathTokenRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)

	// Determine if we have a lease configuration
	leaseConfig, err := b.LeaseConfig(ctx, req.Storage)
	if err != nil {
		b.Logger().Warn("get leaseconfig", "error", err)
		return nil, err
	}
	if leaseConfig == nil {
		leaseConfig = &configLease{}
	}

	role, err := b.Role(ctx, req.Storage, name)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role: %w", err)
	}
	if role == nil {
		return logical.ErrorResponse(fmt.Sprintf("role %q not found", name)), nil
	}

	c, err := b.client(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error retrieving appCredentialClient: %w", err)
	}

	// Create it
	tokenName := fmt.Sprintf("vault-%s-%s-%d", name, req.DisplayName, time.Now().UnixNano()/int64(time.Millisecond))
	id, secret, err := c.Create(tokenName, role.Roles, leaseConfig.TTL) //, role.AccessRules)
	if err != nil {
		b.Logger().Warn("Create applicationcredential", "error", err)
		return nil, err
	}
	// Use the helper to create the secret
	resp := b.Secret(SecretTokenType).Response(map[string]interface{}{
		"application_credential_id":     id,
		"application_credential_secret": secret,
	}, map[string]interface{}{
		"application_credential_id": id,
	})
	resp.Secret.TTL = leaseConfig.TTL

	return resp, nil
}
