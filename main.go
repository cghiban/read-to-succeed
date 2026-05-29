package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"read2succeed/data"
	"read2succeed/handlers"
	"read2succeed/web"
	"syscall"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/nicholasjackson/env"
)

//go:embed var/templates/* var/static
var resources embed.FS

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

	// will store everything in UTC and we'll convert times to whatever TZ is set when reading from DB
	*dbPath += "?_fk=1&_synchronous=NORMAL&_journal=WAL&_cache_size=16000&_loc=UTC"
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatal(err)
	}

	if db == nil || db.Ping() != nil {
		log.Fatal("unable to get a db connection")
	}
	l := log.New(os.Stdout, "reading 2 succeed", log.LstdFlags)
	dataStore = &data.DataStore{DB: db, L: l}

	// ----------------------------------------------------------------
	// DEBUG
	sqliteVersion, _ := dataStore.GetSQLiteVersion()
	l.Println("using SQLite version", sqliteVersion)

	if users, err := dataStore.ListUsers(); err != nil {
		l.Println("could not list users:", err)
	} else {
		for _, u := range users {
			l.Printf("user: id=%d name=%q email=%s", u.ID, u.Name, u.Email)
		}
	}

	varDir := path.Dir(*dbPath)
	if entries, err := os.ReadDir(varDir); err != nil {
		l.Println(varDir, ":", err)
	} else {
		for _, e := range entries {
			l.Printf("%s/%s", varDir, e.Name())
		}
	}

	// /DEBUG
	// ----------------------------------------------------------------
}

func main() {
	slog := log.New(os.Stdout, "reading 2 succeed", log.LstdFlags)

	r2sservice := handlers.NewService(slog, dataStore, sessionKey, resources)

	// auth midleware...
	authMw := handlers.Auth{
		Service: r2sservice,
	}

	sm := mux.NewRouter()
	getRouter := sm.Methods("GET").Subrouter()
	getRouter.HandleFunc("/", r2sservice.GetReadings)

	postRouter := sm.Methods("POST").Subrouter()
	postRouter.Handle("/add", web.WrapMiddleware(r2sservice.AddReading, authMw.UserViaSession, authMw.RequireUser))

	sm.Handle("/groupreadings", web.WrapMiddleware(r2sservice.GetGroupReadings, authMw.UserViaSession, authMw.RequireUser))
	sm.Handle("/groupreaders/{id:[0-9]+}", web.WrapMiddleware(r2sservice.GetGroupReaders, authMw.UserViaSession, authMw.RequireUser))
	sm.Handle("/settings", web.WrapMiddleware(r2sservice.Settings, authMw.UserViaSession, authMw.RequireUser))

	sm.Handle("/addreader", web.WrapMiddleware(r2sservice.AddReader, authMw.UserViaSession, authMw.RequireUser)).Methods("POST")
	sm.Handle("/addgroup", web.WrapMiddleware(r2sservice.AddGroup, authMw.UserViaSession, authMw.RequireUser)).Methods("POST")
	sm.Handle("/updategroup/{id:[0-9]+}", web.WrapMiddleware(r2sservice.UpdateGroup, authMw.UserViaSession, authMw.RequireUser)).Methods("POST").HeadersRegexp("Content-Type", "application/json")
	sm.Handle("/findgroups", web.WrapMiddleware(r2sservice.FindAvailableGroups, authMw.UserViaSession, authMw.RequireUser)).Methods("POST")
	sm.Handle("/joingroup", web.WrapMiddleware(r2sservice.JoinGroup, authMw.UserViaSession, authMw.RequireUser)).Methods("POST")
	sm.Handle("/leavegroup", web.WrapMiddleware(r2sservice.LeaveGroup, authMw.UserViaSession, authMw.RequireUser)).Methods("POST")

	sm.Handle("/dailystats", web.WrapMiddleware(r2sservice.GetDailyStats, authMw.UserViaSession, authMw.RequireUser))
	sm.HandleFunc("/about", r2sservice.About)

	envir := os.Getenv("APP_ENV")
	if envir == "local" || envir == "dev" {
		sm.HandleFunc("/search_books", r2sservice.SearchGoogleBooks)
		sm.HandleFunc("/add_book", r2sservice.AddBook)
		sm.HandleFunc("/library", r2sservice.Library)
	}
	// https://www.alexedwards.net/blog/preventing-csrf-in-go
	// http.CrossOriginProtection
	// it needs go1.25
	csrf := CsrfMiddleware([]byte(*csrfKey))
	userRouter := sm.Methods("POST", "GET").Subrouter()
	userRouter.Use(csrf)
	userRouter.HandleFunc("/register", r2sservice.UserSignUp)
	userRouter.HandleFunc("/login", r2sservice.UserLogIn)
	userRouter.HandleFunc("/logout", r2sservice.UserLogOut)
	userRouter.HandleFunc("/forgot-password", r2sservice.ForgotPassword)
	userRouter.HandleFunc("/reset-password", r2sservice.ResetPassword)

	//sm.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("var/static/"))))

	staticFS, err := fs.Sub(resources, "var/static")
	if err != nil {
		slog.Fatalln(err)
	}

	sm.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	sm.Handle("/favicon.ico", http.NotFoundHandler())

	s := &http.Server{
		Addr:         *bindAddress,
		Handler:      sm,
		IdleTimeout:  60 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		slog.Printf("Starting webserver on %s\n", *bindAddress)
		err = s.ListenAndServe()
		if err != nil {
			slog.Fatalln(err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	slog.Println("Received terminate, graceful shutdown", sig)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.Shutdown(ctx)
}

func CsrfMiddleware(key []byte) func(http.Handler) http.Handler {
	var opts []csrf.Option
	env := os.Getenv("APP_ENV")
	if env == "local" || env == "dev" {
		opts = append(opts, csrf.Secure(false))
	}
	csrfFn := csrf.Protect(key, opts...)

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Need this starting from gorilla/csrf 1.7.3.
			if env == "local" || env == "dev" {
				r = csrf.PlaintextHTTPRequest(r)
			}
			csrfFn(next).ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
