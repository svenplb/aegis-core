package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/svenplb/aegis-core/internal/redactor"
	"github.com/svenplb/aegis-core/internal/scanner"
)

// View states.
const (
	stateInput = iota
	stateResults
	stateSettings
)

// Lipgloss color mapping per entity type.
func entityColor(entityType string) lipgloss.Color {
	switch entityType {
	case "PERSON":
		return lipgloss.Color("5") // magenta
	case "PHONE", "IP_ADDRESS":
		return lipgloss.Color("3") // yellow
	case "DATE":
		return lipgloss.Color("4") // blue
	case "EMAIL", "URL":
		return lipgloss.Color("6") // cyan
	case "SECRET", "FINANCIAL", "CREDIT_CARD":
		return lipgloss.Color("1") // red
	case "ADDRESS", "IBAN":
		return lipgloss.Color("2") // green
	default:
		return lipgloss.Color("3") // yellow
	}
}

// Styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("7")).
			Background(lipgloss.Color("5")).
			Padding(0, 1)

	headerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("5")).
			Padding(0, 1).
			Width(45)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("8"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))
)

type model struct {
	state    int
	textarea textarea.Model
	viewport viewport.Model
	result   *redactor.RedactResult
	width    int
	height   int
	ready    bool // viewport dimensions set
	scanTime time.Duration

	// Settings.
	thresholdPct   int // 0–100, displayed as 0.00–1.00
	allowlist      []string
	settingsFocus  int // 0=threshold, 1..n=allowlist items
	allowlistInput textinput.Model
	addingPattern  bool
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Paste or type text here..."
	ta.ShowLineNumbers = false
	ta.SetHeight(12)
	ta.SetWidth(70)
	ta.Focus()
	ta.CharLimit = 0 // unlimited

	ti := textinput.New()
	ti.Placeholder = "regex pattern..."
	ti.CharLimit = 200
	ti.Width = 40

	return model{
		state:          stateInput,
		textarea:       ta,
		allowlistInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		taWidth := min(msg.Width-4, 80)
		m.textarea.SetWidth(taWidth)

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-6)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 6
		}
		if m.state == stateResults && m.result != nil {
			m.viewport.SetContent(m.renderResults())
		}

	case tea.KeyMsg:
		switch m.state {
		case stateInput:
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyCtrlD:
				return m.doScan()
			case tea.KeyTab:
				m.textarea.Blur()
				m.state = stateSettings
				m.settingsFocus = 0
				return m, nil
			}
		case stateResults:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "n":
				m.textarea.Reset()
				m.textarea.Focus()
				m.state = stateInput
				m.result = nil
				return m, textarea.Blink
			}
		case stateSettings:
			return m.updateSettings(msg)
		}
	}

	switch m.state {
	case stateInput:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	case stateResults:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Adding pattern mode — textinput captures all keys.
	if m.addingPattern {
		switch msg.Type {
		case tea.KeyEnter:
			pattern := strings.TrimSpace(m.allowlistInput.Value())
			if pattern != "" {
				if _, err := regexp.Compile(pattern); err == nil {
					m.allowlist = append(m.allowlist, pattern)
				}
			}
			m.allowlistInput.SetValue("")
			m.allowlistInput.Blur()
			m.addingPattern = false
			return m, nil
		case tea.KeyEscape:
			m.allowlistInput.SetValue("")
			m.allowlistInput.Blur()
			m.addingPattern = false
			return m, nil
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.allowlistInput, cmd = m.allowlistInput.Update(msg)
			return m, cmd
		}
	}

	// Navigation mode.
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyTab:
		m.textarea.Focus()
		m.state = stateInput
		return m, textarea.Blink
	case tea.KeyUp:
		if m.settingsFocus > 0 {
			m.settingsFocus--
		}
	case tea.KeyDown:
		if m.settingsFocus < len(m.allowlist) {
			m.settingsFocus++
		}
	case tea.KeyLeft:
		if m.settingsFocus == 0 {
			m.thresholdPct = max(0, m.thresholdPct-5)
		}
	case tea.KeyRight:
		if m.settingsFocus == 0 {
			m.thresholdPct = min(100, m.thresholdPct+5)
		}
	}

	switch msg.String() {
	case "a":
		m.addingPattern = true
		m.allowlistInput.Focus()
		return m, textinput.Blink
	case "d", "x":
		if m.settingsFocus >= 1 && m.settingsFocus-1 < len(m.allowlist) {
			idx := m.settingsFocus - 1
			m.allowlist = append(m.allowlist[:idx], m.allowlist[idx+1:]...)
			if m.settingsFocus > len(m.allowlist) {
				m.settingsFocus = max(0, len(m.allowlist))
			}
		}
	}

	return m, nil
}

func (m model) doScan() (tea.Model, tea.Cmd) {
	text := m.textarea.Value()
	if strings.TrimSpace(text) == "" {
		return m, nil
	}

	// Build allowlist regexps from settings.
	var allowlist []*regexp.Regexp
	for _, pattern := range m.allowlist {
		if re, err := regexp.Compile(pattern); err == nil {
			allowlist = append(allowlist, re)
		}
	}

	start := time.Now()
	s := scanner.DefaultScanner(allowlist)
	entities := s.Scan(text)

	// Apply score threshold.
	if m.thresholdPct > 0 {
		threshold := float64(m.thresholdPct) / 100.0
		var filtered []scanner.Entity
		for _, e := range entities {
			if e.Score >= threshold {
				filtered = append(filtered, e)
			}
		}
		entities = filtered
	}

	result := redactor.Redact(text, entities)
	m.scanTime = time.Since(start)

	m.result = &result
	m.state = stateResults
	m.textarea.Blur()

	if m.ready {
		m.viewport.SetContent(m.renderResults())
		m.viewport.GotoTop()
	}

	return m, nil
}

func (m model) thresholdDesc() string {
	switch {
	case m.thresholdPct == 0:
		return "all detections"
	case m.thresholdPct <= 75:
		return "minimal filtering"
	case m.thresholdPct <= 85:
		return "moderate"
	case m.thresholdPct <= 90:
		return "strict"
	case m.thresholdPct <= 95:
		return "very strict"
	default:
		return "highest confidence only"
	}
}

func (m model) renderAnnotated() string {
	text := m.result.OriginalText
	entities := m.result.Entities

	// Sort entities by Start ascending.
	sorted := make([]scanner.Entity, len(entities))
	copy(sorted, entities)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Start < sorted[j].Start
	})

	var b strings.Builder
	pos := 0
	for _, e := range sorted {
		if e.Start < pos {
			continue // skip overlapping
		}
		// Write text before entity.
		if e.Start > pos {
			b.WriteString(text[pos:e.Start])
		}
		// Write highlighted entity with type tag.
		clr := entityColor(e.Type)
		highlighted := lipgloss.NewStyle().
			Foreground(clr).
			Bold(true).
			Underline(true).
			Render(e.Text)
		tag := lipgloss.NewStyle().
			Foreground(clr).
			Render("⟨" + e.Type + "⟩")
		b.WriteString(highlighted + tag)
		pos = e.End
	}
	// Write remaining text.
	if pos < len(text) {
		b.WriteString(text[pos:])
	}

	return b.String()
}

func (m model) renderResults() string {
	if m.result == nil {
		return ""
	}

	var b strings.Builder
	r := m.result

	// --- Annotated section ---
	b.WriteString(sectionStyle.Render("─── ANNOTATED ") + sectionStyle.Render(strings.Repeat("─", max(m.width-16, 20))))
	b.WriteString("\n")
	b.WriteString(m.renderAnnotated())
	b.WriteString("\n\n")

	// --- Sanitized section ---
	b.WriteString(sectionStyle.Render("─── SANITIZED ") + sectionStyle.Render(strings.Repeat("─", max(m.width-16, 20))))
	b.WriteString("\n")
	b.WriteString(r.SanitizedText)
	b.WriteString("\n\n")

	// --- Mappings section ---
	if len(r.Mappings) > 0 {
		b.WriteString(sectionStyle.Render("─── MAPPINGS ") + sectionStyle.Render(strings.Repeat("─", max(m.width-15, 20))))
		b.WriteString("\n")

		// Calculate column widths.
		maxToken, maxOrig := 0, 0
		for _, m := range r.Mappings {
			if len(m.Token) > maxToken {
				maxToken = len(m.Token)
			}
			if len(m.Original) > maxOrig {
				maxOrig = len(m.Original)
			}
		}

		for _, mp := range r.Mappings {
			clr := entityColor(mp.Type)
			tokenStyled := lipgloss.NewStyle().Foreground(clr).Bold(true).Render(mp.Token)
			typeStyled := lipgloss.NewStyle().Foreground(clr).Render(mp.Type)

			// Pad token and original for alignment.
			tokenPad := strings.Repeat(" ", maxToken-len(mp.Token))
			origPad := strings.Repeat(" ", maxOrig-len(mp.Original))

			b.WriteString(fmt.Sprintf("  %s%s    %s%s    %s\n",
				tokenStyled, tokenPad,
				mp.Original, origPad,
				typeStyled))
		}
		b.WriteString("\n")
	}

	// --- Statistics section ---
	typeCounts := make(map[string]int)
	for _, e := range r.Entities {
		typeCounts[e.Type]++
	}

	if len(typeCounts) > 0 {
		b.WriteString(sectionStyle.Render("─── STATISTICS ") + sectionStyle.Render(strings.Repeat("─", max(m.width-17, 20))))
		b.WriteString("\n")

		// Sort types by count descending.
		type typeStat struct {
			name  string
			count int
		}
		var stats []typeStat
		maxCount := 0
		for name, count := range typeCounts {
			stats = append(stats, typeStat{name, count})
			if count > maxCount {
				maxCount = count
			}
		}
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].count > stats[j].count
		})

		maxBarWidth := 20
		maxName := 0
		for _, s := range stats {
			if len(s.name) > maxName {
				maxName = len(s.name)
			}
		}

		for _, s := range stats {
			clr := entityColor(s.name)
			barLen := s.count * maxBarWidth / maxCount
			if barLen < 1 {
				barLen = 1
			}
			bar := lipgloss.NewStyle().Foreground(clr).Render(strings.Repeat("█", barLen))
			namePad := strings.Repeat(" ", maxName-len(s.name))
			nameStyled := lipgloss.NewStyle().Foreground(clr).Bold(true).Render(s.name)
			b.WriteString(fmt.Sprintf("  %s%s  %d  %s\n", nameStyled, namePad, s.count, bar))
		}
	}

	return b.String()
}

func (m model) View() string {
	switch m.state {
	case stateInput:
		return m.viewInput()
	case stateResults:
		return m.viewResults()
	case stateSettings:
		return m.viewSettings()
	}
	return ""
}

func (m model) viewInput() string {
	header := headerBoxStyle.Render(titleStyle.Render("aegis") + " — PII Scanner")

	var settingsInfo string
	if m.thresholdPct > 0 || len(m.allowlist) > 0 {
		var parts []string
		if m.thresholdPct > 0 {
			parts = append(parts, fmt.Sprintf("threshold:%.2f", float64(m.thresholdPct)/100.0))
		}
		if len(m.allowlist) > 0 {
			parts = append(parts, fmt.Sprintf("allowlist:%d", len(m.allowlist)))
		}
		settingsInfo = "\n" + dimStyle.Render("  "+strings.Join(parts, "  "))
	}

	help := helpStyle.Render("  Ctrl+D scan  •  Tab settings  •  Ctrl+C quit")

	return fmt.Sprintf("\n%s%s\n\n%s\n\n%s\n", header, settingsInfo, m.textarea.View(), help)
}

func (m model) viewResults() string {
	if m.result == nil {
		return ""
	}

	entityCount := len(m.result.Entities)
	ms := m.scanTime.Milliseconds()

	headerText := fmt.Sprintf("%s — %d entities found (%dms)",
		titleStyle.Render("aegis"), entityCount, ms)
	header := headerBoxStyle.Render(headerText)

	help := helpStyle.Render("  n new scan  •  q quit")

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n", header, m.viewport.View(), help)
}

func (m model) viewSettings() string {
	var b strings.Builder

	header := headerBoxStyle.Render(titleStyle.Render("aegis") + " — Settings")
	b.WriteString("\n" + header + "\n\n")

	// Score Threshold.
	thresholdStr := fmt.Sprintf("%.2f", float64(m.thresholdPct)/100.0)
	desc := m.thresholdDesc()

	if m.settingsFocus == 0 {
		b.WriteString(fmt.Sprintf("  %s  %s   ◂ %s ▸   %s\n",
			activeStyle.Render("▸"),
			lipgloss.NewStyle().Bold(true).Render("Score Threshold"),
			valueStyle.Render(thresholdStr),
			dimStyle.Render(desc)))
	} else {
		b.WriteString(fmt.Sprintf("     %s     %s     %s\n",
			"Score Threshold",
			dimStyle.Render(thresholdStr),
			dimStyle.Render(desc)))
	}

	b.WriteString("\n")

	// Allowlist Patterns.
	b.WriteString("  " + lipgloss.NewStyle().Bold(true).Render("Allowlist Patterns") + "\n")

	if m.addingPattern {
		b.WriteString("    " + m.allowlistInput.View() + "\n")
	}

	if len(m.allowlist) == 0 && !m.addingPattern {
		b.WriteString("    " + dimStyle.Render("(no patterns — press a to add)") + "\n")
	}

	for i, pattern := range m.allowlist {
		if m.settingsFocus == i+1 {
			b.WriteString(fmt.Sprintf("    %s %s\n",
				activeStyle.Render("▸"),
				valueStyle.Render(pattern)))
		} else {
			b.WriteString(fmt.Sprintf("      %s\n", dimStyle.Render(pattern)))
		}
	}

	b.WriteString("\n")

	// Help.
	var helpParts []string
	helpParts = append(helpParts, "Tab back")
	helpParts = append(helpParts, "↑↓ navigate")
	if m.settingsFocus == 0 {
		helpParts = append(helpParts, "←→ threshold")
	}
	if !m.addingPattern {
		helpParts = append(helpParts, "a add pattern")
	}
	if m.settingsFocus >= 1 && len(m.allowlist) > 0 {
		helpParts = append(helpParts, "d delete")
	}
	b.WriteString(helpStyle.Render("  " + strings.Join(helpParts, "  •  ")) + "\n")

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
