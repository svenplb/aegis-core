package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/svenplb/aegis-core/internal/config"
	"github.com/svenplb/aegis-core/internal/redactor"
	"github.com/svenplb/aegis-core/internal/scanner"
)

func main() {
	os.Exit(run())
}

func run() int {
	textFlag := flag.String("text", "", "inline text to scan")
	fileFlag := flag.String("file", "", "path to file to scan")
	configFlag := flag.String("config", "", "path to config YAML file")
	jsonFlag := flag.Bool("json", false, "output structured JSON")
	flag.Parse()

	// Read input text.
	text, err := readInput(*textFlag, *fileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	// Load config.
	var cfg *config.Config
	if *configFlag != "" {
		cfg, err = config.Load(*configFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
			return 2
		}
	} else {
		cfg = config.DefaultConfig()
	}

	// Build allowlist from config.
	var allowlist []*regexp.Regexp
	for _, pattern := range cfg.Scanner.Allowlist {
		re, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error compiling allowlist pattern %q: %v\n", pattern, err)
			return 2
		}
		allowlist = append(allowlist, re)
	}

	// Scan.
	s := scanner.DefaultScanner(allowlist)
	entities := s.Scan(text)

	// Redact.
	result := redactor.Redact(text, entities)

	if *jsonFlag {
		return outputJSON(result)
	}
	return outputPretty(result, isTerminal())
}

func readInput(textFlag, fileFlag string) (string, error) {
	switch {
	case textFlag != "":
		return textFlag, nil
	case fileFlag != "":
		data, err := os.ReadFile(fileFlag)
		if err != nil {
			return "", fmt.Errorf("reading file: %w", err)
		}
		return string(data), nil
	default:
		// Check if stdin is piped.
		stat, err := os.Stdin.Stat()
		if err != nil {
			return "", fmt.Errorf("checking stdin: %w", err)
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return "", fmt.Errorf("no input provided (use --text, --file, or pipe to stdin)")
		}
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}
}

func isTerminal() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func outputJSON(result redactor.RedactResult) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
		return 2
	}
	if len(result.Entities) > 0 {
		return 1
	}
	return 0
}

// ANSI color codes.
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
)

func entityColor(entityType string) string {
	switch entityType {
	case "PERSON":
		return colorMagenta
	case "PHONE", "IP_ADDRESS":
		return colorYellow
	case "DATE":
		return colorBlue
	case "EMAIL", "URL":
		return colorCyan
	case "SECRET", "FINANCIAL", "CREDIT_CARD":
		return colorRed
	case "ADDRESS", "IBAN":
		return colorGreen
	default:
		return colorYellow
	}
}

func outputPretty(result redactor.RedactResult, useColor bool) int {
	entityCount := len(result.Entities)

	// --- ORIGINAL section with highlighted entities ---
	header := fmt.Sprintf("─── ORIGINAL (%d entities found) ", entityCount)
	header += strings.Repeat("─", max(0, 56-len(header)))

	if useColor {
		fmt.Printf("%s%s%s\n", colorBold, header, colorReset)
	} else {
		fmt.Println(header)
	}

	if useColor && entityCount > 0 {
		fmt.Println(highlightEntities(result.OriginalText, result.Entities))
	} else {
		fmt.Println(result.OriginalText)
	}

	// --- SANITIZED section ---
	fmt.Println()
	sanitizedHeader := "─── SANITIZED " + strings.Repeat("─", 42)
	if useColor {
		fmt.Printf("%s%s%s\n", colorBold, sanitizedHeader, colorReset)
	} else {
		fmt.Println(sanitizedHeader)
	}
	fmt.Println(result.SanitizedText)

	// --- STATISTICS section ---
	if entityCount > 0 {
		fmt.Println()
		statsHeader := "─── STATISTICS " + strings.Repeat("─", 41)
		if useColor {
			fmt.Printf("%s%s%s\n", colorBold, statsHeader, colorReset)
		} else {
			fmt.Println(statsHeader)
		}
		fmt.Printf("Replaced: %d\n\n", entityCount)

		// Count per type.
		typeCounts := make(map[string]int)
		for _, e := range result.Entities {
			typeCounts[e.Type]++
		}

		// Sort types for stable output.
		types := make([]string, 0, len(typeCounts))
		for t := range typeCounts {
			types = append(types, t)
		}
		sort.Strings(types)

		fmt.Printf("  %-14s %s\n", "Type", "Count")
		for _, t := range types {
			if useColor {
				fmt.Printf("  %s%-14s%s %d\n", entityColor(t), t, colorReset, typeCounts[t])
			} else {
				fmt.Printf("  %-14s %d\n", t, typeCounts[t])
			}
		}
	}

	fmt.Println()

	if entityCount > 0 {
		return 1
	}
	return 0
}

func highlightEntities(text string, entities []scanner.Entity) string {
	if len(entities) == 0 {
		return text
	}

	// Sort by Start ascending.
	sorted := make([]scanner.Entity, len(entities))
	copy(sorted, entities)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Start < sorted[j].Start
	})

	var buf strings.Builder
	lastEnd := 0
	for _, e := range sorted {
		if e.Start < lastEnd {
			continue // skip overlapping
		}
		buf.WriteString(text[lastEnd:e.Start])
		color := entityColor(e.Type)
		buf.WriteString(color)
		buf.WriteString(colorBold)
		buf.WriteString(text[e.Start:e.End])
		buf.WriteString(colorReset)
		lastEnd = e.End
	}
	buf.WriteString(text[lastEnd:])
	return buf.String()
}
