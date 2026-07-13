package output

import (
	"slices"
	"time"

	"github.com/web-infra-dev/rslint/internal/rule"
)

type Mode uint8

const (
	ModeLint Mode = iota
	ModeLintAndTypeCheck
	ModeTypeCheckOnly
)

// Metadata contains run-level values that cannot be derived from diagnostics.
// StartedAt is retained instead of a precomputed duration so the default
// summary preserves the existing end-to-end timing boundary, including
// diagnostic rendering before the summary is written.
type Metadata struct {
	Mode             Mode
	LintedFiles      int
	TypeCheckedFiles int
	Rules            int
	Threads          int
	FixedIssues      int
	StartedAt        time.Time
}

type Counts struct {
	Errors     int
	Warnings   int
	LintErrors int
	TypeErrors int
}

type Report struct {
	diagnostics []rule.RuleDiagnostic
	metadata    Metadata
	counts      Counts
}

// NewReport snapshots one completed CLI run. Render consumes the report once:
// the default formatter finalizes elapsed time only after diagnostic output is
// flushed, preserving the CLI's end-to-end timing boundary.
func NewReport(diagnostics []rule.RuleDiagnostic, metadata Metadata) Report {
	counts := Counts{}
	for _, diagnostic := range diagnostics {
		switch diagnostic.Severity {
		case rule.SeverityError:
			counts.Errors++
			if (metadata.Mode == ModeLintAndTypeCheck || metadata.Mode == ModeTypeCheckOnly) &&
				diagnostic.Origin == rule.DiagnosticOriginTypeScript {
				counts.TypeErrors++
			}
		case rule.SeverityWarning:
			counts.Warnings++
		}
	}
	counts.LintErrors = counts.Errors - counts.TypeErrors

	return Report{
		diagnostics: slices.Clone(diagnostics),
		metadata:    metadata,
		counts:      counts,
	}
}

func (r Report) Counts() Counts {
	return r.counts
}
