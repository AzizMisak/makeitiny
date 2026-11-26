package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSVSafe(t *testing.T) {
	cases := map[string]string{
		"=HYPERLINK(\"x\")":     "'=HYPERLINK(\"x\")",
		"+90 555":               "'+90 555",
		"-1":                    "'-1",
		"@SUM(A1)":              "'@SUM(A1)",
		"\t=1+1":                "'\t=1+1",
		"\r+1":                  "'\r+1",
		"\n=1+1":                "'\n=1+1",
		"  =HYPERLINK(\"x\")":   "'  =HYPERLINK(\"x\")",
		"\n  +SUM(1)":           "'\n  +SUM(1)",
		" \t\r\n@SUM(A1)":       "' \t\r\n@SUM(A1)",
		"  normal text":         "  normal text",
		"https://ornek.com":     "https://ornek.com",
		"":                      "",
	}
	for in, want := range cases {
		if got := csvSafe(in); got != want {
			t.Errorf("%q: %q beklendi, %q geldi", in, want, got)
		}
	}
}

func TestWriteCSVHasBOMAndAttachment(t *testing.T) {
	rec := httptest.NewRecorder()
	writeCSV(rec, "deneme.csv", []string{"Başlık"}, [][]string{{"=1+1"}, {"Çağrı"}})

	body := rec.Body.String()
	bom := string([]byte{0xEF, 0xBB, 0xBF})
	if !strings.HasPrefix(body, bom) {
		t.Error("UTF-8 BOM eksik")
	}
	if !strings.Contains(body, "'=1+1") || !strings.Contains(body, "Çağrı") {
		t.Errorf("içerik hatalı: %q", body)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, `filename="deneme.csv"`) {
		t.Errorf("Content-Disposition hatalı: %s", cd)
	}
}

func TestExportFormat(t *testing.T) {
	s := func(v string) *string { return &v }
	if f, ok := exportFormat(nil); !ok || f != "csv" {
		t.Error("varsayılan csv olmalı")
	}
	if f, ok := exportFormat(s("JSON")); !ok || f != "json" {
		t.Error("büyük harfli json kabul edilmeli")
	}
	if _, ok := exportFormat(s("xml")); ok {
		t.Error("xml reddedilmeli")
	}
}
