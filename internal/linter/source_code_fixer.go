package linter

import (
	"slices"
	"strings"

	"github.com/web-infra-dev/rslint/internal/rule"
)

type LintMessage interface {
	Fixes() []rule.RuleFix
}

func ApplyRuleFixes[M LintMessage](code string, diagnostics []M) (string, []M, bool) {
	unapplied := []M{}
	withFixes := []M{}

	fixed := false

	for _, diagnostic := range diagnostics {
		if len(diagnostic.Fixes()) > 0 {
			slices.SortFunc(diagnostic.Fixes(), func(a rule.RuleFix, b rule.RuleFix) int {
				start := a.Range.Pos() - b.Range.Pos()
				if start == 0 {
					return a.Range.End() - b.Range.End()
				}
				return start
			})
			withFixes = append(withFixes, diagnostic)
		} else {
			unapplied = append(unapplied, diagnostic)
		}
	}

	slices.SortFunc(withFixes, func(a M, b M) int {
		aFixes, bFixes := a.Fixes(), b.Fixes()

		start := aFixes[0].Range.Pos() - bFixes[0].Range.Pos()
		if start == 0 {
			return aFixes[len(aFixes)-1].Range.End() - bFixes[len(bFixes)-1].Range.End()
		}
		return start
	})

	var builder strings.Builder

	lastFixEnd := 0
	lastWasInsertion := false
	for _, diagnostic := range withFixes {
		fixes := diagnostic.Fixes()
		firstFix := fixes[0]

		isCurrentFixInsertion := firstFix.Range.Pos() == firstFix.Range.End()

		// Check for overlapping fixes (e.g., [0,5] and [2,7])
		isOverlapping := lastFixEnd > firstFix.Range.Pos()

		// Check for adjacent conflicts. This happens when a fix starts exactly where the last one ended,
		// and at least one of them is an insertion. Adjacent replacements are allowed.
		//   - Insertion followed by insertion at same pos: duplicate
		//   - Insertion followed by replacement at same pos: conflict (replacement starts where insertion happened)
		//   - Replacement followed by insertion at same pos: conflict (ambiguous position after replacement)
		//   - Replacement followed by replacement at same pos: OK (adjacent, non-overlapping)
		isAdjacentConflict := fixed &&
			lastFixEnd == firstFix.Range.Pos() &&
			(isCurrentFixInsertion || lastWasInsertion)

		if isOverlapping || isAdjacentConflict {
			unapplied = append(unapplied, diagnostic)
			continue
		}

		for _, fix := range fixes {
			fixed = true
			lastWasInsertion = fix.Range.Pos() == fix.Range.End()

			builder.WriteString(code[lastFixEnd:fix.Range.Pos()])
			builder.WriteString(fix.Text)

			lastFixEnd = fix.Range.End()
		}
	}

	builder.WriteString(code[lastFixEnd:])

	return builder.String(), unapplied, fixed
}
