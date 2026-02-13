package handlers

import (
	"net/http"

	"autoimport-kz/internal/middleware"
	"autoimport-kz/internal/services"
	"github.com/gorilla/sessions"
)

type AuthHandler struct {
	auth        *services.AuthService
	renderer    *TemplateRenderer
	store       *sessions.CookieStore
	sessionName string
}

func NewAuthHandler(auth *services.AuthService, renderer *TemplateRenderer, store *sessions.CookieStore, sessionName string) *AuthHandler {
	return &AuthHandler{auth: auth, renderer: renderer, store: store, sessionName: sessionName}
}

type authPageData struct {
	Title           string
	ContentTemplate string
	Error           string
}

type profilePageData struct {
	Title           string
	ContentTemplate string
	User            any
}

func (h *AuthHandler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, "register.html", authPageData{Title: "Register", ContentTemplate: "register_content"})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderer.Render(w, "register.html", authPageData{Title: "Register", ContentTemplate: "register_content", Error: "Invalid form"})
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.auth.Register(r.Context(), name, email, password)
	if err != nil {
		h.renderer.Render(w, "register.html", authPageData{Title: "Register", ContentTemplate: "register_content", Error: err.Error()})
		return
	}

	session, _ := h.store.Get(r, h.sessionName)
	session.Values["user_id"] = user.ID
	session.Values["role"] = user.Role
	_ = session.Save(r, w)

	http.Redirect(w, r, "/cars", http.StatusFound)
}

func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, "login.html", authPageData{Title: "Login", ContentTemplate: "login_content"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderer.Render(w, "login.html", authPageData{Title: "Login", ContentTemplate: "login_content", Error: "Invalid form"})
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.auth.Login(r.Context(), email, password)
	if err != nil {
		h.renderer.Render(w, "login.html", authPageData{Title: "Login", ContentTemplate: "login_content", Error: "Invalid credentials"})
		return
	}

	session, _ := h.store.Get(r, h.sessionName)
	session.Values["user_id"] = user.ID
	session.Values["role"] = user.Role
	_ = session.Save(r, w)

	http.Redirect(w, r, "/cars", http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, h.sessionName)
	session.Options.MaxAge = -1
	_ = session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	h.renderer.Render(w, "profile.html", profilePageData{Title: "Profile", ContentTemplate: "profile_content", User: user})
}
