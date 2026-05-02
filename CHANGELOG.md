# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-05-01

### Added
- Initial CLI tool with `remember`, `search`, `list`, `delete`, `save-all`, `skill`, `version` commands
- Bleve full-text search engine with BM25 scoring
- Project-local `.gomem` storage directory (0700 permissions)
- `gomem save-all` — index entire project as concise structural summaries (70-95% token savings)
- Structural summarizers for 50+ file formats (Go, Python, Java, TS/JS, Rust, C/C++, C#, Ruby, PHP, Swift, Kotlin, Scala, Dart, Lua, SQL, Dockerfile, Makefile, Terraform, etc.)
- `gomem skill` — auto-install GoMem skill for 11 AI coding agents (claude, pi, cline, codex, copilot, cursor, windsurf, zed, kilo, continue, gemini)
- AGENTS.md auto-generation — agents automatically use GoMem before filesystem tools
- Standard Go project layout: `cmd/gomem/main.go` + `package gomem`
- CI/CD pipeline with GitHub Actions (build + test + cross-platform release)
- Changelog-based release notes generation
- Benchmark: 10k entries, ~5ms P50 query latency

### Changed
- `save-all` stores summaries instead of raw file content
- Switched from `QueryStringQuery` to `MatchQuery` for faster search
- Removed HTTP server mode — pure CLI tool
- `Delete` now verifies document existence before deletion
- Replaced `save-all.bat` with `save-all.ps1`

### Fixed
- Delete no longer reports success for non-existent documents
- `.gomem/`, `.oh-my-pi/`, `.pi/` removed from git tracking

### Removed
- HTTP server (`serve`, `handler.go`, `serve.go`)
- Unused `created_at` field from MemoryDoc
- Redundant sort-by-score in search
