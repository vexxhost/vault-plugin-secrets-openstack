package mock

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// maxTokenNameLength is the maximum length for the name of a Nomad access
// token
const maxTokenNameLength = 256

func pathCredsCreate(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the credential",
			},
			"region": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the region",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathTokenRead,
		},
	}
}

func (b *backend) pathTokenRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)
	region_name := d.Get("region").(string)
	conf, _ := b.readConfigAccess(ctx, req.Storage)

	// Determine if we have a lease configuration
	leaseConfig, err := b.LeaseConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if leaseConfig == nil {
		leaseConfig = &configLease{}
	}

	// Get the service client
	serviceClient, err := b.openstackClient(conf, region_name)
	if err != nil {
		return nil, err
	}

	// Create it
	id, secret, err := b.CreateApplicationCredential(serviceClient, conf.UserID, name)
	if err != nil {
		return nil, err
	}

	// Use the helper to create the secret
	resp := b.Secret(SecretTokenType).Response(map[string]interface{}{
		"application_credential_id":     id,
		"application_credential_secret": secret,
	}, map[string]interface{}{
		"application_credential_secret": secret,
	})
	resp.Secret.TTL = leaseConfig.TTL
	resp.Secret.MaxTTL = leaseConfig.MaxTTL

	return resp, nil
}
