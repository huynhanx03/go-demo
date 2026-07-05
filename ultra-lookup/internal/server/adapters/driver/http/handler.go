package http

import (
	"encoding/json"
	"net/http"
	"time"

	"ultra-lookup/internal/metrics"
	"ultra-lookup/internal/server/core/domain"
	"ultra-lookup/internal/server/core/dto"
	"ultra-lookup/internal/server/core/mapper"
	"ultra-lookup/internal/server/core/model"
	"ultra-lookup/internal/server/core/service"
)

const maxShardValue = 255

type Handler struct {
	svc *service.LookupService
}

func NewHandler(svc *service.LookupService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Lookup(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	kindRaw, idRaw, encodedID := mapper.LookupInputFromRequest(r)

	kind, err := domain.ParseKind(kindRaw)
	if err != nil {
		observeLookup("unknown", "bad_request", domain.LookupSourceNone, start)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := domain.ParseID(idRaw, encodedID)
	if err != nil {
		observeLookup(kind.String(), "bad_request", domain.LookupSourceNone, start)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := h.svc.Lookup(kind, id)
	if result.Found {
		observeLookup(kind.String(), "hit", result.Source, start)
		writeJSON(w, http.StatusOK, mapper.LookupResponseFromResult(result))
		return
	}

	observeLookup(kind.String(), "miss", domain.LookupSourceNone, start)
	writeJSON(w, http.StatusOK, dto.LookupResponse{Found: false, Source: domain.LookupSourceNone})
}

func (h *Handler) Append(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req dto.AppendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		observeAppend("unknown", "bad_request", start)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entry, err := mapper.EntryFromAppendRequest(req)
	if err != nil {
		observeAppend("unknown", "bad_request", start)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Shard > maxShardValue {
		observeAppend(entry.Kind.String(), "bad_request", start)
		http.Error(w, "shard must be in 0..255", http.StatusBadRequest)
		return
	}

	added, deltaEntries, err := h.svc.Append(entry)
	if err != nil {
		observeAppend(entry.Kind.String(), "duplicate", start)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	observeAppend(entry.Kind.String(), "ok", start)
	writeJSON(w, http.StatusOK, dto.AppendResponse{
		Added:        added,
		DeltaEntries: deltaEntries,
	})
}

func (h *Handler) Healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, model.Status{Status: "ok"})
}

func observeLookup(kind, result, source string, start time.Time) {
	metrics.LookupRequests.WithLabelValues(kind, result, source).Inc()
	metrics.LookupLatency.WithLabelValues(kind, result, source).Observe(time.Since(start).Seconds())
}

func observeAppend(kind, result string, start time.Time) {
	metrics.AppendRequests.WithLabelValues(kind, result).Inc()
	metrics.AppendLatency.WithLabelValues(kind, result).Observe(time.Since(start).Seconds())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
