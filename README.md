# read to succeed

Tiny webapp for recording ones read book, articles, etc

##todo

- [x] start using Gorilla toolkit
- [x] add authentication (users, login, session)
- [x] show some stats
- [x] add readers management and stop using READERS env var
- [ ] add user settings (privacy / readers and groups)
- [ ] more stats (daily and monthly)
- [x] show groups readings
- [x] add groups
- [x] join/leave a group
- [x] make use of csrf
- [x] add midlewares
- [x] make some users admins (add is_admin column)
- [x] only admins create groups
- [ ] groups should have a start / ending dates
- [ ] remember last reader on home
- [x] make a proper menu (just text)
- [ ] users may reset password (integrate with mailgun)
- [x] fix groups readings (start with a list of groups and let the user choose a group)
- [x] pagination for user readings
- [x] pagination for group readings
- [x] fix user sign up

database (sqlite):

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
    FOREIGN KEY(user_id) REFERENCES auth_user(user_id)
);
CREATE INDEX `user_settings_ndx1` ON `user_settings` (`user_id`,`id`);

CREATE TABLE books (
    id integer not null primary key,
    user_id integer not null,
    title text,
    authors text,
    thumb_url text,
    isbn text,
    added_on DATETIME
);
CREATE INDEX books_ndx1 ON books(user_id);

CREATE TABLE readers (
    reader_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    grade_lvl TEXT,
    created DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES auth_user(user_id)
);
CREATE INDEX readers_ndx1 ON readers (user_id, reader_id);
CREATE UNIQUE INDEX readers_unq1 ON readers(user_id, name);

INSERT INTO readers (user_id, name, grade_lvl, created) VALUES
(1, 'Cornel', 'grownup', datetime('now','localtime')),
(1, 'Daniel', '2nd', datetime('now','localtime')),
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
    FOREIGN KEY(user_id) REFERENCES auth_user(user_id),
    FOREIGN KEY(reader_id) REFERENCES readers(reader_id)
);

CREATE INDEX readings_ndx1 ON readings (user_id, reader_id);
CREATE INDEX readings_ndx2 ON readings (user_id, reader);

CREATE TABLE groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT CHECK( LENGTH(name) > 3) NOT NULL,
    code TEXT CHECK( LENGTH(code) > 4) NOT NULL,
    status TEXT CHECK( status IN ('private', 'public', 'locked')) NOT NULL DEFAULT 'private',
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES auth_user(user_id)
);

CREATE INDEX groups_ndx1 ON groups (user_id);
CREATE INDEX groups_ndx2 ON groups (code);

CREATE TABLE group_readers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    reader_id INTEGER NOT NULL,
    joined_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(group_id) REFERENCES groups(id),
    FOREIGN KEY(reader_id) REFERENCES readers(reader_id)
);

CREATE INDEX group_readers_ndx1 ON group_readers (group_id);
CREATE INDEX group_readers_ndx2 ON group_readers (reader_id);

INSERT INTO groups (user_id, name, code)
VALUES (1, "Best Group", "ABC-123"), (1, "The Group", "ABC-124");

INSERT INTO group_readers (group_id, reader_id)
VALUES (1, 1), (1, 2), (1, 3);

insert into readings
select id, 1, reader, 3, book_author, book_title, day, duration, created
FROM readings_old
WHERE reader = 'Elena';

-- alter table readings add column user_id integer not null default 1 references auth(user_id);
-- alter table readings add FOREIGN KEY(user_id) REFERENCES auth(user_id);

CREATE INDEX readings1 ON readings(id);
--CREATE INDEX readings_user_id_ndx ON readings(user_id, reader, created);
CREATE INDEX readings_user_id_ndx ON readings(user_id, reader_id, created desc);

INSERT INTO readings (reader, book_author, book_title, day, duration, created) VALUES 
("Cornel", "Ion Creangă", "Povestea poveștilor", "2020-10-01", 3, datetime('now','localtime')),
("Cornel", "Will Wight", "Unsouled", "2020-12-30", 72, datetime('now','localtime'));

-- stats

Title: "Totals",
Readers: [
        Cornel, Stats: [{ Label: 100,Daniel: 340
}

######

Title: "Monthly Totals", or "Daily Totals"
Readers: [{
        Name: "Cornel",
        Stats: [
                { Label: "2020-12", Value: 99},
                { Label: "2021-01", Value: 19},
                { Label: "2021-02", Value: 44},
        ]},{
        Name: "Daniel",
        Stats: [
                { Label: "2020-12", Value: 19},
                { Label: "2021-01", Value: 129},
                { Label: "2021-02", Value: 340},
        ]},
}


SELECT sum(duration), reader
FROM readings
LEFT JOIN readers on readings.reader = readers.name
WHERE readings.user_id = 1
GROUP BY reader;

SELECT sum(duration), reader, strftime("%Y-%m", day) as m
FROM readings
LEFT JOIN readers on readings.reader_id = readers.reader_id
WHERE readings.user_id = 1
GROUP BY reader, m;

-- weekly; weekday 1 == monday, 4 == thursday
SELECT sum(duration), DATE(day, 'weekday 1') as monday 
FROM readings WHERE reader = 'Daniel' group by monday;

SELECT strftime('%W', day) AS week_no, DATE(day, 'weekday 4') as thursday, sum(duration) FROM readings WHERE reader = 'Daniel' group by thursday, week_no;

-- current week
SELECT strftime('%w', day) AS week_no, DATE(day, 'weekday 1') as monday, date('now', 'start of day', 'weekday 0', '-7 day') FROM readings
WHERE reader = 'Daniel'
AND day >= date('now', 'start of day', 'weekday 0', '-7 day');

-- current week starts on Sunday
SELECT strftime('%w', day) AS week_no, 
    date('now', 'start of day', 'weekday 0', '-7 day') AS week_start, 
    sum(duration)
FROM readings
WHERE reader = 'Daniel'
AND day >= date('now', 'start of day', 'weekday 0', '-7 day')
GROUP BY week_no, week_start

-- past 5 weeks
SELECT strftime('%W', day) AS week_no, 
    date(day, 'start of day', 'weekday 0', '-31 day') AS week_start,
    date('now', 'start of day', 'weekday 0', '-31 day')
FROM readings
WHERE reader = 'Daniel'
AND day >= date('now', 'start of day', 'weekday 0', '-31 day')
--GROUP BY week_no, week_start

WITH RECURSIVE last31days(date) AS (
  VALUES(DATE('now', '-31 day'))
  UNION ALL
  SELECT DATE(date, '+1 day')
  FROM last31days
  WHERE date <= date('now')
)
SELECT date, sum(NULLIF(duration, 0)) as duration
FROM last31days LEFT JOIN readings ON date(day)=date
WHERE reader = 'Daniel'
GROUP BY date;

WITH RECURSIVE last31days(date) AS (
  VALUES(DATE('now', '-31 day', 'localtime'))
  UNION ALL
  SELECT DATE(date, '+1 day')
  FROM last31days
  WHERE date <= date('now')
), reader_readings(day, reader, duration) AS (
    SELECT DATE(day) as day, reader, sum(duration)
    FROM readings
    WHERE DATE('now', '-31 day', 'localtime') < DATE(day) AND user_id=1--reader = 'Daniel'
    GROUP BY day, reader
)
SELECT date, CASE WHEN reader IS NULL THEN '-' ELSE reader END AS reader,
    CASE WHEN duration IS NULL THEN 0 ELSE duration END AS daily_duration
FROM last31days LEFT JOIN reader_readings ON date = day
WHERE date <= CURRENT_DATE;

WITH RECURSIVE generate_series(value) AS (
    SELECT 0
    UNION ALL
    SELECT value + 1 FROM generate_series
    LIMIT 60
) SELECT date('now', '-' || value || ' day') FROM generate_series

-- list group readings
SELECT g.name, g.code, r.reader, r.book_title
FROM readings r 
JOIN group_readers gr ON r.reader_id = gr.reader_id
JOIN groups g ON gr.group_id = g.id
where g.id = xx

SELECT g.name, g.id, gr.group_id, gr.reader_id, r.reader_id, r.reader, r.book_title
FROM readings r
JOIN group_readers gr ON r.reader_id = gr.reader_id
JOIN groups g ON gr.group_id = g.id
where g.id = 3;

SELECT g.name, g.id, gr.group_id, gr.reader_id, rdr.reader_id, r.reader_id, rdr.name, r.reader, r.id, r.book_title
FROM readers rdr
JOIN readings r on rdr.reader_id = r.reader_id
JOIN group_readers gr ON r.reader_id = gr.reader_id
JOIN groups g ON gr.group_id = g.id
where g.id = 3;



select g.id, g.name, g.user_id, r.user_id, r.reader_id,  r.name
FROM groups g
JOIN group_readers gr ON g.id = gr.group_id
JOIN readers r ON gr.reader_id = r.reader_id
WHERE r.user_id = 1
```

Run as:

```bash
BIND_ADDRESS=:8080 DB_PATH=var/db.db go run .
```

