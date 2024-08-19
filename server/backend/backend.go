package backend

import (
	"context"
	"fmt"
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
		log.Printf("could not login: error: %v", err)
		log.Println(err)
	}
	return client
}

func AutoAuth(ctx context.Context) (*vault.Client, error) {
	if os.Getenv("KUBERNETES_ROLE") != "" {
		log.Println("using kubernetes authentication")
		return useKubernetes(ctx), nil
	}
	if os.Getenv("APPROLE_ROLE_ID") != "" && os.Getenv("APPROLE_SECRET_ID") != "" {
		log.Println("using approle authentication")
		return useAppRole(ctx), nil
	}
	if os.Getenv("VAULT_TOKEN") != "" {
		log.Println("using token authentication")
		return UseToken(os.Getenv("VAULT_TOKEN")), nil
	}
	return nil, fmt.Errorf("no authentication method found")
}

func UseToken(token string) *vault.Client {
	client := SetupConnection()

	if err := client.SetToken(token); err != nil {
		log.Println(err)
	}
	return client
}

func useAppRole(ctx context.Context) *vault.Client {
	client := SetupConnection()

	resp, err := client.Auth.AppRoleLogin(
		ctx,
		schema.AppRoleLoginRequest{
			RoleId:   os.Getenv("APPROLE_ROLE_ID"),
			SecretId: os.Getenv("APPROLE_SECRET_ID"),
		},
	)
	if err != nil {
		log.Println(err)
	}

	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		log.Println(err)
	}
	return client
}

func useKubernetes(ctx context.Context) *vault.Client {
	client := SetupConnection()

	file, _ := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")

	resp, err := client.Auth.KubernetesLogin(
		ctx,
		schema.KubernetesLoginRequest{
			Role: os.Getenv("KUBERNETES_ROLE"),
			Jwt:  string(file),
		},
		vault.WithMountPath(os.Getenv("KUBERNETES_PATH")),
	)
	if err != nil {
		log.Println(err)
	}

	if err := client.SetToken(resp.Auth.ClientToken); err != nil {
		log.Println(err)
	}
	return client
}
