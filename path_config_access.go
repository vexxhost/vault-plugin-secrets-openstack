package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const configAccessKey = "config/auth"

func pathConfigAccess(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: configAccessKey,
		Fields: map[string]*framework.FieldSchema{
			"auth_url": {
				Type:        framework.TypeString,
				Description: "OpenStack authentication URL",
			},
			"user_id": {
				Type:        framework.TypeString,
				Description: "User ID for authentication",
			},
			"username": {
				Type:        framework.TypeString,
				Description: "Username for authentication",
			},
			"password": {
				Type:        framework.TypeString,
				Description: "Password for authentication",
			},
			"user_domain_id": {
				Type:        framework.TypeString,
				Description: "Domain ID for user authentication",
			},
			"user_domain_name": {
				Type:        framework.TypeString,
				Description: "Domain name for user authentication",
			},
			"application_credential_id": {
				Type:        framework.TypeString,
				Description: "Application credential ID for authentication",
			},
			"application_credential_name": {
				Type:        framework.TypeString,
				Description: "Application credential name for authentication",
			},
			"application_credential_secret": {
				Type:        framework.TypeString,
				Description: "Application credential secret for authentication",
			},
			"region_name": {
				Type:        framework.TypeString,
				Description: "Region name for endpoint selection",
			},
			"cacert": {
				Type:        framework.TypeString,
				Description: "PEM-encoded CA certificate for TLS verification",
			},
			"cert": {
				Type:        framework.TypeString,
				Description: "PEM-encoded client certificate for mutual TLS",
			},
			"key": {
				Type:        framework.TypeString,
				Description: "PEM-encoded client key for mutual TLS",
			},
			"insecure": {
				Type:        framework.TypeBool,
				Description: "Skip TLS verification (not recommended for production)",
				Default:     false,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathConfigAccessRead,
			logical.CreateOperation: b.pathConfigAccessWrite,
			logical.UpdateOperation: b.pathConfigAccessWrite,
			logical.DeleteOperation: b.pathConfigAccessDelete,
		},
		ExistenceCheck: b.configExistenceCheck,
	}
}

func (b *backend) configExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	entry, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return false, err
	}

	return entry != nil, nil
}

func (b *backend) readConfigAccess(ctx context.Context, storage logical.Storage) (*Config, error) {
	entry, err := storage.Get(ctx, configAccessKey)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	conf := &Config{}
	if err := entry.DecodeJSON(conf); err != nil {
		return nil, fmt.Errorf("error reading OpenStack access configuration: %w", err)
	}

	return conf, nil
}

func (b *backend) pathConfigAccessRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	conf, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"auth_url":                    conf.AuthURL,
			"user_id":                     conf.UserID,
			"username":                    conf.Username,
			"user_domain_id":              conf.UserDomainID,
			"user_domain_name":            conf.UserDomainName,
			"application_credential_id":   conf.ApplicationCredentialID,
			"application_credential_name": conf.ApplicationCredentialName,
			"region_name":                 conf.RegionName,
			"cacert":                      conf.CACert,
			"cert":                        conf.Cert,
			"insecure":                    conf.Insecure,
		},
	}, nil
}

func (b *backend) pathConfigAccessWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	conf, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &Config{}
	}

	if authURL, ok := data.GetOk("auth_url"); ok {
		conf.AuthURL = authURL.(string)
	}
	if userID, ok := data.GetOk("user_id"); ok {
		conf.UserID = userID.(string)
	}
	if username, ok := data.GetOk("username"); ok {
		conf.Username = username.(string)
	}
	if password, ok := data.GetOk("password"); ok {
		conf.Password = password.(string)
	}
	if userDomainID, ok := data.GetOk("user_domain_id"); ok {
		conf.UserDomainID = userDomainID.(string)
	}
	if userDomainName, ok := data.GetOk("user_domain_name"); ok {
		conf.UserDomainName = userDomainName.(string)
	}
	if appCredID, ok := data.GetOk("application_credential_id"); ok {
		conf.ApplicationCredentialID = appCredID.(string)
	}
	if appCredName, ok := data.GetOk("application_credential_name"); ok {
		conf.ApplicationCredentialName = appCredName.(string)
	}
	if appCredSecret, ok := data.GetOk("application_credential_secret"); ok {
		conf.ApplicationCredentialSecret = appCredSecret.(string)
	}
	if regionName, ok := data.GetOk("region_name"); ok {
		conf.RegionName = regionName.(string)
	}
	if cacert, ok := data.GetOk("cacert"); ok {
		conf.CACert = cacert.(string)
	}
	if cert, ok := data.GetOk("cert"); ok {
		conf.Cert = cert.(string)
	}
	if key, ok := data.GetOk("key"); ok {
		conf.Key = key.(string)
	}
	if insecure, ok := data.GetOk("insecure"); ok {
		conf.Insecure = insecure.(bool)
	}

	entry, err := logical.StorageEntryJSON(configAccessKey, conf)
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathConfigAccessDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, configAccessKey); err != nil {
		return nil, err
	}
	return nil, nil
}

type Config struct {
	AuthURL                     string `json:"auth_url"`
	UserID                      string `json:"user_id"`
	Username                    string `json:"username"`
	Password                    string `json:"password"`
	UserDomainID                string `json:"user_domain_id"`
	UserDomainName              string `json:"user_domain_name"`
	ApplicationCredentialID     string `json:"application_credential_id"`
	ApplicationCredentialName   string `json:"application_credential_name"`
	ApplicationCredentialSecret string `json:"application_credential_secret"`
	RegionName                  string `json:"region_name"`
	CACert                      string `json:"cacert"`
	Cert                        string `json:"cert"`
	Key                         string `json:"key"`
	Insecure                    bool   `json:"insecure"`
}

func (c *Config) UsesApplicationCredential() bool {
	return c.ApplicationCredentialID != "" || c.ApplicationCredentialName != ""
}

func (c *Config) AuthOptions(projectID, projectName string) *gophercloud.AuthOptions {
	return &gophercloud.AuthOptions{
		IdentityEndpoint:            c.AuthURL,
		UserID:                      c.UserID,
		Username:                    c.Username,
		Password:                    c.Password,
		DomainID:                    c.UserDomainID,
		DomainName:                  c.UserDomainName,
		TenantID:                    projectID,
		TenantName:                  projectName,
		ApplicationCredentialID:     c.ApplicationCredentialID,
		ApplicationCredentialName:   c.ApplicationCredentialName,
		ApplicationCredentialSecret: c.ApplicationCredentialSecret,
	}
}
