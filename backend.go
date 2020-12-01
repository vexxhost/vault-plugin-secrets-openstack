package openstack

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// backend wraps the backend framework and adds a map for storing key value pairs
type backend struct {
	*framework.Backend

	store map[string][]byte
}

var _ logical.Factory = Factory

// Factory configures and returns openstack backends
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b, err := newBackend()
	if err != nil {
		return nil, err
	}

	if conf == nil {
		return nil, fmt.Errorf("configuration passed into backend is nil")
	}

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
}

func newBackend() (*backend, error) {
	b := &backend{
		store: make(map[string][]byte),
	}

	b.Backend = &framework.Backend{
		Help:        strings.TrimSpace(openstackHelp),
		BackendType: logical.TypeLogical,

		Paths: []*framework.Path{
			pathConfigAccess(b),
			pathCreateCreds(b),
			pathConfigLease(b),
			pathListRoles(b),
			pathRoles(b),
		},
		Secrets: []*framework.Secret{
			secretToken(b),
		},
	}

	return b, nil
}

const openstackHelp = `
The Openstack backend is a secrets backend that interacts with openstack api 
to create ApplicationCredentials.
`
