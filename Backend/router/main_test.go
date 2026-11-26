package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"makeitiny/router/internal/handler"
)

func TestNewRootHandler(t *testing.T) {
	h := handler.New(nil, "http://localhost")
	var gotPath, gotMethod string
	routes := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	})
	root := newRootHandler(h, routes)

	cases := []struct {
		name         string
		method       string
		path         string
		wantCode     int
		wantLocation string
		wantRoute    string
	}{
		{"kök", http.MethodGet, "/", http.StatusFound, "http://localhost/error?reason=not_found", ""},
		{"çoklu slash", http.MethodGet, "///", http.StatusFound, "http://localhost/error?reason=not_found", ""},
		{"sonda slash tek segment GET", http.MethodGet, "/abc/", http.StatusOK, "", "/abc"},
		{"sonda slash tek segment HEAD", http.MethodHead, "/abc/", http.StatusOK, "", "/abc"},
		{"health sonda slash korunur", http.MethodGet, "/health/", http.StatusOK, "", "/health/"},
		{"bilinmeyen çok segmentli GET", http.MethodGet, "/abc/unknown", http.StatusFound, "http://localhost/error?reason=not_found", ""},
		{"bilinmeyen çok segmentli sonda slash GET", http.MethodGet, "/abc/unknown/", http.StatusFound, "http://localhost/error?reason=not_found", ""},
		{"preview yolu geçer", http.MethodGet, "/abc/preview", http.StatusOK, "", "/abc/preview"},
		{"wait yolu geçer", http.MethodGet, "/abc/wait", http.StatusOK, "", "/abc/wait"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotPath, gotMethod = "", ""
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(c.method, c.path, nil)
			root.ServeHTTP(rec, req)

			if rec.Code != c.wantCode {
				t.Fatalf("status %d beklendi, %d geldi", c.wantCode, rec.Code)
			}
			if c.wantLocation != "" && rec.Header().Get("Location") != c.wantLocation {
				t.Fatalf("Location %q beklendi, %q geldi", c.wantLocation, rec.Header().Get("Location"))
			}
			if c.wantRoute != "" && (gotPath != c.wantRoute || gotMethod != c.method) {
				t.Fatalf("route (%s %s) beklendi, (%s %s) geldi", c.method, c.wantRoute, gotMethod, gotPath)
			}
		})
	}
}
