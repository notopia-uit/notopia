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
		// Each test gets an isolated in-memory database (no cache=shared to ensure isolation)
		db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to in-memory database: %w", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get sql.DB: %w", err)
		}
		sqlDB.SetMaxOpenConns(1)
		t.Cleanup(func() {
			_ = sqlDB.Close()
		})

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
		if err := app.LoadPoliciesFromString(enforcer, policy); err != nil {
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
