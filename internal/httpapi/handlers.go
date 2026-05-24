package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/yinkaolotin/tiny/internal/storage"
)

type Handler struct {
	store storage.Store
	log   zerolog.Logger
}

func New(store storage.Store, log zerolog.Logger) *Handler {
	return &Handler{store: store, log: log}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) Ready(w http.ResponseWriter, _ *http.Request) {
	if !h.store.Ready() {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

func (h *Handler) Items(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := h.store.List()
		json.NewEncoder(w).Encode(items)

	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
			TTL  int    `json:"ttl_seconds"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		item := h.store.Create(req.Name, time.Duration(req.TTL)*time.Second)
		h.log.Info().Str("item_id", item.ID).Msg("item created")

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(item)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
