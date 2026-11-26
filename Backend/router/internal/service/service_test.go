package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"testing"
	"time"

	"makeitiny/router/internal/repository"
)

func TestValidateLinkAccessOrder(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		rec  repository.LinkRecord
		want error
	}{
		{"engelli ve kapalı link engelli ekranına gider", repository.LinkRecord{SecurityStatus: "blocked", IsActive: false}, ErrLinkBlocked},
		{"zararlı ve kapalı link engelli ekranına gider", repository.LinkRecord{SecurityStatus: "malicious", IsActive: false}, ErrLinkBlocked},
		{"yalnızca kapalı", repository.LinkRecord{SecurityStatus: "safe", IsActive: false}, ErrLinkDisabled},
		{"süresi dolmuş", repository.LinkRecord{IsActive: true, ExpiresAtUnix: now.Add(-time.Minute).Unix()}, ErrLinkExpired},
		{"tıklama hakkı dolmuş", repository.LinkRecord{IsActive: true, MaxClicks: 3, Clicks: 3}, ErrLinkLimitReached},
		{"şüpheli ama açık", repository.LinkRecord{IsActive: true, SecurityStatus: "suspicious"}, nil},
		{"yayında", repository.LinkRecord{IsActive: true, MaxClicks: 3, Clicks: 2}, nil},
	}
	for _, c := range cases {
		if got := validateLinkAccess(c.rec, now); !errors.Is(got, c.want) {
			t.Errorf("%s: %v beklendi, %v geldi", c.name, c.want, got)
		}
	}
}

// Bekleme biletiyle devam ederken tıklama hakkı denetlenmez: hakkı dolduran son ziyaretçi de hedefe ulaşmalı.
func TestValidateLinkStateIgnoresClickLimit(t *testing.T) {
	rec := repository.LinkRecord{IsActive: true, MaxClicks: 1, Clicks: 1}
	if err := validateLinkState(rec, time.Now()); err != nil {
		t.Fatalf("hakkı dolan son ziyaret engellenmemeli: %v", err)
	}
}

func TestValidateWaitStateRejectsLockedLink(t *testing.T) {
	rec := repository.LinkRecord{IsActive: true, MaxClicks: 1, Clicks: 1}
	if err := validateWaitState(rec, time.Now()); err != nil {
		t.Fatalf("hakkı dolan son ziyaret engellenmemeli: %v", err)
	}
	locked := repository.LinkRecord{IsActive: true, PasswordHash: "$argon2id$v=19$m=65536,t=1,p=4$salt$hash"}
	if err := validateWaitState(locked, time.Now()); !errors.Is(err, ErrLinkLocked) {
		t.Fatalf("sonradan kilitlenen link ErrLinkLocked dönmeliydi, %v geldi", err)
	}
}

func TestIsBlockedIP(t *testing.T) {
	blocked := []string{
		"127.0.0.1", "10.1.2.3", "172.20.0.5", "192.168.1.1", "169.254.169.254",
		"100.64.0.1", "0.0.0.0", "::1", "::1%lo0", "fd00::1", "fe80::1", "fe80::1%eth0", "::ffff:127.0.0.1",
	}
	for _, s := range blocked {
		if !IsBlockedIP(netip.MustParseAddr(s)) {
			t.Errorf("%s engellenmeliydi", s)
		}
	}
	allowed := []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111", "172.32.0.1"}
	for _, s := range allowed {
		if IsBlockedIP(netip.MustParseAddr(s)) {
			t.Errorf("%s engellenmemeliydi", s)
		}
	}
}

func TestDenyInternal(t *testing.T) {
	if err := denyInternal("tcp", "127.0.0.1:80", nil); err == nil {
		t.Error("loopback bağlantısı reddedilmeliydi")
	}
	if err := denyInternal("tcp", "93.184.216.34:443", nil); err != nil {
		t.Errorf("genel adres kabul edilmeliydi: %v", err)
	}
}

func TestReachableStatus(t *testing.T) {
	for code, want := range map[int]bool{
		200: true, 301: true, 401: true, 403: true, 406: true, 429: true,
		404: false, 410: false, 500: false, 503: false,
	} {
		if got := reachableStatus(code); got != want {
			t.Errorf("%d için %v beklendi, %v geldi", code, want, got)
		}
	}
}

func TestIsWAFChallenge(t *testing.T) {
	if !isWAFChallenge(503, http.Header{"Cf-Mitigated": {"challenge"}}) {
		t.Error("Cf-Mitigated 503 WAF sayılmalıydı")
	}
	if !isWAFChallenge(503, http.Header{"Server": {"cloudflare"}}) {
		t.Error("Server: cloudflare 503 WAF sayılmalıydı")
	}
	if isWAFChallenge(503, http.Header{"Server": {"nginx"}}) {
		t.Error("düz 503 WAF sayılmamalıydı")
	}
	if isWAFChallenge(200, http.Header{"Server": {"cloudflare"}}) {
		t.Error("200 WAF sayılmamalıydı")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCheckTargetReachable(t *testing.T) {
	origTransport := targetCheckClient.Transport
	t.Cleanup(func() { targetCheckClient.Transport = origTransport })

	t.Run("HEAD 405 veya 404 sonrası GET Range ile 200 döner", func(t *testing.T) {
		var methods []string
		var gotRange string
		targetCheckClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			methods = append(methods, r.Method)
			if r.Method == http.MethodGet {
				gotRange = r.Header.Get("Range")
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(""))}, nil
			}
			return &http.Response{StatusCode: 405, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(""))}, nil
		})

		if err := checkTargetReachable(context.Background(), "https://example.com"); err != nil {
			t.Fatalf("erişilebilir bekleniyordu: %v", err)
		}
		if len(methods) != 2 || methods[0] != "HEAD" || methods[1] != "GET" {
			t.Fatalf("HEAD ardından GET bekleniyordu: %v", methods)
		}
		if gotRange != "bytes=0-0" {
			t.Fatalf("Range bytes=0-0 beklendi, %q geldi", gotRange)
		}
	})

	t.Run("HEAD ağ hatasında GET fallback çalışır", func(t *testing.T) {
		var methods []string
		targetCheckClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			methods = append(methods, r.Method)
			if r.Method == http.MethodHead {
				return nil, errors.New("eof")
			}
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(""))}, nil
		})

		if err := checkTargetReachable(context.Background(), "https://example.com"); err != nil {
			t.Fatalf("GET fallback ile başarılı olmalıydı: %v", err)
		}
		if len(methods) != 2 {
			t.Fatalf("2 deneme beklendi: %v", methods)
		}
	})

	t.Run("SSRF hatasında GET fallback çalışmaz", func(t *testing.T) {
		calls := 0
		targetCheckClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			return nil, errBlockedAddress
		})

		if err := checkTargetReachable(context.Background(), "https://example.com"); !errors.Is(err, ErrTargetUnreachable) {
			t.Fatalf("ErrTargetUnreachable beklendi, %v geldi", err)
		}
		if calls != 1 {
			t.Fatalf("SSRF engellemesinde tek deneme beklendi, %d oldu", calls)
		}
	})

	t.Run("Cloudflare 503 geçerli, düz 503 erişilemez", func(t *testing.T) {
		targetCheckClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 503,
				Header:     http.Header{"Server": {"cloudflare"}},
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		})
		if err := checkTargetReachable(context.Background(), "https://example.com"); err != nil {
			t.Fatalf("Cloudflare 503 erişilebilir sayılmalıydı: %v", err)
		}

		targetCheckClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 503,
				Header:     http.Header{"Server": {"nginx"}},
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		})
		if err := checkTargetReachable(context.Background(), "https://example.com"); !errors.Is(err, ErrTargetUnreachable) {
			t.Fatalf("düz 503 ErrTargetUnreachable dönmeliydi, %v geldi", err)
		}
	})
}

func TestVerifyPasswordLimitedTimeout(t *testing.T) {
	svc := &RedirectService{unlockSem: make(chan struct{}, 1)}
	svc.unlockSem <- struct{}{}
	defer func() { <-svc.unlockSem }()

	_, err := svc.verifyPasswordLimited(context.Background(), "secret", "$argon2id$v=19$m=65536,t=1,p=4$salt$hash")
	if !errors.Is(err, ErrTooManyRequests) {
		t.Fatalf("ErrTooManyRequests beklendi, %v geldi", err)
	}
}
