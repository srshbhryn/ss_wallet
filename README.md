# SS Wallet

A modular Go (Golang) wallet system for marketplaces and apps where **available** funds can differ from **total** funds due to **block/unblock** and **delayed availability** (e.g., deposit release dates). The project ships with:

- **REST API** for balances, transactions, deposits, and withdrawals
- **Workers**:
  - `deposit_applier`: periodically applies eligible deposits (moves blocked → available)
  - `banker`: polls and processes withdrawals via pluggable bank/PSP integrations
- **PostgreSQL** persistence via **GORM**
- **Database migrations** via **Goose**
- **Structured logging** with `slog` and embedded Git commit hash
- **Docker-ready** with a single image carrying all binaries (run specific binary via `command:`)

---

## Table of Contents

- [SS Wallet](#ss-wallet)
  - [Table of Contents](#table-of-contents)
  - [Architecture](#architecture)
  - [Domain Concepts](#domain-concepts)
  - [Project Layout](#project-layout)
  - [Configuration](#configuration)
    - [Separate config files](#separate-config-files)
      - [Example combined config](#example-combined-config)
    - [Environment variables](#environment-variables)
  - [Database \& Migrations](#database--migrations)
  - [Running Locally (no Docker)](#running-locally-no-docker)
  - [Docker \& Compose](#docker--compose)
    - [Build the image named `ss_wallet`](#build-the-image-named-ss_wallet)
    - [Run with docker compose](#run-with-docker-compose)
  - [API Endpoints](#api-endpoints)
    - [Wallet](#wallet)
    - [Deposits](#deposits)
    - [Withdrawals](#withdrawals)
    - [Transactions (user history)](#transactions-user-history)
  - [Workers](#workers)
    - [deposit\_applier](#deposit_applier)
    - [banker](#banker)
  - [Logging](#logging)
  - [Testing](#testing)
    - [Unit tests](#unit-tests)
    - [Integration tests (Postgres)](#integration-tests-postgres)
  - [Operational Notes](#operational-notes)
  - [Troubleshooting](#troubleshooting)
  - [FAQ](#faq)
  - [Design \& Module Schematic](#design--module-schematic)
    - [High‑Level Component Diagram](#highlevel-component-diagram)
    - [Sequence: **Deposit Apply** (blocked → available)](#sequence-deposit-apply-blocked--available)
    - [Sequence: **Withdrawal Processing**](#sequence-withdrawal-processing)
    - [Module ↔ Package Correspondence](#module--package-correspondence)
    - [Transaction Boundaries \& Idempotency](#transaction-boundaries--idempotency)
    - [Error Handling \& Retries](#error-handling--retries)
    - [Configuration Boundaries](#configuration-boundaries)

---

## Architecture

**High level**
- **REST** service exposes application endpoints.
- **Core** domain implements wallet, transactions, deposits, withdrawals, and repository abstractions.
- **Workers**:
  - `deposit_applier` scans pending deposits whose `ApplyAt < now` and are not yet applied.
  - `banker` fetches withdrawals to send, calls a **BankClient** (PSP integration), updates status.

**Transactionality**
- Repositories use a factory that creates a DB transaction per unit of work.
- Critical updates occur under `SELECT ... FOR UPDATE` to ensure consistency (e.g., balance updates).

**Integrations**
- Pluggable `BankClient` (e.g., a Dummy client for development). Real PSPs can be added with their own configs.

---

## Domain Concepts

- **Wallet**  
  Holds:
  - `AvailableBalance`: funds the user can withdraw now.
  - `BlockedBalance`: funds reserved or pending release.
  - `TotalBalance`: conceptual sum (available + blocked).

- **Deposit**  
  Adds funds to a wallet; funds may be **blocked** until `ApplyAt`. When eligible, `deposit_applier` **applies** and moves amounts to `AvailableBalance`, also writing a ledger **Transaction**.

- **Withdrawal**  
  Requests to send funds via the bank/PSP. Enters a lifecycle: `created → sent → completed` or `reversed/failed`.

- **Transaction**  
  Immutable audit log entry capturing net effects on balances.

---

## Project Layout

```
.
├── bin/
│   ├── rest/               # REST API entrypoint (main package)
│   ├── deposit_applier/    # deposit applier worker (main package)
│   └── banker/             # withdrawal worker (main package)
├── lib/
│   ├── core/               # entities, repositories, service wiring
│   ├── deposits/           # deposit services and repository
│   ├── withdraws/          # withdrawal services, worker, integrations
│   ├── rest/               # HTTP handlers, middleware (Gin)
│   └── utils/
│       ├── db/             # GORM init, DB utilities
│       └── logger/         # slog setup, GitCommit var
├── migrations/             # Goose SQL migrations
├── configs/                # (optional) separate YAML config files
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

## Configuration

The application reads a single YAML file path from `CONFIG_PATH`. You can either:
- keep **one combined YAML**, or
- keep **separate YAMLs** and merge them during your build/deployment process.

### Separate config files

Example minimal files (you can place them under `configs/`):

**`configs/rest.yaml`**
```yaml
rest:
  bind_at: 8080
```

**`configs/db.yaml`**
```yaml
db:
  dsn: "postgres://postgres:postgres@localhost:5432/wallet?sslmode=disable"
```

**`configs/withdraws.dummy.yaml`**
```yaml
withdraws:
  bank_type: "DUMMY"
  dummy:
    api_key: "test"
```

**`configs/deposit_applier.yaml`**
```yaml
deposit_applier:
  id_prefix: ""   # optional; if set, only deposits with ID starting with this prefix are applied
  sleep: "5s"
```

**`configs/banker.yaml`**
```yaml
banker:
  poll_interval: "5s"
  concurrency: 4
  backoff: "2s"
  retries: 3
```

**`configs/logging.yaml`**
```yaml
logging:
  level: "info"
  dir: "./logs"
```

#### Example combined config
If you prefer a single file, merge the sections:

**`config.docker.yaml` (example)**
```yaml
rest:
  bind_at: 8080

db:
  dsn: "postgres://postgres:postgres@db:5432/wallet?sslmode=disable"

withdraws:
  bank_type: "DUMMY"
  dummy:
    api_key: "test"

deposit_applier:
  id_prefix: ""
  sleep: "5s"

banker:
  poll_interval: "5s"
  concurrency: 4
  backoff: "2s"
  retries: 3

logging:
  level: "info"
  dir: "/var/log/ss-wallet"
```

### Environment variables

If your loader maps dots to underscores, you can override any config via env:

```
REST_BIND_AT=8080
DB_DSN=postgres://postgres:postgres@localhost:5432/wallet?sslmode=disable
WITHDRAWS_BANK_TYPE=DUMMY
WITHDRAWS_DUMMY_API_KEY=test
DEPOSIT_APPLIER_ID_PREFIX=
DEPOSIT_APPLIER_SLEEP=5s
BANKER_POLL_INTERVAL=5s
BANKER_CONCURRENCY=4
BANKER_BACKOFF=2s
BANKER_RETRIES=3
LOGGING_LEVEL=info
LOGGING_DIR=./logs
```

---

## Database & Migrations

We use **PostgreSQL** and **Goose** migrations.

- Install goose locally:
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```
- Apply migrations:
  ```bash
  export DB_DSN='postgres://postgres:postgres@localhost:5432/wallet?sslmode=disable'
  goose -dir ./migrations postgres "$DB_DSN" up
  ```
- Status:
  ```bash
  goose -dir ./migrations postgres "$DB_DSN" status
  ```

> In Docker/Compose, migrations are run by a `migrator` service (see below).

---

## Running Locally (no Docker)

1. **Start Postgres** (e.g., Docker standalone):
   ```bash
   docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=wallet postgres:16
   ```

2. **Run migrations**:
   ```bash
   export DB_DSN='postgres://postgres:postgres@localhost:5432/wallet?sslmode=disable'
   goose -dir ./migrations postgres "$DB_DSN" up
   ```

3. **Create a config** (e.g., `config.dev.yaml`) and export `CONFIG_PATH`:
   ```bash
   export CONFIG_PATH=$PWD/config.dev.yaml
   ```

4. **Run services**:
   ```bash
   # REST API
   go run ./bin/rest

   # Deposit applier worker
   go run ./bin/deposit_applier

   # Banker worker
   go run ./bin/banker
   ```

---

## Docker & Compose

### Build the image named `ss_wallet`

> **Why `ss_wallet`?**  
> The docker-compose expects an application image called **`ss_wallet`**. Build and tag your image with this name to let compose start the REST/worker processes without a local build section in compose.

**Basic build:**
```bash
docker build -t ss_wallet .
```

**Embed the Git commit hash into logs (recommended):**
```bash
GIT_COMMIT=$(git rev-parse --short HEAD)
docker build \
  --build-arg GIT_COMMIT="$GIT_COMMIT" \
  -t ss_wallet .
```

If your `Dockerfile` uses Go build flags directly, you can also pass:
```bash
GO_LDFLAGS="-X wallet/lib/utils/logger.GitCommit=$GIT_COMMIT"
docker build --build-arg GO_LDFLAGS="$GO_LDFLAGS" -t ss_wallet .
```

**Multi-arch (optional, with Buildx):**
```bash
docker buildx build --platform linux/amd64,linux/arm64 -t ss_wallet --load .
```

> After building, you’ll have a local image named `ss_wallet` that compose can run.

### Run with docker compose

A minimal compose (excerpt) assuming you already built `ss_wallet`:

```yaml
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: wallet
    ports: ["5432:5432"]
    volumes:
      - dbdata:/var/lib/postgresql/data

  migrator:
    image: golang:1.23
    depends_on: [db]
    working_dir: /app
    volumes:
      - ./:/app
    entrypoint: ["/bin/sh","-lc"]
    command: >
      '
      go install github.com/pressly/goose/v3/cmd/goose@latest &&
      goose -dir ./migrations postgres "$$DB_DSN" up
      '
    environment:
      DB_DSN: postgres://postgres:postgres@db:5432/wallet?sslmode=disable

  rest:
    image: ss_wallet
    depends_on: [db, migrator]
    environment:
      CONFIG_PATH: /app/config.docker.yaml
    volumes:
      - ./config.docker.yaml:/app/config.docker.yaml:ro
    ports: ["8080:8080"]
    command: ["rest"]  # run REST binary from the same image

  # Optional workers:
  banker:
    image: ss_wallet
    depends_on: [db]
    environment:
      CONFIG_PATH: /app/config.docker.yaml
    volumes:
      - ./config.docker.yaml:/app/config.docker.yaml:ro
    command: ["banker"]

  deposit_applier:
    image: ss_wallet
    depends_on: [db]
    environment:
      CONFIG_PATH: /app/config.docker.yaml
    volumes:
      - ./config.docker.yaml:/app/config.docker.yaml:ro
    command: ["deposit_applier"]

volumes:
  dbdata:
```

Bring it up:
```bash
docker compose up
```

---

## API Endpoints

> Exact paths may vary if you change route groups; here’s a representative set.

### Wallet
```
GET /v1/wallets/{user_id}/balance
→ { "user_id": "...", "available": 1200, "blocked": 300, "total": 1500 }
```

### Deposits
```
POST /v1/deposits
Body:
{
  "user_id": "uuid",
  "amount": 1000,
  "apply_at": "2025-09-15T10:00:00Z"   // optional; if omitted, apply immediately
}
→ 201 Created { "id": "...", "status": "created" }
```

### Withdrawals
```
POST /v1/withdrawals
Body:
{
  "user_id": "uuid",
  "amount": 500,
  "iban": "IR......"
}
→ 201 Created { "id": "...", "status": "created" }
```

### Transactions (user history)
```
GET /v1/wallets/{user_id}/transactions?limit=50&cursor=...
→ 200 OK [{ "id": 1, "delta_available": 500, "delta_blocked": -500, ... }]
```

> Auth: add your middleware of choice; headers can be forwarded via Gin middleware.

---

## Workers

### deposit_applier
- Periodically scans for deposits with:
  - `id` starting with `deposit_applier.id_prefix` (if non-empty),
  - `apply_at < now`,
  - `apply_transaction_id = 0` (or null/zero — not applied).
- For each eligible deposit:
  - Moves funds from **blocked** → **available** in a single transaction.
  - Writes a **Transaction** entry (immutable ledger).
- Config:
  ```yaml
  deposit_applier:
    id_prefix: ""
    sleep: "5s"
  ```

### banker
- Polls for withdrawals to send (created/pending state).
- Sends them via `BankClient` based on `withdraws.bank_type`.
- Updates status to `sent/complete/failed`; supports retries/backoff.
- Config:
  ```yaml
  banker:
    poll_interval: "5s"
    concurrency: 4
    backoff: "2s"
    retries: 3
  withdraws:
    bank_type: "DUMMY"
    dummy:
      api_key: "test"
  ```

---

## Logging

- Uses `slog` with leveled logging.
- Git commit hash can be embedded into logs via a variable:
  - `wallet/lib/utils/logger.GitCommit` (set by `-ldflags` during build).
- Example build flags:
  ```bash
  GIT_COMMIT=$(git rev-parse --short HEAD)
  go build -ldflags "-X wallet/lib/utils/logger.GitCommit=$GIT_COMMIT" ./bin/rest
  ```

Log directory configured by:
```yaml
logging:
  level: "info"
  dir: "/var/log/ss-wallet"
```

---

## Testing

### Unit tests
```bash
go test ./...
```

- Use mocks for integrations (e.g., BankClient dummy).
- For services that require DB, prefer in-memory transaction per test (`BEGIN; ROLLBACK`) or a temporary DB schema.

### Integration tests (Postgres)
1. Start a test Postgres (e.g., with Docker).
2. Apply migrations.
3. Run tests with `DB_DSN` pointing to the test instance:
   ```bash
   DB_DSN='postgres://postgres:postgres@localhost:5432/wallet?sslmode=disable' go test ./...
   ```

> Tip: For deterministic tests, seed with fixed timestamps and UUIDs.

---

## Operational Notes

- **Atomicity**: Balance updates and ledger writes should occur within a single DB transaction and use `FOR UPDATE` where relevant.
- **Idempotency**: Workers should tolerate restarts; `deposit_applier` should skip already-applied deposits.
- **Observability**: log request IDs and correlation IDs where possible; expose a `/metrics` endpoint if you add Prometheus later.
- **Graceful shutdown**: all binaries listen to SIGINT/SIGTERM; workers stop taking new work and finish in-flight tasks before exit.

---

## Troubleshooting

- **`exec /app/entrypoint.sh: no such file or directory`**  
  The final image likely doesn’t include `entrypoint.sh`. This project expects running binaries directly:
  - In compose, use `command: ["rest"]` (or `["banker"]`, `["deposit_applier"]`).

- **`OCI runtime create failed: executable file not found`**  
  Ensure the binary exists in the image and is in `PATH`. In the provided Dockerfile, copy built binaries to a known dir (e.g., `/app`) and set `WORKDIR /app`. Then run with `command: ["rest"]`.

- **Goose cannot connect**  
  Check `DB_DSN` (host/port). In compose, the hostname is the **service name**: `db`.

- **GORM locking type undefined**  
  Ensure you’re on **GORM v2** (`gorm.io/gorm`) and use `Clauses(clause.Locking{Strength: "UPDATE"})` from `gorm.io/gorm/clause`.

---

## FAQ

**Q: How do I add a new PSP/bank integration?**  
A: Implement the `BankClient` interface (`Send`, `GetStatus`), add a new enum constant, wire it in the `NewBankClient` factory, and define its config under `withdraws.<your_psp>`.

**Q: Should I use `uuid` or `bigint` IDs?**  
A: Wallet user IDs are often `uuid` (external identity). Transaction IDs can be `bigserial` for ordering. Pick what fits your consistency and sharding strategy.

**Q: How is delayed availability modeled?**  
A: Deposits may be created with `ApplyAt` in the future; until then, amounts are **blocked**. The `deposit_applier` moves them to **available** when due.

---


## Design & Module Schematic

This section explains the **overall design**, how modules interact, and the exact **source package correspondence**.

### High‑Level Component Diagram

```mermaid
flowchart LR
  subgraph Clients
    A[REST Clients\nPostman/Browsers]
  end

  subgraph Service["REST Service (bin/rest)"]
    H[Handlers\nlib/rest/internal]
    MW[Middleware/Auth]
  end

  subgraph Core["Domain Core (lib/core)"]
    Svc[Service Wiring\nTx Factory]
    Rep[Repositories\nWallet/Txn/Deposit/Withdraw]
    Ent[Entities/Models]
  end

  subgraph Deposits["Deposits (lib/deposits)"]
    DSvc[Deposit Service]
    DRepo[Deposit Repo]
  end

  subgraph Withdraws["Withdraws (lib/withdraws)"]
    WSvc[Withdraw Service]
    WRepo[Withdraw Repo]
    subgraph Integrations["Integrations\nlib/withdraws/integrations"]
      BankClient[BankClient Interface]
      Dummy[Dummy PSP Client]
    end
  end

  subgraph Workers
    DA[deposit_applier (bin/deposit_applier)]
    BK[banker (bin/banker)]
  end

  subgraph Infra["Infrastructure"]
    DB[(PostgreSQL)]
    G["GORM\nlib/utils/db"]
    LG["slog logger\nlib/utils/logger"]
    MG["Goose Migrations\nmigrations/*.sql"]
  end

  A -->|HTTP JSON| H
  H -->|calls| Svc
  DA -->|apply cycle| DSvc
  BK -->|poll/send| WSvc
  Svc --> Rep
  Rep -->|gorm| G --> DB
  DSvc --> DRepo --> G
  WSvc --> WRepo --> G
  WSvc --> BankClient
  BankClient --> Dummy
  H --> LG
  DA --> LG
  BK --> LG
```

### Sequence: **Deposit Apply** (blocked → available)

```mermaid
sequenceDiagram
  autonumber
  participant DA as Worker: deposit_applier
  participant D as Deposits Service
  participant R as Repos (Deposit/Wallet/Txn)
  participant G as GORM/DB

  DA->>D: Tick (sleep interval)
  D->>R: GetApplicableDeposits(for update)
  R->>G: SELECT ... FOR UPDATE
  R-->>D: List<Deposit>
  loop For each eligible deposit
    D->>R: Begin Tx
    D->>R: Move blocked → available (update wallet)
    D->>R: Insert immutable transaction
    D->>R: Mark deposit as applied (apply_transaction_id set)
    R->>G: COMMIT
  end
```

### Sequence: **Withdrawal Processing**

```mermaid
sequenceDiagram
  autonumber
  participant BK as Worker: banker
  participant W as Withdraw Service
  participant B as BankClient (PSP)
  participant R as Repos (Withdraw/Wallet/Txn)
  participant G as GORM/DB

  BK->>W: Poll pending withdrawals
  W->>R: Fetch next N to send
  W->>B: Send(iban, amount, trackID)
  alt PSP accepted
    W->>R: Update status SENT
    BK->>W: Periodic GetStatus(trackID)
    W->>B: GetStatus(trackID)
    alt Completed
      W->>R: Update status COMPLETE
    else Failed
      W->>R: Reverse and refund if applicable
    end
  else Rejected/Failure
    W->>R: Reverse and refund if applicable
  end
```

### Module ↔ Package Correspondence

| Module / Concept            | Package & Path                                 | Notes |
|----------------------------|-------------------------------------------------|-------|
| REST Handlers              | `lib/rest/internal`                             | Gin handlers, routing, middleware |
| Domain Entities            | `lib/core`                                      | `Wallet`, `Transaction`, shared types |
| Repo Factory & Tx Scope    | `lib/core`                                      | Starts/commits/rollbacks DB tx per unit of work |
| Deposit Service            | `lib/deposits`                                  | `Create`, `Apply` logic |
| Deposit Repository         | `lib/deposits/repository` (or `.../internal`)   | `GetApplicableDeposits`, `FOR UPDATE` |
| Withdraw Service           | `lib/withdraws`                                 | Lifecycle: `Create/Reverse/MarkAsSent/Complete` |
| Withdraw Repository        | `lib/withdraws/repository`                      | Fetch and update withdrawal rows |
| PSP Integrations           | `lib/withdraws/integrations`                    | `BankClient` interface + concrete clients |
| Workers                    | `bin/deposit_applier`, `bin/banker`             | Daemons executing periodic tasks |
| DB Initialization          | `lib/utils/db`                                  | GORM setup, connection, tuning |
| Logging                    | `lib/utils/logger`                              | `slog`, `GitCommit` ldflags var |
| Migrations                 | `migrations/*.sql`                              | Goose-managed schema changes |
| Entrypoint REST            | `bin/rest`                                      | Main for HTTP server |

### Transaction Boundaries & Idempotency

- **Single DB transaction** per logical operation (service layer starts a tx via the factory).
- **`SELECT ... FOR UPDATE`** on wallet/deposit rows ensures no concurrent balance drift.
- **Idempotent workers**: applying a deposit twice is prevented by checking/setting `apply_transaction_id` (or a unique constraint); withdrawal retries are keyed by `trackID`.

### Error Handling & Retries

- Services return errors up the stack (avoid `return nil` on failure).
- Workers use **exponential backoff** with a max retry count (see `banker.backoff`, `banker.retries`).
- Integrations wrap PSP errors with context (PSP name, trackID, request ID).

### Configuration Boundaries

- `withdraws.bank_type` selects the integration at runtime.
- Each integration has a sub‑config block, e.g., `withdraws.dummy.api_key`.
- Workers have independent schedules and can be scaled horizontally (multiple replicas).

---
