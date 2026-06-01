package infra

import (
	"fmt"
	"log/slog"

	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/otel"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func NewGORMDB(
	databaseCfg *commonconfig.SQL,
	logger *slog.Logger,
	_ otel.Global,
) (*gorm.DB, func(), error) {
	gormLogger := slogGorm.New(
		slogGorm.WithHandler(logger.Handler()),
		slogGorm.WithTraceAll(),
	)
	db, err := gorm.Open(
		postgres.Open(databaseCfg.GetDSN()),
		&gorm.Config{
			Logger: gormLogger,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, nil, fmt.Errorf("failed to use OpenTelemetry plugin for db: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}
	cleanup := func() {
		if err := sqlDB.Close(); err != nil {
			logger.Error(
				"failed to close database connection",
				slog.Any("error", err),
			)
		}
	}
	return db, cleanup, nil
}

var ProvideGORMDB = NewGORMDB
