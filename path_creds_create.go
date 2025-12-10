package openstack

import (
	"context"
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/applicationcredentials"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathCreateCreds(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the role set",
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

	cfg, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error reading access config: %w", err)
	}
	if cfg == nil {
		return logical.ErrorResponse("access config not found"), nil
	}

	identityClient, err := client(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating identity client: %w", err)
	}

	// Create application credential
	tokenName := fmt.Sprintf("vault-%s-%s-%d", name, req.DisplayName, time.Now().UnixMilli())
	expireTime := time.Now().Add(leaseConfig.TTL)
	credential, err := applicationcredentials.Create(ctx, identityClient, cfg.UserID, applicationcredentials.CreateOpts{
		Name:        tokenName,
		Description: fmt.Sprintf("Created by Vault at %s", time.Now().Format(time.RFC3339)),
		Roles:       role.Roles,
		ExpiresAt:   &expireTime,
	}).Extract()
	if err != nil {
		b.Logger().Warn("Create applicationcredential", "error", err)
		return nil, err
	}
	// Use the helper to create the secret
	resp := b.Secret(SecretTokenType).Response(map[string]interface{}{
		"application_credential_id":     credential.ID,
		"application_credential_secret": credential.Secret,
	}, map[string]interface{}{
		"application_credential_id": credential.ID,
	})
	resp.Secret.TTL = leaseConfig.TTL

	return resp, nil
}
