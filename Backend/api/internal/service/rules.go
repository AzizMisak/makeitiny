package service

import (
	"math"
	"net/netip"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// Bu dosya, veritabanına dokunmayan ve birim testle doğrulanan kuralları toplar.

// internalPrefixes: loopback, özel ağlar, link-local (bulut metadata 169.254.169.254 dahil), CGNAT,
// multicast ve ayrılmış aralıklar. Bu adreslere kısa link verilmez.
var internalPrefixes = []netip.Prefix{
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

// IsInternalHost iç ağa işaret eden adresleri yakalar: özel/ayrılmış IP'ler, zone içeren IPv6,
// kısaltılmış/oktal/hex sayısal IPv4 biçimleri, localhost ve tek kelimelik makine adları.
// Alan adının DNS'te iç IP'ye çözülmesi ayrıca router'ın bağlantı anındaki denetimiyle engellenir.
func IsInternalHost(host string) bool {
	h := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if h == "" {
		return true
	}
	if addr, err := netip.ParseAddr(strings.Trim(h, "[]")); err == nil {
		if addr.Zone() != "" {
			return true
		}
		addr = addr.WithZone("").Unmap()
		for _, p := range internalPrefixes {
			if p.Contains(addr) {
				return true
			}
		}
		return false
	}
	if isNumericHost(h) {
		return true
	}
	if h == "localhost" || !strings.Contains(h, ".") {
		return true
	}
	for _, suffix := range []string{".localhost", ".local", ".internal", ".lan", ".home.arpa"} {
		if strings.HasSuffix(h, suffix) {
			return true
		}
	}
	return false
}

func isNumericHost(h string) bool {
	parts := strings.Split(h, ".")
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		if _, err := strconv.ParseUint(part, 0, 64); err != nil {
			return false
		}
	}
	return true
}

// IsSelfRedirect hedef URL'in uygulamanın kendi yönlendirme kapısına (/r/...) işaret edip
// döngü oluşturduğunu denetler.
func IsSelfRedirect(u *url.URL, publicBaseURL string) bool {
	if u == nil || strings.TrimSpace(publicBaseURL) == "" {
		return false
	}
	base, err := url.Parse(strings.TrimSpace(publicBaseURL))
	if err != nil || base.Hostname() == "" {
		return false
	}
	if !strings.EqualFold(u.Hostname(), base.Hostname()) {
		return false
	}
	clean := path.Clean("/" + u.Path)
	return clean == "/r" || strings.HasPrefix(clean, "/r/")
}

// CanResetOthersPassword başka bir hesabın parolasını kimin değiştirebileceğini belirler:
// sahip herkesinkini, yönetici yalnızca yetkisiz (standart) kullanıcınınkini.
func CanResetOthersPassword(actorRoles, targetRoles []string) bool {
	if HasAnyRole(actorRoles, "owner") {
		return true
	}
	return HasAnyRole(actorRoles, "admin") && !IsPrivileged(targetRoles)
}

// checkLinkPassword boş dize "parolayı kaldır" anlamına gelir; yalnızca boşluktan oluşan parola reddedilir
// (router kilit açmada boşluk parolayı kabul etmediği için böyle bir link açılamaz olurdu).
func checkLinkPassword(p *string) error {
	if p != nil && *p != "" && strings.TrimSpace(*p) == "" {
		return ErrBlankPassword
	}
	return nil
}

// ValidatePlanInput paket güncellemesindeki geçersiz değerleri sessizce yok saymak yerine reddeder.
func ValidatePlanInput(in UpdatePlanInput) error {
	if in.Name != nil && strings.TrimSpace(*in.Name) == "" {
		return ErrInvalidPlanValues
	}
	if in.PriceMonthly != nil && (*in.PriceMonthly < 0 || *in.PriceMonthly > 99999999.99 || math.IsNaN(*in.PriceMonthly) || math.IsInf(*in.PriceMonthly, 0)) {
		return ErrInvalidPlanValues
	}
	if in.MaxLinks != nil && *in.MaxLinks < -1 {
		return ErrInvalidPlanValues
	}
	if in.AnalyticsDays != nil && *in.AnalyticsDays < 1 {
		return ErrInvalidPlanValues
	}
	return nil
}
