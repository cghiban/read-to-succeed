package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"read2succeed/google_books"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

type BookSearch struct {
	T *template.Template
}

/*func (s *BookSearch) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	log.Println("path:", uri)
	if r.Method == http.MethodGet {
		if uri == "/search" {
			s.DoSearch(rw, r)
		} else {
			s.Index(rw, r)
		}
		return
	}

	// catch all
	rw.WriteHeader(http.StatusMethodNotAllowed)
}*/

func (s *BookSearch) DoSearch(rw http.ResponseWriter, r *http.Request) {

	uri := r.URL.Path
	log.Println("path:", uri)

	//vars := mux.Vars(r)
	//log.Println("vars:", vars)
	log.Println("query:", r.URL.Query())
	query := r.URL.Query().Get("q")
	log.Printf("q: [%s]", query)

	// https://developers.google.com/books/docs/v1/using

	result := google_books.DoSearch(query)

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.WriteHeader(http.StatusOK)
	//rw.Write([]byte("{\"status\":\"ok\"}"))
	json.NewEncoder(rw).Encode(result)
}

func (s *BookSearch) Index(rw http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	log.Println("path:", uri)

	rw.Header().Add("Cache-Control", "no-cache")
	if err := s.T.ExecuteTemplate(rw, "search.gohtml", nil); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func init() {

}

func main() {
	l := log.New(os.Stdout, "reading 2 succees", log.LstdFlags)

	l.Println("about to start server")

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
	templates := template.Must(template.New("tmpls").Funcs(funcMap).ParseGlob("var/templates/*.gohtml"))

	//templates := template.Must(template.New("tmpls").ParseGlob("var/templates/*.gohtml"))

	bookSearchService := &BookSearch{T: templates}

	//sm := http.NewServeMux()
	sm := mux.NewRouter()

	sm.HandleFunc("/", bookSearchService.Index)
	sm.Handle("/favicon.ico", http.NotFoundHandler())
	sm.HandleFunc("/search", bookSearchService.DoSearch)
	//sm.Handle("/readings/", r2sservice)
	sm.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("var/static/"))))

	s := &http.Server{
		Addr:         ":9090",
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
