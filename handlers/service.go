package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"read2succeed/data"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
)

// Service data struct
type Service struct {
	l       *log.Logger
	store   *data.DataStore
	readers *string
	session *sessions.CookieStore
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

	session := sessions.NewCookieStore([]byte("secret-password"))
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	return &Service{l: l, store: store, t: templates, readers: readers, session: session}
}

// GetReadings - list user's/users' read books
func (s *Service) GetReadings(rw http.ResponseWriter, r *http.Request) {

	session, err := s.session.Get(r, "session")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	isLoggedIn := s.IsLoggedIn(r)
	if !isLoggedIn {
		http.Redirect(rw, r, "/login", http.StatusFound)
		return
	}

	reader := r.URL.Query().Get("reader")
	//userIDv := session.Values["user_id"]
	//userID := userIDv.(int)
	userID := session.Values["user_id"].(int)
	//fmt.Printf("userID: %T\t%q", userID, userID)

	readings, err := s.store.ListUserReadings(userID, reader)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		CurrentReader string
		Readers       []string
		Readings      []data.Reading
		Today         string
	}{
		CurrentReader: reader,
		Readers:       strings.Split(*s.readers, ","),
		Readings:      readings,
		Today:         time.Now().Format("2006-01-02"),
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

	if !s.IsLoggedIn(r) {
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	session, _ := s.session.Get(r, "session")

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

		userIDv := session.Values["user_id"]
		userID, _ := userIDv.(int)
		reading.UserID = userID
		log.Println(reading)
		err = s.store.AddReading(reading)
		if err != nil {
			log.Println(err)
			http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
			return
		}

		rw.Write([]byte("{\"status\":\"ok\"}"))
		return
	}

	err := r.ParseMultipartForm(1_000)
	if r.Method != http.MethodPost {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	data := r.PostForm
	log.Printf("form data: %#v", data)
	rw.Write([]byte("[1,2,3]"))
}
