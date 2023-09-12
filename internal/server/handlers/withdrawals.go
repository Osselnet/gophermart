package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"io"
	"net/http"
)

func (h *handler) postWithdraw(w http.ResponseWriter, r *http.Request) {
	var err error

	ct := r.Header.Get("Content-Type")
	if ct != ContentTypeApplicationJSON {
		err = fmt.Errorf("wrong content type, %s needed", ContentTypeApplicationJSON)
		h.error(w, r, err, http.StatusBadRequest)
		return
	}

	c := h.getSessionFromReqContext(r)
	if c == nil {
		return
	}

	u, err := h.gm.Users.Get(c.UserID)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to get user by ID - %w", err), http.StatusInternalServerError)
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to read request body - %w", err), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	wpr := &gophermart.WithdrawProxy{}
	err = json.Unmarshal(reqBody, &wpr)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to unmarshal body - %w", err), http.StatusBadRequest)
		return
	}

	wpr.UserID = u.ID
	err = h.gm.PostWithdraw(wpr)
	if err != nil {
		if errors.Is(err, gophermart.ErrNotEnoughFunds) {
			h.error(w, r, gophermart.ErrNotEnoughFunds, http.StatusPaymentRequired)
			return
		}

		if errors.Is(err, gophermart.ErrOrderInvalidFormat) {
			h.error(w, r, gophermart.ErrOrderInvalidFormat, http.StatusUnprocessableEntity)
			return
		}

		h.error(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf("new withdraw has been made for order ID %s", wpr.Order)
	h.log(r, LogLvlInfo, msg)
}

func (h *handler) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	var err error

	c := h.getSessionFromReqContext(r)
	if c == nil {
		return
	}

	u, err := h.gm.Users.Get(c.UserID)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to get user by ID - %w", err), http.StatusInternalServerError)
		return
	}

	wsPr, err := h.gm.GetWithdrawals(u.ID)
	if err != nil {
		if errors.Is(err, gophermart.ErrNoContent) {
			h.error(w, r, gophermart.ErrNoContent, http.StatusNoContent)
			return
		}

		h.error(w, r, err, http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(&wsPr)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to marshal JSON - %w", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.Write(body)
}
