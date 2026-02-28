## Bitespeed Identify Service

This project is a small Go backend that implements the Bitespeed "identity reconciliation" problem.
Given an email address and/or phone number, it looks up existing contact records, links them into a
single identity, and returns the **primary contact** together with all related emails,
phone numbers, and secondary contact IDs.

The service is built with:

- Go 
- `gin-gonic` for the HTTP API
- `sqlx` and PostgreSQL for persistence
- A simple `contacts` table that models primary/secondary links between records

---

## How it works

When a request hits `POST /identify` with an `email`, `phoneNumber`, or both:

1. The service looks up all existing `contacts` that share that email or phone number.
2. If nothing is found, it creates a **new primary** contact.
3. If matches exist:
   - It chooses the oldest contact (by `created_at`, then by `id`) as the **primary**.
   - Any other contact in the cluster is converted to a **secondary** that points to the primary via `linked_id`.
   - If the incoming email/phone is new information, it creates an additional **secondary** contact for it.
4. It returns a normalized view:
   - `primaryContactId`
   - unique list of `emails`
   - unique list of `phoneNumbers`
   - list of `secondaryContactIds`

The `db.EnsureSchema` helper is called at startup so the `contacts` table and indexes are created automatically if they do not already exist.

---

## Project layout

High‑level directory structure:

- `main.go`  
  Application entrypoint. Reads `DATABASE_URL`, initializes the DB store, ensures the schema exists, wires the Gin router, and starts the HTTP server.

- `handler/handler.go`  
  HTTP transport layer. Defines the JSON request/response types and registers the `POST /identify` route, delegating to the identify service.

- `contact/model.go`  
  Data model for the `contacts` table, including the `Contact` struct and `LinkPrecedence` enum (`primary` or `secondary`).

- `contact/repository.go`  
  Repository abstraction around `sqlx` for working with contacts:
  finding by email/phone, finding by IDs, creating and updating contacts within or outside transactions.

- `service/identify.go`  
  Core business logic for identity reconciliation. Implements the `IdentifyService` that encapsulates the algorithm for selecting a primary contact and merging clusters.

- `db/store.go`  
  Lightweight wrapper around `sqlx.DB` that owns connection lifecycle and pool configuration.

- `db/schema.go`  
  Schema bootstrap helper. Creates the `contacts` table and related indexes if they do not exist.

---

## Testing the service

The service is currently deployed at:

- **Base URL**: `https://bitspeed-production-ab44.up.railway.app/`
- **Endpoint**: `POST /identify`
- **Database**: PostgreSQL instance hosted on **Neon** (`neondb`)

To exercise the live deployment with all parameters set (both email and phone number), you can run:

```bash
curl -s -X POST https://bitspeed-production-ab44.up.railway.app/identify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "a@example.com",
    "phoneNumber": "1111111111"
  }'
```

You can send additional requests with overlapping and non‑overlapping emails and phone numbers to observe how the service merges contacts and updates the primary/secondary relationships in the Neon‑hosted PostgreSQL database.  
The same endpoint can also be tested using tools like Postman or Hoppscotch by configuring a `POST` request to `/identify` with a JSON body in the format shown above.

