# OTScan

OTScan is a real-time blockchain explorer for the **RMC OTS (OpenTimestamps)** system. It monitors OTS batch lifecycle, node health, copyright claim tracking, and BTC anchor verification across all validator nodes.

## Architecture

```
                    +-----------+
                    |  Frontend |  React + Ant Design
                    |  :3000    |  (SPA served by Go)
                    +-----+-----+
                          |
                    +-----v-----+
                    |  REST API  |  Gin framework
                    | WebSocket  |  Real-time updates
                    +-----+-----+
                          |
          +---------------+---------------+
          |               |               |
    +-----v-----+  +-----v-----+  +------v------+
    | PostgreSQL |  |   Redis   |  |   Indexer   |
    |  Storage   |  |   Cache   |  | (background)|
    +-----+------+  +-----------+  +------+------+
                                          |
                              +-----------+-----------+
                              |     |     |     |     |
                            node0 node1 node2 node3 node4
                              (geth RPC endpoints)
```

**Indexer** runs three background loops:
- **Node Poller** (10s) - polls health & block height from all nodes
- **Batch Syncer** (30s) - syncs OTS batches by on-chain ID from geth RPC
- **Claim Syncer** - syncs copyright Claimed/Published events

All status changes are broadcast via WebSocket for real-time frontend updates.

## Features

- **Dashboard** - Node status overview, batch statistics, block height trend chart
- **Batches** - Paginated list with status filter, anchor node display, block range, RUID count
- **Batch Detail** - Full batch info including BTC confirmation, OTS proof, anchor node, RUID list
- **Nodes** - Per-node health, Calendar server status, pending batch monitoring
- **Claims** - Search by RUID/PUID/AUID/Claimant, publish status tracking
- **Conflicts** - AUID conflict detection (multiple claims on same asset)
- **Verify** - RUID verification with Merkle proof + OTS proof chain
- **Real-time** - WebSocket-driven updates for batch status changes and node health

## Tech Stack

| Layer    | Technology                                    |
|----------|-----------------------------------------------|
| Frontend | React 19, TypeScript, Ant Design 6, Recharts  |
| Backend  | Go 1.24, Gin, pgx/v5, go-redis/v9            |
| Database | PostgreSQL 16                                  |
| Cache    | Redis 7                                        |
| Build    | Vite 7 (frontend), Docker multi-stage          |

## Quick Start

### Docker (recommended)

```bash
# Start all services (PostgreSQL, Redis, OTScan)
docker compose up -d

# OTScan will be available at http://localhost:3000
```

The Docker setup uses `config.docker.yaml` which connects to geth nodes via `host.docker.internal`.

### Local Development

**Prerequisites**: Go 1.24+, Node.js 20+, PostgreSQL 16+, Redis 7+

```bash
# 1. Create config
cp config.docker.yaml config.yaml
# Edit config.yaml with your local database/redis/node settings

# 2. Build and run
make build
./otscan --config config.yaml

# Or for development (with hot reload):
make dev
```

**Frontend development** (separate terminal):

```bash
cd web
npm install
npm run dev    # Vite dev server with HMR
```

### Makefile Targets

| Command          | Description                        |
|------------------|------------------------------------|
| `make build`     | Build frontend + Go binary         |
| `make run`       | Build and run                      |
| `make dev`       | Run directly (`go run`)            |
| `make frontend`  | Build frontend only                |
| `make clean`     | Remove build artifacts             |

## Configuration

```yaml
server:
  host: "0.0.0.0"
  port: 3000
  mode: "release"          # "debug" or "release"

chain:
  id: 192                  # RMC chain ID
  name: "RMC"
  breatheBlockInterval: 3600  # OTS trigger interval (seconds)

nodes:                     # Validator nodes to monitor
  - name: "node0"
    rpcUrl: "http://127.0.0.1:8545"
  - name: "node1"
    rpcUrl: "http://127.0.0.1:8547"
  # ...

database:
  host: "127.0.0.1"
  port: 5432
  user: "otscan"
  password: "otscan_pass"
  dbname: "otscan"

redis:
  addr: "127.0.0.1:6379"
  nodeStatusTTL: "30s"

indexer:
  nodePollingInterval: "10s"
  batchSyncInterval: "30s"
  claimSyncInterval: "30s"
```

## API Endpoints

All endpoints are under `/api/v1/`.

### Dashboard & Health

| Method | Path         | Description              |
|--------|-------------|--------------------------|
| GET    | `/health`    | Health check             |
| GET    | `/config`    | Chain & node config      |
| GET    | `/dashboard` | Overview statistics      |

### Batches

| Method | Path                 | Description                  |
|--------|---------------------|------------------------------|
| GET    | `/batches`           | List batches (paginated, filterable by status) |
| GET    | `/batches/:id`       | Get batch detail (by batchID or onChainID) |
| GET    | `/batches/:id/ruids` | Get RUIDs in a batch         |
| GET    | `/proof/:batchId`    | Get OTS proof for a batch    |

### Nodes

| Method | Path                              | Description                    |
|--------|----------------------------------|--------------------------------|
| GET    | `/nodes`                          | List all nodes                 |
| GET    | `/nodes/:name`                    | Get node status                |
| GET    | `/nodes/:name/history`            | Node status history (trend)    |
| GET    | `/nodes/:name/calendar`           | Pending batches on this node   |
| GET    | `/nodes/:name/calendar-url-status`| Calendar server statistics     |

### Claims & Verification

| Method | Path                      | Description                    |
|--------|--------------------------|--------------------------------|
| GET    | `/claims`                 | Search claims (by ruid/puid/auid/claimant) |
| GET    | `/claims/:ruid`           | Get claim detail               |
| GET    | `/claims/conflicts/:auid` | Check AUID conflicts           |
| GET    | `/stats/claims`           | Claim statistics               |
| GET    | `/conflicts`              | List all conflicts             |
| POST   | `/verify`                 | Verify RUID (Merkle + OTS)     |

### WebSocket

| Path   | Description                                |
|--------|--------------------------------------------|
| `/ws`  | Real-time events (batch_update, node_status) |

## Database Schema

| Table                  | Description                              |
|------------------------|------------------------------------------|
| `nodes`                | Node registry (name, rpc_url)            |
| `node_status`          | Current node state (block, health, components) |
| `node_status_history`  | Time-series node data for trend charts   |
| `batches`              | OTS batches (status, BTC info, anchor node) |
| `batch_ruids`          | RUID-to-batch mapping                    |
| `claims`               | Copyright claims (ruid, claimant, auid, puid) |

Migrations are in `migrations/` and are auto-applied on startup.

## Project Structure

```
otscan/
  cmd/otscan/          # Entry point (main.go)
  internal/
    api/               # Gin handlers and router
    cache/             # Redis cache layer
    config/            # YAML config parser
    indexer/           # Background sync (nodes, batches, claims)
    rpc/               # JSON-RPC client and types
    service/           # Business logic layer
    store/             # PostgreSQL repositories
  migrations/          # SQL schema migrations
  web/                 # React frontend (Vite + TypeScript)
    src/
      api/             # API client and WebSocket
      pages/           # Page components
  Dockerfile           # Multi-stage build
  docker-compose.yml   # Full stack (PostgreSQL + Redis + OTScan)
  config.docker.yaml   # Docker environment config
```

## License

Copyright 2024 The RMC Authors. All rights reserved.
