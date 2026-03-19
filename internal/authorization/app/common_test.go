package app_test

import (
	_ "embed"

	"github.com/casbin/casbin/v3"
	stringadapter "github.com/casbin/casbin/v3/persist/string-adapter"
)

//go:embed policy.csv
var policyCSV string

//go:embed policy_test.csv
var policyTestCSV string

func GetLocalEnforcer(loadTestPolicies bool) (*casbin.TransactionalEnforcer, error) {
	policy := policyCSV
	if loadTestPolicies {
		policy += "\n" + policyTestCSV
	}
	adapter := stringadapter.NewAdapter(policy)
	e, err := casbin.NewTransactionalEnforcer("model.conf", adapter)
	if err != nil {
		return nil, err
	}
	return e, nil
}
