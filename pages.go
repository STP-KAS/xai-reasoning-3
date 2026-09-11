package main

const pages = `
{{define "css"}}
:root { font-family: Georgia, serif; color: #111; background: #f6f4ef; }
body { max-width: 42rem; margin: 2rem auto; padding: 0 1rem; }
h1 { font-size: 1.4rem; }
.muted { color: #555; font-size: 0.95rem; }
.err { color: #8b1a1a; }
label, input, textarea, select, button { display: block; width: 100%; box-sizing: border-box; margin: 0.3rem 0 0.8rem; padding: 0.45rem; font: inherit; }
button { width: auto; cursor: pointer; }
.row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.qr { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.qr figure { margin: 0; background: #fff; padding: 0.6rem; border: 1px solid #ddd; text-align: center; }
.qr img { width: 100%; height: auto; image-rendering: pixelated; }
.mono { font-family: ui-monospace, Consolas, monospace; font-size: 0.85rem; word-break: break-all; }
a { color: #123; }
table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
td, th { text-align: left; padding: 0.25rem 0.4rem; border-bottom: 1px solid #ddd; }
.banner { background: #fff; border: 1px solid #ccc; padding: 0.7rem 1rem; margin-bottom: 1.2rem; }
@media (max-width: 640px) { .row, .qr { grid-template-columns: 1fr; } }
@media print { nav, form, .noprint { display: none; } body { background: #fff; } }
{{end}}

{{define "home"}}
<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Cfg.Name}} — desk</title><style>{{template "css"}}</style>
<body>
<div class="banner">Self-hosted invoice. Price in EUR. Settle SEPA, cash, or KAS. This process does not hold money. Not a dollar. Not Kaspa core.</div>
<h1>{{.Cfg.Name}}</h1>
<p class="muted">Create a ticket. Customer scans EPC (bank app) or kaspa: (wallet). You mark it paid. CSV is for the accountant.</p>
{{if .Err}}<p class="err">{{.Err}}</p>{{end}}
<form method="post" action="/invoices">
  <label>What for <input name="description" required placeholder="Website work, September"></label>
  <div class="row">
    <label>EUR <input name="eur" required inputmode="decimal" placeholder="120.00"></label>
    <label>VAT % (optional) <input name="vat" inputmode="numeric" placeholder="0"></label>
  </div>
  <label>KAS amount if they pay on-chain (typed, not an oracle) <input name="kas" placeholder="leave empty if unused"></label>
  <button type="submit">Make invoice</button>
</form>
<p class="muted noprint">IBAN set: {{if .HasIBAN}}yes{{else}}no — set DESK_IBAN{{end}}. Kaspa address set: {{if .HasKAS}}yes{{else}}no — set DESK_KASPA{{end}}. <a href="/csv">Download CSV</a></p>
{{if .List}}
<h2>Tickets</h2>
<table>
<tr><th>id</th><th>EUR</th><th>paid</th></tr>
{{range .List}}
<tr><td class="mono"><a href="/i/{{.ID}}">{{.ID}}</a></td><td>{{eur .EURCents}}</td><td>{{if .Paid}}{{.PaidHow}}{{else}}open{{end}}</td></tr>
{{end}}
</table>
{{end}}
</body></html>
{{end}}

{{define "invoice"}}
<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Inv.ID}}</title><style>{{template "css"}}</style>
<body>
<nav class="noprint"><a href="/">← desk</a></nav>
<h1>Invoice {{.Inv.ID}}</h1>
<p><strong>{{.Inv.Description}}</strong></p>
<p>EUR {{eur .Inv.EURCents}}{{if .Inv.VATPercent}} (VAT {{.Inv.VATPercent}}% included in this number as you typed it){{end}}</p>
{{if .Inv.KASAmount}}<p>KAS {{.Inv.KASAmount}} <span class="muted">merchant-typed</span></p>{{end}}
<p class="mono">Remittance / payment ref: {{.Inv.Remittance}}</p>
{{if .Inv.Paid}}
<p>Paid via {{.Inv.PaidHow}} {{.Inv.PaidRef}} at {{.Inv.PaidAt}}. <a href="/i/{{.Inv.ID}}/receipt">Receipt</a></p>
{{else}}
<div class="qr">
  {{if .HasIBAN}}
  <figure>
    <img src="/i/{{.Inv.ID}}/qr/epc.png" alt="EPC QR">
    <figcaption>SEPA / EPC QR<br><span class="muted">Scan with a European banking app</span></figcaption>
  </figure>
  {{end}}
  {{if .HasKAS}}
  <figure>
    <img src="/i/{{.Inv.ID}}/qr/kaspa.png" alt="Kaspa QR">
    <figcaption>kaspa: URI<br><span class="mono">{{.KasURI}}</span></figcaption>
  </figure>
  {{end}}
</div>
{{if not .HasIBAN}}{{if not .HasKAS}}<p class="err">No IBAN and no Kaspa address. Cash / mark-paid still works. Set DESK_IBAN and/or DESK_KASPA.</p>{{end}}{{end}}
<form class="noprint" method="post" action="/i/{{.Inv.ID}}/paid">
  <p>You confirm you got the money. The desk does not watch the chain or the bank.</p>
  <label>How
    <select name="how" required>
      <option value="sepa">SEPA / bank</option>
      <option value="cash">Cash</option>
      <option value="kas">KAS (I checked the tx)</option>
      <option value="other">Other</option>
    </select>
  </label>
  <label>Ref (txid, end-to-end id, or note) <input name="ref" placeholder="optional"></label>
  <button type="submit">Mark paid</button>
</form>
<p class="muted noprint">Agents: GET /api/invoices/{{.Inv.ID}}/resource → 402. Fake <span class="mono">X-Kaspa-Payment</span> is refused.</p>
{{end}}
</body></html>
{{end}}

{{define "receipt"}}
<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Receipt {{.Inv.ID}}</title><style>{{template "css"}}</style>
<body>
<nav class="noprint"><a href="/">← desk</a> · <a href="/i/{{.Inv.ID}}">invoice</a> · <button onclick="window.print()">Print / PDF</button></nav>
<h1>Receipt</h1>
<p>{{.Cfg.Name}}</p>
<table>
<tr><th>id</th><td class="mono">{{.Inv.ID}}</td></tr>
<tr><th>what</th><td>{{.Inv.Description}}</td></tr>
<tr><th>amount</th><td>EUR {{eur .Inv.EURCents}}</td></tr>
{{if .Inv.KASAmount}}<tr><th>KAS noted</th><td>{{.Inv.KASAmount}}</td></tr>{{end}}
<tr><th>paid how</th><td>{{.Inv.PaidHow}}</td></tr>
<tr><th>ref</th><td class="mono">{{.Inv.PaidRef}}</td></tr>
<tr><th>when</th><td>{{.Inv.PaidAt}}</td></tr>
</table>
<p class="muted">This is a bookkeeping receipt from software you run. It is not a bank statement and not a Kaspa confirmation. Check the ref on your bank or explorer if you need proof of settlement.</p>
</body></html>
{{end}}
`
