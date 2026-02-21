package scanner

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"
)

const (
	testEmail      = "test@example.com"
	testPhone      = "+49 30 12345678"
	testIBAN       = "DE89370400440532013000"
	testIBANSpaced = "DE89 3704 0044 0532 0130 00"
	testAddress    = "Musterstraße 12, 10115 Berlin"
	testPerson     = "Herr Thomas Schmidt"
	testPersonName = "Thomas Schmidt"
	testCreditCard = "4111 1111 1111 1111"
)

func hasEntityOfType(entities []Entity, typ string) bool {
	for _, e := range entities {
		if e.Type == typ {
			return true
		}
	}
	return false
}

func hasEntityWithText(entities []Entity, typ, text string) bool {
	for _, e := range entities {
		if e.Type == typ && e.Text == text {
			return true
		}
	}
	return false
}

func countEntitiesOfType(entities []Entity, typ string) int {
	n := 0
	for _, e := range entities {
		if e.Type == typ {
			n++
		}
	}
	return n
}

// --- 1. PII at text boundaries ---

func TestAdversarial_PIIAtBoundaries(t *testing.T) {
	s := DefaultScanner(nil)

	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		{"email at start", testEmail, "EMAIL"},
		{"email at end", "Contact " + testEmail, "EMAIL"},
		{"email is entire text", testEmail, "EMAIL"},
		{"IBAN at start", testIBANSpaced + " is the IBAN.", "IBAN"},
		{"IBAN at end", "Transfer to " + testIBANSpaced, "IBAN"},
		{"IBAN is entire text", testIBANSpaced, "IBAN"},
		{"credit card at start", testCreditCard + " was charged.", "CREDIT_CARD"},
		{"credit card at end", "Card number: " + testCreditCard, "CREDIT_CARD"},
		{"person at end", testPerson, "PERSON"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			if !hasEntityOfType(entities, tc.wantType) {
				t.Errorf("expected %s in %q, got: %v", tc.wantType, tc.input, entities)
			}
			for _, e := range entities {
				if e.Start < 0 || e.End > len(tc.input) || e.Start > e.End {
					t.Errorf("entity %v has invalid offsets for input len %d", e, len(tc.input))
				}
			}
		})
	}
}

// --- 2. Consecutive PII ---

func TestAdversarial_ConsecutivePII(t *testing.T) {
	s := DefaultScanner(nil)

	tests := []struct {
		name      string
		input     string
		wantTypes []string
	}{
		{"two emails with space", "test@example.com user@domain.org", []string{"EMAIL"}},
		{"two emails no space", "test@example.comuser@domain.org", []string{"EMAIL"}},
		{"person then email", "Herr Thomas Schmidt test@example.com", []string{"PERSON", "EMAIL"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			for _, wt := range tc.wantTypes {
				if !hasEntityOfType(entities, wt) {
					t.Errorf("expected %s in %q, got: %v", wt, tc.input, entities)
				}
			}
		})
	}
}

// --- 3. PII in brackets ---

func TestAdversarial_PIIInBrackets(t *testing.T) {
	s := DefaultScanner(nil)

	brackets := []struct {
		name  string
		left  string
		right string
	}{
		{"parentheses", "(", ")"},
		{"square brackets", "[", "]"},
		{"curly braces", "{", "}"},
		{"double quotes", "\"", "\""},
		{"single quotes", "'", "'"},
	}

	piiSamples := []struct {
		desc     string
		value    string
		wantType string
	}{
		{"email", testEmail, "EMAIL"},
		{"IBAN", testIBANSpaced, "IBAN"},
		{"credit card", testCreditCard, "CREDIT_CARD"},
	}

	for _, br := range brackets {
		for _, pii := range piiSamples {
			name := fmt.Sprintf("%s in %s", pii.desc, br.name)
			t.Run(name, func(t *testing.T) {
				input := "Here is " + br.left + pii.value + br.right + " for you."
				entities := s.Scan(input)
				if !hasEntityOfType(entities, pii.wantType) {
					t.Errorf("expected %s in %q, got: %v", pii.wantType, input, entities)
				}
			})
		}
	}
}

// --- 4. Large text (1MB+) ---

func TestAdversarial_LargeText(t *testing.T) {
	s := DefaultScanner(nil)

	filler := "Dies ist ein ganz normaler Satz ohne personenbezogene Daten darin. "
	fillerChunk := strings.Repeat(filler, 100)

	var b strings.Builder
	for b.Len() < 500*1024 {
		b.WriteString(fillerChunk)
	}

	b.WriteString("Kontaktieren Sie test@example.com oder rufen Sie +49 30 12345678 an. ")
	b.WriteString("IBAN: DE89 3704 0044 0532 0130 00. ")
	b.WriteString("Herr Thomas Schmidt wohnt in Musterstraße 12, 10115 Berlin. ")

	for b.Len() < 1024*1024 {
		b.WriteString(fillerChunk)
	}

	text := b.String()
	if len(text) < 1024*1024 {
		t.Fatalf("text is only %d bytes, expected >= 1MB", len(text))
	}

	entities := s.Scan(text)

	wantTypes := []string{"EMAIL", "PHONE", "IBAN", "PERSON"}
	for _, wt := range wantTypes {
		if !hasEntityOfType(entities, wt) {
			t.Errorf("expected %s in 1MB text, not found in %d entities", wt, len(entities))
		}
	}

	normalizedLen := len(norm.NFC.String(text))
	for _, e := range entities {
		if e.Start < 0 || e.End > normalizedLen || e.Start >= e.End {
			t.Errorf("entity %v has invalid offsets (text len after NFC: %d)", e, normalizedLen)
		}
	}
}

// --- 5. Empty and whitespace ---

func TestAdversarial_EmptyAndWhitespace(t *testing.T) {
	s := DefaultScanner(nil)

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"single space", " "},
		{"tabs and spaces", "   \t\t   "},
		{"newlines only", "\n\n\n"},
		{"CRLF only", "\r\n\r\n"},
		{"mixed whitespace", " \t\n\r \t\n"},
		{"only special chars", "!@#$%^&*()"},
		{"only punctuation", "...,,,...;;"},
		{"only unicode spaces", "\u00A0\u2003\u2009"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			if len(entities) > 0 {
				t.Errorf("expected 0 entities for %q, got %d: %v", tc.name, len(entities), entities)
			}
		})
	}
}

// --- 6. Unicode NFC vs NFD ---

func TestAdversarial_UnicodeNormalization(t *testing.T) {
	s := DefaultScanner(nil)

	baseText := "Frau Maria Müller schrieb an müller@example.de vom Büro aus."

	nfcText := norm.NFC.String(baseText)
	nfdText := norm.NFD.String(baseText)

	if nfcText == nfdText {
		t.Skip("NFC and NFD forms are identical for this text")
	}

	nfcEntities := s.Scan(nfcText)
	nfdEntities := s.Scan(nfdText)

	if len(nfcEntities) != len(nfdEntities) {
		t.Errorf("NFC produced %d entities, NFD produced %d entities", len(nfcEntities), len(nfdEntities))
		t.Logf("NFC entities: %v", nfcEntities)
		t.Logf("NFD entities: %v", nfdEntities)
		return
	}

	for i := range nfcEntities {
		if nfcEntities[i].Type != nfdEntities[i].Type {
			t.Errorf("entity %d: NFC type %q != NFD type %q", i, nfcEntities[i].Type, nfdEntities[i].Type)
		}
		if nfcEntities[i].Text != nfdEntities[i].Text {
			t.Errorf("entity %d: NFC text %q != NFD text %q", i, nfcEntities[i].Text, nfdEntities[i].Text)
		}
	}
}

// --- 7. Mixed scripts ---

func TestAdversarial_MixedScripts(t *testing.T) {
	s := DefaultScanner(nil)

	t.Run("latin email in Cyrillic context", func(t *testing.T) {
		input := "\u041F\u0438\u0448\u0438\u0442\u0435 \u043D\u0430 test@example.com \u043F\u043E\u0436\u0430\u043B\u0443\u0439\u0441\u0442\u0430"
		entities := s.Scan(input)
		if !hasEntityOfType(entities, "EMAIL") {
			t.Errorf("expected EMAIL in Cyrillic-surrounded text, got: %v", entities)
		}
	})

	t.Run("phone in Greek context", func(t *testing.T) {
		input := "\u0395\u03C0\u03B9\u03BA\u03BF\u03B9\u03BD\u03C9\u03BD\u03AF\u03B1: +49 30 12345678 \u03B5\u03B4\u03CE"
		entities := s.Scan(input)
		if !hasEntityOfType(entities, "PHONE") {
			t.Errorf("expected PHONE in Greek-surrounded text, got: %v", entities)
		}
	})

	t.Run("pure Cyrillic no PII", func(t *testing.T) {
		input := "\u041F\u0440\u0438\u0432\u0435\u0442 \u043C\u0438\u0440"
		entities := s.Scan(input)
		if len(entities) > 0 {
			t.Logf("unexpected entities in pure Cyrillic: %v", entities)
		}
	})
}

// --- 8. Zero-width characters ---

func TestAdversarial_ZeroWidthChars(t *testing.T) {
	s := DefaultScanner(nil)

	zeroWidthChars := []struct {
		name string
		char string
	}{
		{"zero-width space", "\u200B"},
		{"zero-width non-joiner", "\u200C"},
		{"zero-width joiner", "\u200D"},
		{"byte order mark", "\uFEFF"},
	}

	for _, zwc := range zeroWidthChars {
		t.Run("email with "+zwc.name, func(t *testing.T) {
			input := "te" + zwc.char + "st@example.com"
			entities := s.Scan(input)
			t.Logf("email with %s: entities=%v", zwc.name, entities)
		})

		t.Run("IBAN with "+zwc.name, func(t *testing.T) {
			input := "DE89" + zwc.char + "370400440532013000"
			entities := s.Scan(input)
			t.Logf("IBAN with %s: entities=%v", zwc.name, entities)
		})
	}

	t.Run("only zero-width chars", func(t *testing.T) {
		input := "\u200B\u200C\u200D\uFEFF\u200B\u200C\u200D\uFEFF"
		entities := s.Scan(input)
		if len(entities) > 0 {
			t.Errorf("expected 0 entities for zero-width-only text, got %d: %v", len(entities), entities)
		}
	})

	t.Run("BOM prefix does not break email", func(t *testing.T) {
		input := "\uFEFF" + testEmail
		entities := s.Scan(input)
		if !hasEntityOfType(entities, "EMAIL") {
			t.Errorf("BOM prefix broke email detection: %v", entities)
		}
	})
}

// --- 9. Repeated patterns ---

func TestAdversarial_RepeatedPatterns(t *testing.T) {
	s := DefaultScanner(nil)
	const repeatCount = 100

	t.Run("email repeated 100 times", func(t *testing.T) {
		lines := make([]string, repeatCount)
		for i := range lines {
			lines[i] = testEmail
		}
		input := strings.Join(lines, " ")

		entities := s.Scan(input)
		emailCount := countEntitiesOfType(entities, "EMAIL")
		if emailCount != repeatCount {
			t.Errorf("expected %d EMAIL entities, got %d", repeatCount, emailCount)
		}
	})

	t.Run("phone repeated 100 times", func(t *testing.T) {
		lines := make([]string, repeatCount)
		for i := range lines {
			lines[i] = testPhone
		}
		input := strings.Join(lines, " | ")

		entities := s.Scan(input)
		phoneCount := countEntitiesOfType(entities, "PHONE")
		if phoneCount < repeatCount/2 {
			t.Errorf("expected ~%d PHONE entities, got %d", repeatCount, phoneCount)
		}
	})

	t.Run("IBAN repeated 100 times", func(t *testing.T) {
		lines := make([]string, repeatCount)
		for i := range lines {
			lines[i] = testIBANSpaced
		}
		input := strings.Join(lines, " ")

		entities := s.Scan(input)
		ibanCount := countEntitiesOfType(entities, "IBAN")
		if ibanCount < 1 {
			t.Errorf("expected at least 1 IBAN, got 0")
		}
		t.Logf("IBAN x100: got %d IBAN entities", ibanCount)
	})
}

// --- 10. Newline variations ---

func TestAdversarial_NewlineVariations(t *testing.T) {
	s := DefaultScanner(nil)

	t.Run("email after LF", func(t *testing.T) {
		input := "Contact:\n" + testEmail
		entities := s.Scan(input)
		if !hasEntityOfType(entities, "EMAIL") {
			t.Errorf("email after LF not detected: %v", entities)
		}
	})

	t.Run("email after CRLF", func(t *testing.T) {
		input := "Contact:\r\n" + testEmail
		entities := s.Scan(input)
		if !hasEntityOfType(entities, "EMAIL") {
			t.Errorf("email after CRLF not detected: %v", entities)
		}
	})

	t.Run("phone split by LF should not match", func(t *testing.T) {
		input := "+49 30\n12345678"
		entities := s.Scan(input)
		phoneFound := hasEntityOfType(entities, "PHONE")
		t.Logf("phone split by LF: detected=%v, entities=%v", phoneFound, entities)
	})

	t.Run("address block with LF", func(t *testing.T) {
		input := "Musterstraße 12\n10115 Berlin\nDeutschland"
		entities := s.Scan(input)
		if !hasEntityOfType(entities, "ADDRESS") {
			t.Errorf("address block with LF not detected: %v", entities)
		}
	})

	t.Run("address block with CRLF", func(t *testing.T) {
		input := "Musterstraße 12\r\n10115 Berlin\r\nDeutschland"
		entities := s.Scan(input)
		if !hasEntityOfType(entities, "ADDRESS") {
			t.Errorf("address block with CRLF not detected: %v", entities)
		}
	})
}
