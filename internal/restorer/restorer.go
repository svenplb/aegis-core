package restorer

import (
	"sort"
	"strings"

	"github.com/svenplb/aegis-core/internal/redactor"
)

// Restore replaces every placeholder token in text with its original value.
// Tokens are replaced longest-first to avoid partial matches
// (e.g. [PERSON_10] is replaced before [PERSON_1]).
func Restore(text string, mappings []redactor.Mapping) string {
	if len(mappings) == 0 {
		return text
	}

	sorted := make([]redactor.Mapping, len(mappings))
	copy(sorted, mappings)
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i].Token) > len(sorted[j].Token)
	})

	for _, m := range sorted {
		text = strings.ReplaceAll(text, m.Token, m.Original)
	}
	return text
}

// StreamRestorer incrementally restores tokens from streaming chunks.
// It buffers incomplete tokens (an opening '[' without a matching ']').
type StreamRestorer struct {
	mappings []redactor.Mapping
	buffer   string
}

// NewStreamRestorer returns a StreamRestorer configured with the given mappings.
func NewStreamRestorer(mappings []redactor.Mapping) *StreamRestorer {
	sorted := make([]redactor.Mapping, len(mappings))
	copy(sorted, mappings)
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i].Token) > len(sorted[j].Token)
	})
	return &StreamRestorer{mappings: sorted}
}

// Process accepts the next chunk of streamed text. It returns any text that
// can be emitted immediately, buffering incomplete tokens for later.
func (sr *StreamRestorer) Process(chunk string) string {
	sr.buffer += chunk

	// Find the last '[' that has no matching ']' after it.
	lastOpen := strings.LastIndex(sr.buffer, "[")
	if lastOpen != -1 && !strings.Contains(sr.buffer[lastOpen:], "]") {
		// Everything before the '[' is safe to emit; keep the rest buffered.
		safe := sr.buffer[:lastOpen]
		sr.buffer = sr.buffer[lastOpen:]
		return sr.replaceMappings(safe)
	}

	// No incomplete token — emit everything.
	out := sr.replaceMappings(sr.buffer)
	sr.buffer = ""
	return out
}

// Flush returns any remaining buffered text after applying replacements.
func (sr *StreamRestorer) Flush() string {
	out := sr.replaceMappings(sr.buffer)
	sr.buffer = ""
	return out
}

func (sr *StreamRestorer) replaceMappings(text string) string {
	for _, m := range sr.mappings {
		text = strings.ReplaceAll(text, m.Token, m.Original)
	}
	return text
}
