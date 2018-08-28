package data

import (
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	ConfigPath string
)

var Db *sql.DB

func init() {
	var (
		err      error
		dbExists bool = true
	)
	configPath := flag.String("config", "./config.toml", "use -config to specify the config file to load")
	dbPath := flag.String("dbpath", "./cache.db", "use -dbpath to point to a database to use. If the database does not exist, cirrup will create one")
	flag.Parse()
	ConfigPath = *configPath
	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		dbExists = false
	}
	Db, err = sql.Open("sqlite3", *dbPath)
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
	if dbExists == false {
		_, err = Db.Exec(sqlStmt)
		if err != nil {
			log.WithFields(log.Fields{
				"sqlStmt": sqlStmt,
			}).Fatal(err)
			return
		}
	}
	return
}

func GetDBSize() float64 {
	var numPages, pageSize float64
	row := Db.QueryRow("pragma page_count")
	err := row.Scan(&numPages)
	if err != nil {
		log.Fatal(err)
	}
	row = Db.QueryRow("pragma page_size")
	err = row.Scan(&pageSize)
	if err != nil {
		log.Fatal(err)
	}
	return numPages * pageSize
}
