package orders

import (
	"encoding/json"
	"errors"
	"net/http"
)

const dropResponseHeader = "X-Lab-Drop-After-Commit"

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost || request.URL.Path != "/orders" {
		http.NotFound(response, request)
		return
	}

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

	order, err := h.store.Create(operationID, create)
	if errors.Is(err, ErrOperationConflict) {
		http.Error(response, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	if request.Header.Get(dropResponseHeader) == "true" {
		connection, _, err := http.NewResponseController(response).Hijack()
		if err != nil {
			panic("episode fault seam cannot hijack HTTP connection: " + err.Error())
		}
		_ = connection.Close()
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(response).Encode(order)
}
