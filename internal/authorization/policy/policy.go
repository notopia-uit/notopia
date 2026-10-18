package policy

import (
	"fmt"
	"slices"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
)

func parsePolicyCSV(csv string) (model.Model, error) {
	m, err := model.NewModelFromString(ModelConf)
	if err != nil {
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}

	for line := range strings.SplitSeq(csv, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if err := persist.LoadPolicyLine(line, m); err != nil {
			return nil, fmt.Errorf("failed to parse policy line %q: %w", line, err)
		}
	}

	return m, nil
}

// syncRules adds every rule in csv to the enforcer. With prune, `p` rules
// missing from csv are removed. `g` rules are runtime memberships and never pruned.
func syncRules(e *casbin.TransactionalEnforcer, csv string, prune bool) error {
	desired, err := parsePolicyCSV(csv)
	if err != nil {
		return err
	}

	policies, err := desired.GetPolicy("p", "p")
	if err != nil {
		return fmt.Errorf("failed to read p rules: %w", err)
	}

	if prune {
		current, err := e.GetPolicy()
		if err != nil {
			return fmt.Errorf("failed to read current p rules: %w", err)
		}

		var stale [][]string
		for _, rule := range current {
			if !slices.ContainsFunc(policies, func(r []string) bool { return slices.Equal(r, rule) }) {
				stale = append(stale, rule)
			}
		}
		if len(stale) > 0 {
			if _, err := e.RemovePolicies(stale); err != nil {
				return fmt.Errorf("failed to remove stale p rules: %w", err)
			}
		}
	}

	if len(policies) > 0 {
		if _, err := e.AddPoliciesEx(policies); err != nil {
			return fmt.Errorf("failed to add p rules: %w", err)
		}
	}

	for _, ptype := range []string{"g", "g2"} {
		rules, err := desired.GetPolicy("g", ptype)
		if err != nil {
			return fmt.Errorf("failed to read %s rules: %w", ptype, err)
		}
		if len(rules) == 0 {
			continue
		}
		if _, err := e.AddNamedGroupingPoliciesEx(ptype, rules); err != nil {
			return fmt.Errorf("failed to add %s rules: %w", ptype, err)
		}
	}

	return nil
}

func Seed(e *casbin.TransactionalEnforcer) error {
	if err := syncRules(e, CSV, true); err != nil {
		return fmt.Errorf("failed to seed policies: %w", err)
	}
	return nil
}

func LoadFixtures(e *casbin.TransactionalEnforcer) error {
	if err := syncRules(e, FixturesCSV, false); err != nil {
		return fmt.Errorf("failed to load fixture policies: %w", err)
	}
	return nil
}

// MissingRules checks the enforcer's in-memory policy, which is loaded once at
// startup, so a seed applied afterwards is only picked up by a restart.
func MissingRules(e *casbin.TransactionalEnforcer) ([][]string, error) {
	desired, err := parsePolicyCSV(CSV)
	if err != nil {
		return nil, err
	}

	var missing [][]string

	policies, err := desired.GetPolicy("p", "p")
	if err != nil {
		return nil, fmt.Errorf("failed to read p rules: %w", err)
	}
	for _, rule := range policies {
		ok, err := e.HasPolicy(rule)
		if err != nil {
			return nil, fmt.Errorf("failed to check p rule: %w", err)
		}
		if !ok {
			missing = append(missing, append([]string{"p"}, rule...))
		}
	}

	rules, err := desired.GetPolicy("g", "g2")
	if err != nil {
		return nil, fmt.Errorf("failed to read g2 rules: %w", err)
	}
	for _, rule := range rules {
		ok, err := e.HasNamedGroupingPolicy("g2", rule)
		if err != nil {
			return nil, fmt.Errorf("failed to check g2 rule: %w", err)
		}
		if !ok {
			missing = append(missing, append([]string{"g2"}, rule...))
		}
	}

	return missing, nil
}
