package handlers

import (
	"log"
	"net/http"
	"read2succeed/data"
	"strings"

	"github.com/gorilla/csrf"
)

//IsLoggedIn will check if the user has an active session and return True
func (s *Service) IsLoggedIn(r *http.Request) bool {
	session, err := s.session.Get(r, "session")
	if err != nil {
		s.l.Println("error in IsLoggedIn():", err)
		return false
	}
	if session.Values["logged_in"] == true {
		return true
	}
	return false
}

// UserSignUp - handles user signup
func (s *Service) UserSignUp(rw http.ResponseWriter, r *http.Request) {

	formData := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	}
	if r.Method == "GET" {
		if err := s.t.ExecuteTemplate(rw, "register.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err := s.t.ExecuteTemplate(rw, "register.gohtml", formData); err != nil {
			log.Println(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		name := strings.Trim(r.Form.Get("name"), " ")
		email := strings.Trim(r.Form.Get("email"), " ")
		password := strings.Trim(r.Form.Get("password"), " ")

		log.Println(email, password)

		user, err := s.store.GetUser(email)
		if user != nil {
			formData["Message"] = "This email is already in use."
			if err := s.t.ExecuteTemplate(rw, "register.gohtml", formData); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		user = &data.AuthUser{
			Name:  name,
			Email: email,
			Pass:  password,
		}

		err = s.store.CreateUser(user)
		if err != nil {
			http.Error(rw, "Unable to sign user up", http.StatusInternalServerError)
		} else {
			s.l.Printf("user: %#v", user)
			http.Redirect(rw, r, "/login", 302)
		}
	}
}

func (s *Service) UserLogIn(rw http.ResponseWriter, r *http.Request) {

	rw.Header().Add("Cache-Control", "no-cache")

	formData := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	if r.Method == "GET" {
		if err := s.t.ExecuteTemplate(rw, "login.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err := s.t.ExecuteTemplate(rw, "login.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		email := strings.Trim(r.Form.Get("email"), " ")
		password := strings.Trim(r.Form.Get("password"), " ")

		user, err := s.store.GetUser(email)
		if err != nil {
			//http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else if user != nil && user.CheckPasswd(password) {
			session, _ := s.session.Get(r, "session")

			session.Values["logged_in"] = true
			session.Values["user_id"] = user.ID
			session.Values["name"] = user.Name

			//readers, _ := s.store.GetUserReaders(user.ID)
			//session.Values["readers"] = readers //.([]data.Reader)
			err = session.Save(r, rw)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(rw, r, "/", http.StatusFound)
			return
		}

		formData["Message"] = "Invalid email or password!"
		if err := s.t.ExecuteTemplate(rw, "login.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}
}

// UserLogOut - clears the session
func (s *Service) UserLogOut(rw http.ResponseWriter, r *http.Request) {

	session, err := s.session.Get(r, "session")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["logged_in"] = false
	session.Options.MaxAge = -1

	err = session.Save(r, rw)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(rw, r, "/", http.StatusFound)
	return
}
