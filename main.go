package main

import (
	"context"
	"database/sql"
	"encoding/json"
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
	ID         uint      `json:"id,omitempty"`
	ReaderName string    `json:"reader"`
	BookAuthor string    `json:"author"`
	BookTitle  string    `json:"title"`
	Day        string    `json:"day"`
	Duration   int       `json:"duration"`
	CreatedOn  time.Time `json:-`
}

func init() {
	err := env.Parse()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// init template
	funcMap := template.FuncMap{
		"dayToDate": func(s string) string {
			t, err := time.Parse("2006-01-02", s)
			if err != nil {
				return ""
			}
			return t.Format("Jan 2, 2006")
		},
		"dateISOish": func(t time.Time) string { return t.Format("2006-01-02 3:04p") },
	}
	templates = template.Must(template.New("tmpls").Funcs(funcMap).ParseGlob("var/templates/*.gohtml"))
	//templates = templates.Funcs(funcMap)

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
	//sm.HandleFunc("/list", listReadings)
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
		Today    string
	}{
		Readers:  strings.Split(*readers, ","),
		Readings: readings,
		Today:    time.Now().Format("2006-01-02"),
	}

	if err := templates.ExecuteTemplate(rw, "index.gohtml", data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func addReading(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Invalid request", http.StatusBadRequest)
		return
	}

	contentType := r.Header["Content-Type"]
	log.Println(contentType, len(contentType) == 1, contentType[0])

	if len(contentType) == 1 && contentType[0] == "application/json" {

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		reading := &Reading{}
		err := decoder.Decode(reading)
		if err != nil {
			log.Println(err)
			http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
			return
		}

		log.Println(reading)
		err = dataStore.AddReading(reading)
		if err != nil {
			log.Println(err)
			http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
			return
		}

		//fmt.Fprintf(rw, "{s}")
		rw.Write([]byte("{\"status\":\"ok\"}"))
		return
	}

	err := r.ParseMultipartForm(10000)
	if r.Method != http.MethodPost {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	data := r.PostForm
	log.Printf("form data: %#v", data)
	rw.Write([]byte("[1,2,3]"))
}
