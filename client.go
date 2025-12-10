package openstack

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/config"
)

func client(ctx context.Context, cfg *Config, role *RoleSet) (*gophercloud.ServiceClient, error) {
	authOpts := cfg.AuthOptions(role.ProjectID, role.ProjectName)

	// Build TLS config from stored configuration
	if (cfg.Cert != "" && cfg.Key == "") || (cfg.Cert == "" && cfg.Key != "") {
		return nil, errors.New("either both cert and key or none must be provided")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.Insecure,
	}

	if cfg.Cert != "" {
		cert, err := tls.X509KeyPair([]byte(cfg.Cert), []byte(cfg.Key))
		if err != nil {
			return nil, fmt.Errorf("parse TLS cert: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if cfg.CACert != "" {
		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM([]byte(cfg.CACert)) {
			return nil, errors.New("failed to parse CA certificates")
		}
		tlsConfig.RootCAs = roots
	}

	providerClient, err := config.NewProviderClient(ctx, *authOpts, config.WithTLSConfig(tlsConfig))
	if err != nil {
		return nil, err
	}

	identityClient, err := openstack.NewIdentityV3(providerClient, gophercloud.EndpointOpts{Region: cfg.RegionName})
	if err != nil {
		return nil, err
	}

	return identityClient, nil
}
