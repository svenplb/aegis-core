package scanner

// Entity represents a detected PII entity in text.
// JSON shape matches the existing aegis-software frontend contract.
type Entity struct {
	Start    int     `json:"start"`    // byte offset in text
	End      int     `json:"end"`      // byte offset (exclusive)
	Type     string  `json:"type"`     // "PERSON", "EMAIL", "PHONE", etc.
	Text     string  `json:"text"`     // matched substring
	Score    float64 `json:"score"`    // confidence (0.0–1.0)
	Detector string  `json:"detector"` // detection method, e.g. "regex"
}
