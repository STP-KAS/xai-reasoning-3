package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	store, err := openStore(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		Name:       "Test desk",
		IBAN:       compactIBAN("DE89 3704 0044 0532 0130 00"),
		Kaspa:      "qpjm8kzpcj5he3hg9msrdc78a3k46zda866pucwetprgtgc7s3ry2kq38atpq",
		PrepaidKey: "secret-key",
		DataDir:    dir,
	}
	s := newServer(cfg, store)
	return httptest.NewServer(s.routes())
}

func createInvoice(t *testing.T, ts *httptest.Server) string {
	t.Helper()
	res, err := http.PostForm(ts.URL+"/invoices", url.Values{
		"description": {"Website work"},
		"eur":         {"12.50"},
		"kas":         {"1.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 && res.Request.URL.Path == "/invoices" {
		t.Fatalf("status %d path %s", res.StatusCode, res.Request.URL.Path)
	}
	parts := strings.Split(strings.Trim(res.Request.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "i" {
		t.Fatalf("redirect path %s", res.Request.URL.Path)
	}
	return parts[1]
}

func TestCreateAndCSV(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	id := createInvoice(t, ts)
	res, err := http.Get(ts.URL + "/csv")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if !strings.Contains(string(b), id) || !strings.Contains(string(b), "12.50") {
		t.Fatalf("csv: %s", b)
	}
}

func Test402RefusesUnverifiedKaspa(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	id := createInvoice(t, ts)
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/invoices/"+id+"/resource", nil)
	req.Header.Set("X-Kaspa-Payment", "deadbeef")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("got %d", res.StatusCode)
	}
	var body map[string]string
	_ = json.NewDecoder(res.Body).Decode(&body)
	if !strings.Contains(body["error"], "unverified") {
		t.Fatalf("%v", body)
	}
}

func Test402ThenPrepaid(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	id := createInvoice(t, ts)
	res, err := http.Get(ts.URL + "/api/invoices/" + id + "/resource")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("got %d", res.StatusCode)
	}
	res.Body.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/invoices/"+id+"/resource", nil)
	req.Header.Set("X-Prepaid-Key", "secret-key")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("got %d", res.StatusCode)
	}
	var body map[string]any
	_ = json.NewDecoder(res.Body).Decode(&body)
	if body["paid"] != true {
		t.Fatalf("%v", body)
	}
}

func TestQREndpoints(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	id := createInvoice(t, ts)
	for _, path := range []string{"/qr/epc.png", "/qr/kaspa.png"} {
		res, err := http.Get(ts.URL + "/i/" + id + path)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode != 200 || len(b) < 100 || string(b[:8]) != "\x89PNG\r\n\x1a\n" {
			t.Fatalf("%s status %d len %d", path, res.StatusCode, len(b))
		}
	}
}

func TestMarkPaidReceipt(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	id := createInvoice(t, ts)
	res, err := http.PostForm(ts.URL+"/i/"+id+"/paid", url.Values{"how": {"cash"}, "ref": {"drawer"}})
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if !strings.HasSuffix(res.Request.URL.Path, "/receipt") {
		t.Fatalf("path %s", res.Request.URL.Path)
	}
	res, err = http.Get(ts.URL + "/i/" + id + "/receipt")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 || !strings.Contains(string(b), "drawer") {
		t.Fatalf("receipt %d %s", res.StatusCode, b)
	}
}

func TestParseEUR(t *testing.T) {
	n, err := parseEUR("12,5")
	if err != nil || n != 1250 {
		t.Fatalf("%d %v", n, err)
	}
	if _, err := parseEUR(""); err == nil {
		t.Fatal("empty")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
