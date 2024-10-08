package main

import (
	"context"
	"errors"
	"github.com/google/uuid"
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
		log.Println(err)
	}
	var policies = make([]models.Policy, 0)
	for _, rawPolicy := range response.Data.Keys {
		if rawPolicy == "root" {
			// skip the root policy, as it is not a real policy
			continue
		}
		policy, err := parsePolicy(ctx, client, rawPolicy)
		if err != nil {
			log.Println("could not parse policy", rawPolicy)
		} else {
			policies = append(policies, policy)
		}
	}
	return policies, nil
}

func parsePolicy(ctx context.Context, client *vault.Client, name string) (policy models.Policy, err error) {
	p, err := client.System.PoliciesReadAclPolicy(ctx, name)
	if err != nil {
		log.Default().Println("error reading policy", name)
		log.Println(err)
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

func recursivelyGetGraphPaths(ctx context.Context, client *vault.Client, path, kvEngine string, stopAtRecursion int) ([]models.GraphEntry, error) {
	response, err := client.Secrets.KvV2List(ctx, path, vault.WithMountPath(kvEngine))
	if stopAtRecursion == 0 {
		return []models.GraphEntry{}, nil
	}
	if stopAtRecursion > 0 {
		stopAtRecursion--
	}
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
		id := uuid.New().String()
		if !strings.HasSuffix(subPath, "/") {
			secrets = append(secrets, models.GraphEntry{AbsolutePath: path + subPath, Id: id, Name: subPath})
			continue
		}
		subSecrets, err := recursivelyGetGraphPaths(ctx, client, path+subPath, kvEngine, stopAtRecursion)
		if err != nil {
			log.Default().Println("error listing paths of ", path+subPath)
			continue
		}
		subPath = strings.Replace(subPath, "/", "", 1)
		secrets = append(secrets, models.GraphEntry{AbsolutePath: path + subPath, Id: id, Name: subPath, Children: subSecrets})
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
		log.Println(err)
	}

	return secrets, err
}

func getGraphPaths(ctx context.Context, client *vault.Client, stopAtRecursion int) (models.GraphEntry, error) {
	kvEngine := "secret"

	val, ok := os.LookupEnv("VAULT_KV_ENGINE")
	if ok {
		kvEngine = val
	}

	secrets, err := recursivelyGetGraphPaths(ctx, client, "/", kvEngine, stopAtRecursion)
	if err != nil {
		log.Println(err)
	}

	return models.GraphEntry{AbsolutePath: "/", Id: "/", Name: "/", Children: secrets}, err
}
