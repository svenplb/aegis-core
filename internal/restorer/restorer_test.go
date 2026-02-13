package restorer

import (
	"testing"

	"github.com/svenplb/aegis-core/internal/redactor"
	"github.com/svenplb/aegis-core/internal/scanner"
)

func TestRestore_SingleToken(t *testing.T) {
	text := "Call [PERSON_1] tomorrow."
	mappings := []redactor.Mapping{
		{Token: "[PERSON_1]", Original: "Thomas Schmidt", Type: "PERSON"},
	}

	got := Restore(text, mappings)
	want := "Call Thomas Schmidt tomorrow."
	if got != want {
		t.Errorf("Restore = %q, want %q", got, want)
	}
}

func TestRestore_MultipleTokens(t *testing.T) {
	text := "[PERSON_1] emailed [EMAIL_1]."
	mappings := []redactor.Mapping{
		{Token: "[PERSON_1]", Original: "Alice", Type: "PERSON"},
		{Token: "[EMAIL_1]", Original: "alice@example.com", Type: "EMAIL"},
	}

	got := Restore(text, mappings)
	want := "Alice emailed alice@example.com."
	if got != want {
		t.Errorf("Restore = %q, want %q", got, want)
	}
}

func TestRestore_LongestFirst(t *testing.T) {
	// [PERSON_10] must be replaced before [PERSON_1] to avoid partial match.
	text := "Hello [PERSON_1] and [PERSON_10]."
	mappings := []redactor.Mapping{
		{Token: "[PERSON_1]", Original: "Alice", Type: "PERSON"},
		{Token: "[PERSON_10]", Original: "Bob", Type: "PERSON"},
	}

	got := Restore(text, mappings)
	want := "Hello Alice and Bob."
	if got != want {
		t.Errorf("Restore = %q, want %q", got, want)
	}
}

func TestRestore_EmptyMappings(t *testing.T) {
	text := "Nothing to restore."
	got := Restore(text, nil)
	if got != text {
		t.Errorf("Restore = %q, want %q", got, text)
	}
}

func TestRoundTrip(t *testing.T) {
	original := "Alice met Bob at the park."
	entities := []scanner.Entity{
		{Start: 0, End: 5, Type: "PERSON", Text: "Alice", Score: 0.9, Detector: "regex"},
		{Start: 10, End: 13, Type: "PERSON", Text: "Bob", Score: 0.9, Detector: "regex"},
	}

	result := redactor.Redact(original, entities)
	restored := Restore(result.SanitizedText, result.Mappings)

	if restored != original {
		t.Errorf("round-trip failed: got %q, want %q", restored, original)
	}
}

func TestStreamRestore_CompleteToken(t *testing.T) {
	mappings := []redactor.Mapping{
		{Token: "[PERSON_1]", Original: "Alice", Type: "PERSON"},
	}
	sr := NewStreamRestorer(mappings)

	got := sr.Process("Hello [PERSON_1]!")
	want := "Hello Alice!"
	if got != want {
		t.Errorf("Process = %q, want %q", got, want)
	}
}

func TestStreamRestore_SplitToken(t *testing.T) {
	mappings := []redactor.Mapping{
		{Token: "[PERSON_1]", Original: "Alice", Type: "PERSON"},
	}
	sr := NewStreamRestorer(mappings)

	// First chunk has incomplete token.
	out1 := sr.Process("Hello [PERS")
	if out1 != "Hello " {
		t.Errorf("Process chunk1 = %q, want %q", out1, "Hello ")
	}

	// Second chunk completes the token.
	out2 := sr.Process("ON_1] rest")
	if out2 != "Alice rest" {
		t.Errorf("Process chunk2 = %q, want %q", out2, "Alice rest")
	}
}

func TestStreamRestore_Flush(t *testing.T) {
	mappings := []redactor.Mapping{
		{Token: "[PERSON_1]", Original: "Alice", Type: "PERSON"},
	}
	sr := NewStreamRestorer(mappings)

	// Chunk ends with incomplete bracket — held in buffer.
	out := sr.Process("end [")
	if out != "end " {
		t.Errorf("Process = %q, want %q", out, "end ")
	}

	// Flush emits the remaining buffer as-is.
	flushed := sr.Flush()
	if flushed != "[" {
		t.Errorf("Flush = %q, want %q", flushed, "[")
	}
}
