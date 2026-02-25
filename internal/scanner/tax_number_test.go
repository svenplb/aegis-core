package scanner

import (
	"testing"
)

func TestTaxNumber_TruePositives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		// DE
		{"DE Steuernummer", "Steuernummer: 143/262/10560", "143/262/10560"},
		{"DE Steuer-Nr", "Steuer-Nr. 21/815/08150", "21/815/08150"},
		// AT
		{"AT Steuernummer", "Steuernummer: 12-345/6789", "12-345/6789"},
		{"AT Abgabenkontonr", "Abgabenkontonr. 123456789", "123456789"},
		// FR
		{"FR numéro fiscal", "numéro fiscal: 1234567890123", "1234567890123"},
		{"FR SPI", "SPI: 9876543210123", "9876543210123"},
		// IT
		{"IT Partita IVA", "Partita IVA: 12345678901", "12345678901"},
		{"IT P.IVA", "P.IVA: 12345678901", "12345678901"},
		// ES
		{"ES NIF", "NIF: A1234567B", "A1234567B"},
		{"ES CIF", "CIF: B12345670", "B12345670"},
		// PL
		{"PL NIP dashes", "NIP: 123-456-78-90", "123-456-78-90"},
		{"PL NIP plain", "NIP: 1234567890", "1234567890"},
		// HU
		{"HU adószám", "adószám: 12345678-1-42", "12345678-1-42"},
		{"HU adóazonosító", "adóazonosító jel: 12345678142", "12345678142"},
		// BE
		{"BE ondernemingsnummer", "ondernemingsnummer: 1234.567.890", "1234.567.890"},
		{"BE KBO", "KBO: 1234567890", "1234567890"},
		// SK
		{"SK DIČ", "DIČ: 1234567890", "1234567890"},
		{"SK IČ DPH", "IČ DPH: 9876543210", "9876543210"},
		// SI
		{"SI davčna številka", "davčna številka: 12345678", "12345678"},
		// SE
		{"SE organisationsnummer", "organisationsnummer: 556677-8901", "556677-8901"},
		{"SE org.nr.", "org.nr. 5566778901", "5566778901"},
		// DK
		{"DK CVR", "CVR: 12345678", "12345678"},
		{"DK SE-nummer", "SE-nummer: 87654321", "87654321"},
		// FI
		{"FI Y-tunnus", "Y-tunnus: 1234567-8", "1234567-8"},
		{"FI FO-nummer", "FO-nummer: 12345678", "12345678"},
		// NO
		{"NO organisasjonsnummer", "organisasjonsnummer: 123456789", "123456789"},
		// RO
		{"RO CUI", "CUI: 12345678", "12345678"},
		{"RO cod fiscal", "cod fiscal: 1234567890", "1234567890"},
		// BG
		{"BG BULSTAT", "BULSTAT: 123456789", "123456789"},
		{"BG ЕИК", "ЕИК: 1234567890123", "1234567890123"},
		// GR
		{"GR ΑΦΜ", "ΑΦΜ: 123456789", "123456789"},
		{"GR AFM", "AFM: 987654321", "987654321"},
		// LU
		{"LU matricule", "matricule national: 12345678901", "12345678901"},
		// CY
		{"CY TIC", "TIC: 12345678A", "12345678A"},
		// MT
		{"MT TIN", "Malta TIN: 1234567", "1234567"},
		// EE
		{"EE registrikood", "registrikood: 12345678", "12345678"},
		{"EE KMKR", "KMKR: 87654321", "87654321"},
		// LV
		{"LV PVN", "PVN: 12345678901", "12345678901"},
		// LT
		{"LT PVM", "PVM: 1234567890", "1234567890"},
		{"LT įmonės kodas", "įmonės kodas: 1234567", "1234567"},
		// CH
		{"CH UID", "UID: CHE-123.456.789", "CHE-123.456.789"},
		{"CH Unternehmens-ID", "Unternehmens-ID: CHE123456789", "CHE123456789"},
		// GB
		{"GB UTR", "UTR: 1234567890", "1234567890"},
		{"GB tax reference", "tax reference: 9876543210", "9876543210"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Type == "ID_NUMBER" && e.Text == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ID_NUMBER not found in %q: wanted %q, got %v", tc.input, tc.want, entities)
			}
		})
	}
}

func TestTaxNumber_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
	}{
		{"bare digits no keyword", "The reference is 143/262/10560 in the document."},
		{"bare 11 digits", "Account 12345678901 is active."},
		{"bare 9 digits", "Code 123456789 was entered."},
		{"fraction not tax", "The ratio is 21/815/08150 approx."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			for _, e := range entities {
				if e.Type == "ID_NUMBER" && (e.Text == "143/262/10560" || e.Text == "12345678901" || e.Text == "123456789" || e.Text == "21/815/08150") {
					t.Errorf("ID_NUMBER false positive in %q: got %v", tc.input, e)
				}
			}
		})
	}
}
