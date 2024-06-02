package main

import (
	"context"
	"log"
	"secretpath/backend"
)

func main() {
	ctx := context.Background()

	client := backend.SetupClient("my-token")

	policies, err := getPolicies(ctx, client)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("rules:", policies)
	for _, policy := range policies {
		log.Printf("policy: %s has %d rules", policy.Name, policy.AmountOfPolicies())
		if policy.AmountOfPolicies() > 0 {
			for _, rule := range policy.Rules {
				log.Printf("rule: %s has %s capabilities", rule.Path, rule.Capability)
			}
		}
	}
	log.Println("policies:", len(policies))
	log.Println("")
	log.Println("")
	log.Println("")

	needle := "secret/bar/zip"

	for _, policy := range policies {
		if policy.Name != "policy_testing" {
			continue
		}
		if policy.HasAccessTo(needle) {
			log.Printf("policy: %s has access to %s", policy.Name, needle)
		} else {
			log.Printf("policy: %s has no access to %s", policy.Name, needle)
		}
	}
}
