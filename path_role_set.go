package openstack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/identity/v3/applicationcredentials"
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
		return nil, fmt.Errorf("error retrieving roleset: %w", err)
	}
	if entry == nil {
		return nil, nil
	}

	result := &RoleSet{}
	if err := entry.DecodeJSON(result); err != nil {
		return nil, err
	}
	return result, nil
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
			"Roles": role.Roles,
			// "AccessRules": role.AccessRules,
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
		roles := make([]applicationcredentials.Role, 0, 10)
		err := json.Unmarshal([]byte(rawRoles.(string)), &roles)
		if err != nil {
			return nil, err
		}
		role.Roles = roles
	}

	// rawAccessRules, ok := d.GetOk("AccessRules")
	// if ok {
	// 	role.AccessRules = rawAccessRules.([]applicationcredentials.AccessRule)
	// }

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

type RoleSet struct {
	Roles []applicationcredentials.Role `json:"roles,omitempty"`
	// AccessRules []applicationcredentials.AccessRule `json:"access_rules,omitempty"`
}
