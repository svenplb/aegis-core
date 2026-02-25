package redactor

import (
	"testing"

	"github.com/svenplb/aegis-core/internal/scanner"
)

func TestNFC_NFDUmlautBeforeAmount(t *testing.T) {
	// "ü" in NFD form: U+0075 (u) + U+0308 (combining diaeresis) = 3 bytes instead of 2 for NFC "ü".
	// This shifts all byte offsets after the umlaut by +1 per NFD character.
	// The scanner NFC-normalizes internally, so offsets refer to NFC text.
	// Without NFC normalization in Redact(), the replacement would be off-by-one.
	nfdText := "f\u0075\u0308r 1.234,56 \u20AC rest" // "für 1.234,56 € rest" in NFD

	// NFC version: "für 1.234,56 € rest"
	// f(1) ü(2) r(1) space(1) = byte 5; "1.234,56 €" = bytes 5..17 (€ is 3 bytes)
	entities := []scanner.Entity{
		{Start: 5, End: 17, Type: "FINANCIAL", Text: "1.234,56 \u20AC", Score: 0.90, Detector: "regex"},
	}

	result := Redact(nfdText, entities)

	want := "f\u00FCr [FINANCIAL_1] rest"
	if result.SanitizedText != want {
		t.Errorf("SanitizedText = %q, want %q", result.SanitizedText, want)
	}
}

func TestNFC_AlreadyNFCText(t *testing.T) {
	// Text is already NFC — should work unchanged.
	text := "für 1.234,56 \u20AC rest"
	entities := []scanner.Entity{
		{Start: 5, End: 17, Type: "FINANCIAL", Text: "1.234,56 \u20AC", Score: 0.90, Detector: "regex"},
	}

	result := Redact(text, entities)

	want := "für [FINANCIAL_1] rest"
	if result.SanitizedText != want {
		t.Errorf("SanitizedText = %q, want %q", result.SanitizedText, want)
	}
}

func TestNFC_MultipleNFDCharacters(t *testing.T) {
	// Multiple NFD characters: "Müller" with NFD ü and "ö" → "Möller"
	// "Herr Mu\u0308ller zahlt 500,00 \u20AC an Frau Mo\u0308ller"
	nfdText := "Herr M\u0075\u0308ller zahlt 500,00 \u20AC an Frau M\u006F\u0308ller"

	// NFC: "Herr Müller zahlt 500,00 € an Frau Möller"
	// H(0) e(1) r(2) r(3) ' '(4) M(5) ü(6-7) l(8) l(9) e(10) r(11) → "Müller" = [5:12]
	// ' '(12) z(13) a(14) h(15) l(16) t(17) ' '(18) → "500,00 €" starts at 19
	// 5(19) 0(20) 0(21) ,(22) 0(23) 0(24) ' '(25) €(26-28) → "500,00 €" = [19:29]
	// ' '(29) a(30) n(31) ' '(32) F(33) r(34) a(35) u(36) ' '(37) M(38) ö(39-40) l(41) l(42) e(43) r(44) → "Möller" = [38:45]
	entities := []scanner.Entity{
		{Start: 5, End: 12, Type: "PERSON", Text: "M\u00FCller", Score: 0.90, Detector: "regex"},
		{Start: 19, End: 29, Type: "FINANCIAL", Text: "500,00 \u20AC", Score: 0.90, Detector: "regex"},
		{Start: 38, End: 45, Type: "PERSON", Text: "M\u00F6ller", Score: 0.90, Detector: "regex"},
	}

	result := Redact(nfdText, entities)

	want := "Herr [PERSON_1] zahlt [FINANCIAL_1] an Frau [PERSON_2]"
	if result.SanitizedText != want {
		t.Errorf("SanitizedText = %q, want %q", result.SanitizedText, want)
	}
}
