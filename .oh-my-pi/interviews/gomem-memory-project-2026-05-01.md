# Deep Interview Transcript — GoMem

**Date:** 2026-05-01
**Profile:** Standard (target ambiguity ≤ 0.20)
**Final Ambiguity:** ~0.16
**Rounds:** 17 total (including context intake, challenge modes, and crystallization)

## Summary

The user wants a Go-based persistent memory server for AI coding agents (specifically pi). It should be a single binary with an HTTP API, using Bleve for full-text search. The LLM generates search queries; the Go binary executes fast full-text lookups. No embeddings, no graph features, no synonym expansion in V1.

## Key Decision Points

1. **Language:** Go (for speed over Python alternatives)
2. **Deployment:** Standalone HTTP server (single binary, cross-platform)
3. **Concurrency:** Deferred — single-agent for V1
4. **Search model:** LLM generates queries → Go binary does full-text search (Bleve)
5. **Speed target:** ≤50ms for 10k stored entries
6. **Storage engine:** Bleve (pure Go, full-text, BM25, disk-persisted)
7. **V1 scope:** Minimal — just remember + search + delete + health
8. **Synonyms/embeddings:** Deferred beyond V1

## Challenge Modes Applied
- **Contrarian** (Round ~15): Challenged reliance on LLM query quality; proposed fallback mechanisms
- **Simplifier** (Round ~16): Compressed V1 to minimal viable scope (Bleve + HTTP, nothing else)

## Transcript

[See .oh-my-pi/specs/deep-interview-gomem-memory-project.md for the crystallized spec]
