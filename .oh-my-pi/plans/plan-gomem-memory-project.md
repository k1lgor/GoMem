# Plan: GoMem — Go Persistent Memory Server for AI Agents (Revised)

## Intent
A single Go binary that AI coding agents (specifically pi) use as persistent, searchable memory. The LLM generates search queries; the Go binary executes fast full-text lookups via Bleve.

## Success Criteria
- ≤50ms P50 query latency for 10k entries of ~200 bytes each, queried with a 3-word phrase, measured on a modern laptop SSD, after warm-up (10 queries before timing)
- All 6 acceptance criteria passing (spec: remember, search, delete, health, persistence, port config)

---

## 1. Project Structure

```
GoMem/
├── main.go                    # Entry point, CLI flag parsing, server start
├── go.mod
├── go.sum
├── handler.go                 # HTTP handlers (POST /remember, POST /search, etc.)
├── handler_test.go            # Integration tests with httptest
├── store.go                   # Bleve wrapper: interface + implementation
├── store_test.go              # Unit tests for store CRUD
├── bench_test.go              # Benchmark: 10k entries → query latency
├── types.go                   # Request/response structs (colocated)
└── README.md
```

**Architect response:** ✅ Removed `cmd/gomem/` nesting — single binary, flat root. ✅ Removed `internal/` packages — premature for V1. ✅ Removed `models/` — types colocated in `types.go`.

## 2. Technology Choices

| Choice | Option | Rationale |
|---|---|---|
| HTTP framework | `net/http` (Go 1.22+ ServeMux) | Zero deps. Go 1.25 has `GET /remember/{id}` pattern matching natively |
| Search engine | Bleve v2 (blevesearch/bleve/v2) | Pure Go, full-text search, BM25, disk-persisted |
| CLI flags | `flag` (stdlib) | Simple, no cobra needed for ~3 flags |
| Logging | `log/slog` (stdlib, Go 1.21+) | Structured, leveled, zero deps |
| Testing | `testing` + `httptest` | Stdlib, sufficient |

**Architect response:** ✅ Removed `chi` dependency. Go 1.22+ stdlib `http.ServeMux` supports method routing and `{id}` path params. Total external deps: **1** (Bleve).

## 3. Data Flow (unchanged)

### Remember (Write)
```
LLM → POST /remember {"id":"key1","text":"..."}
  → store.Remember(id, text)
    → bleve.Index(id, memoryDoc)
    → 200 OK {"id": "key1"}
```

### Search (Read)
```
LLM → POST /search {"query":"caching strategies"}
  → store.Search(query, limit)
    → bleve.Search(request)
    → 200 OK {"hits": [...], "total": N}
```

### Delete
```
LLM → DELETE /remember/{id}
  → store.Delete(id)
    → bleve.Delete(id)
    → 200 OK
```

## 4. Implementation Order

### Step 1: Project Init
- `go mod init github.com/.../gomem`
- `go get github.com/blevesearch/bleve/v2`
- Write `main.go` with flag parsing (`--port`, `--data-dir`)
- Write `types.go` with all structs

### Step 2: Store Layer (store.go)
- Define index mapping + `MemoryDoc` struct
- Implement `NewStore(path string) (*Store, error)` — open or create index
- Implement `Remember(id, text string) error`
- Implement `Search(query string, limit int) ([]SearchHit, error)`
- Implement `Delete(id string) error`
- Implement `Close() error`
- Error handling: wrap Bleve errors with descriptive messages

### Step 3: HTTP Server (handler.go + main.go wiring)
- `POST /remember` — decode JSON, call store.Remember, respond
- `POST /search` — decode JSON, call store.Search, respond
- `GET /search` — read `?q=` query param, call store.Search, respond
- `DELETE /remember/{id}` — call store.Delete, respond
- `GET /health` — respond `{"status":"ok"}`
- Graceful shutdown with `SIGINT`/`SIGTERM`
- Set `slog` level from `--log-level` flag

### Step 4: Error Contract
All error responses follow:
```json
{"error": "<human message>", "code": "<MACHINE_CODE>"}
```
| HTTP Code | Code | When |
|---|---|---|
| 400 | INVALID_REQUEST | Malformed JSON or missing fields |
| 404 | NOT_FOUND | ID not found on delete |
| 409 | ALREADY_EXISTS | ID already exists on remember |
| 500 | INTERNAL_ERROR | Bleve errors, disk failures, etc. |

### Step 5: Benchmark (bench_test.go)
- Insert 10k documents of ~200 bytes each (random words)
- Warm-up: 10 queries before measuring
- Benchmark function: `BenchmarkSearch` — measure P50 latency over 100 queries
- Queries: 3-word phrases present in the corpus
- Criteria: ≤50ms P50, report P95 and P99 in output
- Document: OS, Go version, SSD vs HDD in benchmark output

### Step 6: Polish
- README.md with example (curl commands)
- Cross-platform: build with `GOOS=linux/darwin/windows`
- Verify zero output on `--log-level=error` under normal operation

## 5. API Contracts

### POST /remember
```json
// Request
{"id": "proj-auth-overview", "text": "Auth uses JWT with RSA256, tokens expire in 24h"}
// 200
{"id": "proj-auth-overview"}
// 409
{"error": "document already exists", "code": "ALREADY_EXISTS"}
```

### POST /search
```json
// Request
{"query": "authentication", "limit": 10}
// 200
{"hits": [{"id": "proj-auth-overview", "score": 2.45, "text": "Auth uses JWT..."}], "total": 1}
```

### DELETE /remember/{id}
```json
// 200
{"status": "deleted"}
// 404
{"error": "not found", "code": "NOT_FOUND"}
```

### GET /health → 200 {"status": "ok"}

## 6. CLI Flags

| Flag | Default | Description |
|---|---|---|
| `--port` | 8080 | HTTP server port |
| `--data-dir` | `./gomem-data` | Directory for Bleve index |
| `--log-level` | `info` | Log level (debug, info, warn, error) |

## 7. Testing Strategy

| Layer | Type | What |
|---|---|---|
| Store | Unit (`store_test.go`) | Remember → Search → Delete → Search (gone). Persistence: reopen and verify data survives. |
| Store | Benchmark (`bench_test.go`) | 10k entries, P50/P95/P99 latency, ≤50ms P50 target |
| Server | Integration (`handler_test.go`) | httptest: full HTTP cycle for each endpoint |
| Error | Integration | Bad JSON → 400, delete non-existent → 404, duplicate → 409 |

## 8. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Bleve index corruption on unclean shutdown | Bleve handles via WAL; add `store.Open()` health check at startup |
| `--data-dir` parent doesn't exist | `os.MkdirAll` on startup with error handling |
| Cross-platform path issues | Use `filepath.Join` — Windows paths with backslashes work |
| Memory grows unbounded with large datasets | Monitored in benchmark; document expected RSS for 10k/100k entries |
