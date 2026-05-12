package infra

import (
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v3/log"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/notopia-uit/notopia/pkg/casbin"
	"gorm.io/gorm"
)

func NewCasbinLogger(logger *slog.Logger) *casbin.SlogLogger {
	return casbin.NewSlogLogger(logger)
}

var ProvideCasbinLogger = NewCasbinLogger

func NewCasbinAdapter(
	gormDB *gorm.DB,
	logger log.Logger,
) (*gormadapter.Adapter, error) {
	adapter, err := gormadapter.NewTransactionalAdapterByDB(gormDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}
	return adapter, nil
}

var ProvideCasbinAdapter = NewCasbinAdapter
