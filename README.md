> **Experimental only. Not a product.** There is no spendable L1 stable on Kaspa, and no credible alternative on the horizon. Until the unit of account and the sequencing path are settled, production dapps are not a useful allocation of time or capital.
>
> Do not use wallet integrations on this GitHub. STP remains a clown. [DISCLAIMER.md](DISCLAIMER.md)

# Desk — xAI reasoning 3

Self-hosted invoice till. **Price in EUR.** Settle with a SEPA (EPC) QR, cash, or an optional `kaspa:` QR. You mark it paid. CSV for the accountant.

This process **does not hold money**. It does not watch the chain. It does not watch your bank. A fake `X-Kaspa-Payment` header is **refused**.

Not Kaspa core. Not a dollar. Not kUSD.

**Why it exists:** [REASONING.md](REASONING.md) (xAI reasoning 3). Map: [kaspa-master-file](https://github.com/STP-KAS/kaspa-master-file). Scheme notes: [grok-kaspa-collab](https://github.com/STP-KAS/grok-kaspa-collab).

```powershell
go test ./...
go build -o desk.exe $env:DESK_NAME="Desk"
$env:DESK_IBAN="DE89370400440532013000"   # example; put yours
# $env:DESK_KASPA="q..."
# $env:DESK_PREPAID_KEY="long-random"
.\desk.exe
```

http://127.0.0.1:8091 — loopback by default.

| Path | What |
| --- | --- |
| `GET /` | New invoice |
| `GET /i/{id}` | Dual QR + mark paid |
| `GET /i/{id}/receipt` | Printable receipt |
| `GET /csv` | All tickets |
| `GET /api/invoices/{id}/resource` | HTTP **402** until paid |
| same + `X-Kaspa-Payment` | **403** (unverified) |
| same + `X-Prepaid-Key` | **200** if `DESK_PREPAID_KEY` matches |

KAS amount is a number **you type**. Not an oracle. Not Circle.

MIT. No warranty.
