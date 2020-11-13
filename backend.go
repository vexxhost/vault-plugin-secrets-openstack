package mock

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// backend wraps the backend framework and adds a map for storing key value pairs
type backend struct {
	*framework.Backend

	store map[string][]byte
}

var _ logical.Factory = Factory

// Factory configures and returns Mock backends
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
		Help:        strings.TrimSpace(mockHelp),
		BackendType: logical.TypeLogical,

		Paths: []*framework.Path{
			pathConfigAccess(b),
			pathConfigCred(b),
			pathConfigLease(b),
		},
		Secrets: []*framework.Secret{
			secretToken(b),
		},
	}

	return b, nil
}

func (b *backend) client(ctx context.Context, s logical.Storage) (*api.Client, error) {
	conf, err := b.readConfigAccess(ctx, s)
	if err != nil {
		return nil, err
	}

	nomadConf := api.DefaultConfig()
	if conf != nil {
		if conf.Address != "" {
			nomadConf.Address = conf.Address
		}
		if conf.Token != "" {
			nomadConf.SecretID = conf.Token
		}
		if conf.CACert != "" {
			nomadConf.TLSConfig.CACertPEM = []byte(conf.CACert)
		}
		if conf.ClientCert != "" {
			nomadConf.TLSConfig.ClientCertPEM = []byte(conf.ClientCert)
		}
		if conf.ClientKey != "" {
			nomadConf.TLSConfig.ClientKeyPEM = []byte(conf.ClientKey)
		}
	}

	client, err := api.NewClient(nomadConf)
	if err != nil {
		return nil, err
	}

	return client, nil
}

const mockHelp = `
The Mock backend is a dummy secrets backend that stores kv pairs in a map.
`
