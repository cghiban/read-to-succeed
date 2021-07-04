package data

import (
	"time"
)

type Book struct {
	ID      int       `json:"id"`
	UserID  int       `json:"user_id"`
	Authors string    `json:"authors"`
	Title   string    `json:"title"`
	ISBN    string    `json:"isbn"`
	AddedOn time.Time `json:"-"`
}

type NewBook struct {
	UserID  int    `json:"user_id"`
	Authors string `json:"authors"`
	Title   string `json:"title"`
	ISBN    string `json:"isbn"`
}

type UpdateBook struct {
	Authors string `json:"authors"`
	Title   string `json:"title"`
	ISBN    string `json:"isbn"`
}

// AddBook - add new reading entry into the db
func (ds *DataStore) AddBook(nb NewBook) (Book, error) {

	/*reader, err := ds.GetReaderByName(nb.ReaderName)
	if err != nil {
		return err
	}
	fmt.Printf("found reader: %+v", reader)*/

	query := `
        INSERT INTO books (user_id, title, authors, isbn, added_on)
        VALUES (?, ?, ?, ?, ?)`

	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return Book{}, err
	}
	defer stmt.Close()

	now := time.Now().UTC().Round(time.Second)

	res, err := stmt.Exec(nb.UserID, nb.Title, nb.Authors, nb.ISBN, now.Format("2006-01-02T15:04:05Z"))
	if err != nil {
		return Book{}, err
	}
	rowNum, _ := res.RowsAffected()
	ds.L.Println(" -- added new book to DB: ", rowNum)

	id, err := res.LastInsertId()
	if err != nil {
		return Book{}, err
	}
	bk := Book{
		ID:      int(id),
		UserID:  nb.UserID,
		Title:   nb.Title,
		Authors: nb.Authors,
		ISBN:    nb.ISBN,
		AddedOn: now,
	}

	return bk, nil
}

// QueryByUserID - retrieve the user's books from the db
func (ds *DataStore) QueryByUserID(userID int) ([]Book, error) {

	/*reader, err := ds.GetReaderByName(nb.ReaderName)
	if err != nil {
		return err
	}
	fmt.Printf("found reader: %+v", reader)*/

	query := `
	SELECT id, user_id, title, authors, isbn, added_on
	FROM books WHERE user_id = ? ORDER BY id DESC`

	rows, err := ds.DB.Query(query, userID)
	if err != nil {
		return []Book{}, err
	}
	defer rows.Close()

	var books []Book
	var b Book
	var added string
	for rows.Next() {
		rows.Scan(&b.ID, &b.UserID, &b.Title, &b.Authors, &b.ISBN, &added)

		t, _ := time.Parse("2006-01-02T15:04:05Z", added)
		b.AddedOn = t.Local()

		books = append(books, b)
	}

	return books, nil
}
