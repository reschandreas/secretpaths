package main

import (
	"context"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/tjarratt/babble"
	"log"
	"math/rand"
	"secretpaths/backend"
	"secretpaths/models"
)

func getRandomPath() string {
	babbler := babble.NewBabbler()
	count := rand.Intn(10)
	if count < 2 {
		count = 2
	}
	babbler.Count = count
	babbler.Separator = "/"
	return babbler.Babble()
}

func getRandomKey() string {
	babbler := babble.NewBabbler()
	babbler.Count = 1
	return babbler.Babble()
}

func getRandomValue() string {
	return getRandomKey()

}

func writeSecret(ctx context.Context, path string, key string, value string, client *vault.Client) {
	_, err := client.Secrets.KvV2Write(ctx, path, schema.KvV2WriteRequest{
		Data: map[string]any{
			key: value,
		}},
		vault.WithMountPath("secret"))
	if err != nil {
		log.Fatal(err)
	}
}

func writePolicy(ctx context.Context, client *vault.Client, name string, paths []string) {
	capabilities := []string{"read", "list"}
	rules := []models.Rule{}
	for _, path := range paths {
		rules = append(rules, models.Rule{Path: path, Capabilities: capabilities})
	}
	policy := models.Policy{Name: name, Rules: rules}
	request := schema.PoliciesWriteAclPolicyRequest{
		Policy: policy.ToRequest(),
	}
	println(policy.ToRequest())
	_, err := client.System.PoliciesWriteAclPolicy(ctx, policy.Name, request)
	if err != nil {
		log.Fatal(err)
	}
}

// Creates a bunch of secrets and policies in Vault to be used in the demo
// The secrets are created with random paths, keys, and values
func main() {
	ctx := context.Background()

	numberOfSecrets := 300
	numberOfPolicies := 30

	client := backend.UseToken("my-token")

	log.Println("Let's create some secrets")

	paths := []string{}

	for i := 0; i < numberOfSecrets; i++ {
		path := getRandomPath()
		key := getRandomKey()
		value := getRandomValue()
		log.Printf("writing secret %s with key %s and value %s", path, key, value)
		writeSecret(ctx, path, key, value, client)
		log.Printf("secret %s written successfully", path)
		paths = append(paths, path)
	}

	log.Println("Let's create a policy")
	for i := 0; i < numberOfPolicies; i++ {
		name := "policy_" + getRandomKey()
		start := rand.Intn(len(paths))
		end := rand.Intn(len(paths))
		if start > end {
			start, end = end, start
		}
		subpaths := paths[start:end]
		log.Printf("writing policy %s", name)
		writePolicy(ctx, client, name, subpaths)
	}
}
