package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	cfg   Config
	store *Store
	tpl   *template.Template
}

func newServer(cfg Config, store *Store) *Server {
	tpl := template.Must(template.New("").Funcs(template.FuncMap{
		"eur": func(cents int) string { return fmt.Sprintf("%d.%02d", cents/100, cents%100) },
	}).Parse(pages))
	return &Server{cfg: cfg, store: store, tpl: tpl}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.home)
	mux.HandleFunc("POST /invoices", s.create)
	mux.HandleFunc("GET /i/{id}", s.show)
	mux.HandleFunc("POST /i/{id}/paid", s.paid)
	mux.HandleFunc("GET /i/{id}/receipt", s.receipt)
	mux.HandleFunc("GET /i/{id}/qr/epc.png", s.qrEPC)
	mux.HandleFunc("GET /i/{id}/qr/kaspa.png", s.qrKaspa)
	mux.HandleFunc("GET /csv", s.csv)
	mux.HandleFunc("GET /api/invoices/{id}", s.apiGet)
	mux.HandleFunc("GET /api/invoices/{id}/resource", s.paywall)
	mux.HandleFunc("POST /api/invoices/{id}/resource", s.paywall)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "ok\n")
	})
	return mux
}

type page struct {
	Cfg     Config
	Inv     Invoice
	List    []Invoice
	Err     string
	HasIBAN bool
	HasKAS  bool
	KasURI  string
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	s.render(w, "home", page{Cfg: s.cfg, List: s.store.all(), HasIBAN: validIBAN(s.cfg.IBAN), HasKAS: s.cfg.Kaspa != ""})
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.fail(w, "bad form")
		return
	}
	cents, err := parseEUR(r.FormValue("eur"))
	if err != nil {
		s.fail(w, err.Error())
		return
	}
	desc := strings.TrimSpace(r.FormValue("description"))
	if desc == "" {
		s.fail(w, "description required")
		return
	}
	vat, _ := strconv.Atoi(r.FormValue("vat"))
	if vat < 0 || vat > 100 {
		vat = 0
	}
	id := newID()
	inv := Invoice{
		ID:          id,
		Created:     time.Now().UTC().Format(time.RFC3339),
		Description: desc,
		EURCents:    cents,
		VATPercent:  vat,
		KASAmount:   strings.TrimSpace(r.FormValue("kas")),
		Remittance:  id,
	}
	if err := s.store.add(inv); err != nil {
		s.fail(w, err.Error())
		return
	}
	http.Redirect(w, r, "/i/"+id, http.StatusSeeOther)
}

func (s *Server) show(w http.ResponseWriter, r *http.Request) {
	inv, ok := s.store.get(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	s.render(w, "invoice", s.invPage(inv, ""))
}

func (s *Server) paid(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		s.fail(w, "bad form")
		return
	}
	how := r.FormValue("how")
	switch how {
	case "sepa", "cash", "kas", "other":
	default:
		s.fail(w, "how must be sepa, cash, kas, or other")
		return
	}
	inv, err := s.store.markPaid(id, how, strings.TrimSpace(r.FormValue("ref")))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/i/"+inv.ID+"/receipt", http.StatusSeeOther)
}

func (s *Server) receipt(w http.ResponseWriter, r *http.Request) {
	inv, ok := s.store.get(r.PathValue("id"))
	if !ok || !inv.Paid {
		http.NotFound(w, r)
		return
	}
	s.render(w, "receipt", s.invPage(inv, ""))
}

func (s *Server) qrEPC(w http.ResponseWriter, r *http.Request) {
	inv, ok := s.store.get(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	payload, err := epcPayload(s.cfg.Name, s.cfg.IBAN, s.cfg.BIC, inv.EURCents, inv.Remittance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	png, err := qrPNG(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

func (s *Server) qrKaspa(w http.ResponseWriter, r *http.Request) {
	inv, ok := s.store.get(r.PathValue("id"))
	if !ok || s.cfg.Kaspa == "" {
		http.NotFound(w, r)
		return
	}
	png, err := qrPNG(kaspaURI(s.cfg.Kaspa, inv.KASAmount))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

func (s *Server) csv(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=desk-invoices.csv")
	fmt.Fprint(w, csv(s.store.all()))
}

func (s *Server) apiGet(w http.ResponseWriter, r *http.Request) {
	inv, ok := s.store.get(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	s.writeJSON(w, http.StatusOK, s.public(inv))
}

func (s *Server) paywall(w http.ResponseWriter, r *http.Request) {
	inv, ok := s.store.get(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	if tx := strings.TrimSpace(r.Header.Get("X-Kaspa-Payment")); tx != "" {
		s.writeJSON(w, http.StatusForbidden, map[string]string{
			"error": "unverified kaspa txid refused",
			"note":  "this till does not trust X-Kaspa-Payment. mark paid in the UI, or send X-Prepaid-Key.",
		})
		return
	}
	if key := strings.TrimSpace(r.Header.Get("X-Prepaid-Key")); key != "" {
		if s.cfg.PrepaidKey == "" || key != s.cfg.PrepaidKey {
			s.writeJSON(w, http.StatusForbidden, map[string]string{"error": "bad prepaid key"})
			return
		}
		if !inv.Paid {
			var err error
			inv, err = s.store.markPaid(inv.ID, "prepaid", "X-Prepaid-Key")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		s.writeJSON(w, http.StatusOK, s.public(inv))
		return
	}
	if inv.Paid {
		s.writeJSON(w, http.StatusOK, s.public(inv))
		return
	}
	accepts := []map[string]string{}
	if validIBAN(s.cfg.IBAN) {
		accepts = append(accepts, map[string]string{
			"rail": "sepa-epc", "iban": s.cfg.IBAN,
			"amount": fmt.Sprintf("EUR%d.%02d", inv.EURCents/100, inv.EURCents%100),
			"remittance": inv.Remittance,
		})
	}
	if s.cfg.Kaspa != "" {
		accepts = append(accepts, map[string]string{
			"rail": "kaspa", "uri": kaspaURI(s.cfg.Kaspa, inv.KASAmount),
			"note": "typed KAS amount, not an oracle. unverified txids are refused.",
		})
	}
	s.writeJSON(w, http.StatusPaymentRequired, map[string]any{
		"status":   402,
		"invoice":  inv.ID,
		"accepts":  accepts,
		"prepaid":  "header X-Prepaid-Key (if DESK_PREPAID_KEY is set)",
		"refused":  "X-Kaspa-Payment is never accepted unverified",
		"human":    "/i/" + inv.ID,
	})
}

func (s *Server) public(inv Invoice) map[string]any {
	return map[string]any{
		"id":          inv.ID,
		"description": inv.Description,
		"eur":         fmt.Sprintf("%d.%02d", inv.EURCents/100, inv.EURCents%100),
		"kas_amount":  inv.KASAmount,
		"paid":        inv.Paid,
		"paid_how":    inv.PaidHow,
		"paid_ref":    inv.PaidRef,
		"paid_at":     inv.PaidAt,
		"receipt":     inv.Paid,
	}
}

func (s *Server) invPage(inv Invoice, err string) page {
	return page{
		Cfg:     s.cfg,
		Inv:     inv,
		Err:     err,
		HasIBAN: validIBAN(s.cfg.IBAN),
		HasKAS:  s.cfg.Kaspa != "",
		KasURI:  kaspaURI(s.cfg.Kaspa, inv.KASAmount),
	}
}

func (s *Server) fail(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	s.render(w, "home", page{Cfg: s.cfg, List: s.store.all(), Err: msg, HasIBAN: validIBAN(s.cfg.IBAN), HasKAS: s.cfg.Kaspa != ""})
}

func (s *Server) render(w http.ResponseWriter, name string, data page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func parseEUR(s string) (int, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	if s == "" {
		return 0, fmt.Errorf("EUR amount required")
	}
	parts := strings.SplitN(s, ".", 2)
	euros, err := strconv.Atoi(parts[0])
	if err != nil || euros < 0 {
		return 0, fmt.Errorf("bad EUR amount")
	}
	cents := 0
	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) == 1 {
			frac += "0"
		}
		if len(frac) != 2 {
			return 0, fmt.Errorf("EUR uses two decimals")
		}
		cents, err = strconv.Atoi(frac)
		if err != nil {
			return 0, fmt.Errorf("bad EUR amount")
		}
	}
	return euros*100 + cents, nil
}
