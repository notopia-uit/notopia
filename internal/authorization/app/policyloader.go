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
	policies, err := tempEnforcer.GetPolicy()
	if err != nil {
		return fmt.Errorf("failed to get policies from temp enforcer: %w", err)
	}
	if len(policies) > 0 {
		if _, err := enforcer.AddPolicies(policies); err != nil {
			return fmt.Errorf("failed to add policies: %w", err)
		}
	}

	// Copy all grouping policies (g - 3 arguments)
	gPolicies, err := tempEnforcer.GetGroupingPolicy()
	if err != nil {
		return fmt.Errorf("failed to get grouping policies from temp enforcer: %w", err)
	}
	if len(gPolicies) > 0 {
		if _, err := enforcer.AddGroupingPolicies(gPolicies); err != nil {
			return fmt.Errorf("failed to add grouping policies: %w", err)
		}
	}

	// Copy g2 grouping policies (2 arguments)
	g2Policies, err := tempEnforcer.GetNamedGroupingPolicy("g2")
	if err != nil {
		return fmt.Errorf("failed to get g2 policies from temp enforcer: %w", err)
	}
	if len(g2Policies) > 0 {
		if _, err := enforcer.AddNamedGroupingPolicies("g2", g2Policies); err != nil {
			return fmt.Errorf("failed to add g2 policies: %w", err)
		}
	}

	// Copy g3 grouping policies if they exist
	g3Policies, err := tempEnforcer.GetNamedGroupingPolicy("g3")
	if err != nil {
		return fmt.Errorf("failed to get g3 policies from temp enforcer: %w", err)
	}
	if len(g3Policies) > 0 {
		if _, err := enforcer.AddNamedGroupingPolicies("g3", g3Policies); err != nil {
			return fmt.Errorf("failed to add g3 policies: %w", err)
		}
	}

	return nil
}
