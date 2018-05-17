package data

import (
	"fmt"
	"os"
)

type UserRecord struct {
	UniqueID    string
	Affiliation string
}

// LookupUser returns true if the user is already in the database or else false
func LookupUser(uid string) bool {
	var count int
	rows, err := Db.Query(fmt.Sprintf("select count(unique_id) from users where unique_id='%s'", uid))
	if err != nil {
		return false
	}
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cirrup/data: %v\n", err)
		}
	}
	if count != 1 {
		return false
	}
	return true
}

// GetUserAff gets the user's affiliation from the cache entry
func GetUserAff(uid string) string {
	var affiliation string
	rows, err := Db.Query(fmt.Sprintf("select affiliation from users where unique_id='%s'", uid))
	if err != nil {
		return ""
	}
	for rows.Next() {
		err = rows.Scan(&affiliation)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cirrup/data: %v\n", err)
		}
	}
	return affiliation
}

// InsertUser creates a new user record
func InsertUser(unique_id string, affiliation string) error {
	tx, err := Db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into users(unique_id, affiliation, timestamp) values(?, ?, julianday('now'))")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(unique_id, affiliation)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// CullUsers removes users added after a specified number of days
func CullUsers(days int) error {
	tx, err := Db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("delete from users where (julianday('now') - julianday(timestamp)) > ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()
	result, err := stmt.Exec(days)
	if err != nil {
		return err
	}
	tx.Commit()
	numCulled, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if numCulled > 0 {
		fmt.Printf("cirrup/data: %d user(s) culled\n", numCulled)
	}
	return nil
}
