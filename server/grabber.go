package main

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/hashicorp/vault-client-go"
	"golang.org/x/sync/errgroup"
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
		return nil, err
	}
	var keys []string
	for _, rawPolicy := range response.Data.Keys {
		if rawPolicy == "root" {
			continue
		}
		keys = append(keys, rawPolicy)
	}

	policies := make([]models.Policy, len(keys))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(16)
	for i, name := range keys {
		i, name := i, name
		g.Go(func() error {
			policy, err := parsePolicy(gctx, client, name)
			if err != nil {
				log.Println("could not parse policy", name)
				return nil
			}
			policies[i] = policy
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	// drop zero values (failed parses)
	out := policies[:0]
	for _, p := range policies {
		if p.Name != "" {
			out = append(out, p)
		}
	}
	return out, nil
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
		if responseError != nil && responseError.StatusCode == 404 {
			log.Default().Println("there is nothing at", path)
			return nil, nil
		}
		return nil, err
	}
	// First pass: collect direct secrets; collect subdirs to recurse in parallel.
	var secrets []models.Secret
	var dirs []string
	for _, subPath := range response.Data.Keys {
		if !strings.HasSuffix(subPath, "/") {
			secrets = append(secrets, models.Secret{Path: path + subPath})
		} else {
			dirs = append(dirs, subPath)
		}
	}

	results := make([][]models.Secret, len(dirs))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(16)
	for i, subPath := range dirs {
		i, subPath := i, subPath
		g.Go(func() error {
			subSecrets, err := recursivelyGetPaths(gctx, client, path+subPath, kvEngine)
			if err != nil {
				log.Default().Println("error listing paths of ", path+subPath)
				return nil
			}
			results[i] = subSecrets
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return secrets, err
	}
	for _, r := range results {
		secrets = append(secrets, r...)
	}
	return secrets, nil
}

func recursivelyGetGraphPaths(ctx context.Context, client *vault.Client, path, kvEngine string, stopAtRecursion int) ([]models.GraphEntry, error) {
	if stopAtRecursion == 0 {
		return []models.GraphEntry{}, nil
	}
	nextDepth := stopAtRecursion
	if nextDepth > 0 {
		nextDepth--
	}
	response, err := client.Secrets.KvV2List(ctx, path, vault.WithMountPath(kvEngine))
	if err != nil {
		var responseError *vault.ResponseError
		errors.As(err, &responseError)
		if responseError != nil && responseError.StatusCode == 404 {
			log.Default().Println("there is nothing at", path)
			return nil, nil
		}
		return nil, err
	}
	type dirJob struct {
		subPath string
		id      string
	}
	var direct []models.GraphEntry
	var dirs []dirJob
	for _, subPath := range response.Data.Keys {
		id := uuid.New().String()
		if !strings.HasSuffix(subPath, "/") {
			direct = append(direct, models.GraphEntry{AbsolutePath: path + subPath, Id: id, Name: subPath})
		} else {
			dirs = append(dirs, dirJob{subPath: subPath, id: id})
		}
	}

	results := make([]models.GraphEntry, len(dirs))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(16)
	for i, d := range dirs {
		i, d := i, d
		g.Go(func() error {
			subSecrets, err := recursivelyGetGraphPaths(gctx, client, path+d.subPath, kvEngine, nextDepth)
			if err != nil {
				log.Default().Println("error listing paths of ", path+d.subPath)
				subSecrets = nil
			}
			name := strings.Replace(d.subPath, "/", "", 1)
			results[i] = models.GraphEntry{AbsolutePath: path + name, Id: d.id, Name: name, Children: subSecrets}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	secrets := append([]models.GraphEntry{}, direct...)
	for _, r := range results {
		secrets = append(secrets, r)
	}
	return secrets, nil
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
