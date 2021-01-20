package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type DataStore struct {
	DB *sql.DB
	l  *log.Logger
}

func (ds *DataStore) AddReading(r *Reading) error {
	query := `                                                                                      
        INSERT INTO readings (reader, book_author, book_title, day, duration, created)                                  
        VALUES (?, ?, ?, ?, ?, datetime('now','localtime'))                                                                         
    `

	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(r.ReaderName, r.BookAuthor, r.BookTitle, r.Day, r.Duration)
	if err != nil {
		return err
	}
	rowNum, _ := res.RowsAffected()
	ds.l.Println(" -- added videos to DB: ", rowNum)

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = uint(id)

	return nil
}

func (ds *DataStore) ListReadings() ([]Reading, error) {
	query := `                                                                                      
        SELECT id, reader, book_author, book_title, day, duration, created
        FROM readings
        ORDER BY id desc
    `
	readings := []Reading{}

	rows, err := ds.DB.Query(query)
	if err != nil {
		return readings, err
	}
	defer rows.Close()
	var r Reading

	//var day, created, duration string
	var created string
	for rows.Next() {
		//rows.Scan(&r.ID, &r.ReaderName, &r.BookAuthor, &r.BookTitle, &day, &duration, &created)
		rows.Scan(&r.ID, &r.ReaderName, &r.BookAuthor, &r.BookTitle, &r.Day, &r.Duration, &created)

		//ds.l.Println(day, duration, created)
		t, _ := time.Parse("2006-01-02T15:04:05Z", created)
		r.CreatedOn = t
		/*t, _ = time.Parse("2006-01-02T00:00:00Z", day)
		r.Day = t
		r.Duration, _ = time.ParseDuration(duration)*/

		//ds.l.Println(r)
		readings = append(readings, r)
	}

	return readings, nil
}
