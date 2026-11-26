package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"makeitiny/router/internal/gen"
	"makeitiny/router/internal/service"
)

func TestParseUserAgentOS(t *testing.T) {
	cases := map[string]string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Safari/604.1":     "iOS",
		"Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) AppleWebKit/605.1.15":                           "iOS",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36":      "Windows",
		"Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 Chrome/120.0 Mobile Safari/537.36":         "Android",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Version/17.0 Safari/605": "macOS",
		"axios/1.6.0":                     "Diğer",
		"Mozilla/5.0 (X11; Linux x86_64)": "Linux",
	}
	for ua, want := range cases {
		if _, _, got := parseUserAgent(ua); got != want {
			t.Errorf("%q: %s beklendi, %s geldi", ua, want, got)
		}
	}
}

func TestReasonFor(t *testing.T) {
	cases := map[error]string{
		service.ErrLinkBlocked:       "blocked",
		service.ErrLinkDisabled:      "disabled",
		service.ErrLinkExpired:       "expired",
		service.ErrLinkLimitReached:  "limit_reached",
		service.ErrLinkNotFound:      "not_found",
		service.ErrTargetUnreachable: "unreachable",
		errors.New("başka"):          "internal",
	}
	for err, want := range cases {
		if got := reasonFor(err); got != want {
			t.Errorf("%v: %s beklendi, %s geldi", err, want, got)
		}
	}
}

func TestRootNotFoundRedirectsToErrorPage(t *testing.T) {
	h := New(nil, "http://localhost")
	rec := httptest.NewRecorder()
	h.RootNotFound(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "http://localhost/error?reason=not_found" {
		t.Fatalf("beklenmeyen yanıt: %d %s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestWriteServiceErrorLinkLocked(t *testing.T) {
	rec := httptest.NewRecorder()
	writeServiceError(rec, service.ErrLinkLocked, "hata")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("403 beklendi, %d geldi", rec.Code)
	}
}

func TestWriteServiceErrorTooManyRequests(t *testing.T) {
	rec := httptest.NewRecorder()
	writeServiceError(rec, service.ErrTooManyRequests, "hata")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("429 beklendi, %d geldi", rec.Code)
	}
	var body gen.ApiError
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("json decode hatası: %v", err)
	}
	if body.Code != "too_many_requests" {
		t.Fatalf("too_many_requests beklendi, %q geldi", body.Code)
	}
}

func TestWaitInfoEmptyTicketReturns400(t *testing.T) {
	h := New(nil, "http://localhost")
	for _, ticket := range []string{"", "   "} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/yaz/wait", nil)
		h.WaitInfo(rec, req, "yaz", gen.WaitInfoParams{T: ticket})
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("t=%q için 400 beklendi, %d geldi", ticket, rec.Code)
		}
		var body gen.ApiError
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("json decode hatası: %v", err)
		}
		if body.Code != "invalid_request" {
			t.Fatalf("invalid_request beklendi, %q geldi", body.Code)
		}
	}
}

func TestUnlockAndContinueWaitBodyLimit(t *testing.T) {
	h := New(nil, "http://localhost")
	big := bytes.Repeat([]byte("a"), 65*1024)

	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/yaz/unlock", bytes.NewReader([]byte(`{"password":"`+string(big)+`"}`)))
	h.Unlock(rec1, req1, "yaz")
	if rec1.Code != http.StatusBadRequest {
		t.Fatalf("Unlock büyük gövdede 400 beklendi, %d geldi", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/yaz/continue", bytes.NewReader([]byte(`{"ticket":"`+string(big)+`"}`)))
	h.ContinueWait(rec2, req2, "yaz")
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("ContinueWait büyük gövdede 400 beklendi, %d geldi", rec2.Code)
	}
}
