package openstack

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestConfigLease_ReadEmpty(t *testing.T) {
	t.Parallel()

	b, reqStorage := getTestBackend(t)

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      leaseConfigKey,
		Storage:   reqStorage,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Fatal("expected nil response for empty lease config")
	}
}

func TestConfigLease_CreateAndRead(t *testing.T) {
	t.Parallel()

	b, reqStorage := getTestBackend(t)

	// Create lease config
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      leaseConfigKey,
		Data: map[string]interface{}{
			"ttl": int64(3600),
		},
		Storage: reqStorage,
	})
	if err != nil {
		t.Fatalf("unexpected error on create: %v", err)
	}
	if resp != nil && resp.IsError() {
		t.Fatalf("unexpected error response: %v", resp.Error())
	}

	// Read and verify lease config
	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      leaseConfigKey,
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

	if resp.Data["ttl"] != int64(3600) {
		t.Errorf("ttl = %v, expected %v", resp.Data["ttl"], int64(3600))
	}
}

func TestConfigLease_Update(t *testing.T) {
	t.Parallel()

	b, reqStorage := getTestBackend(t)

	// Create initial lease config
	_, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      leaseConfigKey,
		Data: map[string]interface{}{
			"ttl": int64(60),
		},
		Storage: reqStorage,
	})
	if err != nil {
		t.Fatalf("unexpected error on create: %v", err)
	}

	// Update lease config
	_, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      leaseConfigKey,
		Data: map[string]interface{}{
			"ttl": int64(120),
		},
		Storage: reqStorage,
	})
	if err != nil {
		t.Fatalf("unexpected error on update: %v", err)
	}

	// Verify update
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      leaseConfigKey,
		Storage:   reqStorage,
	})
	if err != nil {
		t.Fatalf("unexpected error on read: %v", err)
	}
	if resp.Data["ttl"] != int64(120) {
		t.Errorf("ttl = %v, expected %v", resp.Data["ttl"], int64(120))
	}
}

func TestLeaseConfig_Struct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   configLease
		expected time.Duration
	}{
		{
			name:     "zero TTL",
			config:   configLease{TTL: 0},
			expected: 0,
		},
		{
			name:     "one hour TTL",
			config:   configLease{TTL: time.Hour},
			expected: time.Hour,
		},
		{
			name:     "custom TTL",
			config:   configLease{TTL: 30 * time.Minute},
			expected: 30 * time.Minute,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.config.TTL != tc.expected {
				t.Errorf("TTL = %v, expected %v", tc.config.TTL, tc.expected)
			}
		})
	}
}
