package linter

import (
	"slices"
	"strings"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type LintMessage interface {
	Fixes() []rule.RuleFix
}

// ApplyRuleFixes applies non-conflicting fixes to a file's original text and
// returns the result, the diagnostics whose fixes were left for a later pass,
// and whether anything was applied.
//
// `code` is the text as it exists in the file, byte order mark included, while
// fix ranges index the text without one — the mark is never part of what a
// rule sees. That puts the mark at range position -1, which is how a rule asks
// for it to be removed, and is ESLint's convention in SourceCodeFixer.
func ApplyRuleFixes[M LintMessage](code string, diagnostics []M) (string, []M, bool) {
	bom := ""
	text := code
	if strings.HasPrefix(text, utils.BOM) {
		bom = utils.BOM
		text = text[len(utils.BOM):]
	}

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

	// Below every legal fix position, including the byte order mark's -1, so
	// the first fix is never mistaken for an overlap.
	lastFixEnd := -1
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

		// An inverted range is not a range this fixer can honor.
		isInverted := slices.ContainsFunc(fixes, func(fix rule.RuleFix) bool {
			return fix.Range.Pos() > fix.Range.End()
		})

		if isOverlapping || isAdjacentConflict || isInverted {
			unapplied = append(unapplied, diagnostic)
			continue
		}

		for _, fix := range fixes {
			fixed = true
			lastWasInsertion = fix.Range.Pos() == fix.Range.End()

			// A fix reaching back before the text either cuts the mark out or
			// writes over it; either way the mark does not survive. So does a
			// fix that replaces the start of the text with one of its own.
			if (fix.Range.Pos() < 0 && fix.Range.End() >= 0) ||
				(fix.Range.Pos() == 0 && strings.HasPrefix(fix.Text, utils.BOM)) {
				bom = ""
			}

			builder.WriteString(text[max(0, lastFixEnd):max(0, fix.Range.Pos())])
			builder.WriteString(fix.Text)

			lastFixEnd = fix.Range.End()
		}
	}

	builder.WriteString(text[max(0, lastFixEnd):])

	return bom + builder.String(), unapplied, fixed
}
