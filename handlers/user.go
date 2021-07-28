package handlers

import (
	"context"
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
		r.ParseForm()
		if err := s.t.ExecuteTemplate(rw, "register.gohtml", formData); err != nil {
			log.Println(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		name := strings.Trim(r.Form.Get("name"), " ")
		email := strings.Trim(r.Form.Get("email"), " ")
		password := strings.Trim(r.Form.Get("password"), " ")

		log.Println(email, password)

		user, err := s.store.GetUser(email)
		s.l.Println("from GetUser:", err)
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
			http.Redirect(rw, r, "/login", http.StatusFound)
		}
	}
}

func (s *Service) UserLogIn(rw http.ResponseWriter, r *http.Request) {

	formData := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	if r.Method == "GET" {

		rw.Header().Add("Cache-Control", "no-cache")
		if err := s.t.ExecuteTemplate(rw, "login.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {

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

			//s.l.Printf("user checked OK: %v", user)

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
}

// UserAuth provides middleware functions for authorizing users and setting the user
// in the request context.
type Auth struct {
	Service *Service
}

// UserViaSession will retrieve the current user set by the session cookie
// and set it in the request context. UserViaSession will NOT redirect
// to the sign in page if the user is not found. That is left for the
// RequireUser method to handle so that some pages can optionally have
// access to the current user.
func (a *Auth) UserViaSession(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		session, err := a.Service.session.Get(r, "session")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		//a.Service.l.Printf("logged_in: %v\t%T", session.Values["logged_in"], session.Values["logged_in"])
		if session.Values["logged_in"] != true {
			next.ServeHTTP(w, r)
			return
		}

		user_id, _ := session.Values["user_id"].(int)
		user, err := a.Service.store.GetUserByID(user_id)
		if err != nil {
			// If you want you can retain the original functionality to call
			// http.Error if any error aside from app.ErrNotFound is returned,
			// but I find that most of the time we can continue on and let later
			// code error if it requires a user, otherwise it can continue without
			// the user.
			next.ServeHTTP(w, r)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "user", user))
		next.ServeHTTP(w, r)
	}
}

// RequireUser will verify that a user is set in the request context. It if is
// set correctly, the next handler will be called, otherwise it will redirect
// the user to the sign in page and the next handler will not be called.
func (a *Auth) RequireUser(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmp := r.Context().Value("user")
		if tmp == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if _, ok := tmp.(*data.AuthUser); !ok {
			// Whatever was set in the user key isn't a user, so we probably need to
			// sign in.
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	}
}
