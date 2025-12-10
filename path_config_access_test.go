package openstack

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestConfigAccess_ReadEmpty(t *testing.T) {
	t.Parallel()

	b, reqStorage := getTestBackend(t)

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      configAccessKey,
		Storage:   reqStorage,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Fatal("expected nil response for empty config")
	}
}

func TestConfigAccess_CreateAndRead(t *testing.T) {
	t.Parallel()

	b, reqStorage := getTestBackend(t)

	// Create config
	input := map[string]interface{}{
		"auth_url":                      "http://keystone:5000",
		"user_id":                       "admin",
		"username":                      "admin",
		"password":                      "admin",
		"user_domain_id":                "default",
		"user_domain_name":              "Default",
		"application_credential_id":     "appcred123",
		"application_credential_name":   "myappcred",
		"application_credential_secret": "secret123",
		"region_name":                   "RegionOne",
	}

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      configAccessKey,
		Data:      input,
		Storage:   reqStorage,
	})
	if err != nil {
		t.Fatalf("unexpected error on create: %v", err)
	}
	if resp != nil && resp.IsError() {
		t.Fatalf("unexpected error response: %v", resp.Error())
	}

	// Read and verify config
	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      configAccessKey,
		Storage:   reqStorage,
	})
	if err != nil {
		t.Fatalf("unexpected error on read: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
	if resp.IsError() {
		t.Fatalf("unexpected error response: %v", resp.Error())
	}

	// Verify fields (password and app credential secret should not be returned)
	expected := map[string]interface{}{
		"auth_url":                    "http://keystone:5000",
		"user_id":                     "admin",
		"username":                    "admin",
		"user_domain_id":              "default",
		"user_domain_name":            "Default",
		"application_credential_id":   "appcred123",
		"application_credential_name": "myappcred",
		"region_name":                 "RegionOne",
		"cacert":                      "",
		"cert":                        "",
		"insecure":                    false,
	}

	if len(resp.Data) != len(expected) {
		t.Errorf("expected %d fields, got %d", len(expected), len(resp.Data))
	}

	for k, expectedV := range expected {
		actualV, ok := resp.Data[k]
		if !ok {
			t.Errorf("expected field %q not found in response", k)
			continue
		}
		if expectedV != actualV {
			t.Errorf("field %q: expected %v, got %v", k, expectedV, actualV)
		}
	}
}

func TestConfig_UsesApplicationCredential(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name:     "empty config",
			config:   Config{},
			expected: false,
		},
		{
			name:     "username/password only",
			config:   Config{Username: "admin", Password: "secret"},
			expected: false,
		},
		{
			name:     "with application_credential_id",
			config:   Config{ApplicationCredentialID: "appcred123"},
			expected: true,
		},
		{
			name:     "with application_credential_name",
			config:   Config{ApplicationCredentialName: "myappcred"},
			expected: true,
		},
		{
			name: "with both app credential fields",
			config: Config{
				ApplicationCredentialID:   "appcred123",
				ApplicationCredentialName: "myappcred",
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.config.UsesApplicationCredential(); got != tc.expected {
				t.Errorf("UsesApplicationCredential() = %v, expected %v", got, tc.expected)
			}
		})
	}
}

func TestConfig_AuthOptions(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AuthURL:                     "http://keystone:5000",
		UserID:                      "user123",
		Username:                    "admin",
		Password:                    "secret",
		UserDomainID:                "default",
		UserDomainName:              "Default",
		ApplicationCredentialID:     "appcred123",
		ApplicationCredentialName:   "myappcred",
		ApplicationCredentialSecret: "appsecret",
	}

	authOpts := cfg.AuthOptions("project123", "myproject")

	if authOpts.IdentityEndpoint != cfg.AuthURL {
		t.Errorf("IdentityEndpoint = %q, expected %q", authOpts.IdentityEndpoint, cfg.AuthURL)
	}
	if authOpts.UserID != cfg.UserID {
		t.Errorf("UserID = %q, expected %q", authOpts.UserID, cfg.UserID)
	}
	if authOpts.Username != cfg.Username {
		t.Errorf("Username = %q, expected %q", authOpts.Username, cfg.Username)
	}
	if authOpts.Password != cfg.Password {
		t.Errorf("Password = %q, expected %q", authOpts.Password, cfg.Password)
	}
	if authOpts.DomainID != cfg.UserDomainID {
		t.Errorf("DomainID = %q, expected %q", authOpts.DomainID, cfg.UserDomainID)
	}
	if authOpts.DomainName != cfg.UserDomainName {
		t.Errorf("DomainName = %q, expected %q", authOpts.DomainName, cfg.UserDomainName)
	}
	if authOpts.TenantID != "project123" {
		t.Errorf("TenantID = %q, expected %q", authOpts.TenantID, "project123")
	}
	if authOpts.TenantName != "myproject" {
		t.Errorf("TenantName = %q, expected %q", authOpts.TenantName, "myproject")
	}
	if authOpts.ApplicationCredentialID != cfg.ApplicationCredentialID {
		t.Errorf("ApplicationCredentialID = %q, expected %q", authOpts.ApplicationCredentialID, cfg.ApplicationCredentialID)
	}
	if authOpts.ApplicationCredentialName != cfg.ApplicationCredentialName {
		t.Errorf("ApplicationCredentialName = %q, expected %q", authOpts.ApplicationCredentialName, cfg.ApplicationCredentialName)
	}
	if authOpts.ApplicationCredentialSecret != cfg.ApplicationCredentialSecret {
		t.Errorf("ApplicationCredentialSecret = %q, expected %q", authOpts.ApplicationCredentialSecret, cfg.ApplicationCredentialSecret)
	}
}
