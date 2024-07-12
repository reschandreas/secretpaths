package models_test

import (
	"secretpaths/models"
	"strings"
	"testing"
)

// test cases are taken from https://developer.hashicorp.com/vault/docs/concepts/policies

func TestPolicy_ToRequest(t *testing.T) {
	policy := models.Policy{
		Name: "policy_testing",
		Rules: []models.Rule{
			{
				Path:       "secret/foo",
				Capability: []string{"read"},
			},
		},
	}
	shouldBe := "path \"secret/foo\" {\n  capabilities = [\"read\"]\n}\n"
	strings.Compare(shouldBe, policy.ToRequest())
}

func TestPolicy_FromHCL(t *testing.T) {
	hcl := "path \"secret/foo\" {\n  capabilities = [\"read\"]\n}\n"
	policy, err := models.FromHCL("policy_testing", []byte(hcl))
	if err != nil {
		t.Error(err)
	}
	if policy.Name != "policy_testing" {
		t.Errorf("expected: %s, got: %s", "policy_testing", policy.Name)
	}
	if policy.AmountOfPolicies() != 1 {
		t.Errorf("expected: %d, got: %d", 1, policy.AmountOfPolicies())
	}
	if len(policy.Rules[0].Capability) != 1 {
		t.Errorf("expected: %d, got: %d", 1, len(policy.Rules[0].Capability))
	}
	if policy.Rules[0].Path != "secret/foo" {
		t.Errorf("expected: %s, got: %s", "secret/foo", policy.Rules[0].Path)
	}
	if policy.Rules[0].Capability[0] != "read" {
		t.Errorf("expected: %s, got: %s", "read", policy.Rules[0].Capability[0])
	}
}

func TestPolicy_WrongFormat(t *testing.T) {
	hcl := "path \"secret/foo\" {\n  capabilities = [\"read\"]\n"
	_, err := models.FromHCL("policy_testing", []byte(hcl))
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestPolicy_HasAccessTo(t *testing.T) {
	rule := models.NewRule("secret/*", []string{"create", "read", "update", "patch", "delete", "list"})
	super_secret := models.NewRule("secret/super-secret", []string{"deny"})
	restricted := models.NewRule("secret/restricted", []string{"create"})
	policy := models.NewPolicy("policy_testing", []models.Rule{rule, super_secret, restricted})
	if policy.HasAccessTo("secret/super-secret") {
		t.Error("expected: false, got: true, second rule should deny access")
	}
	if !policy.HasAccessTo("secret/foo") {
		t.Error("expected: true, got: false, first rule should allow access")
	}
	if !policy.HasAccessTo("secret/restricted") {
		t.Error("expected: false, got: true, third rule should allow access")
	}
}

func TestPolicy_HasAccessTo2(t *testing.T) {
	rule1 := models.NewRule("secret/foo", []string{"read"})
	rule2 := models.NewRule("secret/bar/*", []string{"read"})
	rule3 := models.NewRule("secret/zip-*", []string{"read"})
	policy := models.NewPolicy("policy_testing", []models.Rule{rule1, rule2, rule3})
	if !policy.HasAccessTo("secret/foo") {
		t.Error("expected: true, got: false, first rule should allow access")
	}
	if policy.HasAccessTo("secret/foo/bar") {
		t.Error("expected: false, got: true, first rule should deny access")
	}
	if !policy.HasAccessTo("secret/bar/*") {
		t.Error("expected: true, got: false, second rule should allow access")
	}
	if policy.HasAccessTo("secret/bars/zip") {
		t.Error("expected: false, got: true, second rule should deny access")
	}
	if !policy.HasAccessTo("secret/zip-zap") {
		t.Error("expected: true, got: false, third rule should allow access")
	}
	if !policy.HasAccessTo("secret/zip-zap/zong") {
		t.Error("expected: true, got: false, third rule should allow access")
	}
	if policy.HasAccessTo("secret/zip/zap") {
		t.Error("expected: false, got: true, third rule should deny access")
	}
}

func TestPolicy_HasAccessTo3(t *testing.T) {
	input := "# Permit reading the \"teamb\" path under any top-level path under secret/\npath \"secret/+/teamb\" {\n  capabilities = [\"read\"]\n}\n# Permit reading secret/foo/bar/teamb, secret/bar/foo/teamb, etc.\npath \"secret/+/+/teamb\" {\n  capabilities = [\"read\"]\n}"
	policy, err := models.FromHCL("policy_testing", []byte(input))
	if err != nil {
		t.Error(err)
	}
	if !policy.HasAccessTo("secret/foo/teamb") {
		t.Error("expected: true, got: false")
	}
	if !policy.HasAccessTo("secret/bar/foo/teamb") {
		t.Error("expected: true, got: false")
	}
}
