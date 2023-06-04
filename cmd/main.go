package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if _, ok := os.LookupEnv("DBUSER"); ok == false {
		return
	}
	if _, ok := os.LookupEnv("DBPASS"); ok == false {
		return
	}

	// cfg := mysql.Config{
	// 	User:   os.Getenv("DBUSER"),
	// 	Passwd: os.Getenv("DBPASS"),
	// 	Net:    "tcp",
	// 	Addr:   "127.0.0.1:3306",
	// 	DBName: "apeironCache",
	// }

	db, err := sql.Open("mysql", "root:Z9@K^2aLQXFQHU58@tcp(127.0.0.1:3306)/apeironCache")
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("errrror")
		return
	}

	// names, err := db.Query("SELECT name FROM users")
	// if err != nil {
	// 	panic(err.Error())
	// }

	// fmt.Println("Getting names!")

	// for names.Next() {
	// 	var name string

	// 	err = names.Scan(&name)

	// 	fmt.Println(name)
	// }

	fmt.Println("Done!")
}
