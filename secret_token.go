package openstack

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/errwrap"
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
			"token": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Request token",
			},
		},

		Renew:  b.secretTokenRenew,
		Revoke: b.secretTokenRevoke,
	}
}

func (b *backend) secretTokenRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	lease, err := b.LeaseConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if lease == nil {
		lease = &configLease{}
	}
	resp := &logical.Response{Secret: req.Secret}
	resp.Secret.TTL = lease.TTL
	return resp, nil
}

func (b *backend) secretTokenRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	c, err := b.client(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if c == nil {
		return nil, errwrap.Wrapf("error retrieving appCredentialClient: {{err}}", err)
	}

	IDRaw, ok := req.Secret.InternalData["application_credential_id"]
	if !ok {
		return nil, fmt.Errorf("application_credential_id is missing on the lease")
	}
	id, ok := IDRaw.(string)
	if !ok {
		return nil, errors.New("unable to convert accessor_id")
	}

	if err := c.Delete(id); err != nil {
		return nil, err
	}

	return nil, nil
}
