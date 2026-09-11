package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Invoice struct {
	ID          string `json:"id"`
	Created     string `json:"created"`
	Description string `json:"description"`
	EURCents    int    `json:"eur_cents"`
	VATPercent  int    `json:"vat_percent,omitempty"`
	KASAmount   string `json:"kas_amount,omitempty"` // merchant-typed, not an oracle
	Remittance  string `json:"remittance"`
	Paid        bool   `json:"paid"`
	PaidAt      string `json:"paid_at,omitempty"`
	PaidHow     string `json:"paid_how,omitempty"` // sepa | cash | kas | prepaid | other
	PaidRef     string `json:"paid_ref,omitempty"`
}

type Store struct {
	mu   sync.Mutex
	path string
	list []Invoice
}

func openStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "invoices.json")}
	b, err := os.ReadFile(s.path)
	if err == nil {
		_ = json.Unmarshal(b, &s.list)
	}
	return s, nil
}

func (s *Store) save() error {
	b, err := json.MarshalIndent(s.list, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) add(inv Invoice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.list = append([]Invoice{inv}, s.list...)
	return s.save()
}

func (s *Store) get(id string) (Invoice, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, inv := range s.list {
		if inv.ID == id {
			return inv, true
		}
	}
	return Invoice{}, false
}

func (s *Store) all() []Invoice {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Invoice, len(s.list))
	copy(out, s.list)
	return out
}

func (s *Store) markPaid(id, how, ref string) (Invoice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.list {
		if s.list[i].ID != id {
			continue
		}
		if s.list[i].Paid {
			return s.list[i], nil
		}
		s.list[i].Paid = true
		s.list[i].PaidAt = time.Now().UTC().Format(time.RFC3339)
		s.list[i].PaidHow = how
		s.list[i].PaidRef = ref
		if err := s.save(); err != nil {
			return Invoice{}, err
		}
		return s.list[i], nil
	}
	return Invoice{}, fmt.Errorf("invoice not found")
}

func newID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return "inv_" + time.Now().UTC().Format("20060102") + "_" + hex.EncodeToString(b[:])
}

func csv(list []Invoice) string {
	out := "id,created,description,eur,kas_amount,paid,paid_how,paid_ref,paid_at,remittance\n"
	for _, inv := range list {
		eur := fmt.Sprintf("%d.%02d", inv.EURCents/100, inv.EURCents%100)
		out += fmt.Sprintf("%s,%s,%q,%s,%s,%t,%s,%q,%s,%q\n",
			inv.ID, inv.Created, inv.Description, eur, inv.KASAmount,
			inv.Paid, inv.PaidHow, inv.PaidRef, inv.PaidAt, inv.Remittance)
	}
	return out
}
