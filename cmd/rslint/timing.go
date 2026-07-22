package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/web-infra-dev/rslint/internal/linter"
)

// timingEnvEnabled reports whether the ESLint-style TIMING environment
// variable requests the per-rule timing table.
func timingEnvEnabled(value string) bool {
	switch strings.ToLower(value) {
	case "", "0", "false":
		return false
	}
	return true
}

// formatRuleTimingTable renders the per-rule timing table, sorted by total
// time descending. Relative is each rule's share of the summed rule time.
func formatRuleTimingTable(timings map[string]linter.RuleTiming) string {
	if len(timings) == 0 {
		return ""
	}

	names := make([]string, 0, len(timings))
	var totalTime time.Duration
	for name, t := range timings {
		names = append(names, name)
		totalTime += t.Time
	}
	sort.Slice(names, func(i, j int) bool {
		a, b := timings[names[i]], timings[names[j]]
		if a.Time != b.Time {
			return a.Time > b.Time
		}
		return names[i] < names[j]
	})

	header := []string{"Rule", "Source", "Time (ms)", "Files", "Relative"}
	rows := make([][]string, 0, len(names))
	for _, name := range names {
		t := timings[name]
		relative := 0.0
		if totalTime > 0 {
			relative = float64(t.Time) / float64(totalTime) * 100
		}
		rows = append(rows, []string{
			name,
			t.Kind,
			fmt.Sprintf("%.1f", float64(t.Time.Microseconds())/1000),
			fmt.Sprintf("%d", t.Files),
			fmt.Sprintf("%.1f%%", relative),
		})
	}

	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var b strings.Builder
	b.WriteString("\n")
	writeRow := func(cells []string) {
		for i, cell := range cells {
			if i > 0 {
				b.WriteString(" | ")
			}
			if i <= 1 { // Rule and Source are left-aligned, numbers right-aligned.
				b.WriteString(cell)
				b.WriteString(strings.Repeat(" ", widths[i]-len(cell)))
			} else {
				b.WriteString(strings.Repeat(" ", widths[i]-len(cell)))
				b.WriteString(cell)
			}
		}
		b.WriteString("\n")
	}
	writeRow(header)
	for i := range header {
		if i > 0 {
			b.WriteString("-|-")
		}
		b.WriteString(strings.Repeat("-", widths[i]))
	}
	b.WriteString("\n")
	for _, row := range rows {
		writeRow(row)
	}
	b.WriteString("Rule times are summed across parallel workers and may exceed wall-clock time.\n")
	return b.String()
}
