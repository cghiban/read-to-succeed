package data

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type Reading struct {
	ID         uint      `json:"id,omitempty"`
	ReaderName string    `json:"reader"`
	BookAuthor string    `json:"author"`
	BookTitle  string    `json:"title"`
	Day        string    `json:"day"`
	Duration   int       `json:"duration"`
	CreatedOn  time.Time `json:-`
}

type DataStore struct {
	DB *sql.DB
	L  *log.Logger
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
	ds.L.Println(" -- added videos to DB: ", rowNum)

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = uint(id)

	return nil
}

// ListReadings - retrieves all records or of the given reader passed as args
func (ds *DataStore) ListReadings(args ...string) ([]Reading, error) {
	readings := []Reading{}
	queryFmt := `
        SELECT id, reader, book_author, book_title, day, duration, created
        FROM readings %s
        ORDER BY id desc
    `
	var query string
	var rows *sql.Rows
	var err error
	if len(args) == 1 && args[0] != "" {
		where := "WHERE reader = ?"
		query = fmt.Sprintf(queryFmt, where)
		rows, err = ds.DB.Query(query, args[0])
	} else {
		query = fmt.Sprintf(queryFmt, "")
		rows, err = ds.DB.Query(query)
	}

	// fmt.Println(query, args[0])
	// fmt.Printf("%#v", args)

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
