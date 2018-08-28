package data

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type ComputerRecord struct {
	ComputerID int
	Username   string
	FsgID      int
}

// LookupComputer returns true if a computer exists in the cache with the specified cid
// or else it returns false
func LookupComputer(computer_id int) bool {
	var count int
	rows, err := Db.Query(fmt.Sprintf("select count(computer_id) from computers where computer_id='%d'", computer_id))
	if err != nil {
		log.Warn(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			log.Warn(err)
		}
	}
	if count != 1 {
		return false
	}
	return true
}

// GetComputerUser returns the Username of a computer record
func GetComputerUser(computer_id int) string {
	var username string
	rows, err := Db.Query(fmt.Sprintf("select username from computers where computer_id='%d'", computer_id))
	if err != nil {
		log.Warn(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&username)
		if err != nil {
			log.Warn(err)
		}
	}
	return username
}

// InsertComputer inserts a new computer record.
func InsertComputer(computer_id, fsg_id int, unique_id string) error {
	tx, err := Db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into computers(computer_id, username, fsg_id) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(computer_id, unique_id, fsg_id)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// UpdateComputerUser udpates a computer record's user.
func UpdateComputerUser(computer_id int, unique_id string) error {
	tx, err := Db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("update computers set username = ? where computer_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(unique_id, computer_id)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// UpdateComputer udpates a computer record.
func UpdateComputer(computer_id, fsg_id int, unique_id string) error {
	tx, err := Db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("update computers set username = ?, fsg_id = ? where computer_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(unique_id, fsg_id, computer_id)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// FindUnmatchedComputers returns a list of computers from the cache that
// belong to the specified unique_id that don't have the specified fsg_id
func FindUnmatchedComputers(username string, fsg_id int) ([]ComputerRecord, error) {
	var count int
	computers := []ComputerRecord{}
	stmt := fmt.Sprintf("select count(*) from computers where username='%s' and fsg_id!='%d'",
		username,
		fsg_id,
	)
	row := Db.QueryRow(stmt)
	err := row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	if count < 1 {
		return nil, errors.New(fmt.Sprintf("No Unmatched Computers Found for %v with %d\n", username, fsg_id))
	}

	stmt = fmt.Sprintf("select * from computers where username='%s' and fsg_id!='%d'",
		username,
		fsg_id,
	)
	rows, err := Db.Query(stmt)
	if err != nil {
		log.Warn(err)
	}
	defer rows.Close()
	for rows.Next() {
		var cr ComputerRecord
		err = rows.Scan(&cr.ComputerID, &cr.Username, &cr.FsgID)
		if err != nil {
			log.Warn(err)
		}
		computers = append(computers, cr)
	}
	return computers, nil
}
