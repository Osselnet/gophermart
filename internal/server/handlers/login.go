package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"io"
	"net/http"
	"time"
)

func (h *handler) register(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != ContentTypeApplicationJSON {
		err := fmt.Errorf("wrong content type, JSON needed")
		h.error(w, r, err, http.StatusBadRequest)
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to read request body - %w", err), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var creds gophermart.Credentials
	err = json.Unmarshal(reqBody, &creds)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to unmarshal body - %w", err), http.StatusBadRequest)
		return
	}

	session, err := h.gm.Register(&creds)
	if err != nil {
		msg := "failed to register new user"
		if errors.Is(err, gophermart.ErrLoginAlreadyTaken) {
			h.error(w, r, fmt.Errorf("%s - %w", msg, err), http.StatusConflict)
			return
		}
		h.error(w, r, fmt.Errorf("%s - %w", msg, err), http.StatusInternalServerError)
		return
	}
	if session == nil {
		h.error(w, r, fmt.Errorf("got nil session"), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   session.Token,
		Expires: session.Expiry,
	})

	msg := fmt.Sprintf("session for user `%s` successfully created", creds.Login)
	h.log(r, LogLvlDebug, msg)
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var creds *gophermart.Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		h.error(w, r, err, http.StatusBadRequest)
		return
	}

	var sessionToken string
	c, err := r.Cookie("session_token")
	if err == nil {
		sessionToken = c.Value
	}

	session, err := h.gm.Login(creds, sessionToken)
	if err != nil {
		if errors.Is(err, gophermart.ErrInvalidPair) || errors.Is(err, gophermart.ErrUserNotFound) {
			h.error(w, r, gophermart.ErrInvalidPair, http.StatusUnauthorized)
			return
		}
		h.error(w, r, err, http.StatusInternalServerError)
		return
	}
	if session == nil {
		h.error(w, r, fmt.Errorf("got nil session"), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   session.Token,
		Expires: session.Expiry,
	})
	msg := fmt.Sprintf("session for user `%s` successfully created", creds.Login)
	h.log(r, LogLvlDebug, msg)
}

func (h *handler) logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			h.error(w, r, gophermart.ErrUnauthorizedAccess, http.StatusUnauthorized)
			return
		}
		h.error(w, r, err, http.StatusBadRequest)
		return
	}

	err = h.gm.Logout(c.Value)
	if err != nil {
		h.log(r, LogLvlError, fmt.Sprintf("failed to delete session - %s", err))
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})

	msg := fmt.Sprintf("logout session %s", c.Value)
	h.log(r, LogLvlDebug, msg)
}

func (h *handler) welcome(w http.ResponseWriter, r *http.Request) {
	session := h.getSessionFromReqContext(r)
	if session == nil {
		return
	}

	u, err := h.gm.Users.Get(session.UserID)
	if err != nil {
		return
	}

	w.Write([]byte(fmt.Sprintf("Welcome, #%d %s!", u.ID, u.Login)))
}
