package data

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

var Db *sql.DB

func init() {
	var err error
	boolPtr := flag.Bool("load", false, "use -load to load the existing cache.db")
	flag.Parse()
	if *boolPtr != true {
		os.Remove("./cache.db")
		fmt.Println("cirrup/data: initializing db")
	}
	Db, err = sql.Open("sqlite3", "./cache.db")
	if err != nil {
		log.Fatal(err)
	}
	sqlStmt := `
 	create table if not exists users 
 	(unique_id varchar(30) not null primary key, affiliation varchar(30), timestamp datetime);
 	create table if not exists computers 
 	(computer_id int not null primary key, username varchar(30) not null, fsg_id int not null);
 	delete from users;
 	delete from computers;
 	`
	if *boolPtr != true {
		_, err = Db.Exec(sqlStmt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cirrup/data: %q: %s\n", err, sqlStmt)
			return
		}
	}
	return
}
