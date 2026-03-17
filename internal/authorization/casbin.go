package authorization

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

func NewCasbinEnforcer(
	databaseCfg *commonconfig.SQL,
) (casbin.IEnforcerContext, error) {
	adapter := gormadapter.NewAdapter(databaseCfg.Scheme, databaseCfg.GetDSN())
	enforcer, err := casbin.NewContextEnforcer("model.conf", adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}
	return enforcer, nil
}

var ProvideCasbinEnforcer = NewCasbinEnforcer
