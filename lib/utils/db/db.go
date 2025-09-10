package db

import (
	"wallet/lib/utils/logger"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dbDsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dbDsn), &gorm.Config{
		Logger: slogGorm.New(
			slogGorm.WithHandler(logger.Get().Handler()), // since v1.3.0
			slogGorm.WithTraceAll(),
		),
	})

}
