package app

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/log"
	"github.com/casbin/casbin/v3/model"
	"github.com/notopia-uit/notopia/internal/authorization/policy"
)

type CasbinAdapter any

func NewCasbinEnforcer(
	adapter CasbinAdapter,
	log log.Logger,
) (*casbin.TransactionalEnforcer, error) {
	model, err := model.NewModelFromString(policy.ModelConf)
	if err != nil {
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}
	enforcer, err := casbin.NewTransactionalEnforcer(model, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}
	enforcer.SetLogger(log)
	return enforcer, nil
}

var ProvideCasbinEnforcer = NewCasbinEnforcer
