# GoMem — Persistent Memory for AI Coding Agents

A fast, single-binary Go tool for persistent, searchable memory. Designed for AI coding agents (pi, Claude Code, Cline, etc.).

**Core principle: Search GoMem FIRST, read files SECOND.** GoMem stores concise structural summaries (not raw file contents), saving 70-95% tokens vs re-reading source files. Memory survives context resets and compaction.

## Quick Start

```bash
# Build
go build -o gomem ./cmd/gomem

# Index your project (stores structural summaries of every file)
gomem save-all

# Search — always do this before read/grep/cat/ls/head
gomem search "authentication"
gomem search "database schema"

# List what's in memory
gomem list

# Store key insights
gomem remember arch-decision "Hexagonal architecture with ports and adapters"

# Delete a memory
gomem delete my-key

# Install skill for your AI agent (auto-creates AGENTS.md too)
gomem skill claude
gomem skill pi
```

## How It Works

GoMem uses [Bleve](https://github.com/blevesearch/bleve) for full-text search with BM25 scoring. Each project gets its own `.gomem` directory in the project root — fully isolated.

### save-all — Project Indexing

`gomem save-all` walks your project and generates **concise structural summaries** for every source file:

| Instead of reading 500 lines of raw code...                                                     | ...GoMem stores this                                                  |
| ----------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| `func (s *Store) Search(q string, limit int) ([]SearchHit, uint64, error) { ... 20 lines ... }` | `(s *Store) Search(q string, limit int) ([]SearchHit, uint64, error)` |

Supported formats: Go, Python, Java, TypeScript, JavaScript, Rust, C/C++, C#, Ruby, PHP, Swift, Kotlin, Scala, Dart, Lua, SQL, Terraform, Dockerfile, Makefile, and 50+ more.

## CLI Commands

| Command                | Description                                     |
| ---------------------- | ----------------------------------------------- |
| `remember <id> <text>` | Store a memory (summary, key insight, decision) |
| `search <query>`       | Search memories (use `*` for all)               |
| `list`                 | List all memories                               |
| `delete <id>`          | Delete a memory by ID                           |
| `save-all`             | Index current project as concise summaries      |
| `skill <agent>`        | Install skill for an AI agent                   |
| `help`                 | Show help                                       |

### skill — Install for Any AI Agent

Installs the GoMem skill so the agent knows to use GoMem before filesystem tools. Also creates `AGENTS.md` at the project root for automatic agent instruction.

```bash
gomem skill                   # List all supported agents
gomem skill claude            # Install for Claude Code
gomem skill pi                # Install for Pi
gomem skill cline             # Install for Cline
gomem skill codex             # Install for OpenAI Codex CLI
gomem skill copilot           # Install for GitHub Copilot
gomem skill cursor            # Install for Cursor
gomem skill windsurf          # Install for Windsurf
gomem skill zed               # Install for Zed
gomem skill kilo              # Install for Kilo Code
gomem skill continue          # Install for Continue
gomem skill gemini            # Install for Gemini CLI
gomem skill claude --global   # Install globally in ~/.claude/
```

## AGENTS.md

Every `gomem save-all` or `gomem skill` run creates/updates `AGENTS.md` at the project root. AI coding agents automatically read this file on session start, instructing them to:

```
1. Search first — Before read, ls, grep, find, cat, head, or memo_search:
   gomem list
   gomem search "<topic>"

2. Filesystem only if needed — Only if GoMem returns nothing useful

3. Store what you learn — gomem remember <id> "<concise summary>"

4. Index new projects — gomem save-all
```

## Performance

### Search Latency (Bleve index)

```
$ go test -bench=BenchmarkSearch -benchtime=30x
BenchmarkSearch-4   30   5.0ms/op   P50: 4.9ms   P95: 5.9ms   P99: 6.0ms
```

10k entries, ~200 bytes each, 3-word queries on Intel i5-7300HQ (2017 laptop).

### Token Savings on Real Repos

| Project          | Language   | Files | Raw chars | GoMem chars | Saved   |
| ---------------- | ---------- | ----- | --------- | ----------- | ------- |
| **GoMem** (self) | Go         | 28    | 96,902    | 9,293       | **90%** |
| **Chalk**        | JavaScript | 35    | 140,731   | 2,562       | **98%** |
| **Express**      | JavaScript | 214   | 734,039   | 16,377      | **97%** |
| **Zod**          | TypeScript | 580   | 5,872,285 | 92,530      | **98%** |

> Benchmarks: `gomem save-all` then sum raw source characters vs stored summary characters.

## When to Use GoMem vs Read/Grep/Ls/Cat/Head

### Use GoMem when you need to KNOW what's in the codebase

GoMem stores **structural summaries** — package names, imports, structs, function signatures, doc comments. This covers ~80% of what an agent needs. Real examples from the GoMem project:

| Question                               | GoMem query                       | What you get                                                                                                                                                      | vs reading raw file                      |
| -------------------------------------- | --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| "What does the store layer do?"        | `gomem search "Store struct"`     | `Package main. Imports: fmt, os, bleve/v2... Structs: Store{}. Functions: NewStore(), Remember(), Search(), Delete(), Close(), DocCount()`                        | 4,103 chars → **1,265** (70% saved)      |
| "What functions does saveall.go have?" | `gomem search "saveall.go"`       | `Functions: cmdSaveAllImpl(), summarizeProject(), buildCompactTree(), summarizeFile(), isBinaryExt(), summarizeGoFile(), summarizeMarkdown(), summarizeText()...` | 20,518 chars → **3,183** (85% saved)     |
| "What imports does the project use?"   | `gomem search "Imports:"`         | `fmt, os, bleve/v2, v2/mapping, search/query, path/filepath, regexp, strings, testing, math/rand, time, sort, runtime`                                            | All 7 source files → one line            |
| "Is there a function called NewStore?" | `gomem search "NewStore"`         | `// NewStore opens an existing Bleve index at path or creates a new one. NewStore(path string) (*Store, error)`                                                   | Must grep + read full function body      |
| "What does this project do?"           | `gomem search "project-overview"` | `Project GoMem: 93 files (7 Go, 9 Markdown, 77 other). Root: ...`                                                                                                 | Read README + scan directory tree        |
| "Find all agent configuration"         | `gomem search "claude"`           | `AgentConfig{Name, Dir, InstallFile, Format, Description}. Agents: pi, claude, cline, codex, copilot, cursor, windsurf, zed, kilo, continue, gemini`              | grep across skill.go + match agent names |

### Use Read/Grep when you need to SEE the exact code

Some tasks genuinely require raw source:

| Situation                  | Why you need read                                          |
| -------------------------- | ---------------------------------------------------------- |
| Debugging a specific line  | Need exact line numbers and surrounding context            |
| Counting occurrences       | `grep -c` for exact frequency                              |
| Regex pattern matching     | GoMem doesn't support regex                                |
| Verifying a fix            | Need to see the full function body, not just the signature |
| The project wasn't indexed | No `.gomem` directory exists yet                           |

### How it saves tokens

**Compression ratio is the key.** A typical function in source code:

```go
// Search performs a full-text query against the index and returns matching hits.
func (s *Store) Search(q string, limit int) ([]SearchHit, uint64, error) {
    if limit <= 0 || limit > 100 {
        limit = 10
    }
    // ... 15+ lines of implementation
}
```

GoMem stores just the **signature + doc comment**:

```
// Search performs a full-text query against the index and returns matching hits.
(s *Store) Search(q string, limit int) ([]SearchHit, uint64, error)
```

That's **90% fewer characters** — and the agent still knows the exact API, parameters, and return types.

### The compounding effect

The real savings come from **cross-session persistence**. You index once with `gomem save-all`, and that knowledge persists across context resets forever. Every time context compacts:

| Approach | Tokens per reset | After 10 resets | Cumulative vs re-reading |
|---|---|---|---|
| Re-read all source files | ~559,878 | ~5,598,780 | — |
| `gomem list` + 2 searches | ~2,827 | ~28,270 | **99.5% fewer tokens** |

GoMem memory survives compaction. On the first session you pay the indexing cost (summaries of all files). Every subsequent session you just pay for `gomem list` + a few searches — a fraction of what re-reading from scratch would cost.

## Files

```
GoMem/
├── main.go         CLI dispatcher and command handlers
├── store.go        Bleve search engine wrapper
├── saveall.go      Project indexing with structural summaries
├── skill.go        Agent skill installer + AGENTS.md writer
├── types.go        Shared types
├── store_test.go   Unit tests
├── bench_test.go   Benchmark
├── skills/         Agent skill files
│   └── gomem/
│       ├── SKILL.md
│       └── scripts/
├── AGENTS.md       Auto-generated agent instructions
└── README.md
```

## License

MIT
