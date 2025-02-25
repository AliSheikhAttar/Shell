package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	dbHost = "localhost"
	dbPort = "5432"
	dbName = "postgres"
	dbUser = "postgres"
	dbPass = "postgres"
)

var db *gorm.DB

func init() {
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		dbHost, dbPort, dbName, dbUser, dbPass)

	sesssion, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		err = fmt.Errorf("unable to connect to database: %w", err)
		panic(err)
	}

	db = sesssion

}

func GetDB() *gorm.DB {
	return db
}
