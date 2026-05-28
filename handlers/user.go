package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"read2succeed/data"
	"strings"
	"time"

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
			for _, m := range msgs {
				switch m.(string) {
				case "AccountCreated":
					formData["AccountCreated"] = true
				case "PasswordReset":
					formData["PasswordReset"] = true
				}
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

// ForgotPassword - renders the forgot-password form and issues a reset token
func (s *Service) ForgotPassword(rw http.ResponseWriter, r *http.Request) {
	formData := map[string]any{
		csrf.TemplateTag: csrf.TemplateField(r),
	}
	switch r.Method {
	case http.MethodGet:
		if err := s.t.ExecuteTemplate(rw, "forgot-password.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		r.ParseForm()
		email := strings.TrimSpace(r.Form.Get("email"))

		user, err := s.store.GetUser(email)
		if err == nil && user != nil {
			raw := make([]byte, 32)
			rand.Read(raw)
			token := hex.EncodeToString(raw)

			s.resetTokenMu.Lock()
			s.resetTokens[token] = resetToken{userID: user.ID, expiresAt: time.Now().Add(time.Hour)}
			s.resetTokenMu.Unlock()

			s.l.Printf("password reset token for %s: %s", email, token)
		}
		// always show the same message to avoid user enumeration
		formData["Message"] = "If that email exists, a reset link has been logged to the server."
		if err := s.t.ExecuteTemplate(rw, "forgot-password.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}
}

// ResetPassword - validates the token and sets a new password
func (s *Service) ResetPassword(rw http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	s.resetTokenMu.Lock()
	rt, ok := s.resetTokens[token]
	s.resetTokenMu.Unlock()

	formData := map[string]any{
		csrf.TemplateTag: csrf.TemplateField(r),
		"Token":          token,
	}

	if !ok || time.Now().After(rt.expiresAt) {
		formData["Error"] = "Reset link is invalid or has expired."
		if err := s.t.ExecuteTemplate(rw, "reset-password.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		if err := s.t.ExecuteTemplate(rw, "reset-password.gohtml", formData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		r.ParseForm()
		password := strings.TrimSpace(r.Form.Get("password"))
		if password == "" {
			formData["Error"] = "Password cannot be empty."
			s.t.ExecuteTemplate(rw, "reset-password.gohtml", formData)
			return
		}

		if err := s.store.UpdatePassword(rt.userID, password); err != nil {
			http.Error(rw, "failed to update password", http.StatusInternalServerError)
			return
		}
		s.l.Printf("password updated for user_id=%d", rt.userID)

		s.resetTokenMu.Lock()
		delete(s.resetTokens, token)
		s.resetTokenMu.Unlock()

		session, err := s.session.Get(r, "session")
		if err == nil {
			session.AddFlash("PasswordReset")
			session.Save(r, rw)
		}
		http.Redirect(rw, r, "/login", http.StatusFound)
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
