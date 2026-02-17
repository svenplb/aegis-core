# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build                  # Build all 3 binaries (aegis-scan, aegis, aegis-server)
make build-server-nlp       # Build server with NLP/NER support (-tags nlp)
make test                   # Run all tests
make test-race              # Run tests with race detector
make test-nlp               # Run tests with NLP build tag
make bench                  # Benchmark scanner performance
make lint                   # go vet
make lint-nlp               # go vet with nlp tag
```

Run a single test:
```bash
go test ./internal/scanner/ -run TestName -v
```

## Build Tags

The `nlp` build tag controls NLP/NER support. Without it, only regex scanning is available (faster builds, smaller binaries). Files guarded by build tags come in pairs:
- `foo.go` (`//go:build nlp`) — real implementation
- `foo_stub.go` (`//go:build !nlp`) — no-op stub returning nil

Affected files: `internal/nlp/session*.go`, `internal/scanner/ner_scanner*.go`, `cmd/aegis-server/nlp*.go`

## Architecture

**Scanner interface** (`internal/scanner/scanner.go`): All detectors implement `Scanner.Scan(text string) []Entity`. Entities carry byte offsets, type, score, and detector origin.

**RegexScanner**: Wraps a compiled regex for one entity type. Options: `WithValidator()` for post-match validation (e.g. Luhn check), `WithExtractGroup()` for capture groups, `WithContextValidator()` for surrounding-text checks.

**CompositeScanner**: Merges results from multiple scanners. Deduplication: sort by (start asc, score desc, length desc), highest score wins overlapping spans. Options: `WithContextEnhancement()`, `WithMinScoreThreshold()`.

**Scanner chain**: `DefaultScanner(allowlist)` → `HybridScanner(allowlist, nil)` → `NewCompositeScanner(BuiltinScanners(), opts...)`. `HybridScanner` accepts an optional NER scanner as second arg.

**NERScanner** (`internal/scanner/ner_scanner.go`): Wraps Hugot BERT pipeline. Converts character offsets to byte offsets via rune mapping. Maps NER labels (PER/LOC/ORG/MISC) to aegis entity types.

**Context enhancer** (`internal/scanner/context_enhancer.go`): Boosts entity scores when contextual keywords (multilingual) appear within a configurable window around the match.

**Redaction** (`internal/redactor/`): Deterministic token assignment (e.g. `[PERSON_1]`), reverse-order replacement to preserve byte offsets.

**Restoration** (`internal/restorer/`): Replaces tokens back to original text, longest-first to avoid prefix collisions. Includes `StreamRestorer` for chunked processing.

## Three Binaries

- **aegis-scan** (`cmd/aegis-scan/`): CLI scanner. Reads from `--text`, `--file`, or stdin. Exit code 1 = PII found.
- **aegis** (`cmd/aegis/`): Interactive TUI (Bubble Tea). States: input → results → settings.
- **aegis-server** (`cmd/aegis-server/`): HTTP API on port 9090. Endpoints: `/health`, `/api/scan`, `/api/redact`, `/api/restore`.

## Configuration

YAML config loaded via `config.Load(path)`. Sections: `scanner` (custom patterns, allowlist, min_score), `nlp` (enabled, model_id, cache_dir), `context` (boost_factor, window_size), `logging` (level). See `config.example.yaml`.

Env vars: `AEGIS_SERVER_PORT`, `AEGIS_CORS_ORIGINS`.

## Key Gotchas

- **Typed nil**: When passing NER scanner to `HybridScanner`, pass as `scanner.Scanner` interface, not `*NERScanner` — a typed nil pointer is not a nil interface.
- **Scanner order**: `BuiltinScanners()` ordering matters for overlap resolution — higher-priority scanners come first.
- **Byte vs character offsets**: Entity offsets are byte-based. NERScanner must convert from character offsets.
- **NFC normalization**: All input text is Unicode-normalized before scanning.
