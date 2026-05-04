package app_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	stringadapter "github.com/casbin/casbin/v3/persist/string-adapter"
	"github.com/notopia-uit/notopia/internal/authorization/app"
)

func GetLocalEnforcer(t testing.TB, loadTestPolicies bool) (*casbin.TransactionalEnforcer, error) {
	t.Helper()

	policy := app.PolicyCSV
	if loadTestPolicies {
		policy += "\n" + app.PolicyTestCSV
	}
	adapter := stringadapter.NewAdapter(policy)
	model, err := model.NewModelFromString(app.ModelConf)
	if err != nil {
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}
	enforcer, err := casbin.NewTransactionalEnforcer(model, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}
	return enforcer, nil
}
