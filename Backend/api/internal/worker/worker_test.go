package worker

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestSanitizePGText(t *testing.T) {
	if got := sanitizePGText("Mozilla/5.0\x00Chrome\x00"); got != "Mozilla/5.0Chrome" {
		t.Fatalf("NUL baytları temizlenmeliydi, gelen: %q", got)
	}
}

func TestIsPermanentPGError(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{&pgconn.PgError{Code: "22021"}, true},
		{&pgconn.PgError{Code: "23505"}, true},
		{&pgconn.PgError{Code: "08006"}, false},
		{errors.New("bağlantı koptu"), false},
	}
	for _, c := range cases {
		if got := isPermanentPGError(c.err); got != c.want {
			t.Errorf("%v için %v beklendi, %v geldi", c.err, c.want, got)
		}
	}
}
