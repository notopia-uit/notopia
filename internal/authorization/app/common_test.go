package app_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	stringadapter "github.com/casbin/casbin/v3/persist/string-adapter"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"gorm.io/gorm"
)

type GetLocalEnforcerParams struct {
	LoadTestPolicies bool
	UseTransaction   bool
}

func GetLocalEnforcer(t testing.TB, params *GetLocalEnforcerParams) (*casbin.TransactionalEnforcer, error) {
	t.Helper()

	if params == nil {
		params = &GetLocalEnforcerParams{}
	}

	m, err := model.NewModelFromString(app.ModelConf)
	if err != nil {
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}

	if params.UseTransaction {
		// Use GORM with in-memory SQLite for transaction support
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to in-memory database: %w", err)
		}

		adapter, err := gormadapter.NewAdapterByDB(db)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize GORM adapter: %w", err)
		}

		enforcer, err := casbin.NewTransactionalEnforcer(m, adapter)
		if err != nil {
			return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
		}

		// Build combined policy CSV
		policy := app.PolicyCSV
		if params.LoadTestPolicies {
			policy += "\n" + app.PolicyTestCSV
		}

		// Load policies using stringadapter to maintain compatibility with g2 rules
		if err := loadPoliciesFromString(enforcer, policy); err != nil {
			return nil, fmt.Errorf("failed to load policies: %w", err)
		}

		return enforcer, nil
	}

	// Use string adapter (non-transactional)
	policy := app.PolicyCSV
	if params.LoadTestPolicies {
		policy += "\n" + app.PolicyTestCSV
	}
	adapter := stringadapter.NewAdapter(policy)

	enforcer, err := casbin.NewTransactionalEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}
	return enforcer, nil
}

// loadPoliciesFromString loads policies from a CSV string using stringadapter,
// then copies them to the enforcer's backend to maintain g2 rule compatibility
func loadPoliciesFromString(enforcer *casbin.TransactionalEnforcer, csv string) error {
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
