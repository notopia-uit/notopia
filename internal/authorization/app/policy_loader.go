package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	stringadapter "github.com/casbin/casbin/v3/persist/string-adapter"
)

// LoadPoliciesFromString loads policies from a CSV string into a Casbin enforcer.
// It uses a temporary string adapter to parse the CSV correctly (handles g, g2, g3 rules)
// and then copies all policies to the target enforcer.
func LoadPoliciesFromString(enforcer *casbin.TransactionalEnforcer, csv string) error {
	// Use string adapter to parse the CSV correctly (handles g2 rules properly)
	tempAdapter := stringadapter.NewAdapter(csv)
	tempModel := enforcer.GetModel()
	tempEnforcer, err := casbin.NewTransactionalEnforcer(tempModel, tempAdapter)
	if err != nil {
		return fmt.Errorf("failed to create temp enforcer: %w", err)
	}

	// Copy all policies from temp enforcer to our enforcer
	policies, _ := tempEnforcer.GetPolicy()
	if len(policies) > 0 {
		if _, err := enforcer.AddPolicies(policies); err != nil {
			return fmt.Errorf("failed to add policies: %w", err)
		}
	}

	// Copy all grouping policies (g - 3 arguments)
	gPolicies, _ := tempEnforcer.GetGroupingPolicy()
	if len(gPolicies) > 0 {
		if _, err := enforcer.AddGroupingPolicies(gPolicies); err != nil {
			return fmt.Errorf("failed to add grouping policies: %w", err)
		}
	}

	// Copy g2 grouping policies (2 arguments)
	g2Policies, _ := tempEnforcer.GetNamedGroupingPolicy("g2")
	if len(g2Policies) > 0 {
		if _, err := enforcer.AddNamedGroupingPolicies("g2", g2Policies); err != nil {
			return fmt.Errorf("failed to add g2 policies: %w", err)
		}
	}

	// Copy g3 grouping policies if they exist
	g3Policies, _ := tempEnforcer.GetNamedGroupingPolicy("g3")
	if len(g3Policies) > 0 {
		if _, err := enforcer.AddNamedGroupingPolicies("g3", g3Policies); err != nil {
			return fmt.Errorf("failed to add g3 policies: %w", err)
		}
	}

	return nil
}
