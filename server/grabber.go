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

func GetPolicies(ctx context.Context, client *vault.Client) ([]models.Policy, error) {
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

func recursivelyGetGraphPaths(ctx context.Context, client *vault.Client, path, kvEngine string) ([]models.GraphEntry, error) {
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
	var secrets []models.GraphEntry
	for _, subPath := range response.Data.Keys {
		if !strings.HasSuffix(subPath, "/") {
			secrets = append(secrets, models.GraphEntry{AbsolutePath: path + subPath, Id: path + subPath, Level: strings.Count(path+subPath, "/"), Name: subPath})
			continue
		}
		subSecrets, err := recursivelyGetGraphPaths(ctx, client, path+subPath, kvEngine)
		if err != nil {
			log.Default().Println("error listing paths of ", path+subPath)
			continue
		}
		subPath = strings.Replace(subPath, "/", "", 1)
		secrets = append(secrets, models.GraphEntry{AbsolutePath: path + subPath, Id: path + subPath, Name: subPath, Level: strings.Count(path+subPath, "/"), Children: subSecrets})
	}
	return secrets, err

}

func GetPaths(ctx context.Context, client *vault.Client) ([]models.Secret, error) {
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

func getGraphPaths(ctx context.Context, client *vault.Client) (models.GraphEntry, error) {
	kvEngine := "secret"

	val, ok := os.LookupEnv("VAULT_KV_ENGINE")
	if ok {
		kvEngine = val
	}

	secrets, err := recursivelyGetGraphPaths(ctx, client, "/", kvEngine)
	if err != nil {
		log.Fatal(err)
	}

	return models.GraphEntry{AbsolutePath: "/", Id: "/", Name: "/", Children: secrets}, err
}
