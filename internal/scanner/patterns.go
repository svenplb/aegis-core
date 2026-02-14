package scanner

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// BuiltinScanners returns all built-in regex-based scanners.
func BuiltinScanners() []Scanner {
	var scanners []Scanner

	// Order matters for overlap: more specific patterns first.
	scanners = append(scanners, secretScanners()...)
	scanners = append(scanners, emailScanners()...)
	scanners = append(scanners, urlScanners()...)
	scanners = append(scanners, ibanScanners()...)
	scanners = append(scanners, creditCardScanners()...)
	scanners = append(scanners, ssnScanners()...)
	scanners = append(scanners, macAddressScanners()...)
	scanners = append(scanners, phoneScanners()...)
	scanners = append(scanners, dateScanners()...)
	scanners = append(scanners, ipScanners()...)
	scanners = append(scanners, medicalScanners()...)
	scanners = append(scanners, ageScanners()...)
	scanners = append(scanners, idNumberScanners()...)
	scanners = append(scanners, orgScanners()...)
	scanners = append(scanners, financialScanners()...)
	scanners = append(scanners, addressScanners()...)
	scanners = append(scanners, personScanners()...)

	return scanners
}

// --- SSN ---

func ssnScanners() []Scanner {
	return []Scanner{
		// US SSN: 123-45-6789
		NewRegexScanner(
			regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			"SSN", 0.95,
			WithValidator(func(s string) bool {
				// Reject 000, 666, 900-999 in area number
				area := s[:3]
				return area != "000" && area != "666" && area[0] != '9'
			}),
		),
		// German Sozialversicherungsnummer (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Sozialversicherungsnummer|SVN|SV-Nummer|Versicherungsnummer)[:\s]+(\d{2}\s?\d{6}\s?[A-Z]\s?\d{3})`),
			"SSN", 0.90,
			WithExtractGroup(1),
		),
		// Swiss AHV: 756.1234.5678.97
		NewRegexScanner(
			regexp.MustCompile(`\b756\.\d{4}\.\d{4}\.\d{2}\b`),
			"SSN", 0.95,
		),
		// UK NINO: AB 12 34 56 C
		NewRegexScanner(
			regexp.MustCompile(`\b[A-CEGHJ-PR-TW-Z][A-CEGHJ-NPR-TW-Z]\s?\d{2}\s?\d{2}\s?\d{2}\s?[A-D]\b`),
			"SSN", 0.90,
		),
		// French INSEE: 1 85 12 75 108 042 36
		NewRegexScanner(
			regexp.MustCompile(`\b[12]\s?\d{2}\s?\d{2}\s?\d{2}\s?\d{3}\s?\d{3}\s?\d{2}\b`),
			"SSN", 0.85,
		),
	}
}

// --- MEDICAL ---

func medicalScanners() []Scanner {
	return []Scanner{
		// ICD-10 codes (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Diagnose|ICD|diagnosis|diagnostic)[:\s]+([A-Z]\d{2}(?:\.\d{1,4})?)`),
			"MEDICAL", 0.90,
			WithExtractGroup(1),
		),
		// Blood pressure: 120/80 mmHg
		NewRegexScanner(
			regexp.MustCompile(`\b\d{2,3}/\d{2,3}\s?(?:mmHg|mm\s?Hg)\b`),
			"MEDICAL", 0.90,
		),
		// Lab values with units
		NewRegexScanner(
			regexp.MustCompile(`\b\d{1,4}(?:[.,]\d{1,2})?\s?(?:mg/dL|mmol/L|g/dL|mL/min|ng/mL|ng/L|µg/L|U/L|IU/L|pg/mL|µmol/L)\b`),
			"MEDICAL", 0.85,
		),
		// BMI values (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:BMI|Body Mass Index)[:\s]+(\d{2}(?:[.,]\d{1,2})?)`),
			"MEDICAL", 0.85,
			WithExtractGroup(1),
		),
		// ICD-10 codes standalone in parentheses: (I21.0), (E11.65)
		NewRegexScanner(
			regexp.MustCompile(`\(([A-Z]\d{2}(?:\.\d{1,4})?)\)`),
			"MEDICAL", 0.85,
			WithExtractGroup(1),
		),
	}
}

// --- AGE ---

func ageScanners() []Scanner {
	return []Scanner{
		// "X years old" / "X-year-old"
		NewRegexScanner(
			regexp.MustCompile(`\b(\d{1,3})\s?(?:-\s?)?(?:years?\s?(?:old)?|year-old)\b`),
			"AGE", 0.85,
			WithExtractGroup(1),
			WithValidator(func(s string) bool {
				n, _ := strconv.Atoi(s)
				return n > 0 && n < 150
			}),
		),
		// "X Jahre alt"
		NewRegexScanner(
			regexp.MustCompile(`\b(\d{1,3})\s?(?:Jahre?\s?(?:alt)?)\b`),
			"AGE", 0.85,
			WithExtractGroup(1),
			WithValidator(func(s string) bool {
				n, _ := strconv.Atoi(s)
				return n > 0 && n < 150
			}),
		),
		// Context-triggered: "age: X" / "Alter: X"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:age|Alter)[:\s]+(\d{1,3})\b`),
			"AGE", 0.80,
			WithExtractGroup(1),
			WithValidator(func(s string) bool {
				n, _ := strconv.Atoi(s)
				return n > 0 && n < 150
			}),
		),
		// Birth year: "born in 1990", "geboren 1985"
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:born\s+(?:in\s+)?|geboren\s+(?:im\s+)?(?:Jahr\s+)?)((?:19|20)\d{2})\b`),
			"AGE", 0.80,
			WithExtractGroup(1),
		),
	}
}

// --- ID_NUMBER ---

func idNumberScanners() []Scanner {
	return []Scanner{
		// German Steuer-ID (context-triggered): 11 digits
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Steuer-?ID|Steueridentifikationsnummer|Tax\s?ID|TIN)[:\s]+(\d{11})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// German Personalausweis (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Personalausweis|Ausweis(?:nummer)?|ID\s?card)[:\s]+([A-Z0-9]{9,10})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// German Reisepass (context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Reisepass|Passport)[:\s]+([A-Z0-9]{9,10})\b`),
			"ID_NUMBER", 0.85,
			WithExtractGroup(1),
		),
		// EU VAT numbers: 2-letter country code + 8-12 alphanumeric (must contain at least one digit)
		NewRegexScanner(
			regexp.MustCompile(`\b(AT|BE|BG|CY|CZ|DE|DK|EE|EL|ES|FI|FR|HR|HU|IE|IT|LT|LU|LV|MT|NL|PL|PT|RO|SE|SI|SK)[A-Z0-9]{8,12}\b`),
			"ID_NUMBER", 0.85,
			WithValidator(func(s string) bool {
				// Must contain at least one digit after country code to avoid matching words like ITALIENISCHES.
				for _, r := range s[2:] {
					if r >= '0' && r <= '9' {
						return true
					}
				}
				return false
			}),
		),
		// German Versichertennummer (insurance number, context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Versichertennummer|Versicherten-?Nr\.?|Versicherungsnr\.?)[:\s]+([A-Z]?\d{6,12})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
		// German Rentenversicherungsnummer (pension number, context-triggered)
		NewRegexScanner(
			regexp.MustCompile(`(?i)(?:Rentenversicherungsnr\.?|Rentenversicherungsnummer|RVNR)[:\s]+(\d{2}\s?\d{6}\s?[A-Z]\s?\d{3})\b`),
			"ID_NUMBER", 0.90,
			WithExtractGroup(1),
		),
	}
}

// --- ORG ---

func orgScanners() []Scanner {
	// Name part for corporate names: allow abbreviations (2-6 uppercase) alongside normal names.
	corpNamePart := `(?:[A-ZÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞ]{2,6}|` + nameComponent + `)(?:-` + nameComponent + `)?`

	// Use [ \t]+ instead of \s+ to prevent matching across newlines.
	sp := `[ \t]+`

	// Corporate suffixes (German)
	corpDE := corpNamePart + `(?:` + sp + corpNamePart + `)*` + sp + `(?:GmbH|AG|SE|KG|OHG|KGaA|UG|e\.G\.|e\.V\.)\b`
	// Corporate suffixes (International)
	corpIntl := corpNamePart + `(?:` + sp + corpNamePart + `)*` + sp + `(?:Ltd|Inc|Corp|LLC|PLC|SA|SAS|SARL|SpA|SRL|BV|NV)\.?\b`
	// German institutions: Universitätsklinikum/Uniklinik/Universität/Klinikum + Name
	deInstitution := `(?:Universitätsklinikum|Uniklinik|Universität|Klinikum)` + sp + namePattern + `(?:` + sp + namePattern + `)*`
	// Klinik + preposition + Name
	klinikPrep := `Klinik` + sp + `(?:am|für|an` + sp + `der|im)` + sp + namePattern + `(?:` + sp + namePattern + `)*`
	// French hospitals
	frHospital := `(?:Hôpital|CHU)` + sp + namePattern + `(?:[ \t\-]+` + namePattern + `)*`
	// Italian hospitals
	itHospital := `(?:Ospedale|Policlinico)` + sp + namePattern + `(?:` + sp + namePattern + `)*`
	// Spanish hospitals
	esHospital := `Hospital` + sp + namePattern + `(?:` + sp + namePattern + `)*`
	// German insurance: AOK + Name
	aok := `AOK` + sp + namePattern
	// German government: Deutsche Rentenversicherung (+ optional Name)
	drv := `Deutsche` + sp + `Rentenversicherung(?:` + sp + namePattern + `)?`
	// University medical centers: Name UMC | UMC Name
	umcSuffix := namePattern + sp + `UMC`
	umcPrefix := `UMC` + sp + namePattern

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(corpDE), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(corpIntl), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(deInstitution), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(klinikPrep), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(frHospital), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(itHospital), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(esHospital), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(aok), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(drv), "ORG", 0.90),
		NewRegexScanner(regexp.MustCompile(umcSuffix), "ORG", 0.85),
		NewRegexScanner(regexp.MustCompile(umcPrefix), "ORG", 0.85),
	}
}

// --- MAC_ADDRESS ---

func macAddressScanners() []Scanner {
	return []Scanner{
		// Standard: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX
		NewRegexScanner(
			regexp.MustCompile(`\b([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}\b`),
			"MAC_ADDRESS", 0.95,
		),
		// Cisco format: XXXX.XXXX.XXXX
		NewRegexScanner(
			regexp.MustCompile(`\b[0-9A-Fa-f]{4}\.[0-9A-Fa-f]{4}\.[0-9A-Fa-f]{4}\b`),
			"MAC_ADDRESS", 0.90,
		),
	}
}

// --- PERSON ---

// Unicode-aware name component: uppercase letter followed by lowercase letters,
// including diacritics (Müller, Ñoño, Ólafsson, etc.).
// Hyphenated names supported: Jean-Pierre, Müller-Schmidt.
const nameComponent = `[A-ZÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞ][a-zàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþß]+`
const namePattern = nameComponent + `(?:-` + nameComponent + `)?`

// Name particles for multi-part surnames (de Groot, van der Berg, von Stein, etc.)
const nameParticle = `(?:de|van|der|von|di|del|della|le|la|da|dos|das|du|ten|ter|het)`

// Full name: 2-4 name components with optional particles between them.
const fullName = namePattern + `(?:[ \t]+(?:` + nameParticle + `[ \t]+)*` + namePattern + `){1,3}`

func personScanners() []Scanner {
	// Context-triggered: keyword + CapFirst CapLast
	// Longer/more specific patterns first to avoid partial matches.
	triggers := []string{
		// Multi-word triggers first
		`Dr\.\s?med\.`, `de\s+heer`,
		`mein Freund`, `meine Freundin`,
		`meinen Patienten`, `meiner Patientin`,
		`my friend`, `my colleague`, `my patient`,
		`mon ami`, `mon amie`,
		// German role triggers
		`Antragsteller(?:in)?`, `Sachbearbeiter(?:in)?`, `Bearbeiter(?:in)?`,
		`Konsiliarius`,
		// Titles (more specific first)
		`Dott\.?\s?ssa`, `Dott\.?`, `Dra\.?`,
		`Prof\.?`, `Dr\.?`,
		// German
		`Herr`, `Frau`, `Patient(?:in)?`, `Kollege`, `Kollegin`,
		// French
		`Monsieur`, `Madame`, `Mademoiselle`,
		// English
		`Mr\.?`, `Mrs\.?`, `Ms\.?`, `colleague`,
		// Dutch
		`Meneer`, `Mevrouw`,
		// Italian
		`Signor(?:a)?`,
		// Spanish
		`Señor(?:a)?`,
	}

	triggerGroup := `(?:` + strings.Join(triggers, `|`) + `)`
	// Use (?i:...) only for the trigger group, keep name pattern case-sensitive.
	// Allow optional colon/comma between trigger and name (e.g. "Antragsteller: Thomas Schmidt").
	contextPattern := `(?i:` + triggerGroup + `)[: \t]+(` + fullName + `)`

	// Verb-triggered: "told/asked/called/emailed Name Name"
	verbs := `(?i:told|asked|called|emailed|contacted|met|visited|informed)`
	verbPattern := verbs + `[ \t]+(` + fullName + `)`

	// Maiden name: "geb. Müller", "geboren Weber"
	maidenPattern := `(?i:geb(?:oren(?:e)?)?\.)[ \t]+(` + namePattern + `)`

	return []Scanner{
		NewRegexScanner(
			regexp.MustCompile(contextPattern),
			"PERSON", 0.95,
			WithExtractGroup(1),
		),
		NewRegexScanner(
			regexp.MustCompile(verbPattern),
			"PERSON", 0.85,
			WithExtractGroup(1),
		),
		NewRegexScanner(
			regexp.MustCompile(maidenPattern),
			"PERSON", 0.85,
			WithExtractGroup(1),
		),
	}
}

// --- EMAIL ---

func emailScanners() []Scanner {
	// RFC 5322 simplified with unicode support for DACH region.
	pattern := `[a-zA-Z0-9._%+\-àáâãäåæçèéêëìíîïðñòóôõöøùúûüýþß]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`
	return []Scanner{
		NewRegexScanner(regexp.MustCompile(pattern), "EMAIL", 0.99),
	}
}

// --- PHONE ---

// ibanPrefixRe matches the leading portion of an IBAN (country code + check digits +
// space-separated groups) that may appear before a phone-like digit sequence.
var ibanPrefixRe = regexp.MustCompile(`[A-Z]{2}\d{2}(?:[\s\-][\dA-Z]{4})*[\s\-]?[\dA-Z]{0,4}$`)

// phoneNotInIBAN rejects a phone match if it sits inside an IBAN-like structure.
// It looks back up to 40 bytes from the match start for an IBAN prefix.
func phoneNotInIBAN(fullText string, start, end int) bool {
	lookback := 40
	from := start - lookback
	if from < 0 {
		from = 0
	}
	prefix := fullText[from:start]
	if ibanPrefixRe.MatchString(prefix) {
		return false
	}
	return true
}

func phoneScanners() []Scanner {
	// International format with + prefix: +49, +43, +41, +33, etc.
	// Supports separators: space, dash, dot, or none.
	// Use [ \t] instead of \s to prevent matching across newlines.
	intl := `\+(?:49|43|41|33|39|34|31|32|351|48|46|358|45|47|353|44)[ \t]?[\d][\d \t.\-]{6,14}\d`

	// Generic 00-prefix international
	generic00 := `00\d{1,3}[ \t.\-]?\d[\d \t.\-]{6,14}\d`

	// German local: 0XXX XXXXXXX
	deLocal := `0[1-9]\d{1,4}[ \t.\-/]?\d[\d \t.\-]{4,10}\d`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(intl), "PHONE", 0.95, WithContextValidator(phoneNotInIBAN)),
		NewRegexScanner(regexp.MustCompile(generic00), "PHONE", 0.90, WithContextValidator(phoneNotInIBAN)),
		NewRegexScanner(regexp.MustCompile(deLocal), "PHONE", 0.85, WithContextValidator(phoneNotInIBAN)),
	}
}

// --- IBAN ---

func ibanScanners() []Scanner {
	// Generic IBAN: 2 letters + 2 digits + 8-30 alphanumeric (with optional spaces/dashes).
	pattern := `\b[A-Z]{2}\d{2}[\s\-]?[\dA-Z]{4}[\s\-]?[\dA-Z]{4}(?:[\s\-]?[\dA-Z]{4}){1,7}(?:[\s\-]?[\dA-Z]{1,4})?\b`

	return []Scanner{
		NewRegexScanner(
			regexp.MustCompile(pattern),
			"IBAN", 0.99,
			WithValidator(validateIBAN),
		),
	}
}

// validateIBAN performs MOD-97 checksum validation.
func validateIBAN(s string) bool {
	// Remove spaces and dashes.
	clean := strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' {
			return -1
		}
		return r
	}, s)

	if len(clean) < 5 || len(clean) > 34 {
		return false
	}

	// Check format: 2 letters + 2 digits + rest alphanumeric.
	for i, r := range clean {
		if i < 2 {
			if !unicode.IsUpper(r) {
				return false
			}
		} else if i < 4 {
			if !unicode.IsDigit(r) {
				return false
			}
		} else {
			if !unicode.IsUpper(r) && !unicode.IsDigit(r) {
				return false
			}
		}
	}

	// Move first 4 chars to end.
	rearranged := clean[4:] + clean[:4]

	// Convert letters to numbers: A=10, B=11, ..., Z=35.
	var numStr strings.Builder
	for _, r := range rearranged {
		if unicode.IsUpper(r) {
			numStr.WriteString(big.NewInt(int64(r - 'A' + 10)).String())
		} else {
			numStr.WriteRune(r)
		}
	}

	n := new(big.Int)
	n.SetString(numStr.String(), 10)
	mod := new(big.Int)
	mod.Mod(n, big.NewInt(97))

	return mod.Int64() == 1
}

// --- CREDIT CARD ---

func creditCardScanners() []Scanner {
	// Visa (16 digits): 4xxx xxxx xxxx xxxx
	// Mastercard (16 digits): 5[1-5]xx or 2[2-7]xx
	visa := `\b4\d{3}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`
	mc := `\b(?:5[1-5]\d{2}|2[2-7]\d{2})[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`

	// Amex (15 digits): 3[47]xx xxxxxx xxxxx
	amex := `\b3[47]\d{2}[\s\-]?\d{6}[\s\-]?\d{5}\b`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(visa), "CREDIT_CARD", 0.95, WithValidator(validateLuhn)),
		NewRegexScanner(regexp.MustCompile(mc), "CREDIT_CARD", 0.95, WithValidator(validateLuhn)),
		NewRegexScanner(regexp.MustCompile(amex), "CREDIT_CARD", 0.95, WithValidator(validateLuhn)),
	}
}

// validateLuhn performs the Luhn algorithm check.
func validateLuhn(s string) bool {
	// Extract digits only.
	var digits []int
	for _, r := range s {
		if unicode.IsDigit(r) {
			digits = append(digits, int(r-'0'))
		}
	}

	if len(digits) < 13 || len(digits) > 19 {
		return false
	}

	sum := 0
	alt := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}

	return sum%10 == 0
}

// --- DATE ---

func dateScanners() []Scanner {
	// DD.MM.YYYY, DD/MM/YYYY, DD-MM-YYYY
	dateCore := `\b(?:0[1-9]|[12]\d|3[01])[./\-](?:0[1-9]|1[0-2])[./\-](?:19|20)\d{2}\b`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(dateCore), "DATE", 0.90),
	}
}

// --- URL ---

func urlScanners() []Scanner {
	pattern := `https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`
	return []Scanner{
		NewRegexScanner(regexp.MustCompile(pattern), "URL", 0.95),
	}
}

// --- IP_ADDRESS ---

func ipScanners() []Scanner {
	// IPv4
	ipv4 := `\b(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\b`

	// IPv6 (simplified: full form and common abbreviations)
	ipv6 := `(?:` +
		`(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}` + // full
		`|(?:[0-9a-fA-F]{1,4}:){1,7}:` + // trailing ::
		`|(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}` + // :: with one group
		`|::1` + // loopback
		`|::` + // unspecified
		`)`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(ipv4), "IP_ADDRESS", 0.90, WithValidator(validateIPv4)),
		NewRegexScanner(regexp.MustCompile(ipv6), "IP_ADDRESS", 0.90),
	}
}

func validateIPv4(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if len(p) > 1 && p[0] == '0' {
			return false // reject leading zeros like 01.02.03.04
		}
	}
	return true
}

// --- FINANCIAL ---

func financialScanners() []Scanner {
	// EUR format: €1.500,00 or 1.500,00 € or 1.500,00€
	eurPrefix := `€\s?\d{1,3}(?:\.\d{3})*,\d{2}`
	eurSuffix := `\d{1,3}(?:\.\d{3})*,\d{2}\s?€`

	// USD/GBP format: $2,500.00 or £2,500.00
	usdGbp := `[$£]\s?\d{1,3}(?:,\d{3})*\.\d{2}`

	// CHF: CHF 1'500.00 or CHF 1500.00
	chf := `CHF\s?\d{1,3}(?:['\x{2019}]\d{3})*\.\d{2}`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(eurPrefix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(eurSuffix), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(usdGbp), "FINANCIAL", 0.90),
		NewRegexScanner(regexp.MustCompile(chf), "FINANCIAL", 0.90),
	}
}

// --- ADDRESS ---

func addressScanners() []Scanner {
	// Use [ \t] instead of \s to prevent matching across newlines.

	// German: Straße/Str./Weg/Platz/Allee/Gasse + Hausnummer (suffix form: Gartenstraße 27)
	deStreetSuffix := `(?:[A-ZÄÖÜ][a-zäöüß]+(?:straße|str\.|weg|platz|allee|gasse|ring|damm|ufer|kai))[ \t]+\d{1,4}[a-zA-Z]?`

	// German: separate-word street name (Berliner Straße 15)
	deStreetSep := namePattern + `(?:[ \t]+` + namePattern + `)?[ \t]+(?:Straße|Str\.|Weg|Platz|Allee|Gasse|Ring|Damm|Ufer|Kai)[ \t]+\d{1,4}[a-zA-Z]?`

	// German: hyphenated street names ending in suffix (Theodor-Stern-Kai 7)
	deStreetHyphen := `(?:[A-ZÄÖÜ][a-zäöüß]+-)+(?:Straße|Str|Weg|Platz|Allee|Gasse|Ring|Damm|Ufer|Kai)[ \t]+\d{1,4}[a-zA-Z]?`

	// City pattern: "Frankfurt", "Bad Homburg", "Frankfurt am Main"
	cityWord := `[A-ZÄÖÜ][a-zäöüß]+`
	cityPattern := cityWord + `(?:[ \t]+` + cityWord + `|[ \t]+[a-z]+[ \t]+` + cityWord + `)?`

	// German with postcode + city (use [ \t] to prevent cross-line matching)
	deWithCitySuffix := deStreetSuffix + `(?:,[ \t]*\d{5}[ \t]+` + cityPattern + `)?`
	deWithCitySep := deStreetSep + `(?:,[ \t]*\d{5}[ \t]+` + cityPattern + `)?`
	deWithCityHyphen := deStreetHyphen + `(?:,[ \t]*\d{5}[ \t]+` + cityPattern + `)?`

	// French: rue/avenue/boulevard + number
	frStreet := `\d{1,4},?[ \t]+(?:rue|avenue|boulevard|place|chemin|impasse)[ \t]+(?:de[ \t]+(?:la[ \t]+)?|du[ \t]+|des[ \t]+|l')?[A-ZÀ-Ü][a-zà-ÿ]+(?:[ \t]+[A-ZÀ-Ü][a-zà-ÿ]+)*`

	// Italian: via/piazza/corso + name + number (with articles: del, della, etc.)
	itStreet := `(?:[Vv]ia|[Pp]iazza|[Cc]orso|[Vv]iale)[ \t]+(?:(?:del|della|dello|dei|degli|delle|di)[ \t]+)?[A-ZÀ-Ü][a-zà-ÿ]+(?:[ \t]+[A-ZÀ-Ü][a-zà-ÿ]+)*[ \t]+\d{1,4}`

	// Spanish: calle/avenida/plaza/paseo
	esStreet := `(?:[Cc]alle|[Aa]venida|[Pp]laza|[Pp]aseo)[ \t]+(?:de[ \t]+(?:la[ \t]+)?|del[ \t]+)?[A-ZÀ-Ü][a-zà-ÿ]+(?:[ \t]+[A-ZÀ-Ü][a-zà-ÿ]+)*[ \t]+\d{1,4}`

	// Dutch: straat/laan/weg/plein/gracht/dreef + number
	nlStreet := `[A-ZÄÖÜ][a-zäöüß]+(?:straat|laan|weg|plein|gracht|kade|singel|dreef)[ \t]+\d{1,4}`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(deWithCitySuffix), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(deWithCitySep), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(deWithCityHyphen), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(frStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(itStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(esStreet), "ADDRESS", 0.85),
		NewRegexScanner(regexp.MustCompile(nlStreet), "ADDRESS", 0.85),
	}
}

// --- SECRET ---

func secretScanners() []Scanner {
	patterns := []struct {
		pattern string
		score   float64
	}{
		// OpenAI
		{`sk-proj-[A-Za-z0-9_\-]{20,}`, 0.99},
		{`sk-[A-Za-z0-9]{20,}`, 0.99},
		// Anthropic
		{`sk-ant-[A-Za-z0-9_\-]{20,}`, 0.99},
		// AWS access key
		{`AKIA[0-9A-Z]{16}`, 0.99},
		// GitHub
		{`gh[patos]_[A-Za-z0-9]{30,}`, 0.99},
		// Slack
		{`xox[bp]-[0-9]{10,}-[A-Za-z0-9\-]+`, 0.99},
		// Bearer token
		{`Bearer\s+[A-Za-z0-9._~+/=\-]{20,}`, 0.95},
		// PEM private key (just the header line)
		{`-----BEGIN (?:RSA |EC |DSA )?PRIVATE KEY-----`, 0.99},
	}

	scanners := make([]Scanner, 0, len(patterns))
	for _, p := range patterns {
		scanners = append(scanners, NewRegexScanner(
			regexp.MustCompile(p.pattern), "SECRET", p.score,
		))
	}
	return scanners
}
