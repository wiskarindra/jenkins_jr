package mysql

import (
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Init returns connector to Spyro database
func Init() *sqlx.DB {
	dbUsername := os.Getenv("DATABASE_USERNAME")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbHost := os.Getenv("DATABASE_HOST")
	dbName := os.Getenv("DATABASE_NAME")
	dbPort := os.Getenv("DATABASE_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	env := os.Getenv("ENV")
	if env == "development" || env == "staging" {
		fmt.Println(fmt.Sprintf("Connecting to [USERNAME]:[PASSWORD]@(%s:%v)/%s?parseTime=true", dbHost, dbPort, dbName))
	}

	dataSourceName := fmt.Sprintf("%s:%v@(%s:%v)/%s?parseTime=true", dbUsername, dbPassword, dbHost, dbPort, dbName)
	db, _ := sqlx.Open("mysql", dataSourceName)
	if err := db.Ping(); err != nil {
		// https://stackoverflow.com/questions/32345124/why-does-sql-open-return-nil-as-error-when-it-should-not
		panic(err.Error())
	}

	if dp, err := strconv.Atoi(os.Getenv("DATABASE_POOL")); err == nil && dp > 0 {
		db.SetMaxIdleConns(dp)
	}
	db.SetConnMaxLifetime(time.Minute)
	db.SetMaxOpenConns(500)
	return db
}
