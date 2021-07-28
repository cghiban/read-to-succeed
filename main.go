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
	"read2succeed/web"
	"syscall"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/nicholasjackson/env"
)

var dataStore *data.DataStore

var bindAddress = env.String("BIND_ADDRESS", true, "", "bind address for server, i.e. localhost")
var sessionKey = env.String("SESSION_KEY", true, "", "Session Key for encoding the session")
var csrfKey = env.String("CSRF_KEY", true, "", "csrf key")
var dbPath = env.String("DB_PATH", true, "", "path to a sqlite DB")

func init() {
	err := env.Parse()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if *sessionKey == "" {
		fmt.Println("missing SESSION_KEY")
		os.Exit(1)
	}

	//db := InitDB(*dbPath)
	*dbPath += "?_fk=1&_journal=WAL&_cache_size=-16000"
	db, err := sql.Open("sqlite3", *dbPath) // Open the created SQLite File

	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("unable to get a db connection")
	}
	l := log.New(os.Stdout, "reading 2 succeed", log.LstdFlags)
	dataStore = &data.DataStore{DB: db, L: l}

	sqliteVersion, _ := dataStore.GetSQLiteVersion()
	l.Println("using SQLite version", sqliteVersion)

	// to keep readers in session
	//gob.Register([]data.Reader{})
}

func main() {
	l := log.New(os.Stdout, "reading 2 succees", log.LstdFlags)

	l.Println("about to start server on ", *bindAddress)

	//dataStore.GetStatsDaily(1)

	//return

	r2sservice := handlers.NewService(l, dataStore, sessionKey)

	// auth midleware...
	authMw := handlers.Auth{
		Service: r2sservice,
	}

	//sm := http.NewServeMux()
	sm := mux.NewRouter()
	getRouter := sm.Methods("GET").Subrouter()
	getRouter.HandleFunc("/", r2sservice.GetReadings)

	postRouter := sm.Methods("POST").Subrouter()
	postRouter.HandleFunc("/add", r2sservice.AddReading)

	sm.Handle("/settings", web.WrapMiddleware(r2sservice.Settings, authMw.UserViaSession, authMw.RequireUser))
	sm.Handle("/addreader", web.WrapMiddleware(r2sservice.AddReader, authMw.UserViaSession, authMw.RequireUser)).Methods("POST")
	sm.Handle("/addgroup", web.WrapMiddleware(r2sservice.AddGroup, authMw.UserViaSession, authMw.RequireUser)).Methods("POST")
	sm.Handle("/updategroup/{id:[0-9]+}", web.WrapMiddleware(r2sservice.UpdateGroup, authMw.UserViaSession, authMw.RequireUser)).Methods("POST").HeadersRegexp("Content-Type", "application/json")

	//sm.HandleFunc("/joingroup", r2sservice.JoinGroup)
	sm.HandleFunc("/dailystats", r2sservice.GetDailyStats)
	sm.HandleFunc("/about", r2sservice.About)

	sm.HandleFunc("/search_books", r2sservice.SearchGoogleBooks)
	sm.HandleFunc("/add_book", r2sservice.AddBook)
	sm.HandleFunc("/library", r2sservice.Library)

	//sm.HandleFunc("/register", r2sservice.UserSignUp)

	// make sure we set Secure to true for production
	csrfMiddleware := csrf.Protect([]byte(*csrfKey), csrf.Secure(false))
	userRouter := sm.Methods("POST", "GET").Subrouter()
	userRouter.Use(csrfMiddleware)
	userRouter.HandleFunc("/register", r2sservice.UserSignUp)
	userRouter.HandleFunc("/login", r2sservice.UserLogIn)
	userRouter.HandleFunc("/logout", r2sservice.UserLogOut)

	sm.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("var/static/"))))

	//sm.Handle("/", r2sservice)
	//sm.Handle("/readings/", r2sservice)
	//sm.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("var/static/"))))
	sm.Handle("/favicon.ico", http.NotFoundHandler())

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	l.Println("Received terminate, graceful shutdown", sig)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)
}
