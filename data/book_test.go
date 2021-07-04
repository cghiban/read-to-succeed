package data_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"read2succeed/data"

	"github.com/google/go-cmp/cmp"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// NewUnit creates a test database. It creates the required table structure
// but the database is otherwise empty. It returns the database to use
// as well as a function to call at the end of the test.
func NewUnit(t *testing.T) (*log.Logger, *sql.DB, func()) {

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	sqlStmt := `
	CREATE TABLE books (
		id integer not null primary key,
		user_id integer not null,
		title text,
		isbn text,
		authors text,
		added_on DATETIME);
	CREATE INDEX books_ndx1 ON books(user_id);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		t.Fatalf("Migrating error: %s", err)
	}

	t.Log("Waiting for database to be ready ...")

	/*ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrating error: %s", err)
	}*/

	teardown := func() {
		t.Helper()
		db.Close()

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = old

		fmt.Println("******************** LOGS ********************")
		fmt.Print(buf.String())
		fmt.Println("******************** LOGS ********************")
	}

	log := log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	return log, db, teardown
}

func TestBook(t *testing.T) {

	log, db, teardown := NewUnit(t)
	t.Cleanup(teardown)

	//DataStore - db operations
	ds := &data.DataStore{
		DB: db,
		L:  log,
	}

	testID := 0
	userID := 1

	nb := data.NewBook{
		UserID:  userID,
		Title:   "New Boook",
		Authors: "A, B",
		ISBN:    "12234",
	}

	book, err := ds.AddBook(nb)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to create entry : %s.", Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to create entry.", Success, testID)

	fmt.Println("just added:", book)

	testID++

	books, err := ds.QueryByUserID(userID)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user's book : %s.", Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to retrieve user's books. Count: %d.", Success, testID, len(books))

	testID++

	if diff := cmp.Diff(book, books[0]); diff != "" {
		t.Fatalf("\t%s\tTest %d:\tShould get back the same book. Diff:\n%s", Failed, testID, diff)
	}
	t.Logf("\t%s\tTest %d:\tShould get back the same book.", Success, testID)

	log.Println("done testing..")
}
