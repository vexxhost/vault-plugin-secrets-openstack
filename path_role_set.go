package openstack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/applicationcredentials"
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
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the role set",
			},
			"roles": {
				Type:        framework.TypeString,
				Description: "JSON array of roles for the application credential",
			},
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

	return &logical.Response{
		Data: map[string]interface{}{
			"roles": role.Roles,
		},
	}, nil
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

	if rawRoles, ok := d.GetOk("roles"); ok {
		var roles []applicationcredentials.Role
		if err := json.Unmarshal([]byte(rawRoles.(string)), &roles); err != nil {
			return nil, fmt.Errorf("invalid roles JSON: %w", err)
		}
		role.Roles = roles
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

type RoleSet struct {
	Roles []applicationcredentials.Role `json:"roles,omitempty"`
}
