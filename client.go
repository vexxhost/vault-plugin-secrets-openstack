package openstack

import (
	"context"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/vexxhost/vault-plugin-secrets-openstack/utils"
)

type appCredentialClient struct {
	serviceClient *gophercloud.ServiceClient
	userID        string
}

func (b *backend) client(ctx context.Context, s logical.Storage) (*appCredentialClient, error) {
	authConfig, err := b.readConfigAccess(ctx, s)
	if err != nil {
		b.Logger().Warn("get access config", "error", err)
		return nil, err
	}
	b.Logger().Warn("Creds", "start!!", err)
	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint:            authConfig.IdentityEndpoint,
		UserID:                      authConfig.UserID,
		Username:                    authConfig.Username,
		Password:                    authConfig.Password,
		TenantID:                    authConfig.TenantID,
		TenantName:                  authConfig.TenantName,
		DomainID:                    authConfig.DomainID,
		DomainName:                  authConfig.DomainName,
		ApplicationCredentialID:     authConfig.ApplicationCredentialID,
		ApplicationCredentialName:   authConfig.ApplicationCredentialName,
		ApplicationCredentialSecret: authConfig.ApplicationCredentialSecret,
	}
	regionName := authConfig.Region

	// Get the service client
	serviceClient, err := utils.OpenstackClient(authOpts, regionName)
	if err != nil {
		b.Logger().Warn("get openstackclient", "error", err)
		return nil, err
	}

	return &appCredentialClient{
		serviceClient: serviceClient,
		userID:        authOpts.UserID,
	}, nil
}

func (c *appCredentialClient) Create(name string, roles []Role, accessrules []AccessRule) (string, string, error) {
	return utils.CreateApplicationCredential(c.serviceClient, c.userID, name, roles, accessrules)
}

func (c *appCredentialClient) Delete(id string) error {
	return utils.DeleteApplicationCredential(c.serviceClient, c.userID, id)
}
