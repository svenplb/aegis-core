package scanner

import (
	"regexp"
	"testing"
)

// --- PERSON tests ---

func TestPerson_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Herr Thomas Schmidt ist hier.", "Thomas Schmidt"},
		{"Frau Maria Müller kommt morgen.", "Maria Müller"},
		{"Patient Jean-Pierre Dubois wird behandelt.", "Jean-Pierre Dubois"},
		{"Mr. John Smith arrived.", "John Smith"},
		{"Mevrouw Anna Bakker belde.", "Anna Bakker"},
		{"Monsieur Pierre Dupont est arrivé.", "Pierre Dupont"},
		{"Madame Sophie Laurent est là.", "Sophie Laurent"},
		{"Mevrouw Anna Bakker woont hier.", "Anna Bakker"},
		{"my colleague Thomas Weber said hello.", "Thomas Weber"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "PERSON" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("PERSON not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestPerson_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The weather is nice today.",
		"Berlin is a city in Germany.",
		"I went to the store yesterday.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "PERSON" {
				t.Errorf("PERSON false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- EMAIL tests ---

func TestEmail_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Contact thomas.schmidt@example.com for info.", "thomas.schmidt@example.com"},
		{"Send to müller@example.de please.", "müller@example.de"},
		{"Email anna.devries@voorbeeld.nl now.", "anna.devries@voorbeeld.nl"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "EMAIL" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("EMAIL not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestEmail_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"This is not an email address.",
		"Use the @ symbol for emails.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "EMAIL" {
				t.Errorf("EMAIL false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- PHONE tests ---

func TestPhone_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Call +49 170 4839201 now.", "+49 170 4839201"},
		{"Ring +44 20 7946 0958 please.", "+44 20 7946 0958"},
		{"Appelez le +33 6 12 34 56 78.", "+33 6 12 34 56 78"},
		{"Chiami +39 06 1234 5678.", "+39 06 1234 5678"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "PHONE" {
				// Allow partial match (phone patterns may vary in boundaries).
				if e.Text == tc.want || containsDigits(e.Text, tc.want) {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("PHONE not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func containsDigits(a, b string) bool {
	da := digitsOnly(a)
	db := digitsOnly(b)
	return da == db
}

func digitsOnly(s string) string {
	var out []byte
	for _, r := range s {
		if r >= '0' && r <= '9' {
			out = append(out, byte(r))
		}
	}
	return string(out)
}

func TestPhone_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The version is 2.0.1 and the build number is 12345.",
		"The year 2024 was great.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "PHONE" {
				t.Errorf("PHONE false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- IBAN tests ---

func TestIBAN_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"IBAN: DE89 3704 0044 0532 0130 00", "DE89 3704 0044 0532 0130 00"},
		{"Transfer to GB29 NWBK 6016 1331 9268 19 please.", "GB29 NWBK 6016 1331 9268 19"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "IBAN" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("IBAN not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestIBAN_ChecksumValidation(t *testing.T) {
	// Valid IBAN
	if !validateIBAN("DE89370400440532013000") {
		t.Error("valid IBAN DE89370400440532013000 rejected")
	}
	if !validateIBAN("GB29 NWBK 6016 1331 9268 19") {
		t.Error("valid IBAN GB29 NWBK 6016 1331 9268 19 rejected")
	}

	// Invalid checksum
	if validateIBAN("DE00370400440532013000") {
		t.Error("invalid IBAN DE00370400440532013000 accepted")
	}
	if validateIBAN("XX123456") {
		t.Error("invalid IBAN XX123456 accepted")
	}
}

func TestIBAN_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The code is ABCD1234.",
		"Product ID: XX99 1234 5678 9999",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "IBAN" {
				t.Errorf("IBAN false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- CREDIT_CARD tests ---

func TestCreditCard_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
	}{
		{"Card: 4111 1111 1111 1111"},      // Visa
		{"Card: 5500 0000 0000 0004"},      // Mastercard
		{"Card: 3782 822463 10005"},        // Amex
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "CREDIT_CARD" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CREDIT_CARD not found in %q, got %v", tc.input, entities)
		}
	}
}

func TestCreditCard_LuhnValidation(t *testing.T) {
	if !validateLuhn("4111111111111111") {
		t.Error("valid Visa card rejected")
	}
	if !validateLuhn("5500 0000 0000 0004") {
		t.Error("valid Mastercard rejected")
	}
	if validateLuhn("4111111111111112") {
		t.Error("invalid card accepted")
	}
}

func TestCreditCard_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"Order #1234567890123456",
		"Tracking: 9999888877776666",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "CREDIT_CARD" {
				t.Errorf("CREDIT_CARD false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- DATE tests ---

func TestDate_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"geboren am 15.03.1990 in Berlin.", "15.03.1990"},
		{"Date: 14/03/2024 confirmed.", "14/03/2024"},
		{"Termin am 01-12-2023.", "01-12-2023"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "DATE" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DATE not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestDate_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"Version 2.0.1 released.",
		"The ratio is 3/4 of the total.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "DATE" {
				t.Errorf("DATE false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- URL tests ---

func TestURL_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Visit https://example.com for info.", "https://example.com"},
		{"Link: http://test.org/path?q=1 here.", "http://test.org/path?q=1"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "URL" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("URL not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestURL_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"Visit example.com for info.", // no scheme
		"ftp is old protocol.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "URL" {
				t.Errorf("URL false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- IP_ADDRESS tests ---

func TestIP_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Server at 192.168.1.1 is down.", "192.168.1.1"},
		{"Localhost is ::1 always.", "::1"},
		{"IP: 10.0.0.1 assigned.", "10.0.0.1"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "IP_ADDRESS" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("IP_ADDRESS not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestIP_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The value is 999.999.999.999 which is invalid.",
		"Version 1.2.3 is stable.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "IP_ADDRESS" {
				t.Errorf("IP_ADDRESS false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- FINANCIAL tests ---

func TestFinancial_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
		want  string
	}{
		{"Rechnung über €1.250,00 gesendet.", "€1.250,00"},
		{"Total: $2,500.00 charged.", "$2,500.00"},
		{"Amount: £1,000.00 paid.", "£1,000.00"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "FINANCIAL" && e.Text == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FINANCIAL not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
		}
	}
}

func TestFinancial_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The temperature is 25 degrees.",
		"Page 42 of the report.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "FINANCIAL" {
				t.Errorf("FINANCIAL false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- ADDRESS tests ---

func TestAddress_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
	}{
		{"Wohnt in Hauptstraße 42, 10115 Berlin."},
		{"Via Roma 42, 00184 Roma."},
		{"Keizersgracht 100 in Amsterdam."},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "ADDRESS" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ADDRESS not found in %q, got %v", tc.input, entities)
		}
	}
}

func TestAddress_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The weather is sunny today.",
		"We discussed the project plan.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "ADDRESS" {
				t.Errorf("ADDRESS false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- SECRET tests ---

func TestSecret_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		input string
	}{
		{"Key: sk-proj-abc123def456ghi789jkl012mno345pqr678"},
		{"AWS: AKIA1234567890ABCDEF"},
		{"Token: ghp_abcdefghijklmnopqrstuvwxyz0123456789"},
		{"API: sk-ant-abcdefghij1234567890abcd"},
	}
	for _, tc := range cases {
		entities := s.Scan(tc.input)
		found := false
		for _, e := range entities {
			if e.Type == "SECRET" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SECRET not found in %q, got %v", tc.input, entities)
		}
	}
}

func TestSecret_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []string{
		"The sky is blue and the grass is green.",
		"Use ssh-keygen to generate keys.",
	}
	for _, input := range cases {
		entities := s.Scan(input)
		for _, e := range entities {
			if e.Type == "SECRET" {
				t.Errorf("SECRET false positive in %q: got %v", input, e)
			}
		}
	}
}

// --- Composite scanner tests ---

func TestOverlapDedup(t *testing.T) {
	// Two scanners that both match the same region with different lengths.
	short := NewRegexScanner(regexp.MustCompile(`test`), "X", 0.9)
	long := NewRegexScanner(regexp.MustCompile(`test case`), "Y", 0.9)

	cs := NewCompositeScanner([]Scanner{short, long}, nil)
	entities := cs.Scan("this is a test case here")

	// Should keep only the longer match "test case".
	if len(entities) != 1 {
		t.Fatalf("expected 1 entity after dedup, got %d: %v", len(entities), entities)
	}
	if entities[0].Text != "test case" {
		t.Errorf("expected longer match 'test case', got %q", entities[0].Text)
	}
}

func TestAllowlistFiltering(t *testing.T) {
	emailScanner := NewRegexScanner(
		regexp.MustCompile(`[a-z]+@[a-z]+\.[a-z]+`),
		"EMAIL", 0.99,
	)
	allowlist := []*regexp.Regexp{
		regexp.MustCompile(`example\.com`),
	}

	cs := NewCompositeScanner([]Scanner{emailScanner}, allowlist)
	entities := cs.Scan("Send to test@example.com and real@secret.org")

	// test@example.com should be filtered out by allowlist.
	for _, e := range entities {
		if e.Text == "test@example.com" {
			t.Error("allowlisted email was not filtered out")
		}
	}

	found := false
	for _, e := range entities {
		if e.Text == "real@secret.org" {
			found = true
		}
	}
	if !found {
		t.Error("non-allowlisted email was incorrectly filtered")
	}
}

func TestEmptyText(t *testing.T) {
	s := DefaultScanner(nil)
	entities := s.Scan("")
	if len(entities) != 0 {
		t.Errorf("expected 0 entities for empty text, got %d", len(entities))
	}
}

func TestUTF8MultibyteOffsets(t *testing.T) {
	s := DefaultScanner(nil)
	input := "Herr Müller schrieb an müller@example.com"
	entities := s.Scan(input)

	for _, e := range entities {
		// Verify that byte offsets produce the correct substring.
		extracted := input[e.Start:e.End]
		if extracted != e.Text {
			t.Errorf("offset mismatch for %q: input[%d:%d] = %q, but Text = %q",
				input, e.Start, e.End, extracted, e.Text)
		}
	}
}

// --- Benchmark ---

func BenchmarkScan100KB(b *testing.B) {
	// Build a ~100KB text with some PII scattered.
	base := "Der Patient Thomas Schmidt, geboren am 15.03.1990, wohnt in Hauptstraße 42, 10115 Berlin. " +
		"Seine E-Mail ist thomas.schmidt@example.com und seine Telefonnummer ist +49 170 4839201. " +
		"Die Rechnung über €1.250,00 wurde per IBAN DE89 3704 0044 0532 0130 00 bezahlt. " +
		"Das Wetter ist heute schön und die Temperatur beträgt 25 Grad. " +
		"Wir erwarten moderate Winde aus dem Nordwesten mit 15 km/h. "
	var sb []byte
	for len(sb) < 100*1024 {
		sb = append(sb, base...)
	}
	text := string(sb)

	s := DefaultScanner(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Scan(text)
	}
}
