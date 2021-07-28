package data

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"read2succeed/utils"
	"strconv"
	"strings"
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
	UserID     int       `json:"user_id,omitempty"`
	ReaderName string    `json:"reader"`
	BookAuthor string    `json:"author"`
	BookTitle  string    `json:"title"`
	Day        string    `json:"day"`
	Duration   int       `json:"duration"`
	CreatedOn  time.Time `json:"-"`
}

// Reader - type for handling readers
type Reader struct {
	ID     int    `json:"id,omitempty"`
	UserID int    `json:"user_id,omitempty"`
	Name   string `json:"name"`
}

type Readers []Reader

// Group ...
type Group struct {
	ID         int
	UserID     int
	Name       string
	AccessCode string
	Status     string
	CreatedOn  time.Time
}

// GroupReaders ...
type GroupReaders struct {
	ID         int
	GroupID    int
	GroupName  string
	ReaderID   int
	ReaderName string
}

//DataStore - db operations
type DataStore struct {
	DB *sql.DB
	L  *log.Logger
}

// GetSQLiteVersion -
func (ds *DataStore) GetSQLiteVersion() (string, error) {
	query := `SELECT sqlite_version()`

	//var row *sql.Row
	row := ds.DB.QueryRow(query)

	var version string
	err := row.Scan(&version)

	return version, err
}

// CreateUser - add new user into db
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
		WHERE email = ?`
	row := ds.DB.QueryRow(query, email)

	if row.Err() != nil {
		ds.L.Println(row.Err())
		return nil, row.Err()
	}
	//ds.L.Println(query, email)

	//var day, created, duration string
	var userID, created string
	var u AuthUser
	err := row.Scan(&userID, &u.Name, &u.Email, &u.Pass, &created)
	if err != nil {
		ds.L.Println("nope...")
		ds.L.Println("****", err)
		return nil, err
	}
	UserID, _ := strconv.Atoi(userID)
	u.ID = UserID
	t, _ := time.Parse("2006-01-02T15:04:05Z", created)
	u.CreatedOn = t

	return &u, nil
}

// GetUserByID - return given user
func (ds *DataStore) GetUserByID(user_id int) (*AuthUser, error) {

	query := `
        SELECT user_id, name, email, passw, created
		FROM auth_user
		WHERE user_id = ?`
	row := ds.DB.QueryRow(query, user_id)

	if row.Err() != nil {
		ds.L.Println(row.Err())
		return nil, row.Err()
	}

	var userID, created string
	var u AuthUser
	err := row.Scan(&userID, &u.Name, &u.Email, &u.Pass, &created)
	if err != nil {
		ds.L.Println("nope...")
		ds.L.Println("****", err)
		return nil, err
	}
	UserID, _ := strconv.Atoi(userID)
	u.ID = UserID
	t, _ := time.Parse("2006-01-02T15:04:05Z", created)
	u.CreatedOn = t

	return &u, nil
}

// AddReading - add new reading entry into the db
func (ds *DataStore) AddReading(r *Reading) error {

	reader, err := ds.GetReaderByName(r.ReaderName)
	if err != nil {
		return err
	}
	fmt.Printf("found reader: %+v", reader)

	query := `
        INSERT INTO readings (user_id, reader, reader_id, book_author, book_title, day, duration, created)
        VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'))
    `

	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// XXX reader_id is not 1.. must fix this!!!
	res, err := stmt.Exec(r.UserID, r.ReaderName, reader.ID, r.BookAuthor, r.BookTitle, r.Day, r.Duration)
	if err != nil {
		return err
	}
	rowNum, _ := res.RowsAffected()
	ds.L.Println(" -- added new reading to DB: ", rowNum)

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

type ReaderStat struct {
	ReaderName string
	Name       string
	Labels     []string
	Values     []int
}

//TotalReadingStat - ....
type TotalReadingStat struct {
	ReaderName    string
	TotalDuration int
}

type DailyReaderStat struct {
	ReaderName string
	Label      string
	Value      int
}

//DailyReadingStats
type DailyReadingStats map[string][]DailyReaderStat

// GetStatsTotals - retrieves all records or of the given reader passed as args
func (ds *DataStore) GetStatsTotals(userID int) ([]TotalReadingStat, error) {
	totals := []TotalReadingStat{}
	query := `SELECT sum(duration) AS total, reader
		FROM readings
		WHERE user_id = ?
		GROUP BY reader
		ORDER BY total DESC`

	var rows *sql.Rows
	var err error
	rows, err = ds.DB.Query(query, userID)

	//fmt.Println(query, userID)
	// fmt.Printf("%#v", args)

	if err != nil {
		return totals, err
	}
	defer rows.Close()
	var stat TotalReadingStat

	//var created string
	for rows.Next() {
		rows.Scan(&stat.TotalDuration, &stat.ReaderName)
		totals = append(totals, stat)
	}

	return totals, nil
}

// GetStatsDaily - retrieves daily readers' stats for the pas 30 days of the given user
func (ds *DataStore) GetStatsDaily(userID int) (DailyReadingStats, error) {
	//type DailyReadingStats map[string][]DailyReaderStat
	dailyStats := DailyReadingStats{}

	query := `WITH RECURSIVE last31days(date) AS (
        VALUES(DATE('now', '-31 day', 'localtime'))
        UNION ALL
        SELECT DATE(date, '+1 day')
        FROM last31days
        WHERE date <= date('now')
        ), reader_readings(day, reader, duration) AS (
            SELECT DATE(day) as day, reader, sum(duration)
            FROM readings
            WHERE DATE('now', '-31 day', 'localtime') < DATE(day) AND user_id = ?
            GROUP BY day, reader
        )
        SELECT date, CASE WHEN reader IS NULL THEN '-' ELSE reader END AS reader,
            CASE WHEN duration IS NULL THEN 0 ELSE duration END AS daily_duration
        FROM last31days LEFT JOIN reader_readings ON date = day
        WHERE date <= CURRENT_DATE`

	var rows *sql.Rows
	var err error
	rows, err = ds.DB.Query(query, userID)

	//fmt.Println(query, userID)
	// fmt.Printf("%#v", args)

	if err != nil {
		return dailyStats, err
	}
	defer rows.Close()

	var entry DailyReaderStat
	for rows.Next() {
		rows.Scan(&entry.Label, &entry.ReaderName, &entry.Value)
		if _, ok := dailyStats[entry.Label]; !ok {
			dailyStats[entry.Label] = []DailyReaderStat{entry}
		} else {
			dailyStats[entry.Label] = append(dailyStats[entry.Label], entry)
		}
		fmt.Printf("###\t%+v\n", entry)
	}

	fmt.Printf("%+v", dailyStats)

	return dailyStats, nil
}

// GetUserReaders - retrieves all readers attached to this user
func (ds *DataStore) GetUserReaders(userID int) ([]Reader, error) {
	var readers []Reader
	query := `SELECT reader_id, name FROM readers
		WHERE user_id = ? ORDER BY name ASC`

	var rows *sql.Rows
	var err error
	rows, err = ds.DB.Query(query, userID)

	//fmt.Println(query, userID)
	// fmt.Printf("%#v", args)

	if err != nil {
		return readers, err
	}
	defer rows.Close()
	var reader Reader

	for rows.Next() {
		rows.Scan(&reader.ID, &reader.Name)
		readers = append(readers, reader)
	}

	return readers, nil
}

// AddReader - add new reader into the db
func (ds *DataStore) AddReader(r *Reader) error {
	query := `INSERT INTO readers (user_id, name, created)
        VALUES (?, ?, datetime('now','localtime'))`

	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(r.UserID, r.Name)
	if err != nil {
		return err
	}
	rowNum, _ := res.RowsAffected()
	ds.L.Println(" -- added new reader to DB: ", rowNum)

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = int(id)

	return nil
}

// GetReaderByName - fetch reader by name
func (ds *DataStore) GetReaderByName(name string) (Reader, error) {
	var reader Reader
	query := `SELECT reader_id, user_id, name FROM readers WHERE name = ?`

	row := ds.DB.QueryRow(query, name)
	//rows.Scan(&reader.ID, &reader.Name)
	err := row.Scan(&reader.ID, &reader.UserID, &reader.Name)
	if err != nil {
		return reader, err
	}

	return reader, nil
}

// AddGroup - add new group
func (ds *DataStore) AddGroup(g *Group) error {
	query := `
	INSERT INTO groups (user_id, name, code)
	VALUES (?, ?, ?)`
	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rand.Seed(time.Now().UnixNano())
	g.AccessCode = utils.RandStringRunes(5)

	res, err := stmt.Exec(g.UserID, g.Name, g.AccessCode)
	if err != nil {
		return err
	}
	rowNum, _ := res.RowsAffected()
	ds.L.Println(" -- added group to DB: ", rowNum)

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	g.ID = int(id)

	return nil
}

// UpdateGroup - update group
func (ds *DataStore) UpdateGroup(g *Group) error {
	query := `UPDATE groups SET name = ?, status = ? WHERE id = ?`
	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(g.Name, g.Status, g.ID)
	if err != nil {
		return err
	}
	rowNum, _ := res.RowsAffected()
	ds.L.Println(" -- update groups: ", rowNum)

	return nil
}

// GroupAddReader - add reader to group
func (ds *DataStore) GroupAddReader(groupID, readerID int) error {
	query := `INSERT INTO group_readers (group_id, reader_id) VALUES (?, ?)`
	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(groupID, readerID)
	if err != nil {
		return err
	}

	return nil
}

// GetGroupsAndReaders - retrieves all user's readers groups
func (ds *DataStore) GetGroupsAndReaders(userID int) (map[string][]Reader, error) {
	groups := map[string][]Reader{}

	query := `SELECT g.id, g.name, group_concat(r.reader_id||'.'||r.name)
	FROM groups g
	JOIN group_readers gr ON g.id = gr.group_id
	JOIN readers r ON gr.reader_id = r.reader_id
	WHERE r.user_id = ?
	GROUP BY g.id, g.name`

	var rows *sql.Rows
	var err error
	rows, err = ds.DB.Query(query, userID)

	if err != nil {
		return groups, err
	}
	defer rows.Close()

	var gID, gName, readerData string

	for rows.Next() {
		rows.Scan(&gID, &gName, &readerData)
		//readers = append(readers, reader)

		readersList := strings.Split(readerData, ",")
		groups[gName] = []Reader{}
		for _, r := range readersList {
			readerInfo := strings.Split(r, ".")
			readerID, _ := strconv.Atoi(readerInfo[0])
			reader := Reader{ID: readerID, UserID: userID, Name: readerInfo[1]}
			groups[gName] = append(groups[gName], reader)
		}
		fmt.Printf("###\t%s: %+v\n", gName, groups[gName])
	}

	return groups, nil
}

// GetUserGroups - retrieves user's groups
func (ds *DataStore) GetUserGroups(userID int) ([]Group, error) {
	var groups []Group

	query := `SELECT id, name, code, status, created
	FROM groups WHERE user_id = ?`

	var rows *sql.Rows
	var err error
	rows, err = ds.DB.Query(query, userID)

	if err != nil {
		return groups, err
	}
	defer rows.Close()

	var gID, gName, gCode, gStatus, gCreated string

	for rows.Next() {
		rows.Scan(&gID, &gName, &gCode, &gStatus, gCreated)
		t, _ := time.Parse("2006-01-02T15:04:05Z", gCreated)
		groupID, _ := strconv.Atoi(gID)

		groups = append(groups, Group{
			ID:         groupID,
			UserID:     userID,
			Name:       gName,
			AccessCode: gCode,
			Status:     gStatus,
			CreatedOn:  t,
		})
	}

	return groups, nil
}

// GetGroupByID - retrieves group
func (ds *DataStore) GetGroupByID(groupID int) (Group, error) {
	var g Group

	query := `SELECT name, user_id, code, status, created FROM groups WHERE id = ?`
	row := ds.DB.QueryRow(query, groupID)

	var gCreated, gUserID string

	err := row.Scan(&g.Name, &gUserID, &g.AccessCode, &g.Status, &gCreated)
	if err != nil {
		return g, nil
	}
	g.UserID, _ = strconv.Atoi(gUserID)
	g.CreatedOn, _ = time.Parse("2006-01-02T15:04:05Z", gCreated)
	g.ID = groupID

	return g, nil
}
