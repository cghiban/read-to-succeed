package data

import (
	"log"
	"time"
)

type Book struct {
	ID       int       `json:"id"`
	UserID   int       `json:"user_id"`
	Authors  string    `json:"authors"`
	Title    string    `json:"title"`
	ISBN     string    `json:"isbn"`
	ThumbURL string    `json:"thumb_url"`
	AddedOn  time.Time `json:"-"`
}

type NewBook struct {
	UserID   int    `json:"user_id"`
	Authors  string `json:"authors"`
	Title    string `json:"title"`
	ThumbURL string `json:"thumb_url"`
	ISBN     string `json:"isbn"`
	Reader   string `json:"reader"`
}

type UpdateBook struct {
	Authors  string `json:"authors"`
	Title    string `json:"title"`
	ISBN     string `json:"isbn"`
	ThumbURL string `json:"thumb_url"`
}

// AddBook - add new reading entry into the db
func (ds *DataStore) AddBook(nb NewBook) (Book, error) {

	query := `
        INSERT INTO books (user_id, title, authors, isbn, thumb_url, added_on)
        VALUES (?, ?, ?, ?, ?, ?)`

	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return Book{}, err
	}
	defer stmt.Close()

	now := time.Now().Round(time.Second)

	res, err := stmt.Exec(nb.UserID, nb.Title, nb.Authors, nb.ISBN, nb.ThumbURL, now.Format("2006-01-02T15:04:05Z07:00"))
	if err != nil {
		return Book{}, err
	}
	//log.Printf("now: %s", now.Format("2006-01-02T15:04:05Z07:00"))
	//rowNum, _ := res.RowsAffected()
	//ds.L.Println(" -- added new book to DB: ", rowNum)

	id, err := res.LastInsertId()
	if err != nil {
		return Book{}, err
	}
	bk := Book{
		ID:       int(id),
		UserID:   nb.UserID,
		Title:    nb.Title,
		Authors:  nb.Authors,
		ISBN:     nb.ISBN,
		ThumbURL: nb.ThumbURL,
		AddedOn:  now,
	}

	return bk, nil
}

// QueryBookByUserID - retrieve user's books from the db
func (ds *DataStore) QueryBookByUserID(userID int) ([]Book, error) {

	query := `
	SELECT id, user_id, title, authors, isbn, thumb_url, added_on
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
		rows.Scan(&b.ID, &b.UserID, &b.Title, &b.Authors, &b.ISBN, &b.ThumbURL, &added)

		//log.Printf("added: %s", added)
		t, err := time.Parse("2006-01-02T15:04:05Z07:00", added)
		if err != nil {
			log.Panicln(err)
		}

		//log.Printf("t: %s", t.Format("2006-01-02T15:04:05Z07:00"))
		b.AddedOn = t.Local()

		books = append(books, b)
	}

	return books, nil
}

// QueryBookByID - retrieve given book
func (ds *DataStore) QueryBookByID(bookID int) (Book, error) {
	query := `
	SELECT id, user_id, title, authors, isbn, thumb_url, added_on
	FROM books WHERE id = ?`

	row := ds.DB.QueryRow(query, bookID)

	var book Book
	var added string
	err := row.Scan(&book.ID, &book.UserID, &book.Title, &book.Authors, &book.ISBN, &book.ThumbURL, &added)
	if err != nil {
		return book, err
	}
	t, _ := time.Parse("2006-01-02T15:04:05Z07:00", added)
	book.AddedOn = t

	return book, nil
}

// UpdateBook - retrieve the user's books from the db
func (ds *DataStore) UpdateBook(bookID int, ub UpdateBook) error {

	/*book, err := ds.QueryBookByID(bookID)
	if err != nil {
		return book, err
	}*/

	query := `
		UPDATE books
		SET authors = ?, title = ?,  isbn = ?, thumb_url = ?
		WHERE id = ?`
	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	//fmt.Println(query, []interface{}{ub.Authors, ub.Title, ub.ISBN, ub.ThumbURL, bookID})
	_, err = stmt.Exec(ub.Authors, ub.Title, ub.ISBN, ub.ThumbURL, bookID)
	if err != nil {
		return err
	}

	return err
}
