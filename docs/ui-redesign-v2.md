# OTScan UI Redesign v2

## Overview

Redesign OTScan from a backend admin panel style to a public blockchain explorer style, inspired by Etherscan.

**Core change**: Left dark sidebar → Top horizontal navigation bar (white background, dropdown menus).

## Navigation Menu

```
[Home]  [Batches ▾]       [Claims ▾]          [Publish ▾]       [Verify]  [More ▾]
         ├ All Batches     ├ All Claims         ├ Asset (AUID)              ├ Nodes
         ├ Non-Empty       ├ Anchored           ├ Person (PUID)             ├ API Docs
         ├ Empty           ├ Non-Anchored        └ Conflicts                └ About
         └ Pending         ├ Published
                           └ Claimant
```

### Menu Descriptions

- **Batches**: OTS batch lifecycle
  - All Batches: All batches ordered by onChainID desc
  - Non-Empty: Batches with ruidCount > 0 (contain at least one RUID)
  - Empty: Batches with ruidCount == 0 (only advance lastAnchoredBlock)
  - Pending: Batches not yet anchored (status != "anchored")

- **Claims**: Copyright claim records (each RUID = one claim)
  - All Claims: All RUIDs ordered by submit time desc (newest first)
  - Anchored: RUIDs whose batch is already anchored on-chain
  - Non-Anchored: RUIDs still in pending/submitted/confirmed batches
  - Published: RUIDs that have published identity + content hash
  - Claimant: List of unique claimant addresses with claim counts

- **Publish**: Published identity and content hash features
  - Asset (AUID): All published asset unique IDs with claim counts
  - Person (PUID): All person unique IDs with associated asset counts
  - Conflicts: AUIDs with multiple PUID claimants (requires publish to detect)

- **Verify**: RUID verification with Merkle + OTS proof chain (standalone page)

- **More**: Secondary pages
  - Nodes: Validator node monitoring (operational view)
  - API Docs: API documentation
  - About: Project information

## Homepage Layout

```
┌────────────────────────────────────────────────────────────────┐
│  OTScan              [Home] [Batches▾] [Claims▾] [Publish▾]   │
│                                         [Verify] [More▾]      │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│   OTScan - RMC Copyright Timestamp Explorer                    │
│   ┌──────────────────────────────────────────────────┐         │
│   │ 🔍 Search by RUID / Batch ID / AUID / Address   │         │
│   └──────────────────────────────────────────────────┘         │
│   Block: 3,501 | Validators: 5/5 | BTC Anchored: 12           │
│                                                                │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌────────────────┐ │
│  │   TOTAL   │ │   TOTAL   │ │ PUBLISHED │ │ RUID Trends    │ │
│  │  BATCHES  │ │   CLAIMS  │ │  / TOTAL  │ │ In 16 Batches  │ │
│  │    12     │ │  22,118   │ │ 150/22118 │ │  ╭─╮           │ │
│  │ 10⬆ 2📭  │ │ anchored  │ │           │ │ ╭╯ ╰─╮ ╭╮     │ │
│  └───────────┘ └───────────┘ └───────────┘ │╯     ╰╯ ╰─    │ │
│                                            └────────────────┘ │
│                                                                │
│  ┌─── Latest Batches ──────────┐ ┌── Latest Claims ─────────┐ │
│  │ #1  batch-1-3201            │ │ 0xab12..  Claimant 0xfe..│ │
│  │     Blk 1-3201 | 22118 RUID│ │ Block 3100 | Published ✓ │ │
│  │     submitted    3m ago     │ │                          │ │
│  │     ⚓ 0xbcdd..0f (node0)   │ │ 0xcd34..  Claimant 0x5e..│ │
│  │                             │ │ Block 3098 | Pending     │ │
│  │ #2  batch-3202-6402         │ │                          │ │
│  │     Blk 3202-6402 | 0 RUID │ │ 0xef56..  Claimant 0x3a..│ │
│  │     pending       just now  │ │ Block 3095 | Published ✓ │ │
│  │                             │ │                          │ │
│  │      [View All Batches →]   │ │   [View All Claims →]   │ │
│  └─────────────────────────────┘ └──────────────────────────┘ │
│                                                                │
│  ┌─── Node Status ───────────────────────────────────────────┐ │
│  │ ● node0 ✓ #3501 | ● node1 ✓ #3501 | ● node2 ✓ #3501    │ │
│  │ ● node3 ✓ #3501 | ● node4 ✓ #3501                       │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                │
├────────────────────────────────────────────────────────────────┤
│  OTScan © 2024 RMC  |  About  |  API Docs                     │
└────────────────────────────────────────────────────────────────┘
```

### Homepage Sections

1. **Hero Section** - Blue gradient background, global search bar, chain stats summary
2. **Statistics Row** - 3 stat cards + 1 mini chart (responsive: 4-col on large, 2x2 on small)
3. **Dual Column** - Latest Batches (left) + Latest Claims (right)
4. **Node Status Bar** - Compact horizontal node health indicators
5. **Footer** - Copyright, links

### Global Search Bar

Smart input recognition:
- **RUID** (0x + 64 hex chars) → Verify page
- **Batch ID** ("batch-..." or pure number) → Batch detail page
- **AUID** (0x + 64 hex, with type selector) → Claims search
- **Address** (0x + 40 hex chars) → Claims search by claimant

### RUID Trends Mini Chart

- Position: Right of stat cards (responsive: below on small screens)
- Size: ~280x120px
- Data: Last 16 batches, each showing ruidCount
- Component: Recharts AreaChart or LineChart
- Source: Existing `GET /batches` response ruidCount field

## Page Routing

### Batches Pages

| Route | Page | Data Source | Description |
|-------|------|-------------|-------------|
| `/batches` | All Batches | Existing API | All batches, ordered by onChainID desc |
| `/batches?filter=non-empty` | Non-Empty | Filter `ruidCount > 0` | Batches containing at least 1 RUID |
| `/batches?filter=empty` | Empty | Filter `ruidCount == 0` | Empty batches |
| `/batches?filter=pending` | Pending | Existing API `status != anchored` | Unfinished batches |
| `/batches/:id` | Batch Detail | Existing API | Single batch detail (unchanged) |

### Claims Pages

| Route | Page | Data Source | Description |
|-------|------|-------------|-------------|
| `/claims` | All Claims | **New API** `GET /claims?sort=latest` | All RUIDs, newest first |
| `/claims?filter=anchored` | Anchored | **New API** `GET /claims?filter=anchored` | RUIDs in anchored batches |
| `/claims?filter=non-anchored` | Non-Anchored | **New API** `GET /claims?filter=non-anchored` | RUIDs in pending batches |
| `/claims?filter=published` | Published | Existing logic `published=true` | Published RUIDs |
| `/claimants` | Claimant List | **New API** `GET /claimants` | Unique claimant addresses with counts |

### Publish Pages

| Route | Page | Data Source | Description |
|-------|------|-------------|-------------|
| `/publish/assets` | Asset (AUID) | **New API** `GET /assets` | All published AUIDs with claim counts |
| `/publish/persons` | Person (PUID) | **New API** `GET /persons` | All PUIDs with asset counts |
| `/publish/conflicts` | Conflicts | Existing API | AUIDs with multiple PUID claimants |

### Other Pages

| Route | Page | Change |
|-------|------|--------|
| `/verify` | Verify | Unchanged, standalone page |
| `/nodes` | Nodes | Moved under More menu, content unchanged |

## New Backend APIs Required

| API | Method | Purpose | Implementation |
|-----|--------|---------|----------------|
| `GET /api/v1/claims?filter=anchored` | Query | Anchored RUIDs | SQL JOIN claims → batches WHERE batches.status = 'anchored' |
| `GET /api/v1/claims?filter=non-anchored` | Query | Non-anchored RUIDs | SQL JOIN claims → batches WHERE batches.status != 'anchored' |
| `GET /api/v1/claims?sort=latest` | Query | All RUIDs by time desc | ORDER BY submit_block DESC |
| `GET /api/v1/claimants` | New | Claimant address list | SELECT claimant, COUNT(*) FROM claims GROUP BY claimant |
| `GET /api/v1/assets` | New | AUID asset list | SELECT auid, COUNT(*) FROM claims WHERE published=true GROUP BY auid |
| `GET /api/v1/persons` | New | PUID person list | SELECT puid, COUNT(DISTINCT auid) FROM claims WHERE published=true GROUP BY puid |
| `GET /api/v1/batches?filter=non-empty` | Query | Non-empty batches | WHERE ruid_count > 0 |
| `GET /api/v1/batches?filter=empty` | Query | Empty batches | WHERE ruid_count = 0 |

All are simple SQL queries on existing tables. No schema changes needed.

## Color & Style

| Element | Value |
|---------|-------|
| Page background | `#f8f9fa` light gray |
| Navigation bar | White top bar with blue bottom border |
| Hero section | Blue gradient `#0784c3` → `#21325b` |
| Primary color | `#3498db` |
| Link color | `#1e88e5` |
| Tables | White card, border-radius, light shadow |
| Status tags | Compact badge style |
| Status colors | pending=blue, submitted=orange, confirmed=green, anchored=gold, failed=red |

## Implementation Phases

### Phase 1: Layout Restructure
- Remove left sidebar, add top navigation bar with dropdowns
- Homepage: Hero section + search bar + stat cards + RUID Trends chart
- Homepage: Dual-column Latest Batches / Latest Claims
- Homepage: Node status bar
- Footer
- Responsive design

### Phase 2: Batches & Claims Pages
- Batches page with filter tabs (All / Non-Empty / Empty / Pending)
- Claims page with filter tabs (All / Anchored / Non-Anchored / Published)
- Backend: Add filter params to existing batch/claim APIs
- Batch detail page style update (Etherscan block detail style)

### Phase 3: Publish Menu & New APIs
- Backend: New endpoints (claimants, assets, persons)
- Claimant list page
- Asset (AUID) list page
- Person (PUID) list page
- Move Conflicts page under Publish menu

### Phase 4: Polish
- Global search smart recognition
- Responsive fine-tuning
- Loading states and empty states
- WebSocket real-time updates integration
