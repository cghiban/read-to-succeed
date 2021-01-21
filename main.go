package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"read2succeed/data"
	"read2succeed/handlers"
	"time"

	"github.com/nicholasjackson/env"
)

var dataStore *data.DataStore

var bindAddress = env.String("BIND_ADDRESS", true, "", "bind address for server, i.e. localhost")
var dbPath = env.String("DB_PATH", true, "", "path to a sqlite DB")
var readers = env.String("READERS", true, "", "list of readers (comma sepparated")

func init() {
	err := env.Parse()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//db := InitDB(*dbPath)
	db, err := sql.Open("sqlite3", *dbPath) // Open the created SQLite File

	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("unable to get a db connection")
	}
	l := log.New(os.Stdout, "reading 2 succees", log.LstdFlags)
	dataStore = &data.DataStore{DB: db, L: l}
}

func main() {
	l := log.New(os.Stdout, "reading 2 succees", log.LstdFlags)

	l.Println("about to start server on ", *bindAddress)

	sm := http.NewServeMux()

	r2sservice := handlers.NewService(l, dataStore, readers)
	sm.Handle("/", r2sservice)
	sm.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("var/static/"))))

	/*
		sm.HandleFunc("/", index)
		sm.HandleFunc("/add", addReading)
		//sm.HandleFunc("/list", listReadings)

	*/
	s := &http.Server{
		Addr:         *bindAddress,
		Handler:      sm,
		IdleTimeout:  60 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatalln(err)
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Kill)
	signal.Notify(sigChan, os.Interrupt)

	sig := <-sigChan
	l.Println("Received terminate, graceful shutdown", sig)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)
}
