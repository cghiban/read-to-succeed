package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/nicholasjackson/env"
)

var dataStore DataStore
var templates *template.Template

var bindAddress = env.String("BIND_ADDRESS", true, "", "bind address for server, i.e. localhost")
var dbPath = env.String("DB_PATH", true, "", "path to a sqlite DB")
var readers = env.String("READERS", true, "", "list of readers (comma sepparated")

//var bindPort = env.Integer("BIND_Port",true,0,"bind port for server, i.e. 9090")

type Reading struct {
	ID         uint   `json:"id"`
	ReaderName string `json:"reader"`
	BookAuthor string `json:"author"`
	BookTitle  string
	Day        time.Time
	Duration   time.Duration
	CreatedOn  time.Time
}

/*var readings = []*Reading{
	&Reading{
		ID:         1,
		ReaderName: "Cornel",
		BookAuthor: "Ion Iliescu",
		BookTitle:  "Măi animalelor",
		Day:        time.Date(2009, time.November, 10, 0, 0, 0, 0, time.Local),
		Duration:   2,
		CreatedOn:  time.Now(),
	},
}*/

// initialize our db
func init() {
	err := env.Parse()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	templates = template.Must(template.ParseGlob("var/templates/*.gohtml"))

	//db := InitDB(*dbPath)
	db, err := sql.Open("sqlite3", *dbPath) // Open the created SQLite File

	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("unable to get a db connection")
	}
	l := log.New(os.Stdout, "reading 2 succees", log.LstdFlags)
	dataStore = DataStore{DB: db, l: l}
}

func main() {
	l := log.New(os.Stdout, "reading 2 succees", log.LstdFlags)

	l.Println("about to start server on ", *bindAddress)

	//gh := handlers.NewGoodbye(l)
	//dh := handlers.NewData(l)

	sm := http.NewServeMux()
	//sm.Handle("/", dh)
	//sm.Handle("/goodbye", gh)
	sm.HandleFunc("/", index)
	sm.HandleFunc("/add", addReading)
	sm.HandleFunc("/list", listReadings)
	sm.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("var/static/"))))

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

func index(rw http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(rw, "OK")

	readings, err := dataStore.ListReadings()
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Readers  []string
		Readings []Reading
	}{
		Readers:  strings.Split(*readers, ","),
		Readings: readings,
	}

	if err := templates.ExecuteTemplate(rw, "index.gohtml", data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func addReading(rw http.ResponseWriter, r *http.Request) {

}

func listReadings(rw http.ResponseWriter, r *http.Request) {

}
