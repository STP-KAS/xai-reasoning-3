package main

import (
	"fmt"
	"strings"
	"unicode"
)

// EPC payload per EPC069-12 (GiroCode). Euro SCT only.
// Empty BIC is allowed. Amount is EUR with a dot.
func epcPayload(name, iban, bic string, eurCents int, remittance string) (string, error) {
	iban = compactIBAN(iban)
	if !validIBAN(iban) {
		return "", fmt.Errorf("invalid IBAN")
	}
	name = clip(strings.TrimSpace(name), 70)
	if name == "" {
		return "", fmt.Errorf("beneficiary name required")
	}
	if eurCents <= 0 {
		return "", fmt.Errorf("amount must be positive")
	}
	if eurCents > 999999999 {
		return "", fmt.Errorf("amount too large for EPC")
	}
	remittance = clip(strings.TrimSpace(remittance), 140)
	bic = strings.ToUpper(strings.TrimSpace(bic))
	amount := fmt.Sprintf("EUR%d.%02d", eurCents/100, eurCents%100)
	lines := []string{
		"BCD",
		"002",
		"1",
		"SCT",
		bic,
		name,
		iban,
		amount,
		"", // purpose
		remittance,
	}
	return strings.Join(lines, "\n"), nil
}

func compactIBAN(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

func validIBAN(iban string) bool {
	if len(iban) < 15 || len(iban) > 34 {
		return false
	}
	for _, r := range iban {
		if !unicode.IsDigit(r) && (r < 'A' || r > 'Z') {
			return false
		}
	}
	rearr := iban[4:] + iban[:4]
	var n strings.Builder
	for _, r := range rearr {
		if r >= 'A' && r <= 'Z' {
			n.WriteString(fmt.Sprintf("%d", int(r-'A'+10)))
		} else {
			n.WriteRune(r)
		}
	}
	return mod97(n.String()) == 1
}

func mod97(digits string) int {
	acc := 0
	for _, r := range digits {
		acc = (acc*10 + int(r-'0')) % 97
	}
	return acc
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func kaspaURI(addr string, kas string) string {
	addr = strings.TrimSpace(addr)
	addr = strings.TrimPrefix(addr, "kaspa:")
	addr = strings.TrimPrefix(addr, "kaspatest:")
	if addr == "" {
		return ""
	}
	uri := "kaspa:" + addr
	kas = strings.TrimSpace(kas)
	if kas != "" {
		uri += "?amount=" + kas
	}
	return uri
}
