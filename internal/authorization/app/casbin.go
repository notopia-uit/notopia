package app

import (
	_ "embed"
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
)

//go:embed model.conf
var ModelConf string

//go:embed policy.csv
var PolicyCSV string

//go:embed policy_test.csv
var PolicyTestCSV string

type CasbinAdapter any

func NewCasbinEnforcer(
	adapter CasbinAdapter,
) (*casbin.TransactionalEnforcer, error) {
	model, err := model.NewModelFromString(ModelConf)
	if err != nil {
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}
	enforcer, err := casbin.NewTransactionalEnforcer(model, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}
	return enforcer, nil
}

var ProvideCasbinEnforcer = NewCasbinEnforcer
