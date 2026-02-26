package scanner

import (
	"testing"
)

func TestBusinessID_CustomerNumber(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		// DE
		{"DE Kundennummer", "Kundennummer: K-2026-12345", "K-2026-12345"},
		{"DE Kunden-Nr", "Kunden-Nr. 123456", "123456"},
		// EN
		{"EN Customer ID", "Customer ID: CUST-001", "CUST-001"},
		{"EN Account Number", "Account Number: ACC-98765", "ACC-98765"},
		// FR
		{"FR numéro client", "Numéro client: CL-44021", "CL-44021"},
		// IT
		{"IT numero cliente", "Numero cliente: 789012", "789012"},
		// ES
		{"ES número de cliente", "Número de cliente: ES-2026-001", "ES-2026-001"},
		// NL
		{"NL klantnummer", "Klantnummer: KL-12345", "KL-12345"},
		// PL
		{"PL numer klienta", "Numer klienta: PL-99001", "PL-99001"},
		// HU
		{"HU ügyfélszám", "Ügyfélszám: UGY-2026-001", "UGY-2026-001"},
		// SE
		{"SE kundnummer", "Kundnummer: 556677", "556677"},
		// EL
		{"EL αριθμός πελάτη", "Αριθμός πελάτη: GR-12345", "GR-12345"},
		// RO
		{"RO număr client", "Număr client: RO-5678", "RO-5678"},
		// FI
		{"FI asiakasnumero", "Asiakasnumero: FI-2026-01", "FI-2026-01"},
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

func TestBusinessID_EmployeeNumber(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		// DE
		{"DE Personalnummer", "Personalnummer: PER-2026-001", "PER-2026-001"},
		{"DE Mitarbeiternummer", "Mitarbeiternummer: MA-12345", "MA-12345"},
		{"DE Pers-Nr", "Pers-Nr. 234567", "234567"},
		// EN
		{"EN Employee ID", "Employee ID: EMP-00123", "EMP-00123"},
		{"EN Staff Number", "Staff Number: STF-789", "STF-789"},
		// FR
		{"FR matricule", "Matricule: M-20260001", "M-20260001"},
		// IT
		{"IT matricola", "Matricola: 456789", "456789"},
		// ES
		{"ES número de empleado", "Número de empleado: E-2026-42", "E-2026-42"},
		// NL
		{"NL personeelsnummer", "Personeelsnummer: P-12345", "P-12345"},
		// PL
		{"PL numer pracownika", "Numer pracownika: PR-001", "PR-001"},
		// SE
		{"SE anställningsnummer", "Anställningsnummer: AN-5678", "AN-5678"},
		// HR
		{"HR matični broj", "Matični broj: MB-123456", "MB-123456"},
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

func TestBusinessID_ContractNumber(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		// DE
		{"DE Vertragsnummer", "Vertragsnummer: V-2026-001", "V-2026-001"},
		{"DE Vertrags-Nr", "Vertrags-Nr. VTG-12345", "VTG-12345"},
		// EN
		{"EN Contract Number", "Contract Number: CON-2026-789", "CON-2026-789"},
		{"EN Agreement No", "Agreement No. AGR-001", "AGR-001"},
		// FR
		{"FR numéro de contrat", "Numéro de contrat: CTR-44021", "CTR-44021"},
		// IT
		{"IT numero contratto", "Numero di contratto: IT-2026-001", "IT-2026-001"},
		// NL
		{"NL contractnummer", "Contractnummer: NL-2026-42", "NL-2026-42"},
		// PL
		{"PL numer umowy", "Numer umowy: UM-2026-001", "UM-2026-001"},
		// HU
		{"HU szerződésszám", "Szerződésszám: SZ-12345", "SZ-12345"},
		// SE
		{"SE avtalsnummer", "Avtalsnummer: AVT-2026-01", "AVT-2026-01"},
		// BG
		{"BG номер договор", "Номер на договор: BG-2026-001", "BG-2026-001"},
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

func TestBusinessID_Salary(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"DE Bruttogehalt", "Bruttogehalt: 4.500,00 €", "4.500,00 €"},
		{"DE Stundensatz", "Stundensatz: €25,00", "€25,00"},
		{"DE Nettolohn", "Nettolohn: 3.200,00 €", "3.200,00 €"},
		{"EN Gross salary", "Gross salary: €5,000.00", "€5,000.00"},
		{"EN Hourly rate", "Hourly rate: $45.00", "$45.00"},
		{"FR Salaire brut", "Salaire brut: 3.500,00 €", "3.500,00 €"},
		{"IT Stipendio lordo", "Stipendio lordo: 2.800,00 €", "2.800,00 €"},
		{"ES Salario bruto", "Salario bruto: 3.000,00 €", "3.000,00 €"},
		{"NL Brutoloon", "Brutoloon: 4.200,00 €", "4.200,00 €"},
		{"PL Wynagrodzenie brutto", "Wynagrodzenie brutto: 8500,00 PLN", "8500,00 PLN"},
		{"SE Bruttolön", "Bruttolön: 45000 SEK", "45000 SEK"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}
}

func TestBusinessID_SEPACreditorRef(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name     string
		input    string
		want     string
		wantType string
	}{
		// RF references are caught by the IBAN scanner (which is correct — same format family)
		{"RF basic", "Referenz: RF18539007547034", "RF18539007547034", "IBAN"},
		{"RF with letters", "Payment ref: RF9149UB0750", "RF9149UB0750", "FINANCIAL"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			found := false
			for _, e := range entities {
				if e.Text == tc.want && e.Type == tc.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s not found in %q: wanted %q, got %v", tc.wantType, tc.input, tc.want, entities)
			}
		})
	}
}

func TestBusinessID_TrueNegatives(t *testing.T) {
	s := DefaultScanner(nil)
	cases := []struct {
		name  string
		input string
	}{
		{"bare alphanumeric", "The code K-2026-12345 is internal."},
		{"bare digits no keyword", "Reference 123456 in the system."},
		{"salary without keyword", "He earns 4500 per month."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := s.Scan(tc.input)
			for _, e := range entities {
				if e.Type == "ID_NUMBER" && (e.Text == "K-2026-12345" || e.Text == "123456") {
					t.Errorf("false positive in %q: got %v", tc.input, e)
				}
			}
		})
	}
}
