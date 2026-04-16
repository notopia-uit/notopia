package infra

import (
	"fmt"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewGORMDB(databaseCfg *commonconfig.SQL) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseCfg.GetDSN()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

var ProvideGORMDB = NewGORMDB

func NewCasbinAdapter(
	gormDB *gorm.DB,
) (*gormadapter.Adapter, error) {
	adapter, err := gormadapter.NewTransactionalAdapterByDB(gormDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}
	return adapter, nil
}

var ProvideCasbinAdapter = NewCasbinAdapter
