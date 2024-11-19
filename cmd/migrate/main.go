package main

import (
	"github.com/iota-agency/iota-sdk/internal/configuration"
	"github.com/iota-agency/iota-sdk/pkg/dbutils"
	"gorm.io/gorm/logger"
)

func main() {
	db, err := dbutils.ConnectDB(configuration.Use().DBOpts, logger.Warn)
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	if err := dbutils.RunMigrations(sqlDB); err != nil {
		panic(err)
	}
}
