
# read to succeed

Tiny webapp for recording ones read book, articles, etc

set user list using ENV vars
```shell
READERS=costel,cornelia,purcel
```

database:

```sql
CREATE TABLE readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    reader TEXT,
    book_author TEXT,
    book_title TEXT,
    day DATE NOT NULL,
    duration INTEGER NOT NULL DEFAULT 0,
    created DATETIME NOT NULL
);

INSERT INTO readings (reader, book_author, book_title, day, duration, created) VALUES 
("Cornel", "Ion Creangă", "Povestea poveștilor", "2020-10-01", "3m", datetime('now','localtime')),
("Cornel", "Will Wight", "Unsouled", "2020-12-30", "1h12m", datetime('now','localtime'));
```

Run as:

```bash
BIND_ADDRESS=:8080 DB_PATH=var/db.db READERS=Cornel,Gigel go run .
```

