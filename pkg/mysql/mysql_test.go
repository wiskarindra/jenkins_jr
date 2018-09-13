package mysql_test

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/subosito/gotenv"

	"github.com/bukalapak/jenkins_jr/pkg/mysql"
)

func init() {
	gotenv.MustLoad(os.Getenv("GOPATH") + "/src/github.com/bukalapak/jenkins_jr/.env")
	os.Setenv("ENV", "test")
	os.Setenv("DATABASE_NAME", os.Getenv("DATABASE_TEST_NAME"))
	os.Setenv("DATABASE_PORT", os.Getenv("DATABASE_TEST_PORT"))
	os.Setenv("DATABASE_USERNAME", os.Getenv("DATABASE_TEST_USERNAME"))
	os.Setenv("DATABASE_PASSWORD", os.Getenv("DATABASE_TEST_PASSWORD"))
	os.Setenv("DATABASE_HOST", os.Getenv("DATABASE_TEST_HOST"))
}

func TestMySQLInitiated(t *testing.T) {
	db := mysql.Init()
	assert.IsType(t, &sqlx.DB{}, db)
}

func TestMySQLNotOpened(t *testing.T) {
	temp := os.Getenv("DATABASE_NAME")
	os.Setenv("DATABASE_NAME", "fake_database")
	defer os.Setenv("DATABASE_NAME", temp)

	assert.Panics(t, func() { mysql.Init() }, "mysql.Init() should raise panic")
}
