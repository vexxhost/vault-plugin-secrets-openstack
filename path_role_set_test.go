package openstack

import (
	"context"
	"testing"

	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/applicationcredentials"
	"github.com/hashicorp/vault/sdk/logical"
)

func TestRoleSet_CRUD(t *testing.T) {
	t.Parallel()

	b, reqStorage := getTestBackend(t)

	// Test reading non-existent roleset returns nil
	t.Run("read non-existent", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "roleset/test",
			Storage:   reqStorage,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Fatal("expected nil response for non-existent roleset")
		}
	})

	// Test creating a roleset
	t.Run("create", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      "roleset/test",
			Data: map[string]interface{}{
				"project_id":          "project123",
				"project_name":        "myproject",
				"project_domain_id":   "default",
				"project_domain_name": "Default",
				"roles":               `[{"id": "role123"}, {"name": "member"}]`,
			},
			Storage: reqStorage,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp != nil && resp.IsError() {
			t.Fatal(resp.Error())
		}
	})

	// Test reading the created roleset
	t.Run("read", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "roleset/test",
			Storage:   reqStorage,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatal("expected response")
		}
		if resp.IsError() {
			t.Fatal(resp.Error())
		}

		if resp.Data["project_id"] != "project123" {
			t.Errorf("expected project_id=project123, got %v", resp.Data["project_id"])
		}
		if resp.Data["project_name"] != "myproject" {
			t.Errorf("expected project_name=myproject, got %v", resp.Data["project_name"])
		}
		if resp.Data["project_domain_id"] != "default" {
			t.Errorf("expected project_domain_id=default, got %v", resp.Data["project_domain_id"])
		}
		if resp.Data["project_domain_name"] != "Default" {
			t.Errorf("expected project_domain_name=Default, got %v", resp.Data["project_domain_name"])
		}

		roles, ok := resp.Data["roles"].([]applicationcredentials.Role)
		if !ok {
			t.Fatalf("expected roles to be []applicationcredentials.Role, got %T", resp.Data["roles"])
		}
		if len(roles) != 2 {
			t.Errorf("expected 2 roles, got %d", len(roles))
		}
	})

	// Test updating the roleset
	t.Run("update", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.UpdateOperation,
			Path:      "roleset/test",
			Data: map[string]interface{}{
				"project_id": "newproject456",
			},
			Storage: reqStorage,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp != nil && resp.IsError() {
			t.Fatal(resp.Error())
		}

		// Verify update
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "roleset/test",
			Storage:   reqStorage,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp.Data["project_id"] != "newproject456" {
			t.Errorf("expected project_id=newproject456, got %v", resp.Data["project_id"])
		}
		// Other fields should be preserved
		if resp.Data["project_name"] != "myproject" {
			t.Errorf("expected project_name=myproject to be preserved, got %v", resp.Data["project_name"])
		}
	})

	// Test deleting the roleset
	t.Run("delete", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.DeleteOperation,
			Path:      "roleset/test",
			Storage:   reqStorage,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp != nil && resp.IsError() {
			t.Fatal(resp.Error())
		}

		// Verify deletion
		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "roleset/test",
			Storage:   reqStorage,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Fatal("expected nil response after deletion")
		}
	})
}

func TestRoleSet_List(t *testing.T) {
	t.Parallel()

	b, reqStorage := getTestBackend(t)

	// Test listing empty rolesets
	t.Run("list empty", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ListOperation,
			Path:      "roleset/",
			Storage:   reqStorage,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatal("expected response")
		}
		if resp.Data["keys"] != nil {
			keys := resp.Data["keys"].([]string)
			if len(keys) != 0 {
				t.Errorf("expected 0 keys, got %d", len(keys))
			}
		}
	})

	// Create some rolesets
	for _, name := range []string{"role1", "role2", "role3"} {
		_, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.CreateOperation,
			Path:      "roleset/" + name,
			Data: map[string]interface{}{
				"roles": `[{"id": "test"}]`,
			},
			Storage: reqStorage,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test listing rolesets
	t.Run("list with entries", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ListOperation,
			Path:      "roleset/",
			Storage:   reqStorage,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatal("expected response")
		}
		keys := resp.Data["keys"].([]string)
		if len(keys) != 3 {
			t.Errorf("expected 3 keys, got %d", len(keys))
		}
	})
}

func TestRoleSet_InvalidJSON(t *testing.T) {
	t.Parallel()

	b, reqStorage := getTestBackend(t)

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "roleset/test",
		Data: map[string]interface{}{
			"roles": `invalid json`,
		},
		Storage: reqStorage,
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if resp != nil && !resp.IsError() {
		t.Fatal("expected error response")
	}
}

func TestRoleSet_HasProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		roleset  RoleSet
		expected bool
	}{
		{
			name:     "empty",
			roleset:  RoleSet{},
			expected: false,
		},
		{
			name:     "with project_id",
			roleset:  RoleSet{ProjectID: "project123"},
			expected: true,
		},
		{
			name:     "with project_name",
			roleset:  RoleSet{ProjectName: "myproject"},
			expected: true,
		},
		{
			name:     "with both",
			roleset:  RoleSet{ProjectID: "project123", ProjectName: "myproject"},
			expected: true,
		},
		{
			name:     "with only domain fields",
			roleset:  RoleSet{ProjectDomainID: "default", ProjectDomainName: "Default"},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.roleset.HasProject() != tc.expected {
				t.Errorf("expected HasProject()=%v, got %v", tc.expected, tc.roleset.HasProject())
			}
		})
	}
}

func TestRole_EmptyName(t *testing.T) {
	t.Parallel()

	b, reqStorage := getTestBackend(t)

	_, err := b.(*backend).Role(context.Background(), reqStorage, "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if err.Error() != "invalid roleset name" {
		t.Errorf("expected 'invalid roleset name' error, got: %v", err)
	}
}
