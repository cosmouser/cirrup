package data

import (
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	Trace      *log.Logger
	Info       *log.Logger
	Warn       *log.Logger
	Error      *log.Logger
	ConfigPath string
)

func logs(traceHandle, infoHandle, warnHandle, errHandle io.Writer) {
	Trace = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime)
	Info = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime)
	Warn = log.New(warnHandle, "WARNING: ", log.Ldate|log.Ltime)
	Error = log.New(errHandle, "ERROR: ", log.Ldate|log.Ltime)
}

var Db *sql.DB

func init() {
	var (
		err      error
		dbExists bool = true
	)
	intPtr := flag.Int("v", 3, "-v sets the verbosity level, scale is 1-4. 4 is most verbose")
	configPath := flag.String("config", "./config.toml", "use -config to specify the config file to load")
	dbPath := flag.String("dbpath", "./cache.db", "use -dbpath to point to a database to use. If the database does not exist, cirrup will create one")
	flag.Parse()
	switch *intPtr {
	case 1:
		logs(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	case 2:
		logs(ioutil.Discard, ioutil.Discard, os.Stderr, os.Stderr)
	case 3:
		logs(ioutil.Discard, os.Stdout, os.Stderr, os.Stderr)
	case 4:
		logs(os.Stdout, os.Stdout, os.Stderr, os.Stderr)
	default:
		logs(ioutil.Discard, os.Stdout, os.Stderr, os.Stderr)
	}
	if *configPath == "" {
		Error.Fatalf("please specify a config to load\n")
	} else {
		ConfigPath = *configPath
	}
	if *dbPath == "" {
		Error.Fatalf("please specify a database file to load\n")
	}
	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		dbExists = false
	}
	Db, err = sql.Open("sqlite3", *dbPath)
	if err != nil {
		Error.Fatal(err)
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
			Error.Printf("cirrup/data: %q: %s\n", err, sqlStmt)
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
		Error.Fatal(err)
	}
	row = Db.QueryRow("pragma page_size")
	err = row.Scan(&pageSize)
	if err != nil {
		Error.Fatal(err)
	}
	return numPages * pageSize
}
