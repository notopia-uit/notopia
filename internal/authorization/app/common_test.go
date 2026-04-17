package app_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	stringadapter "github.com/casbin/casbin/v3/persist/string-adapter"
)

//go:embed model.conf
var modelConf string

//go:embed policy.csv
var policyCSV string

//go:embed policy_test.csv
var policyTestCSV string

func GetLocalEnforcer(t testing.TB, loadTestPolicies bool) (*casbin.TransactionalEnforcer, error) {
	t.Helper()

	policy := policyCSV
	if loadTestPolicies {
		policy += "\n" + policyTestCSV
	}
	adapter := stringadapter.NewAdapter(policy)
	model, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}
	enforcer, err := casbin.NewTransactionalEnforcer(model, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}
	return enforcer, nil
}
