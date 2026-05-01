# GoMem — Persistent Memory for AI Coding Agents

A fast, single-binary Go tool for persistent, searchable memory. Designed for AI coding agents (pi, Claude Code, Cline, etc.).

**CLI-first** — use it directly from the terminal, or start the HTTP server for agent integration.

## Quick Start

```bash
# Build
go build -o gomem .

# Store a memory
gomem remember my-key "The API uses JWT with RSA256, tokens expire in 24h"

# Search memories
gomem search "JWT authentication"

# List all memories
gomem list

# Delete a memory
gomem delete my-key

# Start HTTP server (for AI agent integration)
gomem serve --port 8080
```

## CLI Commands

| Command | Usage | Description |
|---|---|---|
| `remember` | `gomem remember <id> <text>` | Store a memory |
| `search` | `gomem search <query>` | Search memories (use `*` for all) |
| `list` | `gomem list` | List all memories |
| `delete` | `gomem delete <id>` | Delete a memory by ID |
| `serve` | `gomem serve [--port N]` | Start HTTP server |
| `help` | `gomem help` | Show help |

## HTTP API (Server Mode)

Start with `gomem serve --port 8080`.

| Method | Endpoint | Body | Description |
|---|---|---|---|
| POST | `/remember` | `{"id":"...", "text":"..."}` | Store a memory |
| POST | `/search` | `{"query":"...", "limit":10}` | Search memories |
| GET | `/search?q=...` | — | Search by query param |
| DELETE | `/remember/{id}` | — | Delete a memory |
| GET | `/health` | — | Health check |

## Configuration

Set the data directory via:
- `GOMEM_DATA_DIR` environment variable
- `--data-dir` flag (serve mode only)
- Defaults to `~/.gomem`

## Performance

```
$ go test -bench=BenchmarkSearch -benchtime=100x
BenchmarkSearch-4   100   5.5ms/op   P50: 5.4ms   P95: 6.8ms   P99: 7.8ms
```

10k entries, ~200 bytes each, 3-word queries on Intel i5-7300HQ (2017 laptop).

## Integration with AI Agents

### pi
Add to your pi configuration to let it call GoMem as a tool.

### Claude Code
Use the HTTP server mode and configure MCP or a custom skill.

### Any agent
Direct CLI calls from the agent:
```
gomem remember session-123 "Key decisions made: ..."
gomem search "previous decisions about caching"
```

## Files

```
GoMem/
├── main.go       CLI dispatcher
├── serve.go      HTTP server mode
├── handler.go    HTTP handlers
├── store.go      Bleve search engine wrapper
├── types.go      Shared types
├── *_test.go     Tests + benchmark
└── README.md
```

## License

MIT
