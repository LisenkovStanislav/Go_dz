package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	h := newRouter(NewStore())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type=%q", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing X-Request-ID header")
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status in body=%q", body.Status)
	}
}

func TestCreateAndGetOrder(t *testing.T) {
	store := NewStore()
	h := newRouter(store)

	// 1. Создание первого заказа
	reqBody := `{"customer":"  Анна  "}`
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d, want 201, body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type=%q, want application/json", rec.Header().Get("Content-Type"))
	}
	if loc := rec.Header().Get("Location"); loc != "/orders/1" {
		t.Fatalf("Location=%q, want /orders/1", loc)
	}

	var created Order
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID != 1 || created.Customer != "Анна" || created.Status != "new" {
		t.Fatalf("unexpected order: %+v", created)
	}

	// 2. Повторное создание выдаёт другой ID
	req2 := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"customer":"Иван"}`))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusCreated {
		t.Fatalf("status=%d, want 201", rec2.Code)
	}
	if loc := rec2.Header().Get("Location"); loc != "/orders/2" {
		t.Fatalf("Location=%q, want /orders/2", loc)
	}

	// 3. Чтение созданного заказа GET /orders/1
	getReq := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", getRec.Code)
	}
	if getRec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type=%q, want application/json", getRec.Header().Get("Content-Type"))
	}
	var fetched Order
	if err := json.Unmarshal(getRec.Body.Bytes(), &fetched); err != nil {
		t.Fatal(err)
	}
	if fetched != created {
		t.Fatalf("fetched=%+v, want %+v", fetched, created)
	}
}

func TestCreateOrderValidationErrors(t *testing.T) {
	cases := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
	}{
		{
			name:        "empty_body",
			contentType: "application/json",
			body:        "",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "malformed_json",
			contentType: "application/json",
			body:        `{"customer":`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "wrong_customer_type",
			contentType: "application/json",
			body:        `{"customer":123}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty_customer_name",
			contentType: "application/json",
			body:        `{"customer":"   "}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "unknown_field_id_or_status",
			contentType: "application/json",
			body:        `{"id":99,"customer":"Анна","status":"paid"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "second_json_value",
			contentType: "application/json",
			body:        `{"customer":"Анна"}{"customer":"Иван"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "unsupported_content_type",
			contentType: "text/plain",
			body:        `{"customer":"Анна"}`,
			wantStatus:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "body_too_large_over_1024",
			contentType: "application/json",
			body:        `{"customer":"` + strings.Repeat("a", 1100) + `"}`,
			wantStatus:  http.StatusRequestEntityTooLarge,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := NewStore()
			h := newRouter(store)

			req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status=%d, want %d, body=%s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if rec.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("Content-Type=%q, want application/json", rec.Header().Get("Content-Type"))
			}

			var errBody errorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
				t.Fatalf("invalid error JSON: %v", err)
			}
			if errBody.Error == "" {
				t.Fatal("expected non-empty error message")
			}

			// Проверяем, что заказ не был создан при ошибке
			if _, ok := store.ByID(1); ok {
				t.Fatal("store must not be modified on failed request")
			}
		})
	}
}

func TestGetOrderErrors(t *testing.T) {
	store := NewStore()
	store.Create("Анна")
	h := newRouter(store)

	cases := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{name: "unknown_order", path: "/orders/999", wantStatus: http.StatusNotFound},
		{name: "zero_id", path: "/orders/0", wantStatus: http.StatusBadRequest},
		{name: "negative_id", path: "/orders/-5", wantStatus: http.StatusBadRequest},
		{name: "non_numeric_id", path: "/orders/abc", wantStatus: http.StatusBadRequest},
		{name: "overflow_id", path: "/orders/99999999999999999999999999999", wantStatus: http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status=%d, want %d", rec.Code, tc.wantStatus)
			}
			if rec.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("Content-Type=%q, want application/json", rec.Header().Get("Content-Type"))
			}

			var errBody errorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
				t.Fatalf("invalid error JSON: %v", err)
			}
			if errBody.Error == "" {
				t.Fatal("expected non-empty error message")
			}
		})
	}
}

func TestRequestIDAndRecovery(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	r := newRouterWithLogger(NewStore(), logger)
	var seenCtxReqID string

	r.Get("/test-panic", func(w http.ResponseWriter, r *http.Request) {
		seenCtxReqID = RequestIDFromContext(r.Context())
		panic("unexpected failure")
	})

	req := httptest.NewRequest(http.MethodGet, "/test-panic", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, want 500", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type=%q, want application/json", rec.Header().Get("Content-Type"))
	}

	var errBody errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
		t.Fatal(err)
	}
	if errBody.Error == "" {
		t.Fatal("expected error message in 500 response")
	}

	headerReqID := rec.Header().Get("X-Request-ID")
	if headerReqID == "" || headerReqID != seenCtxReqID {
		t.Fatalf("X-Request-ID=%q, ctx request_id=%q", headerReqID, seenCtxReqID)
	}

	var logEntry struct {
		RequestID string `json:"request_id"`
		Method    string `json:"method"`
		Path      string `json:"path"`
		Status    int    `json:"status"`
	}
	if err := json.Unmarshal(logBuf.Bytes(), &logEntry); err != nil {
		t.Fatalf("invalid log JSON: %v, raw=%s", err, logBuf.String())
	}
	if logEntry.Status != http.StatusInternalServerError {
		t.Fatalf("logged status=%d, want 500", logEntry.Status)
	}
	if logEntry.RequestID != headerReqID {
		t.Fatalf("logged request_id=%q, want %q", logEntry.RequestID, headerReqID)
	}
}
