package aegis_test

import (
	"testing"

	"github.com/svenplb/aegis-core/pkg/aegis"
)

func TestDefaultScannerDetectsEmail(t *testing.T) {
	sc := aegis.DefaultScanner(nil)
	entities := sc.Scan("Contact john@example.com for info.")
	if len(entities) == 0 {
		t.Fatal("expected at least one entity")
	}
	found := false
	for _, e := range entities {
		if e.Type == "EMAIL" && e.Text == "john@example.com" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected EMAIL entity john@example.com, got %v", entities)
	}
}

func TestRedactAndRestore(t *testing.T) {
	sc := aegis.DefaultScanner(nil)
	text := "Email me at alice@test.org please."
	entities := sc.Scan(text)

	result := aegis.Redact(text, entities)
	if result.SanitizedText == text {
		t.Fatal("expected redaction to change text")
	}
	if len(result.Mappings) == 0 {
		t.Fatal("expected at least one mapping")
	}

	restored := aegis.Restore(result.SanitizedText, result.Mappings)
	if restored != text {
		t.Errorf("restore failed: got %q, want %q", restored, text)
	}
}

func TestStreamRestorer(t *testing.T) {
	mappings := []aegis.Mapping{
		{Token: "[PERSON_1]", Original: "Alice", Type: "PERSON"},
	}
	sr := aegis.NewStreamRestorer(mappings)

	out := sr.Process("Hello [PER")
	out += sr.Process("SON_1], how are you?")
	out += sr.Flush()

	want := "Hello Alice, how are you?"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}
