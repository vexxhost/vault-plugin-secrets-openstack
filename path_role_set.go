package openstack

import (
	"context"
	"errors"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathListRoles(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "roleset/?$",

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ListOperation: b.pathRoleList,
		},
	}
}

func pathRoles(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "roleset/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the roleset",
			},
			"Roles": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "List of the roles that the application credential has associated",
			},
			// "AccessRules": &framework.FieldSchema{
			// 	Type:        framework.TypeString,
			// 	Description: "List of the access rules",
			// },
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathRolesRead,
			logical.CreateOperation: b.pathRolesWrite,
			logical.UpdateOperation: b.pathRolesWrite,
			logical.DeleteOperation: b.pathRolesDelete,
		},

		ExistenceCheck: b.rolesExistenceCheck,
	}
}

// Establishes dichotomy of request operation between CreateOperation and UpdateOperation.
// Returning 'true' forces an UpdateOperation, CreateOperation otherwise.
func (b *backend) rolesExistenceCheck(ctx context.Context, req *logical.Request, d *framework.FieldData) (bool, error) {
	name := d.Get("name").(string)
	entry, err := b.Role(ctx, req.Storage, name)
	if err != nil {
		return false, err
	}
	return entry != nil, nil
}

func (b *backend) Role(ctx context.Context, storage logical.Storage, name string) (*RoleSet, error) {
	if name == "" {
		return nil, errors.New("invalid roleset name")
	}

	entry, err := storage.Get(ctx, "roleset/"+name)
	if err != nil {
		return nil, errwrap.Wrapf("error retrieving roleset: {{err}}", err)
	}
	if entry == nil {
		return nil, nil
	}

	var result RoleSet
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (b *backend) pathRoleList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	entries, err := req.Storage.List(ctx, "roleset/")
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(entries), nil
}

func (b *backend) pathRolesRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)

	role, err := b.Role(ctx, req.Storage, name)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, nil
	}

	// Generate the response
	resp := &logical.Response{
		Data: map[string]interface{}{
			"Roles":       role.Roles,
			"AccessRules": role.AccessRules,
		},
	}
	return resp, nil
}

func (b *backend) pathRolesWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)

	role, err := b.Role(ctx, req.Storage, name)
	if err != nil {
		return nil, err
	}
	if role == nil {
		role = new(RoleSet)
	}

	rawRoles, ok := d.GetOk("Roles")
	if ok {
		role.Roles = rawRoles.([]Role)
	}

	rawAccessRules, ok := d.GetOk("AccessRules")
	if ok {
		role.AccessRules = rawAccessRules.([]AccessRule)
	}

	entry, err := logical.StorageEntryJSON("roleset/"+name, role)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathRolesDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)
	if err := req.Storage.Delete(ctx, "roleset/"+name); err != nil {
		return nil, err
	}
	return nil, nil
}

type Role struct {
	// DomainID is the domain ID the role belongs to.
	DomainID string `json:"domain_id,omitempty"`
	// ID is the unique ID of the role.
	ID string `json:"id,omitempty"`
	// Name is the role name
	Name string `json:"name,omitempty"`
}

type AccessRule struct {
	// The ID of the access rule
	ID string `json:"id,omitempty"`
	// The API path that the application credential is permitted to access
	Path string `json:"path,omitempty"`
	// The request method that the application credential is permitted to use for a
	// given API endpoint
	Method string `json:"method,omitempty"`
	// The service type identifier for the service that the application credential
	// is permitted to access
	Service string `json:"service,omitempty"`
}

type RoleSet struct {
	Roles       []Role       `json:"roles,omitempty"`
	AccessRules []AccessRule `json:"access_rules,omitempty"`
}
