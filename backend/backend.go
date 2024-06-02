package backend

import (
	"github.com/hashicorp/vault-client-go"
	"log"
	"os"
	"time"
)

func SetupClient(token string) *vault.Client {
	serverAddress := "http://127.0.0.1:8200"

	val, ok := os.LookupEnv("VAULT_ADDR")

	if ok {
		serverAddress = val
	}

	os.Getenv("VAULT_ADDR")

	client, err := vault.New(
		vault.WithAddress(serverAddress),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.SetToken(token); err != nil {
		log.Fatal(err)
	}
	return client
}
