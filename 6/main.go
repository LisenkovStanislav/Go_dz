package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

const maxBodyBytes = 1024

type errorResponse struct {
	Error string `json:"error"`
}

type createOrderRequest struct {
	Customer string `json:"customer"`
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (sr *statusRecorder) WriteHeader(code int) {
	if !sr.wroteHeader {
		sr.status = code
		sr.wroteHeader = true
		sr.ResponseWriter.WriteHeader(code)
	}
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if !sr.wroteHeader {
		sr.WriteHeader(http.StatusOK)
	}
	return sr.ResponseWriter.Write(b)
}

func generateRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b[:])
}

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if reqID == "" {
			reqID = generateRequestID()
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func accessLogMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			logger.Info("http_request",
				slog.String("request_id", RequestIDFromContext(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			)
		})
	}
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				writeJSON(w, http.StatusInternalServerError, errorResponse{
					Error: "internal server error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func newRouter(store *Store) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return newRouterWithLogger(store, logger)
}

func newRouterWithLogger(store *Store, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(requestIDMiddleware)
	r.Use(accessLogMiddleware(logger))
	r.Use(recoverMiddleware)

	r.Get("/health", health)
	r.Post("/orders", createOrder(store))
	r.Get("/orders/{id}", getOrder(store))

	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func createOrder(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "" {
			mediaType, _, err := mime.ParseMediaType(ct)
			if err != nil || mediaType != "application/json" {
				writeJSON(w, http.StatusUnsupportedMediaType, errorResponse{
					Error: "Content-Type должен быть application/json",
				})
				return
			}
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		defer r.Body.Close()

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var req createOrderRequest
		if err := dec.Decode(&req); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeJSON(w, http.StatusRequestEntityTooLarge, errorResponse{
					Error: "тело запроса превышает 1024 байта",
				})
				return
			}
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: "некорректный JSON",
			})
			return
		}

		// Проверка на отсутствие второго JSON-значения после первого
		if err := dec.Decode(&struct{}{}); err != io.EOF {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeJSON(w, http.StatusRequestEntityTooLarge, errorResponse{
					Error: "тело запроса превышает 1024 байта",
				})
				return
			}
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: "лишние данные после JSON-объекта",
			})
			return
		}

		customer := strings.TrimSpace(req.Customer)
		if customer == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: "имя клиента не может быть пустым",
			})
			return
		}

		order := store.Create(customer)
		w.Header().Set("Location", fmt.Sprintf("/orders/%d", order.ID))
		writeJSON(w, http.StatusCreated, order)
	}
}

func getOrder(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: "некорректный идентификатор заказа",
			})
			return
		}

		order, ok := store.ByID(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, errorResponse{
				Error: "заказ не найден",
			})
			return
		}

		writeJSON(w, http.StatusOK, order)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("encode response: %v", err)
		status = http.StatusInternalServerError
		data = []byte(`{"error":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		log.Printf("write response: %v", err)
	}
}

func main() {
	server := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           newRouter(NewStore()),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
