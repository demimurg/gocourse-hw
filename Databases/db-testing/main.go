package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const DSN = "root@tcp(localhost:3306)/golang?charset=utf8"

func main() {
	db, _ := sql.Open("mysql", DSN)
	db.Ping() // вот тут будет первое подключение к базе
	defer db.Close()

}
