package redactor

import (
	"sort"
	"time"

	"github.com/svenplb/aegis-core/internal/scanner"
)

// RedactResult holds the output of a Redact call.
type RedactResult struct {
	OriginalText   string           `json:"original_text"`
	SanitizedText  string           `json:"sanitized_text"`
	Entities       []scanner.Entity `json:"entities"`
	Mappings       []Mapping        `json:"mappings"`
	ProcessingTime int64            `json:"processing_time_ms"`
}

// Redact replaces every entity span in text with a placeholder token and
// returns the sanitised text together with the mapping table.
func Redact(text string, entities []scanner.Entity) RedactResult {
	start := time.Now()

	if len(entities) == 0 {
		return RedactResult{
			OriginalText:   text,
			SanitizedText:  text,
			Entities:       entities,
			Mappings:       nil,
			ProcessingTime: time.Since(start).Milliseconds(),
		}
	}

	// Sort entities by Start ascending to assign tokens in reading order.
	sorted := make([]scanner.Entity, len(entities))
	copy(sorted, entities)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Start < sorted[j].Start
	})

	// First pass: assign tokens in forward order so numbering matches reading order.
	counter := NewCounter()
	type tagged struct {
		ent   scanner.Entity
		token string
	}
	tags := make([]tagged, len(sorted))
	for i, ent := range sorted {
		tags[i] = tagged{ent: ent, token: counter.Next(ent.Type, ent.Text)}
	}

	// Second pass: replace in reverse order to preserve byte offsets.
	buf := []byte(text)
	mappings := make([]Mapping, 0, len(tags))
	for i := len(tags) - 1; i >= 0; i-- {
		t := tags[i]
		tokenBytes := []byte(t.token)
		newBuf := make([]byte, 0, len(buf)-t.ent.End+t.ent.Start+len(tokenBytes))
		newBuf = append(newBuf, buf[:t.ent.Start]...)
		newBuf = append(newBuf, tokenBytes...)
		newBuf = append(newBuf, buf[t.ent.End:]...)
		buf = newBuf

		mappings = append(mappings, Mapping{
			Token:    t.token,
			Original: t.ent.Text,
			Type:     t.ent.Type,
		})
	}

	// Deduplicate mappings (same token may appear multiple times).
	seen := make(map[string]bool, len(mappings))
	deduped := make([]Mapping, 0, len(mappings))
	for _, m := range mappings {
		if !seen[m.Token] {
			seen[m.Token] = true
			deduped = append(deduped, m)
		}
	}

	return RedactResult{
		OriginalText:   text,
		SanitizedText:  string(buf),
		Entities:       entities,
		Mappings:       deduped,
		ProcessingTime: time.Since(start).Milliseconds(),
	}
}
