package g42

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	account *Account
	client  *http.Client
}

func NewHandler(account *Account) *Handler {
	return &Handler{
		account: account,
		client:  &http.Client{Timeout: 2 * time.Second},
	}
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	switch {
	case request.Method == http.MethodPost && request.URL.Path == "/redeem":
		h.redeem(response, request)
	case request.Method == http.MethodGet && request.URL.Path == "/state":
		writeJSON(response, http.StatusOK, h.account.Snapshot())
	case request.Method == http.MethodGet && request.URL.Path == "/export":
		writeJSON(response, http.StatusOK, h.account.Snapshot())
	case request.Method == http.MethodPost && request.URL.Path == "/sync":
		h.sync(response, request)
	default:
		http.NotFound(response, request)
	}
}

func (h *Handler) redeem(response http.ResponseWriter, request *http.Request) {
	var redemption Redemption
	if err := decodeJSON(request, &redemption); err != nil ||
		redemption.OperationID == "" || redemption.Person == "" || redemption.Amount <= 0 {
		http.Error(response, "invalid redemption", http.StatusBadRequest)
		return
	}

	err := h.account.Redeem(redemption)
	switch {
	case errors.Is(err, ErrOperationConflict):
		http.Error(response, err.Error(), http.StatusConflict)
	case errors.Is(err, ErrInsufficientValue):
		http.Error(response, err.Error(), http.StatusUnprocessableEntity)
	case err != nil:
		http.Error(response, err.Error(), http.StatusInternalServerError)
	default:
		writeJSON(response, http.StatusCreated, h.account.Snapshot())
	}
}

func (h *Handler) sync(response http.ResponseWriter, request *http.Request) {
	var input struct {
		PeerURL string `json:"peer_url"`
	}
	if err := decodeJSON(request, &input); err != nil || input.PeerURL == "" {
		http.Error(response, "invalid sync request", http.StatusBadRequest)
		return
	}

	peerResponse, err := h.client.Get(strings.TrimRight(input.PeerURL, "/") + "/export")
	if err != nil {
		http.Error(response, fmt.Sprintf("fetch peer state: %v", err), http.StatusBadGateway)
		return
	}
	defer peerResponse.Body.Close()
	if peerResponse.StatusCode != http.StatusOK {
		http.Error(response, fmt.Sprintf("fetch peer state: status %d", peerResponse.StatusCode), http.StatusBadGateway)
		return
	}

	var remote Snapshot
	decoder := json.NewDecoder(peerResponse.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&remote); err != nil {
		http.Error(response, fmt.Sprintf("decode peer state: %v", err), http.StatusBadGateway)
		return
	}
	if err := h.account.Merge(remote); err != nil {
		http.Error(response, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(response, http.StatusOK, h.account.Snapshot())
}

func decodeJSON(request *http.Request, value any) error {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(value); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_, _ = response.Write(body.Bytes())
}
