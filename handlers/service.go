package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"read2succeed/data"
	"strings"
	"text/template"
	"time"
)

// Service data struct
type Service struct {
	l       *log.Logger
	store   *data.DataStore
	readers *string
	t       *template.Template
}

// NewService initializes a new Serivice
func NewService(l *log.Logger, store *data.DataStore, readers *string) *Service {
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
	templates := template.Must(template.New("tmpls").Funcs(funcMap).ParseGlob("var/templates/*.gohtml"))
	//templates = templates.Funcs(funcMap)

	return &Service{l: l, store: store, t: templates, readers: readers}
}

func (s *Service) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log.Println("path:", r.URL.Path)
	//uri := r.URL.Path
	if r.Method == http.MethodGet {
		s.GetReadings(rw, r)
		return
	}

	if r.Method == http.MethodPost {
		s.AddReading(rw, r)
		return
	}

	// catch all
	rw.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *Service) GetReadings(rw http.ResponseWriter, r *http.Request) {
	readings, err := s.store.ListReadings()
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Readers  []string
		Readings []data.Reading
		Today    string
	}{
		Readers:  strings.Split(*s.readers, ","),
		Readings: readings,
		Today:    time.Now().Format("2006-01-02"),
	}

	if err := s.t.ExecuteTemplate(rw, "index.gohtml", data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Service) AddReading(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Invalid request", http.StatusBadRequest)
		return
	}

	contentType := r.Header["Content-Type"]
	log.Println(contentType, len(contentType) == 1, contentType[0])

	if len(contentType) == 1 && contentType[0] == "application/json" {

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		reading := &data.Reading{}
		err := decoder.Decode(reading)
		if err != nil {
			log.Println(err)
			http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
			return
		}

		log.Println(reading)
		err = s.store.AddReading(reading)
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
