# Context Snapshot: GoMem Memory Project

## Task Statement
Build a Go memory project for AI agents — a single binary that acts as persistent memory ("a brain that never forgets") that LLMs can reference.

## Desired Outcome
A single Go binary that AI coding agents can install and use as persistent, queryable memory. Inspired by concepts like graphify, mempalace, and other memory tools for AI agents.

## Known Facts / Evidence
- Fresh project directory (D:/Coding/test/GoMem/)
- No existing code yet
- Current working directory is the GoMem project root
- User wants Go language, single binary deployment
- Concepts mentioned: graphify, mempalace — graph-based and spatial/loci-based memory patterns

## Constraints
- Must be Go language
- Single binary deployment
- Must work as persistent memory for AI coding agents
- LLMs must be able to reference/query the memory

## Unknowns / Open Questions
- What specific memory features are most important?
- What's the API surface (REST, CLI, gRPC, library)?
- What storage backend (SQLite, Badger, Bolt, custom)?
- What data model (graphs, key-value, vector embeddings)?
- Who are the target AI agents (Cline, Claude Code, custom)?
- What's the interaction model (read/write/search/delete)?
- Performance requirements?
- Is there an existing project this extends or is it from scratch?
