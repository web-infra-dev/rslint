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

// FileWarning is a file-scoped warning without a source range or rule. Config
// discovery uses it for explicit files excluded by ignores/base-path/config
// selection; it must count like an ESLint warning without fabricating an AST.
type FileWarning struct {
	FilePath string
	Message  string
}

type Report struct {
	diagnostics  []rule.RuleDiagnostic
	fileWarnings []FileWarning
	metadata     Metadata
	counts       Counts
}

// NewReport snapshots one completed CLI run. Render consumes the report once:
// the default formatter finalizes elapsed time only after diagnostic output is
// flushed, preserving the CLI's end-to-end timing boundary.
func NewReport(diagnostics []rule.RuleDiagnostic, metadata Metadata) Report {
	return NewReportWithFileWarnings(diagnostics, nil, metadata)
}

func NewReportWithFileWarnings(diagnostics []rule.RuleDiagnostic, fileWarnings []FileWarning, metadata Metadata) Report {
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
	counts.Warnings += len(fileWarnings)
	counts.LintErrors = counts.Errors - counts.TypeErrors

	return Report{
		diagnostics:  slices.Clone(diagnostics),
		fileWarnings: slices.Clone(fileWarnings),
		metadata:     metadata,
		counts:       counts,
	}
}

func (r Report) Counts() Counts {
	return r.counts
}
