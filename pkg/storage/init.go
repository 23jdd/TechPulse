package storage

import (
	"config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db sqlx.DB

func InitStore(config config.Config) {
	var err error
	db, err := sqlx.Open("mysql", config.DataSourceName)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
}
