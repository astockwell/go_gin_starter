package main

// import (
// 	"fmt"
//
// 	"gorm.io/driver/postgres"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// )

// bootstrapPostgresDb initializes a PostgreSQL database connection using GORM
// func bootstrapPostgresDb(connStr string) (*gorm.DB, error) {
// 	// Open connection to PostgreSQL database
// 	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
// 	if err != nil {
// 		return nil, fmt.Errorf("gorm.Open(postgres.Open()): %w", err)
// 	}
//
// 	err = db.AutoMigrate(&Book{})
// 	if err != nil {
// 		return nil, fmt.Errorf("db.AutoMigrate(): %w", err)
// 	}
//
// 	return db, nil
// }

// bootstrapSqliteDb initializes a SQLite database connection using GORM
// func bootstrapSqliteDb(connStr string) (*gorm.DB, error) {
// 	// Open connection to SQLite database
// 	db, err := gorm.Open(sqlite.Open(connStr), &gorm.Config{})
// 	if err != nil {
// 		return nil, fmt.Errorf("gorm.Open(sqlite.Open()): %w", err)
// 	}

// 	err = db.AutoMigrate(&Book{})
// 	if err != nil {
// 		return nil, fmt.Errorf("db.AutoMigrate(): %w", err)
// 	}

// 	return db, nil
// }
