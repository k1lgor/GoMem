---
name: gomem
description: Persistent memory for AI coding agents. Search stored knowledge before reading files. Remember project context across sessions. Compatible with pi, Claude Code, Cline, Codex, and any agent with CLI access.
---

# GoMem — Persistent Memory for AI Coding Agents

GoMem stores and retrieves persistent memory for AI coding agents. Each project gets its own isolated `.gomem` directory. Data survives restarts and context resets.

**Core principle: Search memory FIRST, read files SECOND.** Before scanning a project with `read`, `ls`, `grep`, `find`, `cat`, `head`, or `memo_search`, always check GoMem — it's faster and cheaper.

## Setup

### 1. Install

Build the binary from the GoMem project root:

```bash
cd /path/to/gomem-project
go build -o gomem ./cmd/gomem
# Or download a prebuilt binary
```

Place `gomem` in your PATH:

```bash
# Linux/Mac
mv gomem ~/.local/bin/

# Windows
# Add the directory to PATH, or keep it in the project root
```

### 2. Auto-Install (Easiest)

```bash
# List all supported agents:
gomem skill

# Install for your agent (run from project root):
gomem skill pi            # Pi
gomem skill claude        # Claude Code
gomem skill cline         # Cline
gomem skill codex         # OpenAI Codex CLI
gomem skill copilot       # GitHub Copilot
gomem skill cursor        # Cursor
gomem skill windsurf      # Windsurf
gomem skill zed           # Zed
gomem skill kilo          # Kilo Code
gomem skill continue      # Continue
gomem skill gemini        # Gemini CLI

# Install globally (in ~/.agent/ instead of project dir):
gomem skill claude --global
gomem skill pi -g
```

### 3. Manual Install

**pi**: Add to `.pi/skills/gomem/`:
```bash
cp -r skills/gomem .pi/skills/
```

**Claude Code**: Copy to `.claude/skills/gomem/`:
```bash
cp -r skills/gomem .claude/skills/
```

**Cline**: Copy to `.cline/skills/gomem/`:
```bash
cp -r skills/gomem .cline/skills/
```

**Any agent**: Add this instruction to the system prompt or agent rules:
```
Before reading files with read/ls/grep/find/memo_search, 
always search GoMem memory first using the gomem CLI.
```

## Usage

### Prefix: Before ANY filesystem scan

When starting work on a project or exploring an unfamiliar codebase:

```bash
# 1. Check what's already in memory
gomem list

# 2. Search for relevant context
gomem search "architecture patterns"
gomem search "authentication implementation"
gomem search "database schema"

# 3. Only if nothing useful found, use filesystem tools
ls -la
# ... read files, grep, etc.

# 4. Store what you learn for next time
gomem remember arch-decision "We chose hexagonal architecture with ports and adapters"
```

### Store a Memory

```bash
gomem remember <id> <text>
```

The `<id>` should be a meaningful key (kebab-case) that you'll remember later. Examples:
- `project-arch` — project architecture overview
- `auth-flow` — authentication flow description
- `db-schema` — database schema summary
- `api-endpoints` — API endpoint list
- `key-insight-1` — numbered key insights
- `decision-cache-strategy` — architectural decisions

### Search Memories

```bash
gomem search "query terms here"
gomem search "*"        # list all
```

Search uses full-text matching with BM25 scoring. Results are ranked by relevance.

### List All Memories

```bash
gomem list
```

### Delete a Memory

```bash
gomem delete <id>
```

## Save All — Project Snapshot

The `gomem save-all` command reads an entire project and stores **concise structural summaries** of each file into GoMem memory. This gives the agent a complete picture of the codebase without having to re-read raw files.

**This is the key feature.** Instead of storing raw file contents (which wastes tokens), it extracts:
- Package/module names
- Imports/dependencies
- Structs, classes, interfaces, types
- Function and method signatures with doc comments
- For docs: headings, first paragraphs

Typical savings: **70-95% fewer tokens** vs reading raw files.

### Usage

```bash
# From the project root:
gomem save-all
```

### What It Stores

| Memory ID | Content |
|---|---|
| `project-overview` | Project name, file counts by type |
| `project-structure` | Directory tree (2 levels) |
| `file:<relative-path>` | Structural summary of each source file |
| `AGENTS.md` | Auto-generated at project root after save-all |

### Agent Workflow After Save-All

After running `save-all`, the agent can:

```bash
# Find everything about a topic
gomem search "authentication"

# Find which file contains what
gomem search "JWT token validation"

# Get project overview
gomem search "project-overview"

# Find specific files
gomem search "full-text"    # searches summary content, not filenames
```

## Best Practices

### DO: Search First

```
✅ gomem search "deployment pipeline"
   → Found 2 results. No need to grep the whole project.
```

### DON'T: Blindly Scan

```
❌ find . -name "*.go" | xargs grep "deploy"
   → Expensive. Use GoMem first.
```

### DO: Store After Discovery

When you find something important:

```bash
# After understanding a complex module
gomem remember module-parser "The parser module uses recursive descent. Entry point: parse/parser.go"

# After making a decision
gomem remember decision-cache "We chose Redis over Memcached because we need persistence"

# After a debugging session
gomem remember bug-fix-null-ptr "Null pointer in user.go:42 when User.Email is empty. Fixed with nil check."
```

### DON'T: Let Knowledge Die

If you don't store it, the next agent session starts from zero. Store key findings every time.

## save-all Built-In

The `gomem save-all` command is built into the binary (no scripts needed):

```bash
gomem save-all
```

It walks the project, reads each source file, and generates **concise structural summaries** using language-aware parsers. Supported formats include Go, Python, Java, TypeScript, JavaScript, Rust, C/C++, C#, Ruby, PHP, Swift, Kotlin, Scala, Dart, Lua, SQL, Terraform, Dockerfile, Makefile, and 50+ more.

Shell scripts (`skills/gomem/scripts/save-all.sh` and `save-all.ps1`) are also provided for convenience.

After indexing, GoMem also writes `AGENTS.md` at the project root so agents automatically use GoMem before filesystem tools.

## Example Workflows

### Onboarding a New Project

```bash
# 1. Save the entire project
./skills/gomem/scripts/save-all.sh .

# 2. Ask questions
gomem search "project architecture"
gomem search "file:main.go"
gomem search "dependencies"
```

### Resume After Context Reset

```bash
# 1. Check what you know
gomem list

# 2. Refresh on key topics
gomem search "decisions made"
gomem search "current task"

# 3. Continue working
```

### Code Review Session

```bash
# 1. Store what you're reviewing
gomem remember review-target "Analyzing PR #42: auth module refactor"

# 2. Search for affected areas
gomem search "authentication"
gomem search "user model"

# 3. Store findings
gomem remember review-finding-1 "Duplicate JWT validation in auth.go and middleware.go"
```

## Compatibility

GoMem works with any AI coding agent that can:
- Execute CLI commands
- Read files
- Be given instructions via a skill or system prompt

| Agent | Install Command |
|---|---|
| pi | `gomem skill pi` |
| Claude Code | `gomem skill claude` |
| Cline | `gomem skill cline` |
| OpenAI Codex CLI | `gomem skill codex` |
| GitHub Copilot | `gomem skill copilot` |
| Cursor | `gomem skill cursor` |
| Windsurf | `gomem skill windsurf` |
| Zed | `gomem skill zed` |
| Kilo Code | `gomem skill kilo` |
| Continue | `gomem skill continue` |
| Gemini CLI | `gomem skill gemini` |

All commands also write `AGENTS.md` at the project root, which automatically instructs any AI agent to use GoMem before filesystem tools.

Use `--global` or `-g` to install globally in `~/.<agent>/skills/gomem/` instead of the current project.

## Troubleshooting

**"gomem: command not found"** → Build the binary or add it to PATH:
```bash
export PATH=$PATH:/path/to/gomem-project
```

**"No memories stored yet"** → Run `save-all` first to populate memory, or use `gomem remember` to store specific facts.

**Memory is empty for this project** → Each project has its own `.gomem` directory. Make sure you're in the right project root.

**Search results not relevant** → Try different query terms. GoMem uses full-text BM25 matching — be specific with your query words.
