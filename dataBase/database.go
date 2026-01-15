// Package database used for database connections
package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

type User struct {
	ID           int
	Name         string
	Surname      string
	PhoneNumber  string
	Balance      float64
	PasswordHash string
	Session      string
}

var UserName = "root"
var Password = "rustam1122"
var Address = "localhost"
var Port = "3306"
var DBName = "simple_bank"

func init() {
	var err error

	dsn := UserName + ":" + Password + "@tcp(" + Address + ":" + Port + ")/" + DBName + "?charset=utf8mb4&parseTime=True&loc=Local"

	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	if err = DB.Ping(); err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	log.Println("✅ Database connection established")
}

func Close() {
	if DB != nil {
		DB.Close()
		log.Println("📴 Database connection closed")
	}
}
