package main

import (
	"context"
	"github.com/hashicorp/vault-client-go"
	"log"
	"secretpath/models"
)

func getPolicies(ctx context.Context, client *vault.Client) ([]models.Policy, error) {
	response, err := client.System.PoliciesListAclPolicies(ctx)

	if err != nil {
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
		log.Fatal(err)
	}
	policy, err = models.FromHCL(name, []byte(p.Data.Policy))
	return
}
