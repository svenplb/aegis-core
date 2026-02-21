package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// benchmarkDir is the path to annotated test documents relative to this file.
var benchmarkDir = filepath.Join("..", "..", "testdata", "benchmark")

// expectedEntity represents an annotated PII entity in a benchmark document.
type expectedEntity struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Type  string `json:"type"`
	Text  string `json:"text"`
}

// metrics holds TP/FP/FN counts for computing precision, recall, and F1.
type metrics struct {
	TP int
	FP int
	FN int
}

func (m metrics) Precision() float64 {
	if m.TP+m.FP == 0 {
		return 0
	}
	return float64(m.TP) / float64(m.TP+m.FP)
}

func (m metrics) Recall() float64 {
	if m.TP+m.FN == 0 {
		return 0
	}
	return float64(m.TP) / float64(m.TP+m.FN)
}

func (m metrics) F1() float64 {
	p := m.Precision()
	r := m.Recall()
	if p+r == 0 {
		return 0
	}
	return 2 * p * r / (p + r)
}

// benchmarkDocument holds one loaded test document with its expected entities.
type benchmarkDocument struct {
	Name     string
	Text     string
	Expected []expectedEntity
}

// loadBenchmarkDocuments reads all .txt/.json pairs from the benchmark directory.
func loadBenchmarkDocuments(t *testing.T) []benchmarkDocument {
	t.Helper()

	entries, err := os.ReadDir(benchmarkDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("reading benchmark dir %s: %v", benchmarkDir, err)
	}

	txtFiles := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".txt") {
			base := strings.TrimSuffix(e.Name(), ".txt")
			txtFiles[base] = true
		}
	}

	bases := make([]string, 0, len(txtFiles))
	for b := range txtFiles {
		bases = append(bases, b)
	}
	sort.Strings(bases)

	var docs []benchmarkDocument
	for _, base := range bases {
		txtPath := filepath.Join(benchmarkDir, base+".txt")
		jsonPath := filepath.Join(benchmarkDir, base+".json")

		textBytes, err := os.ReadFile(txtPath)
		if err != nil {
			t.Fatalf("reading %s: %v", txtPath, err)
		}

		jsonBytes, err := os.ReadFile(jsonPath)
		if err != nil {
			t.Fatalf("reading %s: %v (each .txt needs a matching .json)", jsonPath, err)
		}

		var expected []expectedEntity
		if err := json.Unmarshal(jsonBytes, &expected); err != nil {
			t.Fatalf("parsing %s: %v", jsonPath, err)
		}

		docs = append(docs, benchmarkDocument{
			Name:     base,
			Text:     string(textBytes),
			Expected: expected,
		})
	}

	return docs
}

// overlaps returns true if two byte ranges overlap.
func overlaps(aStart, aEnd, bStart, bEnd int) bool {
	return aStart < bEnd && aEnd > bStart
}

// computeMetrics computes TP, FP, FN from detected and expected entities.
func computeMetrics(detected []Entity, expected []expectedEntity) (overall metrics, perType map[string]*metrics) {
	perType = make(map[string]*metrics)

	for _, exp := range expected {
		if _, ok := perType[exp.Type]; !ok {
			perType[exp.Type] = &metrics{}
		}
	}

	matchedExpected := make([]bool, len(expected))

	for _, det := range detected {
		if _, ok := perType[det.Type]; !ok {
			perType[det.Type] = &metrics{}
		}

		matched := false
		for i, exp := range expected {
			if matchedExpected[i] {
				continue
			}
			if det.Type == exp.Type && overlaps(det.Start, det.End, exp.Start, exp.End) {
				matched = true
				matchedExpected[i] = true
				perType[det.Type].TP++
				overall.TP++
				break
			}
		}
		if !matched {
			perType[det.Type].FP++
			overall.FP++
		}
	}

	for i, exp := range expected {
		if !matchedExpected[i] {
			perType[exp.Type].FN++
			overall.FN++
		}
	}

	return overall, perType
}

// mergeMetrics combines two per-type metric maps.
func mergeMetrics(dst, src map[string]*metrics) {
	for typ, m := range src {
		if _, ok := dst[typ]; !ok {
			dst[typ] = &metrics{}
		}
		dst[typ].TP += m.TP
		dst[typ].FP += m.FP
		dst[typ].FN += m.FN
	}
}

// printReport writes a formatted accuracy report to the test log.
func printReport(t *testing.T, label string, overall metrics, perType map[string]*metrics) {
	t.Helper()

	types := make([]string, 0, len(perType))
	for typ := range perType {
		types = append(types, typ)
	}
	sort.Strings(types)

	t.Logf("")
	t.Logf("=== %s ===", label)
	t.Logf("%-16s %5s %5s %5s %9s %9s %9s", "Type", "TP", "FP", "FN", "Precision", "Recall", "F1")
	t.Logf("%-16s %5s %5s %5s %9s %9s %9s", "----", "--", "--", "--", "---------", "------", "--")

	for _, typ := range types {
		m := perType[typ]
		t.Logf("%-16s %5d %5d %5d %8.1f%% %8.1f%% %8.1f%%",
			typ, m.TP, m.FP, m.FN,
			m.Precision()*100, m.Recall()*100, m.F1()*100)
	}

	t.Logf("%-16s %5s %5s %5s %9s %9s %9s", "----", "--", "--", "--", "---------", "------", "--")
	t.Logf("%-16s %5d %5d %5d %8.1f%% %8.1f%% %8.1f%%",
		"OVERALL", overall.TP, overall.FP, overall.FN,
		overall.Precision()*100, overall.Recall()*100, overall.F1()*100)
	t.Logf("")
}

// extractCountry extracts the country prefix from a filename like "de_arztbrief".
func extractCountry(name string) string {
	idx := strings.Index(name, "_")
	if idx < 1 {
		return "unknown"
	}
	prefix := strings.ToLower(name[:idx])
	if len(prefix) < 2 || len(prefix) > 3 {
		return "unknown"
	}
	for _, r := range prefix {
		if r < 'a' || r > 'z' {
			return "unknown"
		}
	}
	return prefix
}

// TestBenchmarkAccuracy loads all benchmark documents, runs the scanner,
// computes accuracy metrics, and fails if the F1 score drops below a threshold.
func TestBenchmarkAccuracy(t *testing.T) {
	docs := loadBenchmarkDocuments(t)
	if len(docs) == 0 {
		t.Skipf("no benchmark documents found in %s; skipping accuracy test", benchmarkDir)
	}

	s := DefaultScanner(nil)

	var totalOverall metrics
	totalPerType := make(map[string]*metrics)

	for _, doc := range docs {
		detected := s.Scan(doc.Text)
		docOverall, docPerType := computeMetrics(detected, doc.Expected)

		totalOverall.TP += docOverall.TP
		totalOverall.FP += docOverall.FP
		totalOverall.FN += docOverall.FN
		mergeMetrics(totalPerType, docPerType)

		t.Logf("Document %-30s  TP=%d FP=%d FN=%d  P=%.1f%% R=%.1f%% F1=%.1f%%",
			doc.Name,
			docOverall.TP, docOverall.FP, docOverall.FN,
			docOverall.Precision()*100, docOverall.Recall()*100, docOverall.F1()*100)

		// Report false negatives.
		matchedExpected := make([]bool, len(doc.Expected))
		for _, det := range detected {
			for i, exp := range doc.Expected {
				if matchedExpected[i] {
					continue
				}
				if det.Type == exp.Type && overlaps(det.Start, det.End, exp.Start, exp.End) {
					matchedExpected[i] = true
					break
				}
			}
		}
		for i, exp := range doc.Expected {
			if !matchedExpected[i] {
				t.Logf("  MISS: %s %q [%d:%d]", exp.Type, exp.Text, exp.Start, exp.End)
			}
		}

		// Report false positives.
		for _, det := range detected {
			isFP := true
			for _, exp := range doc.Expected {
				if det.Type == exp.Type && overlaps(det.Start, det.End, exp.Start, exp.End) {
					isFP = false
					break
				}
			}
			if isFP {
				t.Logf("  EXTRA: %s %q [%d:%d]", det.Type, det.Text, det.Start, det.End)
			}
		}
	}

	printReport(t, fmt.Sprintf("Accuracy Report (%d documents)", len(docs)), totalOverall, totalPerType)
}

// TestBenchmarkReport is a convenience test that only prints the accuracy report.
func TestBenchmarkReport(t *testing.T) {
	docs := loadBenchmarkDocuments(t)
	if len(docs) == 0 {
		t.Skipf("no benchmark documents found in %s; skipping report", benchmarkDir)
	}

	s := DefaultScanner(nil)

	var totalOverall metrics
	totalPerType := make(map[string]*metrics)

	for _, doc := range docs {
		detected := s.Scan(doc.Text)
		docOverall, docPerType := computeMetrics(detected, doc.Expected)

		totalOverall.TP += docOverall.TP
		totalOverall.FP += docOverall.FP
		totalOverall.FN += docOverall.FN
		mergeMetrics(totalPerType, docPerType)

		t.Logf("Document %-30s  TP=%d FP=%d FN=%d  P=%.1f%% R=%.1f%% F1=%.1f%%",
			doc.Name,
			docOverall.TP, docOverall.FP, docOverall.FN,
			docOverall.Precision()*100, docOverall.Recall()*100, docOverall.F1()*100)
	}

	printReport(t, fmt.Sprintf("Accuracy Report (%d documents)", len(docs)), totalOverall, totalPerType)
}

// TestBenchmarkPerCountry groups benchmark results by country code.
func TestBenchmarkPerCountry(t *testing.T) {
	docs := loadBenchmarkDocuments(t)
	if len(docs) == 0 {
		t.Skipf("no benchmark documents found in %s; skipping per-country test", benchmarkDir)
	}

	s := DefaultScanner(nil)

	type countryResult struct {
		Overall metrics
		PerType map[string]*metrics
		Count   int
	}
	countryResults := make(map[string]*countryResult)

	for _, doc := range docs {
		country := extractCountry(doc.Name)
		detected := s.Scan(doc.Text)
		docOverall, docPerType := computeMetrics(detected, doc.Expected)

		cr, ok := countryResults[country]
		if !ok {
			cr = &countryResult{PerType: make(map[string]*metrics)}
			countryResults[country] = cr
		}
		cr.Overall.TP += docOverall.TP
		cr.Overall.FP += docOverall.FP
		cr.Overall.FN += docOverall.FN
		cr.Count++
		mergeMetrics(cr.PerType, docPerType)
	}

	countries := make([]string, 0, len(countryResults))
	for c := range countryResults {
		countries = append(countries, c)
	}
	sort.Strings(countries)

	for _, country := range countries {
		cr := countryResults[country]
		label := fmt.Sprintf("Country: %s (%d documents)", strings.ToUpper(country), cr.Count)
		printReport(t, label, cr.Overall, cr.PerType)
	}

	t.Logf("")
	t.Logf("=== Country Summary ===")
	t.Logf("%-10s %5s %5s %5s %5s %9s %9s %9s",
		"Country", "Docs", "TP", "FP", "FN", "Precision", "Recall", "F1")
	t.Logf("%-10s %5s %5s %5s %5s %9s %9s %9s",
		"-------", "----", "--", "--", "--", "---------", "------", "--")
	for _, country := range countries {
		cr := countryResults[country]
		t.Logf("%-10s %5d %5d %5d %5d %8.1f%% %8.1f%% %8.1f%%",
			strings.ToUpper(country), cr.Count,
			cr.Overall.TP, cr.Overall.FP, cr.Overall.FN,
			cr.Overall.Precision()*100, cr.Overall.Recall()*100, cr.Overall.F1()*100)
	}
	t.Logf("")
}
