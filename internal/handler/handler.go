package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/luqmanrafi/urlshortener/internal/model"
	"github.com/luqmanrafi/urlshortener/internal/service"
)

type Handler struct {
	svc *service.Shortener
	baseURL string
}

func New(svc *service.Shortener, baseURL string) *Handler{
	return &Handler{svc: svc, baseURL: baseURL}
}

func (h *Handler) Routes()http.Handler{
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/shorten", h.shorten)
	mux.HandleFunc("GET /{code}", h.redirect)
	mux.HandleFunc("GET /api/stats/{code}",h.stats)
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /readyz", h.ready)
	return mux
}

type shortenRequest struct {
	URL string `json:"url"`
	Alias string `json:"alias"`
	ExpiresIn int `json:"expires_in"`
}

type shortenResponse struct{
	Code string `json:"code"`
	ShortURL string `json:"short_url"`
}

type statsResponse struct{
	Code string `json:"code"`
	OriginalURL string `json:"original_url"`
	ClickCount int `json:"click_count"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request){
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.ExpiresIn < 0 {
		writeError(w, http.StatusBadRequest, "expires_in must not be negative")
		return
	}

	code, err := h.svc.Shorten(r.Context(), req.URL, req.Alias, time.Duration(req.ExpiresIn)*time.Second)
	if err != nil {
		if errors.Is(err, model.ErrInvalidUrl){
			writeError(w, http.StatusBadRequest, "url must be valid http/https URL")
			return
		}
		if errors.Is(err, model.ErrInvalidAlias){
			writeError(w, http.StatusBadRequest, "alias must 3-30 chars (letters, numbers, - or _ ) and not reserved yet")
			return
		}
		if errors.Is(err, model.ErrCodeExists){
			writeError(w, http.StatusConflict, "alias already in use")
			return
		}
		
		log.Printf("shorten error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, shortenResponse{
		Code: code,
		ShortURL: h.baseURL + "/" + code,
	})

}
func (h *Handler) redirect(w http.ResponseWriter, r *http.Request){
		code := r.PathValue("code")

		original, err := h.svc.Resolve(r.Context(),code)
		if err != nil{
			if errors.Is(err, model.ErrNotFound){
				http.NotFound(w, r)
				return
			}
			if errors.Is(err, model.ErrExpired){
				http.Error(w, "link has expired", http.StatusGone)
				return
			}
			log.Printf("resolve error: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, original, http.StatusFound)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request){
	code := r.PathValue("code")
	u, err := h.svc.Stats(r.Context(), code)
	if err != nil{
		if errors.Is(err, model.ErrNotFound){
			http.NotFound(w, r)
			return
		}
		log.Printf("stats error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, statsResponse{
		Code: u.Code,
		OriginalURL: u.Original,
		ClickCount: u.ClickCount,
		CreatedAt: u.CreatedAt,
	})
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request){
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.svc.Ready(ctx); err != nil{
		log.Printf("ready: %v", err)
		writeError(w, http.StatusServiceUnavailable, "not ready")
		// return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}


func writeJSON(w http.ResponseWriter, status int, v any){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string){
	writeJSON(w, status, map[string]string{"error": msg})
}