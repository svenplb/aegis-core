package scanner

import (
	"math/big"
	"regexp"
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
	scanners = append(scanners, phoneScanners()...)
	scanners = append(scanners, dateScanners()...)
	scanners = append(scanners, ipScanners()...)
	scanners = append(scanners, financialScanners()...)
	scanners = append(scanners, addressScanners()...)
	scanners = append(scanners, personScanners()...)

	return scanners
}

// --- PERSON ---

// Unicode-aware name component: uppercase letter followed by lowercase letters,
// including diacritics (Müller, Ñoño, Ólafsson, etc.).
// Hyphenated names supported: Jean-Pierre, Müller-Schmidt.
const nameComponent = `[A-ZÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞ][a-zàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþß]+`
const namePattern = nameComponent + `(?:-` + nameComponent + `)?`
const fullName = namePattern + `\s+` + namePattern

func personScanners() []Scanner {
	// Context-triggered: keyword + CapFirst CapLast
	triggers := []string{
		// German
		`Herr`, `Frau`, `Patient`, `Patientin`, `Kollege`, `Kollegin`,
		`Dr\.`, `Prof\.`, `mein Freund`, `meine Freundin`,
		`meinen Patienten`, `meiner Patientin`,
		// French
		`Monsieur`, `Madame`, `Mademoiselle`, `mon ami`, `mon amie`,
		// English
		`Mr\.?`, `Mrs\.?`, `Ms\.?`, `Dr\.?`, `Prof\.?`,
		`my friend`, `my colleague`, `my patient`, `colleague`,
		// Dutch
		`Meneer`, `Mevrouw`,
		// Italian
		`Signor`, `Signora`,
		// Spanish
		`Señor`, `Señora`,
	}

	triggerGroup := `(?:` + strings.Join(triggers, `|`) + `)`
	// Use (?i:...) only for the trigger group, keep name pattern case-sensitive.
	contextPattern := `(?i:` + triggerGroup + `)\s+(` + fullName + `)`

	// Verb-triggered: "told/asked/called/emailed Name Name"
	verbs := `(?i:told|asked|called|emailed|contacted|met|visited|informed)`
	verbPattern := verbs + `\s+(` + fullName + `)`

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

func phoneScanners() []Scanner {
	// International format with + prefix: +49, +43, +41, +33, etc.
	// Supports separators: space, dash, dot, or none.
	intl := `\+(?:49|43|41|33|39|34|31|32|351|48|46|358|45|47|353|44)\s?[\d][\d\s.\-]{6,14}\d`

	// Generic 00-prefix international
	generic00 := `00\d{1,3}[\s.\-]?\d[\d\s.\-]{6,14}\d`

	// German local: 0XXX XXXXXXX
	deLocal := `0[1-9]\d{1,4}[\s.\-/]?\d[\d\s.\-]{4,10}\d`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(intl), "PHONE", 0.95),
		NewRegexScanner(regexp.MustCompile(generic00), "PHONE", 0.90),
		NewRegexScanner(regexp.MustCompile(deLocal), "PHONE", 0.85),
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
	// German: Straße/Str./Weg/Platz/Allee/Gasse + Hausnummer
	deStreet := `(?:[A-ZÄÖÜ][a-zäöüß]+(?:straße|str\.|weg|platz|allee|gasse|ring|damm|ufer))\s+\d{1,4}[a-zA-Z]?`

	// German with postcode + city
	deWithCity := deStreet + `(?:,\s*\d{5}\s+[A-ZÄÖÜ][a-zäöüß]+(?:\s+[A-ZÄÖÜ][a-zäöüß]+)?)?`

	// French: rue/avenue/boulevard + number
	frStreet := `\d{1,4},?\s+(?:rue|avenue|boulevard|place|chemin|impasse)\s+(?:de\s+(?:la\s+)?|du\s+|des\s+|l')?[A-ZÀ-Ü][a-zà-ÿ]+(?:\s+[A-ZÀ-Ü][a-zà-ÿ]+)*`

	// Italian: via/piazza/corso + name + number
	itStreet := `(?:[Vv]ia|[Pp]iazza|[Cc]orso|[Vv]iale)\s+[A-ZÀ-Ü][a-zà-ÿ]+(?:\s+[A-ZÀ-Ü][a-zà-ÿ]+)*\s+\d{1,4}`

	// Spanish: calle/avenida/plaza
	esStreet := `(?:[Cc]alle|[Aa]venida|[Pp]laza)\s+(?:de\s+(?:la\s+)?|del\s+)?[A-ZÀ-Ü][a-zà-ÿ]+(?:\s+[A-ZÀ-Ü][a-zà-ÿ]+)*\s+\d{1,4}`

	// Dutch: straat/laan/weg/plein + number
	nlStreet := `[A-ZÄÖÜ][a-zäöüß]+(?:straat|laan|weg|plein|gracht|kade|singel)\s+\d{1,4}`

	return []Scanner{
		NewRegexScanner(regexp.MustCompile(deWithCity), "ADDRESS", 0.85),
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
		{`gh[patos]_[A-Za-z0-9]{36}`, 0.99},
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
