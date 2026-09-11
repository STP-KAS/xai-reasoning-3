package main

import (
	"strings"
	"testing"
)

func TestValidIBAN(t *testing.T) {
	// Public example IBAN (Bundesbank / Wikipedia-style DE test).
	if !validIBAN(compactIBAN("DE89 3704 0044 0532 0130 00")) {
		t.Fatal("expected valid DE IBAN")
	}
	if validIBAN("DE00") {
		t.Fatal("short IBAN")
	}
	if validIBAN(compactIBAN("DE89 3704 0044 0532 0130 01")) {
		t.Fatal("bad checksum should fail")
	}
}

func TestEPCPayload(t *testing.T) {
	p, err := epcPayload("Desk", "DE89370400440532013000", "", 12000, "inv_1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(p, "BCD\n002\n1\nSCT\n") {
		t.Fatalf("header: %q", p)
	}
	if !strings.Contains(p, "EUR120.00") {
		t.Fatalf("amount: %q", p)
	}
	if !strings.Contains(p, "inv_1") {
		t.Fatalf("remittance: %q", p)
	}
}

func TestKaspaURI(t *testing.T) {
	u := kaspaURI("qpjm8kzpcj5he3hg9msrdc78a3k46zda866pucwetprgtgc7s3ry2kq38atpq", "1.5")
	if !strings.HasPrefix(u, "kaspa:") || !strings.Contains(u, "amount=1.5") {
		t.Fatal(u)
	}
}
