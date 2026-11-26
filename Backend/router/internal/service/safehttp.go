package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

var errBlockedAddress = errors.New("hedef adres iç ağa işaret ediyor")

// blockedPrefixes: loopback, özel ağlar, link-local (bulut metadata 169.254.169.254 dahil),
// CGNAT, multicast ve ayrılmış aralıklar. Router bu adreslere hiçbir zaman istek atmaz.
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

// IsBlockedIP adresin iç ağa, zone'lu IPv6'ya ya da ayrılmış bir aralığa ait olup olmadığını söyler.
func IsBlockedIP(addr netip.Addr) bool {
	if addr.Zone() != "" {
		return true
	}
	addr = addr.WithZone("").Unmap()
	for _, p := range blockedPrefixes {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

// denyInternal bağlantı kurulmadan hemen önce, DNS çözümlemesinden sonraki gerçek IP'yi denetler.
// Böylece DNS rebinding ve iç adrese giden yönlendirmeler de engellenir.
func denyInternal(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return errBlockedAddress
	}
	addr, err := netip.ParseAddr(host)
	if err != nil || IsBlockedIP(addr) {
		return errBlockedAddress
	}
	return nil
}

var targetCheckClient = &http.Client{
	Transport: &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout: 3 * time.Second,
			Control: denyInternal,
		}).DialContext,
		TLSHandshakeTimeout:   3 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
		DisableKeepAlives:     true,
	},
	CheckRedirect: func(_ *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return http.ErrUseLastResponse
		}
		return nil
	},
}

func probe(ctx context.Context, method, target string) (int, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MakeItTiny/1.0)")
	if method == http.MethodGet {
		req.Header.Set("Range", "bytes=0-0")
	}
	resp, err := targetCheckClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	_ = resp.Body.Close()
	return resp.StatusCode, resp.Header, nil
}
