package backend

import (
	"context"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"log"
	"os"
	"time"
)

func SetupConnection() *vault.Client {
	serverAddress := "http://127.0.0.1:8200"

	val, ok := os.LookupEnv("VAULT_ADDR")

	if ok {
		serverAddress = val
	}

	client, err := vault.New(
		vault.WithAddress(serverAddress),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func UseToken(token string) *vault.Client {
	client := SetupConnection()

	if err := client.SetToken(token); err != nil {
		log.Fatal(err)
	}
	return client
}

func UseAppRole(ctx context.Context) *vault.Client {
	client := SetupConnection()

	resp, err := client.Auth.AppRoleLogin(
		ctx,
		schema.AppRoleLoginRequest{
			RoleId:   os.Getenv("APPROLE_ROLE_ID"),
			SecretId: os.Getenv("APPROLE_SECRET_ID"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		log.Fatal(err)
	}
	return client
}
