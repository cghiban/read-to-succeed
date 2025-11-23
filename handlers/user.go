package handlers

import (
	"net/http"
	"read2succeed/data"
	"strings"

	"github.com/gorilla/csrf"
)

// IsLoggedIn will check if the user has an active session and return True
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

	formData := map[string]any{
		csrf.TemplateTag: csrf.TemplateField(r),
	}
	switch r.Method {
	case http.MethodGet:
		if err := s.t.ExecuteTemplate(rw, "register.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		r.ParseForm()

		name := strings.Trim(r.Form.Get("name"), " ")
		email := strings.Trim(r.Form.Get("email"), " ")
		password := strings.Trim(r.Form.Get("password"), " ")

		user, _ := s.store.GetUser(email)
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

		err := s.store.CreateUser(user)
		if err != nil {
			http.Error(rw, "Unable to sign user up", http.StatusInternalServerError)
		} else {
			s.l.Printf("just signed up: %#v", user.Email)
			session, err := s.session.Get(r, "session")
			if err == nil {
				session.AddFlash("AccountCreated")
				session.Save(r, rw)
			}
			http.Redirect(rw, r, "/login", http.StatusFound)
		}
	}
}

func (s *Service) UserLogIn(rw http.ResponseWriter, r *http.Request) {

	formData := map[string]any{
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	switch r.Method {
	case http.MethodGet:
		session, err := s.session.Get(r, "session")
		if err == nil {
			msgs := session.Flashes()
			session.Save(r, rw) // needs to clear the flashes
			if len(msgs) > 0 && msgs[0].(string) == "AccountCreated" {
				formData["AccountCreated"] = true
			}
		}

		rw.Header().Add("Cache-Control", "no-cache")
		if err := s.t.ExecuteTemplate(rw, "login.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		email := strings.Trim(r.Form.Get("email"), " ")
		password := strings.Trim(r.Form.Get("password"), " ")

		user, err := s.store.GetUser(email)
		if err != nil {
			//http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else if user != nil && user.CheckPasswd(password) {
			session, err := s.session.Get(r, "session")
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			session.Values["logged_in"] = true
			session.Values["user_id"] = user.ID
			//session.Values["name"] = user.Name
			session.Values["is_admin"] = user.IsAdmin

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
	default:
		http.Error(rw, "bad request", http.StatusBadRequest)
		return
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
}
