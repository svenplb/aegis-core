package redactor

import (
	"testing"

	"github.com/svenplb/aegis-core/internal/scanner"
)

func TestRedact_SingleEntity(t *testing.T) {
	text := "Call Thomas Schmidt tomorrow."
	entities := []scanner.Entity{
		{Start: 5, End: 19, Type: "PERSON", Text: "Thomas Schmidt", Score: 0.95, Detector: "regex"},
	}

	result := Redact(text, entities)

	want := "Call [PERSON_1] tomorrow."
	if result.SanitizedText != want {
		t.Errorf("SanitizedText = %q, want %q", result.SanitizedText, want)
	}
	if len(result.Mappings) != 1 {
		t.Fatalf("len(Mappings) = %d, want 1", len(result.Mappings))
	}
	if result.Mappings[0].Token != "[PERSON_1]" {
		t.Errorf("Token = %q, want %q", result.Mappings[0].Token, "[PERSON_1]")
	}
	if result.Mappings[0].Original != "Thomas Schmidt" {
		t.Errorf("Original = %q, want %q", result.Mappings[0].Original, "Thomas Schmidt")
	}
}

func TestRedact_MultipleEntitiesSameType(t *testing.T) {
	text := "Alice met Bob at the park."
	entities := []scanner.Entity{
		{Start: 0, End: 5, Type: "PERSON", Text: "Alice", Score: 0.9, Detector: "regex"},
		{Start: 10, End: 13, Type: "PERSON", Text: "Bob", Score: 0.9, Detector: "regex"},
	}

	result := Redact(text, entities)

	want := "[PERSON_1] met [PERSON_2] at the park."
	if result.SanitizedText != want {
		t.Errorf("SanitizedText = %q, want %q", result.SanitizedText, want)
	}
	if len(result.Mappings) != 2 {
		t.Fatalf("len(Mappings) = %d, want 2", len(result.Mappings))
	}
}

func TestRedact_MultipleDifferentTypes(t *testing.T) {
	text := "Email alice@example.com or call Alice."
	entities := []scanner.Entity{
		{Start: 6, End: 23, Type: "EMAIL", Text: "alice@example.com", Score: 0.99, Detector: "regex"},
		{Start: 32, End: 37, Type: "PERSON", Text: "Alice", Score: 0.9, Detector: "regex"},
	}

	result := Redact(text, entities)

	want := "Email [EMAIL_1] or call [PERSON_1]."
	if result.SanitizedText != want {
		t.Errorf("SanitizedText = %q, want %q", result.SanitizedText, want)
	}
}

func TestRedact_SameTextReusesToken(t *testing.T) {
	text := "Alice and Bob met Alice again."
	entities := []scanner.Entity{
		{Start: 0, End: 5, Type: "PERSON", Text: "Alice", Score: 0.9, Detector: "regex"},
		{Start: 10, End: 13, Type: "PERSON", Text: "Bob", Score: 0.9, Detector: "regex"},
		{Start: 18, End: 23, Type: "PERSON", Text: "Alice", Score: 0.9, Detector: "regex"},
	}

	result := Redact(text, entities)

	want := "[PERSON_1] and [PERSON_2] met [PERSON_1] again."
	if result.SanitizedText != want {
		t.Errorf("SanitizedText = %q, want %q", result.SanitizedText, want)
	}
	// Deduplicated: only 2 unique mappings.
	if len(result.Mappings) != 2 {
		t.Fatalf("len(Mappings) = %d, want 2", len(result.Mappings))
	}
}

func TestRedact_UTF8Multibyte(t *testing.T) {
	// German umlauts are multi-byte in UTF-8: Ä=2 bytes, ö=2, ü=2, ß=2.
	text := "Herr Müller wohnt in Österreich."
	// "Müller" starts at byte 5, 'M'(1) + 'ü'(2) + 'l'(1) + 'l'(1) + 'e'(1) + 'r'(1) = 7 bytes → End=12
	// "Österreich" starts at byte 22 (after "wohnt in "), 'Ö'(2)+s+t+e+r+r+e+i+c+h = 11 bytes → End=33
	muellerStart := len("Herr ")     // 5
	muellerEnd := muellerStart + len("Müller") // 5 + 7 = 12
	oesterreichStart := len("Herr Müller wohnt in ") // 22
	oesterreichEnd := oesterreichStart + len("Österreich") // 22 + 11 = 33

	entities := []scanner.Entity{
		{Start: muellerStart, End: muellerEnd, Type: "PERSON", Text: "Müller", Score: 0.9, Detector: "regex"},
		{Start: oesterreichStart, End: oesterreichEnd, Type: "LOCATION", Text: "Österreich", Score: 0.85, Detector: "regex"},
	}

	result := Redact(text, entities)

	want := "Herr [PERSON_1] wohnt in [LOCATION_1]."
	if result.SanitizedText != want {
		t.Errorf("SanitizedText = %q, want %q", result.SanitizedText, want)
	}
}

func TestRedact_EmptyEntities(t *testing.T) {
	text := "Nothing to redact here."
	result := Redact(text, nil)

	if result.SanitizedText != text {
		t.Errorf("SanitizedText = %q, want %q", result.SanitizedText, text)
	}
	if result.OriginalText != text {
		t.Errorf("OriginalText = %q, want %q", result.OriginalText, text)
	}
	if len(result.Mappings) != 0 {
		t.Errorf("len(Mappings) = %d, want 0", len(result.Mappings))
	}
}

func TestRedact_ReverseOrderProcessing(t *testing.T) {
	// Entities provided in forward order should still be processed correctly.
	text := "AB CD EF"
	entities := []scanner.Entity{
		{Start: 0, End: 2, Type: "X", Text: "AB", Score: 1, Detector: "test"},
		{Start: 3, End: 5, Type: "X", Text: "CD", Score: 1, Detector: "test"},
		{Start: 6, End: 8, Type: "X", Text: "EF", Score: 1, Detector: "test"},
	}

	result := Redact(text, entities)

	want := "[X_1] [X_2] [X_3]"
	if result.SanitizedText != want {
		t.Errorf("SanitizedText = %q, want %q", result.SanitizedText, want)
	}
}

func TestCounter_Next(t *testing.T) {
	c := NewCounter()

	tok1 := c.Next("PERSON", "Alice")
	if tok1 != "[PERSON_1]" {
		t.Errorf("tok1 = %q, want [PERSON_1]", tok1)
	}

	tok2 := c.Next("PERSON", "Bob")
	if tok2 != "[PERSON_2]" {
		t.Errorf("tok2 = %q, want [PERSON_2]", tok2)
	}

	// Same text → same token.
	tok3 := c.Next("PERSON", "Alice")
	if tok3 != "[PERSON_1]" {
		t.Errorf("tok3 = %q, want [PERSON_1]", tok3)
	}

	// Different type starts at 1.
	tok4 := c.Next("EMAIL", "alice@example.com")
	if tok4 != "[EMAIL_1]" {
		t.Errorf("tok4 = %q, want [EMAIL_1]", tok4)
	}
}
