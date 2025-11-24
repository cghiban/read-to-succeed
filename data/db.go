package data

import (
	"crypto/sha256"
	"database/sql"
	"errors"
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
	IsAdmin   bool
	CreatedOn time.Time
}

func encryptPassword(rawPass string) string {
	return fmt.Sprintf("%x", sha256.Sum224([]byte(rawPass)))
}

// CheckPasswd - validates password for login
func (u AuthUser) CheckPasswd(rawPass string) bool {

	// TODO probably not the best way to compare
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
	Pages      int       `json:"pages"`
	Note       string    `json:"note"`
	CreatedOn  time.Time `json:"-"`
}

// Reader - type for handling readers
type Reader struct {
	ID        int       `json:"id,omitempty"`
	UserID    int       `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	CreatedOn time.Time `json:"-"`
}

type Readers []Reader

// Group ...
type Group struct {
	ID         int       `json:"id,omitempty"`
	UserID     int       `json:"-"`
	Name       string    `json:"name"`
	AccessCode string    `json:"-"`
	Status     string    `json:"-"`
	CreatedOn  time.Time `json:"-"`
}

type UpdateGroup struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// GroupReaders ...
type GroupReaders struct {
	ID         int
	GroupID    int
	GroupName  string
	ReaderID   int
	ReaderName string
}

// DataStore - db operations
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
	INSERT INTO auth_user (email, name, passw, is_admin, created)
	VALUES (?, ?, ?, false, datetime('now','localtime'))
	`
	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	u.Pass = encryptPassword(u.Pass)

	res, err := stmt.Exec(u.Email, u.Name, u.Pass)
	if err != nil {
		return err
	}

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
        SELECT user_id, name, email, passw, is_admin, created
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
	err := row.Scan(&userID, &u.Name, &u.Email, &u.Pass, &u.IsAdmin, &created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ds.L.Printf("user [%s] not found\n", email)
		} else {
			ds.L.Println("****", err)
		}
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
        SELECT user_id, name, email, passw, is_admin, created
		FROM auth_user
		WHERE user_id = ?`
	row := ds.DB.QueryRow(query, user_id)

	if row.Err() != nil {
		ds.L.Println(row.Err())
		return nil, row.Err()
	}

	var userID, created string
	var u AuthUser
	err := row.Scan(&userID, &u.Name, &u.Email, &u.Pass, &u.IsAdmin, &created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ds.L.Printf("user %d not found\n", user_id)
		} else {
			ds.L.Println("****", err)
		}
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
	//fmt.Printf("found reader: %+v", reader)

	query := `
        INSERT INTO readings (user_id, reader, reader_id, book_author, book_title, day, duration, pages, note, created)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'))
    `

	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(r.UserID, r.ReaderName, reader.ID, r.BookAuthor, r.BookTitle, r.Day, r.Duration, r.Pages, r.Note)
	if err != nil {
		return err
	}
	//rowNum, _ := res.RowsAffected()
	//ds.L.Println(" -- added new reading to DB: ", rowNum)

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = int(id)

	return nil
}

type Pagination struct {
	Total  int
	Limit  int
	Offset int
}

type UserReadingsRequest struct {
	UserID int
	Reader string
	Limit  *int
	Offset *int
}

type UserReadingsResponse struct {
	Readings   []Reading
	Pagination Pagination
}

// ListUserReadings - retrieves all records or of the given reader passed as args
func (ds *DataStore) ListUserReadings(req UserReadingsRequest) (UserReadingsResponse, error) {
	output := UserReadingsResponse{}

	queryFmt := `
        SELECT id, reader, book_author, book_title, day, duration, pages, note, created
        FROM readings WHERE user_id = ? %s`
	var query, where string
	var rows *sql.Rows
	var err error
	var args []any
	if req.Reader != "" {
		where = " AND reader = ? "
		query = fmt.Sprintf(queryFmt, where)
		args = append(args, req.UserID, req.Reader)
	} else {
		query = fmt.Sprintf(queryFmt, "")
		args = append(args, req.UserID)
	}

	var limit, offset int
	if req.Limit != nil && *req.Limit <= 100 {
		limit = *req.Limit
	} else {
		limit = 20
	}

	if req.Offset != nil {
		offset = *req.Offset
	}

	query += fmt.Sprintf(" ORDER BY id desc LIMIT %d OFFSET %d", limit, offset)

	rows, err = ds.DB.Query(query, args...)
	if err != nil {
		return output, err
	}
	defer rows.Close()

	readings := []Reading{}
	var r Reading
	var created string
	for rows.Next() {
		rows.Scan(&r.ID, &r.ReaderName, &r.BookAuthor, &r.BookTitle, &r.Day, &r.Duration, &r.Pages, &r.Note, &created)
		t, _ := time.Parse("2006-01-02T15:04:05Z", created)
		r.CreatedOn = t
		readings = append(readings, r)
	}

	output.Readings = readings

	totalQuery := "SELECT count(*) FROM readings WHERE user_id = ? " + where
	totalQuery += " ORDER BY id desc"

	//ds.L.Println(totalQuery, args)

	row := ds.DB.QueryRow(totalQuery, args...)
	var total int
	err = row.Scan(&total)
	if err != nil {
		ds.L.Printf("can't query totals: %+v", err)
		return output, err
	}
	output.Pagination = Pagination{
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}

	return output, nil
}

type GroupReading struct {
	GroupID    int    `json:"group_id,omitempty"`
	GroupName  string `json:"group_name,omitempty"`
	ReaderID   int
	ReaderName string    `json:"reader"`
	BookAuthor string    `json:"author"`
	BookTitle  string    `json:"title"`
	Day        string    `json:"day"`
	Duration   int       `json:"duration"`
	Pages      int       `json:"pages"`
	CreatedOn  time.Time `json:"-"`
}

type GroupReadingsRequest struct {
	GroupID int
	UserID  int
	Limit   *int
	Offset  *int
}

type GroupReadingsResponse struct {
	GroupReadings []GroupReading
	Group         Group
	Pagination    Pagination
}

// ListUserGroupsReadings - retrieves all records of the given users' group(s)
// may pass group ID as a filter
func (ds *DataStore) ListUserGroupsReadings(req GroupReadingsRequest) (GroupReadingsResponse, error) {
	output := GroupReadingsResponse{}

	var limit, offset int
	if req.Limit != nil && *req.Limit <= 100 {
		limit = *req.Limit
	} else {
		limit = 20
	}

	if req.Offset != nil {
		offset = *req.Offset
	}

	queryFmt := `
		SELECT %s
		FROM groups g
		JOIN group_readers gr ON g.id = gr.group_id
		JOIN readings r ON gr.reader_id = r.reader_id
		WHERE g.id IN (
			-- my readers' groups
			SELECT group_id
			FROM group_readers
			WHERE reader_id IN (
				SELECT reader_id FROM readers
				WHERE user_id = ?
			) %s
		)`
	queryCols := `g.id, g.name, r.reader, r.reader_id, r.book_author, r.book_title,
	r.day, r.duration, r.pages, r.created`

	var query, where string
	var rows *sql.Rows
	var err error
	args := []any{}
	if req.GroupID != 0 {
		where = " AND group_id = ? "
		query = fmt.Sprintf(queryFmt, queryCols, where)
		args = append(args, req.UserID, req.GroupID)
	} else {
		query = fmt.Sprintf(queryFmt, queryCols, where)
		args = append(args, req.UserID)
	}

	query += fmt.Sprintf(" ORDER BY r.created DESC LIMIT %d OFFSET %d", limit, offset)
	//fmt.Println(query, req.UserID, req.GroupID)

	rows, err = ds.DB.Query(query, args...)

	if err != nil {
		return output, err
	}
	defer rows.Close()

	readings := []GroupReading{}
	var r GroupReading

	//var day, created, duration string
	var created string
	for rows.Next() {
		rows.Scan(
			&r.GroupID,
			&r.GroupName,
			&r.ReaderName,
			&r.ReaderID,
			&r.BookAuthor,
			&r.BookTitle,
			&r.Day,
			&r.Duration,
			&r.Pages,
			&created)

		//ds.l.Println(day, duration, created)
		t, _ := time.Parse("2006-01-02T15:04:05Z", created)
		r.CreatedOn = t
		/*t, _ = time.Parse("2006-01-02T00:00:00Z", day)
		r.Day = t
		r.Duration, _ = time.ParseDuration(duration)*/

		//ds.l.Println(r)
		readings = append(readings, r)
	}

	output.GroupReadings = readings

	totalQuery := fmt.Sprintf(queryFmt, "count(*)", where)
	//ds.L.Println(totalQuery, args)

	row := ds.DB.QueryRow(totalQuery, args...)
	var total int
	err = row.Scan(&total)
	if err != nil {
		ds.L.Printf("can't query totals: %+v", err)
		return output, err
	}
	output.Pagination = Pagination{
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}

	return output, nil
}

type ReaderStat struct {
	ReaderName string
	Name       string
	Labels     []string
	Values     []int
}

// TotalReadingStat - ....
type TotalReadingStat struct {
	ReaderName    string
	TotalDuration int
	TotalPages    int
}

type DailyReaderStat struct {
	ReaderName    string
	Label         string
	TotalDuration int
	TotalPages    int
}

// DailyReadingStats
type DailyReadingStats map[string][]DailyReaderStat

// GetStatsTotals - retrieves all records or of the given reader passed as args
func (ds *DataStore) GetStatsTotals(userID int) ([]TotalReadingStat, error) {
	totals := []TotalReadingStat{}
	query := `SELECT sum(duration) AS total, sum(pages) AS pages, reader
		FROM readings
		WHERE user_id = ?
		GROUP BY reader
		ORDER BY total DESC`

	var rows *sql.Rows
	var err error
	rows, err = ds.DB.Query(query, userID)

	if err != nil {
		return totals, err
	}
	defer rows.Close()
	var stat TotalReadingStat

	//var created string
	for rows.Next() {
		rows.Scan(&stat.TotalDuration, &stat.TotalPages, &stat.ReaderName)
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
        ), reader_readings(day, reader, duration, pages) AS (
            SELECT DATE(day), reader, SUM(duration), SUM(pages)
            FROM readings
            WHERE DATE('now', '-31 day', 'localtime') < DATE(day) AND user_id = ?
            GROUP BY day, reader
        )
        SELECT date, CASE WHEN reader IS NULL THEN '-' ELSE reader END AS reader,
            CASE WHEN duration IS NULL THEN 0 ELSE duration END AS daily_duration,
            CASE WHEN pages IS NULL THEN 0 ELSE pages END AS daily_pages
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
		rows.Scan(&entry.Label, &entry.ReaderName, &entry.TotalDuration, &entry.TotalPages)
		if _, ok := dailyStats[entry.Label]; !ok {
			dailyStats[entry.Label] = []DailyReaderStat{entry}
		} else {
			dailyStats[entry.Label] = append(dailyStats[entry.Label], entry)
		}
		//fmt.Printf("###\t%+v\n", entry)
	}
	//fmt.Printf("%+v", dailyStats)

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

// GetUserReaders - retrieves all readers attached to this user
func (ds *DataStore) GetReaderByID(readerID int) (Reader, error) {

	query := `SELECT reader_id, user_id, name FROM readers WHERE reader_id = ?`
	//ds.L.Printf("Query: %s [%d]", query, readerID)

	var err error
	row := ds.DB.QueryRow(query, readerID)

	if err != nil {
		return Reader{}, err
	}
	var reader Reader
	err = row.Scan(&reader.ID, &reader.UserID, &reader.Name)

	return reader, err
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
	query := `INSERT INTO groups (user_id, name, code, created) VALUES (?, ?, ?, ?)`
	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rand.Seed(time.Now().UnixNano())
	g.AccessCode = utils.RandStringRunes(5)
	g.CreatedOn = time.Now().Truncate(time.Second)

	res, err := stmt.Exec(g.UserID, g.Name, g.AccessCode, g.CreatedOn)
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
func (ds *DataStore) UpdateGroup(groupID int, ug UpdateGroup) error {
	query := `UPDATE groups SET name = ?, status = ? WHERE id = ?`
	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(ug.Name, ug.Status, groupID)
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

// GroupRemoveReader - remove reader from group
func (ds *DataStore) GroupRemoveReader(groupID, readerID int) error {
	query := `DELETE FROM group_readers WHERE group_id = ? AND reader_id = ?`
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

type GReader struct {
	Reader
	GroupID int
}

// GetGroupsAndReaders - retrieves all user's readers' groups
func (ds *DataStore) GetGroupsAndReaders(userID int) (map[string][]GReader, error) {

	query := `SELECT g.id, g.name, group_concat(r.reader_id||'.'||r.name)
	FROM groups g
	JOIN group_readers gr ON g.id = gr.group_id
	JOIN readers r ON gr.reader_id = r.reader_id
	WHERE r.user_id = ?
	GROUP BY g.id, g.name`

	var rows *sql.Rows
	var err error
	rows, err = ds.DB.Query(query, userID)

	groups := map[string][]GReader{}

	if err != nil {
		return groups, err
	}
	defer rows.Close()

	var gID int
	var gName, readerData string

	for rows.Next() {
		rows.Scan(&gID, &gName, &readerData)
		//readers = append(readers, reader)

		readersList := strings.Split(readerData, ",")
		groups[gName] = []GReader{}
		for _, r := range readersList {
			readerInfo := strings.Split(r, ".")
			readerID, _ := strconv.Atoi(readerInfo[0])
			reader := GReader{
				Reader{ID: readerID, UserID: userID, Name: readerInfo[1]},
				gID,
			}
			groups[gName] = append(groups[gName], reader)
		}
		//fmt.Printf("###\t%s: %+v\n", gName, groups[gName])
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
		rows.Scan(&gID, &gName, &gCode, &gStatus, &gCreated)
		t, err := time.Parse("2006-01-02T15:04:05Z07:00", gCreated)
		if err != nil {
			ds.L.Println("timeParse: ", err.Error())
		}
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

type GroupMemberCount struct {
	GroupID     int
	GroupName   string
	MemberCount int
}

// GetUserGroupMemberCount - retrieves user's groups with number of members
func (ds *DataStore) GetUserGroupMemberCount(userID int) ([]GroupMemberCount, error) {
	var groups []GroupMemberCount

	query := `SELECT g.id, g.name, count(*) AS cnt
		FROM groups g
		JOIN group_readers gr ON g.id = gr.group_id
		WHERE g.id IN (
			-- my readers' groups
			SELECT group_id
			FROM group_readers
			WHERE reader_id IN (
				SELECT reader_id FROM readers
				WHERE user_id = ?
			)
		)
		GROUP BY g.id, g.name
		ORDER BY cnt DESC, g.name ASC`

	var rows *sql.Rows
	var err error
	rows, err = ds.DB.Query(query, userID)

	if err != nil {
		return groups, err
	}
	defer rows.Close()

	var gName string
	var gID, gCount int

	for rows.Next() {
		rows.Scan(&gID, &gName, &gCount)

		groups = append(groups, GroupMemberCount{
			GroupID:     int(gID),
			GroupName:   gName,
			MemberCount: gCount,
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
		return g, err
	}
	g.UserID, _ = strconv.Atoi(gUserID)
	g.CreatedOn, _ = time.Parse("2006-01-02T15:04:05Z07:00", gCreated)
	g.ID = groupID

	return g, nil
}

// GetGroupReaders - retrieves a group's members
func (ds *DataStore) GetGroupReaders(groupID int) ([]Reader, error) {
	query := `SELECT r.reader_id, r.user_id, r.name, r.created
		FROM groups g
		JOIN group_readers gr ON g.id = gr.group_id
		JOIN readers r ON gr.reader_id = r.reader_id
		WHERE g.id = ?`

	var rows *sql.Rows
	var err error
	readers := []Reader{}
	rows, err = ds.DB.Query(query, groupID)

	if err != nil {
		return readers, err
	}
	defer rows.Close()

	var r Reader
	var created string

	for rows.Next() {
		err := rows.Scan(&r.ID, &r.UserID, &r.Name, &created)
		if err != nil {
			return readers, err
		}
		//g.UserID, _ = strconv.Atoi(gUserID)
		r.CreatedOn, _ = time.Parse("2006-01-02T15:04:05Z07:00", created)
		//g.ID = groupID
		readers = append(readers, r)
	}

	return readers, nil
}

// FindNewGroupsForReader - find groups base on a query
func (ds *DataStore) FindNewGroupsForReader(q string, readerID int) ([]Group, error) {

	query := `
	WITH readers_groups AS (
		SELECT group_id FROM group_readers WHERE reader_id = ?
	)
	SELECT g.id, g.name, g.status
	FROM groups g
	LEFT JOIN readers_groups rg ON g.id = rg.group_id
	WHERE rg.group_ID IS NULL
		AND ((status='public' AND name LIKE '%'|| ? ||'%') OR code = ?)`
	//fmt.Printf("%s [%s, %s]\n", query, fmt.Sprintf("%%%s%%", q), q)

	groups := []Group{}
	var g Group
	var gCreated string

	//rows, err := ds.DB.Query(query, fmt.Sprintf("%%%s%%", q))
	rows, err := ds.DB.Query(query, readerID, q, q)
	if err != nil {
		return []Group{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&g.ID, &g.Name, &g.Status)
		if err != nil {
			fmt.Println("in for err: ", err)
			return []Group{}, err
		}
		g.CreatedOn, _ = time.Parse("2006-01-02T15:04:05Z07:00", gCreated)
		groups = append(groups, g)
	}
	return groups, nil
}
