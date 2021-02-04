
# read to succeed

Tiny webapp for recording ones read book, articles, etc


##todo

- [x] start using Gorilla toolkit
- [x] add authentication (users, login, session)
- [ ] show some stats
- [ ] add readers management and stop using READERS env var
- [ ] add user settings (privacy / readers and groups)
- [ ] add groups


set user list using ENV vars
```shell
READERS=costel,cornelia,purcel
```

database:

```sql

CREATE TABLE auth_user (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    passw TEXT NOT NULL,
    created DATETIME NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX auth_user1 ON auth_user(user_id);
CREATE UNIQUE INDEX auth_user_id_UNIQUE ON auth_user(email ASC);

CREATE TABLE user_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY(user_id) REFERENCES auth(user_id)
);
CREATE INDEX `user_settings_ndx1` ON `user_settings` (`user_id`,`id`);

CREATE TABLE readers (
    reader_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    grade_lvl TEXT,
    created DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES auth(user_id)
);
CREATE INDEX readers_ndx1 ON readers (user_id, reader_id);

INSERT INTO readers (user_id, name, grade_lvl, created) VALUES
(1, 'Cornel', 'grownup', datetime('now','localtime')),
(1, 'Daniel', '1st', datetime('now','localtime')),
(1, 'Elena', 'grownup', datetime('now','localtime'));

CREATE TABLE readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    reader TEXT,
    reader_id INTEGER NOT NULL,
    book_author TEXT,
    book_title TEXT,
    day TEXT,
    duration INTEGER NOT NULL DEFAULT 0,
    created DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES auth(user_id),
    FOREIGN KEY(reader_id) REFERENCES readers(reader_id)
);

insert into readings
select id, user_id, reader, 2, book_author, book_title, day, duration, created
FROM readings_old
WHERE reader = 'Daniel';

-- alter table readings add column user_id integer not null default 1 references auth(user_id);
-- alter table readings add FOREIGN KEY(user_id) REFERENCES auth(user_id);

CREATE INDEX readings1 ON readings(id);
--CREATE INDEX readings_user_id_ndx ON readings(user_id, reader, created);
CREATE INDEX readings_user_id_ndx ON readings(user_id, reader_id, created desc);

INSERT INTO readings (reader, book_author, book_title, day, duration, created) VALUES 
("Cornel", "Ion Creangă", "Povestea poveștilor", "2020-10-01", 3, datetime('now','localtime')),
("Cornel", "Will Wight", "Unsouled", "2020-12-30", 72, datetime('now','localtime'));
```

Run as:

```bash
BIND_ADDRESS=:8080 DB_PATH=var/db.db READERS=Cornel,Gigel go run .
```

