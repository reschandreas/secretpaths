package main

import (
	"context"
	"errors"
	"github.com/hashicorp/vault-client-go"
	"log"
	"os"
	"secretpaths/models"
	"strings"
)

func getPolicies(ctx context.Context, client *vault.Client) ([]models.Policy, error) {
	response, err := client.System.PoliciesListAclPolicies(ctx)

	if err != nil {
		log.Default().Println("error listing policies")
		log.Fatal(err)
	}
	var policies []models.Policy
	for _, rawPolicy := range response.Data.Keys {
		if rawPolicy == "root" {
			// skip the root policy, as it is not a real policy
			continue
		}
		policy, err := parsePolicy(ctx, client, rawPolicy)
		if err != nil {
			log.Fatal(err)
		}
		policies = append(policies, policy)
	}
	return policies, nil
}

func parsePolicy(ctx context.Context, client *vault.Client, name string) (policy models.Policy, err error) {
	p, err := client.System.PoliciesReadAclPolicy(ctx, name)
	if err != nil {
		log.Default().Println("error reading policy", name)
		log.Fatal(err)
	}
	policy, err = models.FromHCL(name, []byte(p.Data.Policy))
	return
}

func recursivelyGetPaths(ctx context.Context, client *vault.Client, path, kvEngine string) ([]models.Secret, error) {
	response, err := client.Secrets.KvV2List(ctx, path, vault.WithMountPath(kvEngine))
	if err != nil {
		var responseError *vault.ResponseError
		errors.As(err, &responseError)
		if responseError.StatusCode == 404 {
			log.Default().Println("there is nothing at", path)
			return nil, nil
		}
		return nil, err
	}
	var secrets []models.Secret
	for _, subPath := range response.Data.Keys {
		if !strings.HasSuffix(subPath, "/") {
			secrets = append(secrets, models.Secret{Path: path + subPath})
			continue
		}
		subSecrets, err := recursivelyGetPaths(ctx, client, path+subPath, kvEngine)
		if err != nil {
			log.Default().Println("error listing paths of ", path+subPath)
			continue
		}
		secrets = append(secrets, subSecrets...)
	}
	return secrets, err

}

func getPaths(ctx context.Context, client *vault.Client) ([]models.Secret, error) {
	kvEngine := "secret"

	val, ok := os.LookupEnv("VAULT_KV_ENGINE")
	if ok {
		kvEngine = val
	}

	secrets, err := recursivelyGetPaths(ctx, client, "/", kvEngine)
	if err != nil {
		log.Fatal(err)
	}

	return secrets, err
}
