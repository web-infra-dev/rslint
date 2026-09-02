package output

import (
	"errors"
	"slices"
	"time"
)

type Mode uint8

const (
	ModeLint Mode = iota
	ModeLintAndTypeCheck
	ModeTypeCheckOnly
)

// Severity is the presentation-level severity consumed by output formatters.
// It deliberately does not expose the rule framework's diagnostic type.
type Severity uint8

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityOff
)

func (severity Severity) String() string {
	switch severity {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warn"
	case SeverityOff:
		return "off"
	default:
		return "error"
	}
}

// Position is zero-based; Column is measured in UTF-16 code units.
type Position struct {
	Line   int
	Column int
}

// TextRange is a half-open range of UTF-8 byte offsets in the source text.
type TextRange struct {
	Start int
	End   int
}

// DiagnosticSource is an immutable, output-owned source snapshot used for
// code frames and location validation. It retains no compiler object and
// clones the caller-owned mutable line map.
type DiagnosticSource struct {
	text       string
	lineStarts []int
}

// NewDiagnosticSource validates that source presentation data is safe to
// consume and copies its mutable, UTF-8-byte-offset line starts into
// output-owned ints. The narrow integer constraint accepts both native Go
// offsets and ts-go TextPos values without importing compiler-core contracts.
// The producer owns ECMAScript line-boundary semantics; output must not
// rediscover line boundaries from the source text.
func NewDiagnosticSource[T ~int | ~int32](text string, lineStarts []T) (*DiagnosticSource, error) {
	if len(lineStarts) == 0 || lineStarts[0] != 0 {
		return nil, errors.New("source line starts are not structurally valid")
	}
	textLength := len(text)
	ownedLineStarts := make([]int, len(lineStarts))
	previous := -1
	for index, value := range lineStarts {
		lineStart := int(value)
		if lineStart <= previous || lineStart > textLength {
			return nil, errors.New("source line starts are not structurally valid")
		}
		ownedLineStarts[index] = lineStart
		previous = lineStart
	}
	return &DiagnosticSource{text: text, lineStarts: ownedLineStarts}, nil
}

// Diagnostic is the presentation projection of one finding. Location values
// have already been detached from the lint domain; Source is populated only
// when the selected formatter needs a code frame.
type Diagnostic struct {
	FilePath     string
	RuleName     string
	Message      string
	Range        TextRange
	Start        Position
	End          Position
	Source       *DiagnosticSource
	Severity     Severity
	PreFormatted bool
}

// OutcomeKind is the CLI decision that drives the completed status line. The
// command computes it once so rendering and the process exit code cannot
// disagree (notably when --max-warnings is exceeded).
type OutcomeKind uint8

const (
	OutcomePassed OutcomeKind = iota
	OutcomeDiagnosticsFailed
	OutcomeWarningLimitExceeded
)

type Outcome struct {
	Kind         OutcomeKind
	WarningLimit int
}

func (outcome Outcome) Failed() bool {
	return outcome.Kind != OutcomePassed
}

// Summary contains the default formatter's completed-run facts that are not
// already part of the Report. A nil Summary means the report is
// diagnostics-only and avoids computing data that machine formats never use.
type Summary struct {
	Files       int
	Rules       int
	Threads     int
	FixedIssues int
	StartedAt   time.Time
}

type Counts struct {
	Errors     int
	Warnings   int
	LintErrors int
	TypeErrors int
}

type Report struct {
	mode        Mode
	diagnostics []Diagnostic
	summary     Summary
	hasSummary  bool
	counts      Counts
	outcome     Outcome
}

// NewReport snapshots a completed CLI report. Counts and outcome are supplied
// by the command-owned assembly stage; output treats that projection as
// authoritative and only retains and renders it.
func NewReport(
	mode Mode,
	diagnostics []Diagnostic,
	counts Counts,
	summary *Summary,
	outcome Outcome,
) Report {
	var ownedSummary Summary
	hasSummary := summary != nil
	if summary != nil {
		ownedSummary = *summary
	}
	return Report{
		mode:        mode,
		diagnostics: slices.Clone(diagnostics),
		summary:     ownedSummary,
		hasSummary:  hasSummary,
		counts:      counts,
		outcome:     outcome,
	}
}

func (report Report) Counts() Counts {
	return report.counts
}

func (report Report) Outcome() Outcome {
	return report.outcome
}
