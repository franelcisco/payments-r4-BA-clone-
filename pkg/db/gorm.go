package db

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDBSQLHandler creates a new database handler for PostgreSQL using GORM.
func NewDBSQLHandler(conn string) (*gorm.DB, error) {
	sqlDB, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	// Ping the database to check if the connection is successful
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to the PostgreSQL database!")
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	return gormDB, nil
}
