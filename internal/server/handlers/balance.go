package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *handler) getBalance(w http.ResponseWriter, r *http.Request) {
	var err error

	c := h.getSessionFromReqContext(r)
	if c == nil {
		return
	}

	balanceProxy, err := h.gm.GetBalance(c.UserID)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to get balance for user - %w", err), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(&balanceProxy)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to marshal JSON - %w", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.Write(body)
}
