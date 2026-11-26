package service

import (
	"errors"
	"net/url"
	"testing"
)

func TestValidateTargetURLRejectsInternalHosts(t *testing.T) {
	const publicBase = "https://makeitiny.com"
	bad := []string{
		"http://127.0.0.1:6379",
		"http://127.1",
		"http://0177.0.0.1",
		"http://0x7f.0.0.1",
		"http://2130706433",
		"http://169.254.169.254/latest/meta-data",
		"http://10.0.0.5/admin",
		"http://192.168.1.1",
		"http://[::1]:8080",
		"http://[::1%25lo0]:8080",
		"http://[fe80::1%25eth0]",
		"http://localhost:8080",
		"http://api:8080/health",
		"http://postgres",
		"http://servis.internal",
		"http://yazici.local",
		"http://kullanici:parola@ornek.com",
		"https://makeitiny.com/r/abc",
		"https://makeitiny.com/r",
	}
	for _, raw := range bad {
		if _, err := validateTargetURL(raw, publicBase); err == nil {
			t.Errorf("%s reddedilmeliydi", raw)
		}
	}

	good := map[string]string{
		"ornek.com/sayfa":            "https://ornek.com/sayfa",
		"http://magaza.ornek.com/a":  "http://magaza.ornek.com/a",
		"https://8.8.8.8/dns":        "https://8.8.8.8/dns",
		"https://makeitiny.com/hakk": "https://makeitiny.com/hakk",
	}
	for raw, want := range good {
		got, err := validateTargetURL(raw, publicBase)
		if err != nil || got != want {
			t.Errorf("%s: %s beklendi, %s / %v geldi", raw, want, got, err)
		}
	}

	if _, err := validateTargetURL("http://127.0.0.1", publicBase); !errors.Is(err, ErrInternalTargetURL) {
		t.Errorf("iç adres için ErrInternalTargetURL beklendi, %v geldi", err)
	}
	if _, err := validateTargetURL("http://api:8080/health", publicBase); !errors.Is(err, ErrInternalTargetURL) {
		t.Errorf("http://api:8080/health için ErrInternalTargetURL beklendi, %v geldi", err)
	}
	if _, err := validateTargetURL("https://makeitiny.com/r/deneme", publicBase); !errors.Is(err, ErrInternalTargetURL) {
		t.Errorf("kendi /r/ adresine yönlendirme için ErrInternalTargetURL beklendi, %v geldi", err)
	}
	for _, invalid := range []string{"not-a-valid-url", "just-a-word", "bad url.com", ""} {
		if _, err := validateTargetURL(invalid, publicBase); !errors.Is(err, ErrInvalidTargetURL) {
			t.Errorf("%q için ErrInvalidTargetURL beklendi, %v geldi", invalid, err)
		}
	}
}

func TestIsSelfRedirectNonCanonicalPaths(t *testing.T) {
	const publicBase = "https://makeitiny.com"
	selfPaths := []string{
		"https://makeitiny.com/r",
		"https://makeitiny.com/r/",
		"https://makeitiny.com/r/abc",
		"https://makeitiny.com//r/abc",
		"https://makeitiny.com/foo/../r/abc",
		"https://makeitiny.com/r/foo/..",
		"https://makeitiny.com/a/b/../../r/xyz",
	}
	for _, raw := range selfPaths {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("url parse hatası (%s): %v", raw, err)
		}
		if !IsSelfRedirect(u, publicBase) {
			t.Errorf("%s self-redirect olarak algılanmalıydı", raw)
		}
	}

	nonSelf := []string{
		"https://makeitiny.com/register",
		"https://makeitiny.com/r2/abc",
		"https://other.com/r/abc",
	}
	for _, raw := range nonSelf {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("url parse hatası (%s): %v", raw, err)
		}
		if IsSelfRedirect(u, publicBase) {
			t.Errorf("%s self-redirect olmamalıydı", raw)
		}
	}
}

func TestCanResetOthersPassword(t *testing.T) {
	owner := []string{"user", "owner"}
	admin := []string{"user", "admin"}
	user := []string{"user"}

	cases := []struct {
		actor, target []string
		want          bool
	}{
		{owner, admin, true},
		{owner, user, true},
		{admin, user, true},
		{admin, admin, false},
		{admin, owner, false},
		{user, user, false},
	}
	for _, c := range cases {
		if got := CanResetOthersPassword(c.actor, c.target); got != c.want {
			t.Errorf("aktör %v, hedef %v: %v beklendi", c.actor, c.target, c.want)
		}
	}
}

func TestCheckLinkPassword(t *testing.T) {
	str := func(s string) *string { return &s }
	if err := checkLinkPassword(nil); err != nil {
		t.Error("parola gönderilmemesi geçerli")
	}
	if err := checkLinkPassword(str("")); err != nil {
		t.Error("boş dize parolayı kaldırır, geçerli")
	}
	if err := checkLinkPassword(str("   ")); !errors.Is(err, ErrBlankPassword) {
		t.Error("yalnızca boşluk reddedilmeli")
	}
	if err := checkLinkPassword(str(" gizli ")); err != nil {
		t.Error("boşluk içeren gerçek parola geçerli")
	}
}

func TestValidatePlanInput(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	i := func(v int32) *int32 { return &v }
	s := func(v string) *string { return &v }

	invalid := []UpdatePlanInput{
		{Name: s("  ")},
		{PriceMonthly: f(-1)},
		{PriceMonthly: f(100000000.00)},
		{MaxLinks: i(-2)},
		{AnalyticsDays: i(0)},
	}
	for _, in := range invalid {
		if err := ValidatePlanInput(in); !errors.Is(err, ErrInvalidPlanValues) {
			t.Errorf("%+v reddedilmeliydi", in)
		}
	}
	if err := ValidatePlanInput(UpdatePlanInput{Name: s("Pro"), PriceMonthly: f(99999999.99), MaxLinks: i(-1), AnalyticsDays: i(1)}); err != nil {
		t.Errorf("geçerli değerler reddedildi: %v", err)
	}
}
