package openstack

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestConfigAccess(t *testing.T) {
	t.Parallel()

	b, reqStorage := getTestBackend(t)

	testConfigRead(t, b, reqStorage, nil)

	input := map[string]interface{}{
		"IdentityEndpoint":            "http://keycloak:5000",
		"UserID":                      "admin",
		"Username":                    "admin",
		"Password":                    "admin",
		"ApplicationCredentialID":     "admin",
		"ApplicationCredentialName":   "admin",
		"ApplicationCredentialSecret": "admin",
		"TenantID":                    "admin",
		"TenantName":                  "admin",
		"DomainID":                    "admin",
		"DomainName":                  "admin",
		"Region":                      "admin",
	}
	expected := map[string]interface{}{
		"IdentityEndpoint":          "http://keycloak:5000",
		"UserID":                    "admin",
		"Username":                  "admin",
		"ApplicationCredentialID":   "admin",
		"ApplicationCredentialName": "admin",
		"TenantID":                  "admin",
		"TenantName":                "admin",
		"DomainID":                  "admin",
		"DomainName":                "admin",
		"Region":                    "admin",
	}
	testConfigUpdate(t, b, reqStorage, input)
	testConfigRead(t, b, reqStorage, expected)
}

func testConfigUpdate(t *testing.T, b logical.Backend, s logical.Storage, d map[string]interface{}) {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      configAccessKey,
		Data:      d,
		Storage:   s,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp != nil && resp.IsError() {
		t.Fatal(resp.Error())
	}
}

func testConfigRead(t *testing.T, b logical.Backend, s logical.Storage, expected map[string]interface{}) {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      configAccessKey,
		Storage:   s,
	})

	if err != nil {
		t.Fatal(err)
	}

	if resp == nil && expected == nil {
		return
	}

	if resp.IsError() {
		t.Fatal(resp.Error())
	}

	if len(expected) != len(resp.Data) {
		t.Errorf("read data mismatch (expected %d values, got %d)", len(expected), len(resp.Data))
	}

	for k, expectedV := range expected {
		actualV, ok := resp.Data[k]

		if !ok {
			t.Errorf(`expected data["%s"] = %v but was not included in read output"`, k, expectedV)
		} else if expectedV != actualV {
			t.Errorf(`expected data["%s"] = %v, instead got %v"`, k, expectedV, actualV)
		}
	}

	if t.Failed() {
		t.FailNow()
	}
}
