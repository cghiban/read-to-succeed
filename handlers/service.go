package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"read2succeed/data"
	"read2succeed/google_books"
	"sort"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/mux"

	"github.com/gorilla/sessions"
)

// Service data struct
type Service struct {
	l     *log.Logger
	store *data.DataStore
	//readers *string
	//session *sqlitestore.SqliteStore
	session *sessions.CookieStore
	t       *template.Template
}

// NewService initializes a new Serivice
func NewService(l *log.Logger, store *data.DataStore, sessionKey *string) *Service {
	// init template
	funcMap := template.FuncMap{
		"dayToDate": func(s string) string {
			t, err := time.Parse("2006-01-02", s)
			if err != nil {
				return ""
			}
			return t.Format("Jan 2, 2006")
		},
		"dateISOish": func(t time.Time) string { return t.Format("2006-01-02 3:04pm") },
	}
	templates := template.Must(template.New("tmpls").Funcs(funcMap).ParseGlob("var/templates/*.gohtml"))
	//templates = templates.Funcs(funcMap)

	sessStore := sessions.NewCookieStore([]byte(*sessionKey))
	/*sessStore, err := sqlitestore.NewSqliteStoreFromConnection(store.DB, "sessions", "/", 86400, []byte(*sessionKey))
	if err != nil {
		panic(err)
	}*/

	//sessStore.Options = &sessions.Options{HttpOnly: true}

	sessStore.Options = &sessions.Options{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   7 * 86400,
	}

	return &Service{l: l, store: store, t: templates, session: sessStore}
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

	// TODO XXX paginate results:
	// https://github.com/vcraescu/go-paginator
	readings, err := s.store.ListUserReadings(userID, reader)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	stats, err := s.store.GetStatsTotals(userID)
	if err != nil {
		s.l.Panicln("stats err: ", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	//readers := session.Values("readers")
	//val := session.Values["readers"]
	//fmt.Printf("%+v\n", val)
	readers, err := s.store.GetUserReaders(userID)
	if err != nil {
		s.l.Panicln("stats err: ", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	data := struct {
		CurrentReader string
		Readers       []data.Reader
		Readings      []data.Reading
		Today         string
		Stats         []data.TotalReadingStat
	}{
		CurrentReader: reader,
		//Readers:       strings.Split(*s.readers, ","),
		Readers:  readers,
		Readings: readings,
		Today:    time.Now().Format("2006-01-02"),
		Stats:    stats,
	}

	//s.l.Printf("stats: %#v\n", stats)

	if err := s.t.ExecuteTemplate(rw, "index.gohtml", data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// AddReading - add new entry
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

// GetDailyStats - list user's/users' read books
func (s *Service) GetDailyStats(rw http.ResponseWriter, r *http.Request) {
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

	stats, err := s.store.GetStatsTotals(userID)
	if err != nil {
		s.l.Panicln("stats err: ", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	dailyStats, err := s.store.GetStatsDaily(userID)
	if err != nil {
		s.l.Panicln("daily stats err: ", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	readers, err := s.store.GetUserReaders(userID)
	if err != nil {
		s.l.Panicln("stats err: ", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	days := make([]string, 0, len(dailyStats))
	for day := range dailyStats {
		days = append(days, day)
	}
	//sort.Strings(sortedDays)
	sort.Sort(sort.Reverse(sort.StringSlice(days)))

	data := struct {
		CurrentReader string
		Readers       []data.Reader
		Today         string
		Stats         []data.TotalReadingStat
		DailyStats    data.DailyReadingStats
		Days          []string
	}{
		CurrentReader: reader,
		Readers:       readers,
		Today:         time.Now().Format("2006-01-02"),
		Stats:         stats,
		DailyStats:    dailyStats,
		Days:          days,
	}
	for _, day := range days {
		fmt.Printf("** %+s\t%+v\n", day, dailyStats[day])
	}

	//s.l.Printf("stats: %#v\n", stats)
	//s.l.Printf("dailyStats: %#v\n", dailyStats)

	if err := s.t.ExecuteTemplate(rw, "daily-stats.gohtml", data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// Settings - display settings page
func (s *Service) Settings(rw http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*data.AuthUser)

	readers, err := s.store.GetUserReaders(user.ID)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	userGroups, err := s.store.GetUserGroups(user.ID)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	groupReaders, err := s.store.GetGroupsAndReaders(user.ID)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Readers      []data.Reader
		UserGroups   []data.Group
		GroupReaders map[string][]data.Reader
	}{
		Readers:      readers,
		UserGroups:   userGroups,
		GroupReaders: groupReaders,
	}

	log.Printf("data:%v+\n", data)

	if err := s.t.ExecuteTemplate(rw, "settings.gohtml", data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// AddReader - add new reader
func (s *Service) AddReader(rw http.ResponseWriter, r *http.Request) {

	contentType := r.Header.Get("Content-Type")
	log.Println(contentType)

	if contentType != "application/json" {
		http.Error(rw, "Invalid request: expecting JSON.", http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	newReader := &data.Reader{}
	err := decoder.Decode(newReader)
	if err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	user := r.Context().Value("user").(*data.AuthUser)

	newReader.UserID = user.ID
	log.Println(newReader)
	err = s.store.AddReader(newReader)
	if err != nil {
		s.l.Printf("AddReader(%d, %s):", user.ID, newReader.Name)
		s.l.Println(err)
		http.Error(rw, "{\"status\":\"error\", \"message\":\"Unable to add reader\"}", http.StatusInternalServerError)
		return
	}

	rw.Write([]byte("{\"status\":\"ok\"}"))
}

// AddGroup - add new group
func (s *Service) AddGroup(rw http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(rw, "Invalid request: expecting JSON.", http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	newGroup := &data.Group{}
	err := decoder.Decode(newGroup)
	if err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	user := r.Context().Value("user").(*data.AuthUser)
	newGroup.UserID = user.ID
	log.Println(newGroup)
	err = s.store.AddGroup(newGroup)
	if err != nil {
		s.l.Printf("AddGroup(%d, %s):", user.ID, newGroup.Name)
		s.l.Println(err)
		http.Error(rw, "{\"status\":\"error\", \"message\":\"Unable to add group\"}", http.StatusInternalServerError)
		return
	}

	rw.Write([]byte("{\"status\":\"ok\"}"))
}

// UpdateGroup - update the given group
func (s *Service) UpdateGroup(rw http.ResponseWriter, r *http.Request) {

	user := r.Context().Value("user").(*data.AuthUser)
	vars := mux.Vars(r)
	fmt.Printf("vars: %+v\n", vars)

	//contentType == "application/json"
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	groupID, _ := strconv.Atoi(vars["id"])
	group, err := s.store.GetGroupByID(groupID)
	if err != nil {
		s.l.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("groupbyid: %v\n", group)

	fmt.Println("about to compare", user.ID, "with", group.UserID)
	if group.UserID != user.ID {
		s.l.Println(err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	someGroup := &data.Group{}
	if err := decoder.Decode(someGroup); err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}
	//s.l.Printf("someGroup: %+v", someGroup)

	var changed bool
	if someGroup.Name != "" && someGroup.Name != group.Name {
		//fmt.Println("about to update Name:", someGroup.Name, "=>", group.Name)
		group.Name = someGroup.Name
		changed = true
	}
	if someGroup.Status != "" && someGroup.Status != group.Status {
		//fmt.Println("about to update Status:", group.Status, "=>", someGroup.Status)
		group.Status = someGroup.Status
		changed = true
	}
	if changed {
		s.l.Printf("about to update: %+v", group)
		if err = s.store.UpdateGroup(&group); err != nil {
			s.l.Println("UpdateGroup:", err)
			http.Error(rw, "{\"status\":\"error\", \"message\":\"Unable to update group\"}", http.StatusBadRequest)
			return
		}
	}
	rw.Write([]byte("{\"status\":\"ok\"}"))
}

// About - about this site
func (s *Service) About(rw http.ResponseWriter, r *http.Request) {
	if err := s.t.ExecuteTemplate(rw, "about.gohtml", nil); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Service) SearchGoogleBooks(rw http.ResponseWriter, r *http.Request) {

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

// Library - list or add books to user's library
func (s *Service) Library(rw http.ResponseWriter, r *http.Request) {

	isLoggedIn := s.IsLoggedIn(r)
	if !isLoggedIn {
		http.Redirect(rw, r, "/login", http.StatusFound)
		return
	}

	session, err := s.session.Get(r, "session")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := session.Values["user_id"].(int)

	books, err := s.store.QueryByUserID(userID)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Books []data.Book
	}{
		Books: books,
	}

	if err := s.t.ExecuteTemplate(rw, "library.gohtml", data); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// AddBook - add new book to user's library
func (s *Service) AddBook(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Invalid request", http.StatusBadRequest)
		return
	}

	if !s.IsLoggedIn(r) {
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	session, _ := s.session.Get(r, "session")
	contentType := r.Header.Get("Content-Type")
	log.Println(contentType, len(contentType) == 1, contentType[0])
	if contentType != "application/json" {
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	newBook := data.NewBook{}
	//byteValue, _ := ioutil.ReadAll(r.Body)
	//s.l.Println("BODY:\n" + string(byteValue))
	//err := json.Unmarshal(byteValue, &newBook)

	r.Body = http.MaxBytesReader(rw, r.Body, 5120) // 5KB
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close()

	err := decoder.Decode(&newBook)
	if err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	userIDv := session.Values["user_id"]
	userID, _ := userIDv.(int)
	newBook.UserID = userID
	log.Println(newBook)
	book, err := s.store.AddBook(newBook)
	if err != nil {
		log.Println(err)
		http.Error(rw, "{\"status\":\"error\"}", http.StatusBadRequest)
		return
	}

	output := struct {
		Status string `json:"status"`
		Book   data.Book
	}{
		Status: "ok",
		Book:   book,
	}

	encoder := json.NewEncoder(rw)
	encoder.Encode(output)
}
