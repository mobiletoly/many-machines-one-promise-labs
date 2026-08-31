package orders

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/health":
		response.WriteHeader(http.StatusNoContent)
	case request.Method == http.MethodPost && request.URL.Path == "/orders":
		h.create(response, request)
	default:
		http.NotFound(response, request)
	}
}

func (h *Handler) create(response http.ResponseWriter, request *http.Request) {
	operationID := request.Header.Get("Idempotency-Key")
	if operationID == "" {
		http.Error(response, "Idempotency-Key is required", http.StatusBadRequest)
		return
	}

	var create CreateRequest
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&create); err != nil || create.Wine == "" || create.Quantity <= 0 {
		http.Error(response, "invalid create request", http.StatusBadRequest)
		return
	}

	order, err := h.store.Create(request.Context(), operationID, create)
	switch {
	case errors.Is(err, ErrOperationConflict):
		http.Error(response, err.Error(), http.StatusConflict)
	case err != nil:
		http.Error(response, err.Error(), http.StatusInternalServerError)
	default:
		writeJSON(response, http.StatusCreated, order)
	}
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
