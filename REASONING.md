# xAI reasoning 3

**Not Kaspa core. Not a KIP. Not a dollar.**  
Pass 3 of the 11 Sep 2026 desk work. Software in this repo is the till. This file is why it looks like this.

| Pass | Where | Job |
| --- | --- | --- |
| 1 | [kaspa-master-file / GROK-HEAVY-REVIEW](https://github.com/STP-KAS/kaspa-master-file/blob/main/GROK-HEAVY-REVIEW.md) | Pins. What is live vs draft vs wrong. Four compiler holes. |
| 2 | [kaspa-master-file / THINK-BIG](https://github.com/STP-KAS/kaspa-master-file/blob/main/THINK-BIG.md) + [grok-kaspa-collab](https://github.com/STP-KAS/grok-kaspa-collab) | Scheme. Track 0 / Track 1. Beyond crypto: unit of account ≠ settlement ≠ receipt. |
| **3** | **this repo** | Build the smallest thing that is still honest. Dual-rail invoice. 402 that refuses fake txids. |

---

## 1. What had to be true

Kaspa after Toccata can lock coins with a script. That is real (KIPs 16/17/20/21 Active, rusty-kaspa v2.0.1, silverc v1.0.0).

Kaspa cannot mint a dollar. KCC-20 is Draft. Argent has no tag. vProgs are open PRs. DAGKnight is Proposed.

This desk’s own stack, checked against its own docs:

- Gramlane 402 accepts `X-Kaspa-Payment` **at HTTP layer only**.
- stillpay-tn10 is ENGINE_SPEC: broadcast off, **our** lock/claim/reclaim journal empty. Parker’s 1-sompi receipts are not this desk’s.
- Kaspa Till `kUSD` is **reserved**. Merchant rate is a sign, not an oracle.
- `ledger.json` is one process’s book, not Nakamoto.

If pass 3 shipped another README about grams, it would be pass 2 again.

---

## 2. The split (this is the whole scheme)

Three jobs got glued together in crypto marketing:

| Job | Human need | Honest tool |
| --- | --- | --- |
| Unit of account | Tax office and customer share a number | **EUR** (or local fiat) |
| Settlement | Money moves | SEPA Instant / cash / optional KAS |
| Receipt | Dispute, accountant, “I paid” | Ticket id + ref you typed after you checked |

PegLab tried to merge all three into a toy token. Dollars **0–0**. Parker won receipts because 1 sompi is only a **ref**.

Pass 3 builds that split in software: the ticket is always EUR. SEPA QR is the default settle rail in Europe (EPC069-12). `kaspa:` is an extra box. Mark-paid is the confirmation because this process does **not** watch the bank or the chain.

---

## 3. What we refused to build (on purpose)

| Temptation | Why not |
| --- | --- |
| Another gram ledger | Still a JSON file pretending to be L1 |
| kUSD till as if live | Reserved asset. Would be a lie on the first screen |
| 402 that “accepts a txid” | Gramlane already does that. It is theatre without a node check |
| BTCPay clone with plugins, payroll, Shopify | Too big. We do not have that product |
| stillpay broadcast + wasm submitter in this pass | Different repo, still gated. Do not fake txids here |
| Fourth 402 envelope | Bind elldeeone later. Steal k402 lock later. Not today |
| Oracle KAS/EUR | The till’s own vision already calls the rate a sign on the counter |

The 402 *industry* (x402 Foundation) settles mostly in stables. “Skip Circle for dapps” is a path, not a forecast. This till skips Circle by **not being a coin**.

---

## 4. What we did build

A loopback Go binary:

1. Create invoice: description + EUR (+ optional typed KAS amount).
2. Show **EPC QR** if `DESK_IBAN` is a real IBAN (mod-97). Show **kaspa: QR** if `DESK_KASPA` is set.
3. Human marks paid: sepa / cash / kas / other + optional ref.
4. Printable receipt. CSV for the accountant.
5. `GET /api/invoices/{id}/resource` → **402** while open.
6. Same URL with `X-Kaspa-Payment` → **403**. Unverified txids are refused.
7. `X-Prepaid-Key` matching `DESK_PREPAID_KEY` → **200** and the ticket is marked prepaid. That is a tab, labelled as one.

Default bind `127.0.0.1:8091`. Desk keeps 0. No seed. No wallet inject.

Verified: `go test ./...` and a live process (health, create, both QRs as PNG, fake header 403, unpaid 402, prepaid 200, cash receipt). No browser click-test in that pass.

---

## 5. Who it is for (narrow)

- A person in the euro area who already invoices in EUR and wants a scan-to-pay QR.
- The same person if they also hold KAS and want a second QR, not a new unit of account.
- This desk, as Track 1 evidence: software you run, we do not hold the float.

Not for: licensed PSP replacement, card refunds, coffee priced in grams, “agent GDP,” feeding miners.

---

## 6. What is still missing (do not flatten)

- No chain watch. If they pay KAS, **you** check the explorer, then mark paid.
- No bank watch. If they pay SEPA, **you** check the IBAN credit, then mark paid.
- stillpay TN10 journal still empty. Escrow is not this binary.
- WorkCredit `consume()` is not this binary.
- A Kaspa plugin for BTCPay might still be the grown-up Track 1. This is the smallest till that is not a costume.

Kill-if: speaking kUSD as live in this UI · accepting unverified `X-Kaspa-Payment` · `ledger.json` sold as L1 · seed paste · TN12.

---

*Pins from the 11 Sep 2026 freeze. Recheck tags and PRs before quoting protocol state.*
