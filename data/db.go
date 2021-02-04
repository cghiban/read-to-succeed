package data

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

// AuthUser - user
type AuthUser struct {
	ID        int
	Name      string
	Email     string
	Pass      string
	CreatedOn time.Time
}

func encryptPassword(rawPass string) string {
	return fmt.Sprintf("%x", sha256.Sum224([]byte(rawPass)))
}

// CheckPasswd - validates password for login
func (u AuthUser) CheckPasswd(rawPass string) bool {

	return u.Pass == encryptPassword(rawPass)
}

// Reading - type for handling readings data
type Reading struct {
	ID         int       `json:"id,omitempty"`
	UserID     int       `json:"user_id",omitempty`
	ReaderName string    `json:"reader"`
	BookAuthor string    `json:"author"`
	BookTitle  string    `json:"title"`
	Day        string    `json:"day"`
	Duration   int       `json:"duration"`
	CreatedOn  time.Time `json:-`
}

//DataStore - db operations
type DataStore struct {
	DB *sql.DB
	L  *log.Logger
}

func (ds *DataStore) CreateUser(u *AuthUser) error {
	query := `
	INSERT INTO auth_user (email, name, passw, created)
	VALUES (?, ?, ?, datetime('now','localtime'))
	`
	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	u.Pass = encryptPassword(u.Pass)
	//encPass := encryptPassword(u.Pass)

	res, err := stmt.Exec(u.Email, u.Name, u.Pass)
	if err != nil {
		return err
	}
	//rowNum, _ := res.RowsAffected()
	//ds.L.Println(" -- added videos to DB: ", rowNum)

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = int(id)

	return nil
}

// GetUser - retrieves all records or of the given reader passed as args
func (ds *DataStore) GetUser(email string) (*AuthUser, error) {

	query := `
        SELECT user_id, name, email, passw, created
		FROM auth_user
		WHERE email = ?
    `
	row := ds.DB.QueryRow(query, email)

	if row.Err() != nil {
		ds.L.Println(row.Err())
		return nil, row.Err()
	}
	ds.L.Println(query, email)

	//var day, created, duration string
	var userID, created string
	var u AuthUser
	err := row.Scan(&userID, &u.Name, &u.Email, &u.Pass, &created)
	if err != nil {
		ds.L.Println("****", err)
		return nil, err
	}
	UserID, _ := strconv.Atoi(userID)
	u.ID = UserID
	t, _ := time.Parse("2006-01-02T15:04:05Z", created)
	u.CreatedOn = t

	ds.L.Printf("**** %#v\n", u)

	return &u, nil
}

// AddReading - add new reading entry into the db
func (ds *DataStore) AddReading(r *Reading) error {
	query := `
        INSERT INTO readings (user_id, reader, reader_id, book_author, book_title, day, duration, created)
        VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'))
    `

	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(r.UserID, r.ReaderName, 1, r.BookAuthor, r.BookTitle, r.Day, r.Duration)
	if err != nil {
		return err
	}
	rowNum, _ := res.RowsAffected()
	ds.L.Println(" -- added videos to DB: ", rowNum)

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = int(id)

	return nil
}

// ListUserReadings - retrieves all records or of the given reader passed as args
func (ds *DataStore) ListUserReadings(userID int, args ...string) ([]Reading, error) {
	readings := []Reading{}
	queryFmt := `
        SELECT id, reader, book_author, book_title, day, duration, created
        FROM readings WHERE user_id = ? %s
        ORDER BY id desc
    `
	var query string
	var rows *sql.Rows
	var err error
	if len(args) == 1 && args[0] != "" {
		where := " AND reader = ? "
		query = fmt.Sprintf(queryFmt, where)
		rows, err = ds.DB.Query(query, userID, args[0])
	} else {
		query = fmt.Sprintf(queryFmt, "")
		rows, err = ds.DB.Query(query, userID)
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
