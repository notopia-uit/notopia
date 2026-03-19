package app_test

import (
	"github.com/casbin/casbin/v3"
	fileadapter "github.com/casbin/casbin/v3/persist/file-adapter"
)

func GetLocalEnforcer(loadTestPolicies bool) (*casbin.TransactionalEnforcer, error) {
	adapter := fileadapter.NewAdapter("../policy_test.csv")
	e, err := casbin.NewTransactionalEnforcer("../model.conf", adapter)
	if err != nil {
		return nil, err
	}
	if loadTestPolicies {
		err := adapter.LoadPolicy(e.GetModel())
		if err != nil {
			return nil, err
		}
		err = e.BuildRoleLinks()
		if err != nil {
			return nil, err
		}
	}
	return e, nil
}
