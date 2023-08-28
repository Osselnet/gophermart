package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

const (
	ContentTypeApplicationJSON = "application/json"
	ContentTypeTextPlain       = "text/plain"
	LogLvlDebug                = "[DEBUG]"
	LogLvlInfo                 = "[INFO]"
	LogLvlWarning              = "[WARNING]"
	LogLvlError                = "[ERROR]"
	LogLvlFatal                = "[FATAL]"
)

type handler struct {
	router chi.Router
	gm     *gophermart.GopherMart
}

func New(gm *gophermart.GopherMart) *handler {
	h := &handler{
		router: chi.NewRouter(),
		gm:     gm,
	}

	h.router.Use(middleware.Compress(3, "gzip"))
	h.router.Use(middleware.RequestID)
	h.router.Use(middleware.RealIP)
	h.router.Use(middleware.Logger)
	h.router.Use(middleware.Recoverer)

	h.router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)
		r.Get("/logout", h.logout)
		r.Get("/welcome", h.welcome)

		r.Post("/orders", h.postOrders)
		r.Get("/orders", h.getOrders)

		r.Get("/balance", h.getBalance)
		r.Post("/balance/withdraw", h.postWithdraw)
		r.Get("/balance/withdrawals", h.getWithdrawals)
	})

	return h
}

func (h *handler) GetRouter() chi.Router {
	return h.router
}

func (h *handler) log(r *http.Request, lvl, msg string) {
	reqID := middleware.GetReqID(r.Context())
	if reqID != "" {
		reqID = "[" + reqID + "] "
	}
	url := fmt.Sprintf(`"%s %s%s%s"`, r.Method, "http://", r.Host, r.URL)
	log.Printf("%s%s %s %s", reqID, lvl, url, msg)
}

func (h *handler) error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	reqID := middleware.GetReqID(r.Context())

	type errorJSON struct {
		Error      string
		StatusCode int
	}
	e := errorJSON{
		Error:      err.Error(),
		StatusCode: statusCode,
	}

	prefix := "[ERROR]"
	if reqID != "" {
		prefix = fmt.Sprintf("[%s] [ERROR]", reqID)
	}

	b, errMarshal := json.Marshal(e)
	if errMarshal != nil {
		msg := fmt.Sprintf(`{"Error": "Failed to marshal error - %s", "StatusCode": 500`, err)
		w.Write([]byte(msg))
		log.Println(prefix, msg)
		return
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(statusCode)
	w.Write(b)
	log.Println(prefix, e)
}
