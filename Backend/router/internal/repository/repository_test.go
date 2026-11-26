package repository

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestParseWaitTicket(t *testing.T) {
	ready := time.Now().Add(5 * time.Second).UnixMilli()
	got, ok := parseWaitTicket("yaz-26|" + itoa(ready))
	if !ok || got.Alias != "yaz-26" || got.ReadyAt.UnixMilli() != ready {
		t.Fatalf("bilet çözülemedi: %+v %v", got, ok)
	}
	for _, bad := range []string{"", "yaz26", "|123", "yaz26|abc"} {
		if _, ok := parseWaitTicket(bad); ok {
			t.Errorf("%q geçersiz sayılmalıydı", bad)
		}
	}
}

func TestClickDeltaFieldIsNotAnAlias(t *testing.T) {
	// "id:" öneki takma ad kuralında (harf, rakam, tire, alt çizgi) bulunamaz; rakamdan oluşan takma adlarla karışmaz.
	if got := ClickDeltaField(42); got != "id:42" {
		t.Fatalf("id:42 beklendi, %s geldi", got)
	}
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func startFakeRedis(t *testing.T, fn func(cmd []string) string) *redis.Client {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("fake redis listen hatası: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				r := bufio.NewReader(c)
				for {
					args, err := readRESPArray(r)
					if err != nil {
						return
					}
					if len(args) == 0 {
						continue
					}
					upper := strings.ToUpper(args[0])
					if upper == "HELLO" {
						_, _ = c.Write([]byte("-ERR unknown command\r\n"))
						continue
					}
					if upper == "PING" {
						_, _ = c.Write([]byte("+PONG\r\n"))
						continue
					}
					if upper == "CLIENT" {
						_, _ = c.Write([]byte("+OK\r\n"))
						continue
					}
					_, _ = c.Write([]byte(fn(args)))
				}
			}(conn)
		}
	}()

	rdb := redis.NewClient(&redis.Options{
		Addr:         ln.Addr().String(),
		MaxRetries:   0,
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func readRESPArray(r *bufio.Reader) ([]string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	if !strings.HasPrefix(line, "*") {
		return strings.Fields(line), nil
	}
	n, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, n)
	for i := 0; i < n; i++ {
		hdr, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		hdr = strings.TrimSuffix(strings.TrimSuffix(hdr, "\n"), "\r")
		if !strings.HasPrefix(hdr, "$") {
			return nil, fmt.Errorf("unexpected header: %s", hdr)
		}
		sz, err := strconv.Atoi(hdr[1:])
		if err != nil {
			return nil, err
		}
		buf := make([]byte, sz+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		args = append(args, string(buf[:sz]))
	}
	return args, nil
}

func TestWaitTicketErrorsAndConsume(t *testing.T) {
	ctx := context.Background()

	// Redis yokken ErrNoRedis döner
	noRedisRepo := NewLinkRepository(nil, nil)
	if _, err := noRedisRepo.GetWaitTicket(ctx, "abc"); !errors.Is(err, ErrNoRedis) {
		t.Fatalf("ErrNoRedis beklendi, %v geldi", err)
	}
	if ok, err := noRedisRepo.ConsumeWaitTicket(ctx, "abc"); ok || !errors.Is(err, ErrNoRedis) {
		t.Fatalf("(false, ErrNoRedis) beklendi, (%v, %v) geldi", ok, err)
	}

	store := map[string]string{
		"wait:bad":   "bozuk-deger",
		"wait:valid": "yaz|" + itoa(time.Now().Add(time.Minute).UnixMilli()),
	}
	rdb := startFakeRedis(t, func(cmd []string) string {
		switch strings.ToUpper(cmd[0]) {
		case "GET":
			v, ok := store[cmd[1]]
			if !ok {
				return "$-1\r\n"
			}
			return fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
		case "DEL":
			if _, ok := store[cmd[1]]; ok {
				delete(store, cmd[1])
				return ":1\r\n"
			}
			return ":0\r\n"
		default:
			return "-ERR unsupported\r\n"
		}
	})
	repo := NewLinkRepository(nil, rdb)

	if _, err := repo.GetWaitTicket(ctx, ""); !errors.Is(err, ErrTicketNotFound) {
		t.Fatalf("boş bilette ErrTicketNotFound beklendi, %v geldi", err)
	}
	if _, err := repo.GetWaitTicket(ctx, "missing"); !errors.Is(err, ErrTicketNotFound) {
		t.Fatalf("redis.Nil için ErrTicketNotFound beklendi, %v geldi", err)
	}
	if _, err := repo.GetWaitTicket(ctx, "bad"); !errors.Is(err, ErrTicketNotFound) {
		t.Fatalf("bozuk bilette ErrTicketNotFound beklendi, %v geldi", err)
	}

	if ok, err := repo.ConsumeWaitTicket(ctx, "valid"); err != nil || !ok {
		t.Fatalf("ilk ConsumeWaitTicket (true, nil) beklendi, (%v, %v) geldi", ok, err)
	}
	if ok, err := repo.ConsumeWaitTicket(ctx, "valid"); err != nil || ok {
		t.Fatalf("ikinci ConsumeWaitTicket (false, nil) beklendi, (%v, %v) geldi", ok, err)
	}

	// Bağlantı hatasında raw Redis hatası korunur
	deadClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: 0, DialTimeout: 50 * time.Millisecond})
	defer deadClient.Close()
	deadRepo := NewLinkRepository(nil, deadClient)
	if _, err := deadRepo.GetWaitTicket(ctx, "valid"); err == nil || errors.Is(err, ErrTicketNotFound) {
		t.Fatalf("ağ hatasında raw error beklendi, %v geldi", err)
	}
	if ok, err := deadRepo.ConsumeWaitTicket(ctx, "valid"); ok || err == nil {
		t.Fatalf("ağ hatasında (false, err) beklendi, (%v, %v) geldi", ok, err)
	}
}

func TestPendingClicksUsesProcsSet(t *testing.T) {
	ctx := context.Background()
	var procs []string
	scanCalled := false

	rdb := startFakeRedis(t, func(cmd []string) string {
		switch strings.ToUpper(cmd[0]) {
		case "SCAN":
			scanCalled = true
			return "*2\r\n$1\r\n0\r\n*0\r\n"
		case "SMEMBERS":
			if cmd[1] != clickDeltasProcsKey {
				return "*0\r\n"
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("*%d\r\n", len(procs)))
			for _, p := range procs {
				sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(p), p))
			}
			return sb.String()
		case "HMGET":
			key := cmd[1]
			var v1, v2 string
			if key == clickDeltasKey {
				v1 = "3"
			} else if key == "clicks:deltas:proc:1" {
				v2 = "2"
			}
			res := []string{v1, v2}
			var sb strings.Builder
			sb.WriteString("*2\r\n")
			for _, v := range res {
				if v == "" {
					sb.WriteString("$-1\r\n")
				} else {
					sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
				}
			}
			return sb.String()
		default:
			return "-ERR unsupported\r\n"
		}
	})
	repo := NewLinkRepository(nil, rdb)

	total, flushing := repo.pendingClicks(ctx, 42, "yaz")
	if total != 3 || flushing {
		t.Fatalf("procs boşken (3, false) beklendi, (%d, %v) geldi", total, flushing)
	}

	procs = []string{"clicks:deltas:proc:1"}
	total, flushing = repo.pendingClicks(ctx, 42, "yaz")
	if total != 5 || !flushing {
		t.Fatalf("procs doluyken (5, true) beklendi, (%d, %v) geldi", total, flushing)
	}
	if scanCalled {
		t.Fatal("pendingClicks SCAN çağırmamalıydı")
	}
}

func TestAllowUnlockAttempt(t *testing.T) {
	ctx := context.Background()

	// Redis yokken fail-open (true)
	if !NewLinkRepository(nil, nil).AllowUnlockAttempt(ctx, "yaz", "1.2.3.4") {
		t.Fatal("Redis yokken fail-open true olmalıydı")
	}

	counts := map[string]int64{}
	expireSet := false
	rdb := startFakeRedis(t, func(cmd []string) string {
		switch strings.ToUpper(cmd[0]) {
		case "INCR":
			counts[cmd[1]]++
			return fmt.Sprintf(":%d\r\n", counts[cmd[1]])
		case "EXPIRE":
			expireSet = true
			return ":1\r\n"
		default:
			return "-ERR unsupported\r\n"
		}
	})
	repo := NewLinkRepository(nil, rdb)

	for i := 1; i <= 10; i++ {
		if !repo.AllowUnlockAttempt(ctx, "yaz", "1.2.3.4") {
			t.Fatalf("%d. denemeye izin verilmeliydi", i)
		}
	}
	if !expireSet {
		t.Fatal("ilk denemede EXPIRE ayarlanmalıydı")
	}
	if repo.AllowUnlockAttempt(ctx, "yaz", "1.2.3.4") {
		t.Fatal("11. deneme engellenmeliydi")
	}
}
