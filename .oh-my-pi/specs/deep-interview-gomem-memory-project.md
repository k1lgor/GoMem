# GoMem — Deep Interview Spec

## Intent
Python-based memory tools (graphify, mempalace) are too slow. Build a fast, persistent memory server in Go that AI coding agents can query via a simple API.

## Desired Outcome
A single Go binary (≤50ms query latency for 10k entries) that runs as a standalone HTTP server, persists memories to disk via Bleve full-text search, and is queried by an LLM agent that generates search queries.

## API Surface (V1)
| Method | Endpoint | Body | Description |
|---|---|---|---|
| POST | /remember | `{"id": "...", "text": "..."}` | Store a memory |
| GET | /search?q=... | — | Search memories by query string |
| POST | /search | `{"query": "..."}` | Search memories by query string (body) |
| DELETE | /remember/:id | — | Delete a memory |
| GET | /health | — | Health check |

## In-Scope (V1)
- Go single binary, cross-platform (Linux, macOS, Windows)
- HTTP REST API
- Bleve as the storage + full-text search engine (BM25, stemming, stop-word removal)
- LLM generates search queries → Go binary executes full-text search → returns results
- Disk-persisted index
- Single-agent (no concurrency protection needed)
- ≤50ms query latency for 10k entries (target)

## Out-of-Scope (V1)
- Embedding models / vector search in the binary
- Synonym expansion / thesaurus
- Graph data model / graph traversal
- Multi-agent concurrency / locking
- gRPC
- Authentication / authorization
- Clustering / replication

## Decision Boundaries (agent may decide without confirmation)
- HTTP framework (net/http stdlib or chi/gin)
- Bleve index configuration (analyzer choice, index mapping)
- CLI flags (--port, --data-dir)
- Go project layout and package structure
- Logging framework / verbosity levels

## Constraints
- Pure Go (no CGo unless unavoidable)
- Minimal memory footprint
- Open source (public or personal)

## Testable Acceptance Criteria
1. `POST /remember` with `{"id": "...", "text": "..."}` stores and persists the text
2. `POST /search {"query": "..."}` returns ranked results with relevance scores (Bleve hits)
3. 10k stored entries → single query returns in ≤50ms
4. Binary produces zero output under normal operation (log level controls)
5. Binary starts and responds on a configurable port (default 8080)
6. Data survives process restart (disk-persisted index)

## Assumptions Exposed + Resolutions
| Assumption | Resolution |
|---|---|
| "As fast as possible" is a target | → ≤50ms for 10k entries |
| Embeddings needed for semantic search | → LLM generates the query, not the binary |
| Single binary means all-inclusive | → Pure Go + Bleve, no external deps |
| Concurrency may be needed | → Deferred; single-agent for V1 |
