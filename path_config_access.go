package openstack

import (
	"context"

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
			"Region": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the region",
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

func (b *backend) readConfigAccess(ctx context.Context, storage logical.Storage) (*authOptions, error) {
	entry, err := storage.Get(ctx, configAccessKey)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	conf := &authOptions{}
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
			"Region":                    conf.Region,
		},
	}, nil
}

func (b *backend) pathConfigAccessWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	conf, err := b.readConfigAccess(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if conf == nil {
		conf = &authOptions{}
	}

	auth_url, ok := data.GetOk("IdentityEndpoint")
	if ok {
		conf.IdentityEndpoint = auth_url.(string)
	}
	user_id, ok := data.GetOk("UserID")
	if ok {
		conf.UserID = user_id.(string)
	}
	username, ok := data.GetOk("Username")
	if ok {
		conf.Username = username.(string)
	}
	password, ok := data.GetOk("Password")
	if ok {
		conf.Password = password.(string)
	}
	project_id, ok := data.GetOk("TenantID")
	if ok {
		conf.TenantID = project_id.(string)
	}
	project_name, ok := data.GetOk("TenantName")
	if ok {
		conf.TenantName = project_name.(string)
	}
	domain_id, ok := data.GetOk("DomainID")
	if ok {
		conf.DomainID = domain_id.(string)
	}
	domain_name, ok := data.GetOk("DomainName")
	if ok {
		conf.DomainName = domain_name.(string)
	}
	credential_id, ok := data.GetOk("ApplicationCredentialID")
	if ok {
		conf.ApplicationCredentialID = credential_id.(string)
	}
	credential_name, ok := data.GetOk("ApplicationCredentialName")
	if ok {
		conf.ApplicationCredentialName = credential_name.(string)
	}
	credential_secret, ok := data.GetOk("ApplicationCredentialSecret")
	if ok {
		conf.ApplicationCredentialSecret = credential_secret.(string)
	}
	region, ok := data.GetOk("Region")
	if ok {
		conf.Region = region.(string)
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

type authOptions struct {
	IdentityEndpoint            string
	UserID                      string
	Username                    string
	Password                    string
	TenantID                    string
	TenantName                  string
	DomainID                    string
	DomainName                  string
	ApplicationCredentialID     string
	ApplicationCredentialName   string
	ApplicationCredentialSecret string
	Region                      string
}
