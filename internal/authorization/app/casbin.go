package app

import (
	_ "embed"
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
)

//go:embed model.conf
var modelConf string

type CasbinAdapter any

func NewCasbinEnforcer(
	adapter CasbinAdapter,
) (*casbin.TransactionalEnforcer, error) {
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

var ProvideCasbinEnforcer = NewCasbinEnforcer
