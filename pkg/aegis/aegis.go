// Package aegis provides the public API for the aegis-core PII detection engine.
//
// It re-exports the core types and functions so that external Go modules can
// import them without reaching into internal packages.
package aegis

import (
	"regexp"

	"github.com/svenplb/aegis-core/internal/redactor"
	"github.com/svenplb/aegis-core/internal/restorer"
	"github.com/svenplb/aegis-core/internal/scanner"
)

// ---------- Scanner types ----------

// Scanner detects PII entities in text.
type Scanner = scanner.Scanner

// Entity represents a detected PII entity with byte offsets.
type Entity = scanner.Entity

// DefaultScanner returns a CompositeScanner with all built-in regex patterns.
// Allowlist entries are compiled regexes; any entity whose text matches an
// allowlist pattern is dropped.
func DefaultScanner(allowlist []*regexp.Regexp) Scanner {
	return scanner.DefaultScanner(allowlist)
}

// NewCompositeScanner creates a scanner that merges results from multiple
// child scanners, deduplicating overlapping spans.
func NewCompositeScanner(scanners []Scanner, allowlist []*regexp.Regexp) Scanner {
	return scanner.NewCompositeScanner(scanners, allowlist)
}

// BuiltinScanners returns all built-in regex-based scanners.
func BuiltinScanners() []Scanner {
	return scanner.BuiltinScanners()
}

// ---------- Redaction ----------

// RedactResult holds the output of a Redact call.
type RedactResult = redactor.RedactResult

// Mapping links a placeholder token to its original text.
type Mapping = redactor.Mapping

// Redact replaces every entity span in text with a placeholder token
// (e.g. [PERSON_1]) and returns the sanitised text together with the
// mapping table needed for restoration.
func Redact(text string, entities []Entity) RedactResult {
	return redactor.Redact(text, entities)
}

// ---------- Restoration ----------

// Restore replaces every placeholder token in text with its original value.
// Tokens are replaced longest-first to avoid partial matches.
func Restore(text string, mappings []Mapping) string {
	return restorer.Restore(text, mappings)
}

// StreamRestorer incrementally restores tokens from streaming chunks,
// buffering incomplete tokens (an opening '[' without a matching ']').
type StreamRestorer = restorer.StreamRestorer

// NewStreamRestorer returns a StreamRestorer configured with the given mappings.
func NewStreamRestorer(mappings []Mapping) *StreamRestorer {
	return restorer.NewStreamRestorer(mappings)
}
