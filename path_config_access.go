package mock

import (
	"context"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const configAccessKey = "config/auth"

func pathConfigAccess(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: configAccessKey,
		Fields: map[string]*framework.FieldSchema{
			"IdentityEndpoint": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Identity address",
			},

			"UserID": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "UserID",
			},

			"Username": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Username",
			},

			"Password": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Password",
			},

			"ApplicationCredentialID": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "ApplicationCredentialID",
			},

			"ApplicationCredentialName": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "ApplicationCredentialName",
			},

			"ApplicationCredentialSecret": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "ApplicationCredentialSecret",
			},

			"TenantID": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "TenantID",
			},

			"TenantName": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "TenantName",
			},

			"DomainID": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "DomainID",
			},

			"DomainName": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "DomainName",
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

func (b *backend) readConfigAccess(ctx context.Context, storage logical.Storage) (*gophercloud.AuthOptions, error) {
	entry, err := storage.Get(ctx, configAccessKey)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	conf := &gophercloud.AuthOptions{}
	if err := entry.DecodeJSON(conf); err != nil {
		return nil, errwrap.Wrapf("error reading nomad access configuration: {{err}}", err)
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
			"IdentityEndpoint":          conf.IdentityEndpoint,
			"UserID":                    conf.UserID,
			"Username":                  conf.Username,
			"TenantID":                  conf.TenantID,
			"TenantName":                conf.TenantName,
			"DomainID":                  conf.DomainID,
			"DomainName":                conf.DomainName,
			"ApplicationCredentialID":   conf.ApplicationCredentialID,
			"ApplicationCredentialName": conf.ApplicationCredentialName,
		},
	}, nil
}

func (b *backend) pathConfigAccessWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	conf, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &gophercloud.AuthOptions{}
	}

	auth_url, ok := data.GetOk("IdentityEndpoint")
	if ok {
		conf.Address = auth_url.(string)
	}
	user_id, ok := data.GetOk("UserID")
	if ok {
		conf.Token = user_id.(string)
	}
	username, ok := data.GetOk("Username")
	if ok {
		conf.CACert = username.(string)
	}
	password, ok := data.GetOk("Password")
	if ok {
		conf.ClientCert = password.(string)
	}
	project_id, ok := data.GetOk("TenantID")
	if ok {
		conf.ClientKey = project_id.(string)
	}
	project_name, ok := data.GetOk("TenantName")
	if ok {
		conf.Token = project_name.(string)
	}
	domain_id, ok := data.GetOk("DomainID")
	if ok {
		conf.Token = domain_id.(string)
	}
	domain_name, ok := data.GetOk("DomainName")
	if ok {
		conf.Token = domain_name.(string)
	}
	credential_id, ok := data.GetOk("ApplicationCredentialID")
	if ok {
		conf.Token = credential_id.(string)
	}
	credential_name, ok := data.GetOk("ApplicationCredentialName")
	if ok {
		conf.Token = credential_name.(string)
	}
	credential_secret, ok := data.GetOk("ApplicationCredentialSecret")
	if ok {
		conf.Token = credential_secret.(string)
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
