package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/models/view"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
	"net/http"
	"time"
)

type LoginHandler struct {
	config   *conf.Config
	userSrvc services.IUserService
}

func NewLoginHandler(userService services.IUserService) *LoginHandler {
	return &LoginHandler{
		config:   conf.Get(),
		userSrvc: userService,
	}
}

func (h *LoginHandler) RegisterRoutes(router *mux.Router) {
	router.Path("/login").Methods(http.MethodGet).HandlerFunc(h.GetIndex)
	router.Path("/login").Methods(http.MethodPost).HandlerFunc(h.PostLogin)
	router.Path("/logout").Methods(http.MethodPost).HandlerFunc(h.PostLogout)
	router.Path("/signup").Methods(http.MethodGet).HandlerFunc(h.GetSignup)
	router.Path("/signup").Methods(http.MethodPost).HandlerFunc(h.PostSignup)
}

func (h *LoginHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r))
}

func (h *LoginHandler) PostLogin(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	var login models.Login
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r).WithError("missing parameters"))
		return
	}
	if err := loginDecoder.Decode(&login, r.PostForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r).WithError("missing parameters"))
		return
	}

	user, err := h.userSrvc.GetUserById(login.Username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r).WithError("resource not found"))
		return
	}

	if !utils.CompareBcrypt(user.Password, login.Password, h.config.Security.PasswordSalt) {
		w.WriteHeader(http.StatusUnauthorized)
		templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r).WithError("invalid credentials"))
		return
	}

	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, login.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.LoginTemplate].Execute(w, h.buildViewModel(r).WithError("internal server error"))
		return
	}

	user.LastLoggedInAt = models.CustomTime(time.Now())
	h.userSrvc.Update(user)

	http.SetCookie(w, h.config.CreateCookie(models.AuthCookieKey, encoded, "/"))
	http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
}

func (h *LoginHandler) PostLogout(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	http.SetCookie(w, h.config.GetClearCookie(models.AuthCookieKey, "/"))
	http.Redirect(w, r, fmt.Sprintf("%s/", h.config.Server.BasePath), http.StatusFound)
}

func (h *LoginHandler) GetSignup(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r))
}

func (h *LoginHandler) PostSignup(w http.ResponseWriter, r *http.Request) {
	if h.config.IsDev() {
		loadTemplates()
	}

	if !h.config.IsDev() && !h.config.Security.AllowSignup {
		w.WriteHeader(http.StatusForbidden)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r).WithError("registration is disabled on this server"))
		return
	}

	if cookie, err := r.Cookie(models.AuthCookieKey); err == nil && cookie.Value != "" {
		http.Redirect(w, r, fmt.Sprintf("%s/summary", h.config.Server.BasePath), http.StatusFound)
		return
	}

	var signup models.Signup
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r).WithError("missing parameters"))
		return
	}
	if err := signupDecoder.Decode(&signup, r.PostForm); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r).WithError("missing parameters"))
		return
	}

	if !signup.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r).WithError("invalid parameters"))
		return
	}

	numUsers, _ := h.userSrvc.Count()

	_, created, err := h.userSrvc.CreateOrGet(&signup, numUsers == 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r).WithError("failed to create new user"))
		return
	}
	if !created {
		w.WriteHeader(http.StatusConflict)
		templates[conf.SignupTemplate].Execute(w, h.buildViewModel(r).WithError("user already existing"))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/?success=%s", h.config.Server.BasePath, "account created successfully"), http.StatusFound)
}

func (h *LoginHandler) buildViewModel(r *http.Request) *view.LoginViewModel {
	numUsers, _ := h.userSrvc.Count()

	return &view.LoginViewModel{
		Success:    r.URL.Query().Get("success"),
		Error:      r.URL.Query().Get("error"),
		TotalUsers: int(numUsers),
	}
}
