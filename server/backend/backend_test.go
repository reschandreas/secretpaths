package backend

import (
	"context"
	"os"
	"testing"
)

func TestAutoAuth_NoMethod(t *testing.T) {
	for _, k := range []string{"KUBERNETES_ROLE", "APPROLE_ROLE_ID", "APPROLE_SECRET_ID", "VAULT_TOKEN"} {
		os.Unsetenv(k)
	}
	client, err := AutoAuth(context.Background())
	if err == nil {
		t.Fatal("expected error when no auth method configured")
	}
	if client != nil {
		t.Fatal("expected nil client")
	}
}

func TestAutoAuth_Token(t *testing.T) {
	for _, k := range []string{"KUBERNETES_ROLE", "APPROLE_ROLE_ID", "APPROLE_SECRET_ID"} {
		os.Unsetenv(k)
	}
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_TOKEN")
	os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	defer os.Unsetenv("VAULT_ADDR")

	client, err := AutoAuth(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}