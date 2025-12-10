package openstack

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

type backend struct {
	*framework.Backend
}

var _ logical.Factory = Factory

// Factory configures and returns OpenStack backends.
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	if conf == nil {
		return nil, errors.New("configuration passed into backend is nil")
	}

	b := &backend{}
	b.Backend = &framework.Backend{
		Help:        strings.TrimSpace(openstackHelp),
		BackendType: logical.TypeLogical,
		Paths: []*framework.Path{
			pathConfigAccess(b),
			pathConfigLease(b),
			pathListRoles(b),
			pathRoles(b),
			pathCreateCreds(b),
		},
		Secrets: []*framework.Secret{
			secretToken(b),
		},
	}

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
}

const openstackHelp = `
The OpenStack secrets backend generates application credentials for OpenStack.
`
