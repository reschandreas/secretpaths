package models

import (
	"github.com/hashicorp/hcl/v2/hclsimple"
	"regexp"
	"sort"
	"strings"
)

type Policy struct {
	Name  string `hcl:"name,label"`
	Rules []Rule `hcl:"path,block"`
}

type Rule struct {
	Path       string   `hcl:"path,label"`
	Regex      string   // helper field, makes it easier to match paths
	Capability []string `hcl:"capabilities"`
}

func NewRule(path string, capability []string) Rule {
	return Rule{
		Path:       path,
		Regex:      PathToRegex(path),
		Capability: capability,
	}
}

func NewPolicy(name string, rules []Rule) Policy {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].IsHigherPriorityThan(rules[j])
	})
	return Policy{
		Name:  name,
		Rules: rules,
	}
}

func (p Policy) AmountOfPolicies() int {
	return len(p.Rules)
}

func (p Policy) ToRequest() (hcl string) {
	for _, rule := range p.Rules {
		hcl += "path \"" + rule.Path + "\" {\n\tcapabilities = [\"" + strings.Join(rule.Capability, "\", \"") + "\"]\n}\n\n"
	}
	return
}

func (p Policy) HasAccessTo(path string) bool {
	for _, rule := range p.Rules {
		if rule.HasAccessTo(path) {
			return true
		}
		if rule.SpecificallyDeniesAccessTo(path) {
			return false
		}
	}
	return false
}

func (p Policy) containsDenyCapability() bool {
	for _, rule := range p.Rules {
		if contains(rule.Capability, "deny") {
			return true
		}
	}
	return false
}

func contains(slice []string, needle string) bool {
	for _, element := range slice {
		if element == needle {
			return true
		}
	}
	return false
}

func (r Rule) SpecificallyDeniesAccessTo(path string) bool {
	if r.Path != path {
		return false
	}
	return contains(r.Capability, "deny")
}

func (r Rule) HasAccessTo(path string) bool {
	matched, err := regexp.Match(r.Regex, []byte(path))
	if err != nil {
		return false
	}
	if matched {
		return !contains(r.Capability, "deny")
	}
	return matched
}

func FromHCL(name string, hcl []byte) (policy Policy, err error) {
	err = hclsimple.Decode("policy.hcl", hcl, nil, &policy)
	if err != nil {
		return
	}
	policy.Name = name
	var rewrittenRules []Rule
	for _, rule := range policy.Rules {
		rule.Regex = PathToRegex(rule.Path)
		rewrittenRules = append(rewrittenRules, rule)
	}
	policy.Rules = rewrittenRules
	return policy, err
}

func PathToRegex(path string) string {
	regex := strings.ReplaceAll(path, "+", "[^/]+")
	// * is only allowed once and at the end of the path
	regex = strings.Replace(regex, "*", ".*", 1)
	regex += "$"
	return regex
}

// IsHigherPriorityThan Returns true if the rule is higher priority than the other rule, according to the following rules:
// 1. If the first wildcard (+) or glob (*) occurs earlier in P1, P1 is lower priority
// 2. If P1 ends in * and P2 doesn't, P1 is lower priority
// 3. If P1 has more + (wildcard) segments, P1 is lower priority
// 4. If P1 is shorter, it is lower priority
// 5. If P1 is smaller lexicographically, it is lower priority
func (r Rule) IsHigherPriorityThan(other Rule) bool {
	firstRegex := strings.IndexAny(r.Path, "*+")
	secondRegex := strings.IndexAny(other.Path, "*+")
	//If the first wildcard (+) or glob (*) occurs earlier in P1, P1 is lower priority
	if firstRegex != -1 && secondRegex != -1 {
		if firstRegex < secondRegex {
			return false
		}
	}
	//If P1 ends in * and P2 doesn't, P1 is lower priority
	if r.Path[len(r.Path)-1] == '*' && other.Path[len(other.Path)-1] != '*' {
		return false
	}
	//If P1 has more + (wildcard) segments, P1 is lower priority
	if strings.Count(r.Path, "+") > strings.Count(other.Path, "+") {
		return false
	}
	//If P1 is shorter, it is lower priority
	if len(r.Path) < len(other.Path) {
		return false
	}
	//If P1 is smaller lexicographically, it is lower priority
	if r.Path < other.Path {
		return false
	}
	return true
}
