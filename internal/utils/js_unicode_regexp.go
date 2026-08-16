package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/iceisfun/gojs/jsregexp"
)

// JSUnicodeRegexpStepBudget bounds one match against an untrusted rule-option
// pattern. Rule comments are short, and callers may additionally stop trying a
// matcher for the rest of a file after this budget is exhausted.
const JSUnicodeRegexpStepBudget = 100_000

// JSUnicodeRegexpCaptureSlotWorkBudget bounds work the underlying matcher does
// outside its step counter when it copies, restores, or clears capture slots.
// The effective step budget shrinks with the number of capture slots, and an
// additional preflight bounds the slots cleared across unanchored starts.
const JSUnicodeRegexpCaptureSlotWorkBudget = 10_000_000

// JSUnicodeRegexpLookaroundCaptureSlotWorkBudget separately bounds capture
// arrays cloned by zero-width lookarounds. A successful match does not trip the
// file-local step fuse, so this lower limit keeps repeat matches inexpensive.
const JSUnicodeRegexpLookaroundCaptureSlotWorkBudget = 1_000_000

// JSUnicodeRegexpZeroStepWorkBudget also accounts for the zero-step wrappers
// and terms traversed by every nullable branch combination.
const JSUnicodeRegexpZeroStepWorkBudget = 100_000

const (
	// These bounds are deliberately generous for a rule option matched against
	// short source comments. They keep the standalone backtracking compiler
	// from exhausting the process stack or retaining attacker-sized programs.
	JSUnicodeRegexpMaxPatternBytes    = 128 << 10
	JSUnicodeRegexpMaxNestingDepth    = 1_024
	JSUnicodeRegexpMaxGroups          = 4_096
	JSUnicodeRegexpMaxPropertyEscapes = 128
)

// ErrJSUnicodeRegexpResourceLimit marks a syntactically possible pattern that
// rslint refuses to compile because it exceeds a process-safety boundary.
var ErrJSUnicodeRegexpResourceLimit = errors.New("ECMAScript Unicode regular expression exceeds resource limit")

// JSUnicodeRegexp is an immutable ECMAScript RegExp compiled with the `u` flag.
// A compiled value is safe to publish and match concurrently. The underlying
// engine's mutable budget setter is deliberately not exposed.
type JSUnicodeRegexp struct {
	regexp                    *jsregexp.Regexp
	lookaroundCaptureSlotWork int
	zeroStepWork              int
}

// CompileJSUnicodeRegexp compiles pattern with ECMAScript Unicode (`u`)
// semantics and a bounded matcher budget. Its Unicode property data is pinned
// by the jsregexp dependency (Unicode 17 in v0.2.0).
//
// jsregexp v0.2.0 validates named capture identifiers using Go's older Unicode
// tables. normalizeJSRegexpGroupNames first rewrites every valid ECMAScript
// group name and backreference to a generated ASCII name, using the engine's
// own ID_Start / ID_Continue property tables for validation. This preserves
// the original pattern semantics while avoiding that parser-only mismatch.
func CompileJSUnicodeRegexp(pattern string) (*JSUnicodeRegexp, error) {
	_, err := validateJSUnicodeRegexpResourceUsage(pattern)
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeJSRegexpGroupNames(pattern)
	if err != nil {
		return nil, err
	}
	re, err := jsregexp.Compile(normalized, "u")
	if err != nil {
		return nil, err
	}
	zeroStep := jsUnicodeRegexpZeroStepComplexityFor(re.AST().Body)
	if zeroStep.maxWork > JSUnicodeRegexpZeroStepWorkBudget {
		return nil, fmt.Errorf(
			"%w: zero-step nullable work exceeds %d",
			ErrJSUnicodeRegexpResourceLimit,
			JSUnicodeRegexpZeroStepWorkBudget,
		)
	}
	// Capture-heavy lookarounds and quantified groups can copy or reset capture
	// slices on each engine step. Scale the step limit so those operations have
	// the same bounded-work envelope as ordinary patterns. Set the budget exactly
	// once, before the regexp can be shared; Match never mutates this limit.
	captureSlots := 2 * (re.NumSubexp() + 1)
	// gojs lookarounds clone the complete capture slice before their first VM
	// step and restore it on failure. Bound a chain of those otherwise-unmetered
	// operations using the compiler's actual capture count. This also caps the
	// simultaneously retained saved slices at half the slot-operation limit.
	if zeroStep.maxLookarounds > JSUnicodeRegexpLookaroundCaptureSlotWorkBudget/(2*captureSlots) {
		return nil, fmt.Errorf(
			"%w: lookarounds require more than %d capture-slot operations",
			ErrJSUnicodeRegexpResourceLimit,
			JSUnicodeRegexpLookaroundCaptureSlotWorkBudget,
		)
	}
	lookaroundCaptureWork := 2 * zeroStep.maxLookarounds * captureSlots
	stepBudget := min(JSUnicodeRegexpStepBudget, JSUnicodeRegexpCaptureSlotWorkBudget/captureSlots)
	if lookaroundCaptureWork != 0 {
		// A lookaround chain can be revisited by quantified/backtracking paths.
		// Scale again so successful matches cannot repeat the whole chain at the
		// compile-time memory bound on every ordinary engine step.
		stepBudget = min(stepBudget, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget/lookaroundCaptureWork)
	}
	if zeroStep.maxWork != 0 {
		// A quantified or backtracking path can revisit the same zero-step
		// segment. Keep that repeated work inside the same per-match envelope.
		stepBudget = min(stepBudget, JSUnicodeRegexpZeroStepWorkBudget/zeroStep.maxWork)
	}
	re.SetStepBudget(stepBudget)
	return &JSUnicodeRegexp{
		regexp:                    re,
		lookaroundCaptureSlotWork: lookaroundCaptureWork,
		zeroStepWork:              zeroStep.maxWork,
	}, nil
}

// jsUnicodeRegexpZeroStepComplexity models a zero-step matcher segment as the
// affine cost base + paths*continuationCost. This catches both combinatorial
// nullable choices and a large zero-step suffix visited by every choice.
type jsUnicodeRegexpZeroStepComplexity struct {
	paths             int
	work              int
	lookarounds       int
	hasStep           bool
	prefixPaths       int
	prefixWork        int
	prefixLookarounds int
	tailPaths         int
	tailWork          int
	tailLookarounds   int
	maxPaths          int
	maxWork           int
	maxLookarounds    int
}

// jsUnicodeRegexpZeroStepComplexityFor returns the affine zero-step cost for
// node plus the largest path/work segment anywhere below it. Counts saturate
// one past their respective limits. Assertions, consumers, backreferences, and
// non-zero quantifiers charge a step and split segments. Lookarounds are
// conservatively treated as one zero-step path even when their polarity makes
// that path fail, carrying failure exploration into enclosing nodes.
func jsUnicodeRegexpZeroStepComplexityFor(node jsregexp.Node) jsUnicodeRegexpZeroStepComplexity {
	return jsUnicodeRegexpZeroStepComplexityForDirection(node, false)
}

func jsUnicodeRegexpZeroStepComplexityForDirection(node jsregexp.Node, back bool) jsUnicodeRegexpZeroStepComplexity {
	switch node := node.(type) {
	case *jsregexp.Empty:
		return newJSUnicodeRegexpZeroStepComplexity(1, 1, 0)
	case *jsregexp.Group:
		return jsUnicodeRegexpZeroStepComplexityForDirection(node.Body, back)
	case *jsregexp.Capture:
		body := jsUnicodeRegexpZeroStepComplexityForDirection(node.Body, back)
		innerMaxPaths := body.maxPaths
		innerMaxWork := body.maxWork
		if body.paths != 0 {
			// A wholly zero-step body crosses this wrapper once for every
			// nullable path before it invokes the outer continuation.
			wrapperWork := multiplyJSUnicodeRegexpZeroStep(body.paths, 2, JSUnicodeRegexpZeroStepWorkBudget)
			body.work = addJSUnicodeRegexpZeroStep(body.work, wrapperWork, JSUnicodeRegexpZeroStepWorkBudget)
		}
		if body.tailPaths != 0 {
			// A body that charged a step can still end in a zero-step suffix whose
			// paths all cross this wrapper. Preserve that transfer explicitly so an
			// enclosing concat can compose it with its own nullable suffix.
			wrapperWork := multiplyJSUnicodeRegexpZeroStep(body.tailPaths, 2, JSUnicodeRegexpZeroStepWorkBudget)
			body.tailWork = addJSUnicodeRegexpZeroStep(body.tailWork, wrapperWork, JSUnicodeRegexpZeroStepWorkBudget)
		}
		if body.prefixPaths != 0 {
			// A zero-step match before the body's first consumer crosses the
			// capture continuation and restores its slots when the outer
			// continuation fails. Charge those wrapper operations before the next
			// alternative reaches its first metered step.
			wrapperWork := multiplyJSUnicodeRegexpZeroStep(body.prefixPaths, 2, JSUnicodeRegexpZeroStepWorkBudget)
			body.prefixWork = addJSUnicodeRegexpZeroStep(body.prefixWork, wrapperWork, JSUnicodeRegexpZeroStepWorkBudget)
		}
		if innerMaxPaths != 0 {
			// An internal maximum can describe resumption after this capture's
			// outer continuation failed, not just an entry prefix or final tail.
			// That return restores the capture slots before the body reaches its
			// next step. Preserve a conservative, composable upper bound for an
			// enclosing continuation failure.
			wrapperWork := multiplyJSUnicodeRegexpZeroStep(innerMaxPaths, 2, JSUnicodeRegexpZeroStepWorkBudget)
			body.maxWork = max(
				body.maxWork,
				addJSUnicodeRegexpZeroStep(innerMaxWork, wrapperWork, JSUnicodeRegexpZeroStepWorkBudget),
			)
		}
		body.includeJSUnicodeRegexpZeroStepMaxima()
		return body
	case *jsregexp.Disjunction:
		result := newJSUnicodeRegexpZeroStepComplexity(0, 1, 0)
		for _, alternative := range node.Alternatives {
			child := jsUnicodeRegexpZeroStepComplexityForDirection(alternative, back)
			if child.hasStep {
				// Earlier wholly zero-step alternatives are exhausted before this
				// alternative reaches its first consumer. That prefix shares a segment
				// with all of their continuation work.
				result.includeJSUnicodeRegexpZeroStepPrefix(
					addJSUnicodeRegexpZeroStep(result.paths, child.prefixPaths, JSUnicodeRegexpZeroStepWorkBudget),
					addJSUnicodeRegexpZeroStep(result.work, child.prefixWork, JSUnicodeRegexpZeroStepWorkBudget),
					addJSUnicodeRegexpZeroStep(result.lookarounds, child.prefixLookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget),
				)
				result.hasStep = true
				if result.tailPaths != 0 {
					// A stepful earlier alternative can fail, after which this child's
					// zero-step prefix runs before the next metered step.
					result.includeJSUnicodeRegexpZeroStepCandidate(
						addJSUnicodeRegexpZeroStep(result.tailPaths, child.prefixPaths, JSUnicodeRegexpZeroStepWorkBudget),
						addJSUnicodeRegexpZeroStep(result.tailWork, child.prefixWork, JSUnicodeRegexpZeroStepWorkBudget),
						addJSUnicodeRegexpZeroStep(result.tailLookarounds, child.prefixLookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget),
					)
				}
			}
			// If an earlier alternative reached its continuation after a step and
			// that continuation failed, later wholly zero-step alternatives run in
			// the same unmetered segment. Alternatives add rather than multiply.
			var carriedTailPaths, carriedTailWork, carriedTailLookarounds int
			if result.tailPaths != 0 {
				carriedTailPaths = addJSUnicodeRegexpZeroStep(result.tailPaths, child.paths, JSUnicodeRegexpZeroStepWorkBudget)
				carriedTailWork = addJSUnicodeRegexpZeroStep(result.tailWork, child.work, JSUnicodeRegexpZeroStepWorkBudget)
				carriedTailLookarounds = addJSUnicodeRegexpZeroStep(result.tailLookarounds, child.lookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget)
			}
			result.tailPaths = max(carriedTailPaths, child.tailPaths)
			result.tailWork = max(carriedTailWork, child.tailWork)
			result.tailLookarounds = max(carriedTailLookarounds, child.tailLookarounds)
			result.paths = addJSUnicodeRegexpZeroStep(result.paths, child.paths, JSUnicodeRegexpZeroStepWorkBudget)
			result.work = addJSUnicodeRegexpZeroStep(result.work, child.work, JSUnicodeRegexpZeroStepWorkBudget)
			result.lookarounds = addJSUnicodeRegexpZeroStep(result.lookarounds, child.lookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget)
			result.maxPaths = max(result.maxPaths, child.maxPaths)
			result.maxWork = max(result.maxWork, child.maxWork)
			result.maxLookarounds = max(result.maxLookarounds, child.maxLookarounds)
			result.includeJSUnicodeRegexpZeroStepMaxima()
		}
		if result.paths == 0 {
			// work/lookarounds describe paths that reach this node's continuation
			// without a step. Preserve their prefix cost in max*, but don't expose a
			// transfer that cannot actually reach the continuation.
			result.work = 0
			result.lookarounds = 0
		}
		return result
	case *jsregexp.Concat:
		if len(node.Terms) == 0 {
			return newJSUnicodeRegexpZeroStepComplexity(1, 1, 0)
		}
		result := newJSUnicodeRegexpZeroStepComplexity(1, 0, 0)
		if back {
			for i := len(node.Terms) - 1; i >= 0; i-- {
				child := jsUnicodeRegexpZeroStepComplexityForDirection(node.Terms[i], back)
				result = composeJSUnicodeRegexpZeroStep(result, child)
			}
		} else {
			for _, term := range node.Terms {
				child := jsUnicodeRegexpZeroStepComplexityForDirection(term, back)
				result = composeJSUnicodeRegexpZeroStep(result, child)
			}
		}
		return result
	case *jsregexp.Lookaround:
		// A lookahead always executes forward and a lookbehind always executes
		// backward, independent of the surrounding match direction.
		body := jsUnicodeRegexpZeroStepComplexityForDirection(node.Body, node.Behind)
		// Lookarounds are atomic, so at most one body result reaches their
		// caller. Count that result regardless of polarity: a zero-step failure
		// still makes an enclosing disjunction try its next alternative.
		result := newJSUnicodeRegexpZeroStepComplexity(
			1,
			addJSUnicodeRegexpZeroStep(1, body.maxWork, JSUnicodeRegexpZeroStepWorkBudget),
			addJSUnicodeRegexpZeroStep(1, body.maxLookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget),
		)
		result.hasStep = body.hasStep
		if body.hasStep {
			result.prefixPaths = body.prefixPaths
			result.prefixWork = addJSUnicodeRegexpZeroStep(1, body.prefixWork, JSUnicodeRegexpZeroStepWorkBudget)
			result.prefixLookarounds = addJSUnicodeRegexpZeroStep(1, body.prefixLookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget)
		}
		if body.tailPaths != 0 {
			// A step in the atomic body can be the most recent engine step before
			// this lookaround either invokes its outer continuation or fails into an
			// enclosing alternative. The body itself cannot expose more than one
			// match, but all of its failed exploration is charged as work.
			result.tailPaths = 1
			result.tailWork = result.work
			result.tailLookarounds = result.lookarounds
		}
		result.maxPaths = max(result.maxPaths, body.maxPaths)
		result.maxWork = max(result.maxWork, body.maxWork)
		result.maxLookarounds = max(result.maxLookarounds, body.maxLookarounds)
		result.includeJSUnicodeRegexpZeroStepMaxima()
		return result
	case *jsregexp.Quantifier:
		body := jsUnicodeRegexpZeroStepComplexityForDirection(node.Body, back)
		if node.Max == 0 {
			// gojs's simple-consumer fast path calls its continuation directly
			// for `a{0}` without entering the step-counted loop. Treat every
			// zero-maximum shape the same way as a conservative upper bound.
			result := newJSUnicodeRegexpZeroStepComplexity(1, 1, 0)
			result.maxPaths = max(result.maxPaths, body.maxPaths)
			result.maxWork = max(result.maxWork, body.maxWork)
			result.maxLookarounds = max(result.maxLookarounds, body.maxLookarounds)
			return result
		}
		// Other quantifiers charge a step before entering their body. A nullable
		// body may then be explored before the quantifier invokes its outer
		// continuation once, so expose that work as the post-step tail. Candidate
		// continuations from simple quantified consumers each correspond to a
		// separately counted step and therefore intentionally remain one path.
		result := jsUnicodeRegexpZeroStepComplexity{
			hasStep:         true,
			prefixPaths:     1,
			tailPaths:       1,
			tailWork:        body.maxWork,
			tailLookarounds: body.maxLookarounds,
			maxPaths:        body.maxPaths,
			maxWork:         body.maxWork,
			maxLookarounds:  body.maxLookarounds,
		}
		result.includeJSUnicodeRegexpZeroStepMaxima()
		return result
	default:
		// Consumers, assertions, and backreferences charge a step before they
		// can invoke their continuation. Conservatively model the successful
		// edge; a failing edge can only expose later alternatives, which the
		// disjunction recurrence adds to this same tail.
		result := jsUnicodeRegexpZeroStepComplexity{hasStep: true, prefixPaths: 1, tailPaths: 1}
		result.includeJSUnicodeRegexpZeroStepMaxima()
		return result
	}
}

func newJSUnicodeRegexpZeroStepComplexity(paths int, work int, lookarounds int) jsUnicodeRegexpZeroStepComplexity {
	return jsUnicodeRegexpZeroStepComplexity{
		paths:          paths,
		work:           work,
		lookarounds:    lookarounds,
		maxPaths:       paths,
		maxWork:        work,
		maxLookarounds: lookarounds,
	}
}

func composeJSUnicodeRegexpZeroStep(a, b jsUnicodeRegexpZeroStepComplexity) jsUnicodeRegexpZeroStepComplexity {
	paths, work, lookarounds := composeJSUnicodeRegexpZeroStepTransfer(
		a.paths, a.work, a.lookarounds,
		b.paths, b.work, b.lookarounds,
	)
	var carriedTailPaths, carriedTailWork, carriedTailLookarounds int
	if a.tailPaths != 0 && b.paths != 0 {
		carriedTailPaths, carriedTailWork, carriedTailLookarounds = composeJSUnicodeRegexpZeroStepTransfer(
			a.tailPaths, a.tailWork, a.tailLookarounds,
			b.paths, b.work, b.lookarounds,
		)
	}

	var prefixPaths, prefixWork, prefixLookarounds int
	if a.hasStep {
		// The first step can occur inside a, before b is entered at all. Preserve
		// that entry path even when b is not nullable. Any earlier zero-step
		// continuations exposed by a can additionally traverse b and return before
		// the later step in a is attempted.
		prefixPaths = a.prefixPaths
		prefixWork = a.prefixWork
		prefixLookarounds = a.prefixLookarounds
		if a.prefixPaths > 1 && b.paths != 0 {
			continuations := a.prefixPaths - 1
			prefixPaths = addJSUnicodeRegexpZeroStep(
				1,
				multiplyJSUnicodeRegexpZeroStep(continuations, b.paths, JSUnicodeRegexpZeroStepWorkBudget),
				JSUnicodeRegexpZeroStepWorkBudget,
			)
			prefixWork = addJSUnicodeRegexpZeroStep(
				a.prefixWork,
				multiplyJSUnicodeRegexpZeroStep(continuations, b.work, JSUnicodeRegexpZeroStepWorkBudget),
				JSUnicodeRegexpZeroStepWorkBudget,
			)
			prefixLookarounds = addJSUnicodeRegexpZeroStep(
				a.prefixLookarounds,
				multiplyJSUnicodeRegexpZeroStep(continuations, b.lookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget),
				JSUnicodeRegexpLookaroundCaptureSlotWorkBudget,
			)
		}
	}
	if a.paths != 0 && b.hasStep {
		// Or a can complete without a step and invoke b. Every nullable path in
		// a can pay b's prefix before the first consumer in the combined node.
		candidatePaths := multiplyJSUnicodeRegexpZeroStep(a.paths, b.prefixPaths, JSUnicodeRegexpZeroStepWorkBudget)
		candidateWork := addJSUnicodeRegexpZeroStep(
			a.work,
			multiplyJSUnicodeRegexpZeroStep(a.paths, b.prefixWork, JSUnicodeRegexpZeroStepWorkBudget),
			JSUnicodeRegexpZeroStepWorkBudget,
		)
		candidateLookarounds := addJSUnicodeRegexpZeroStep(
			a.lookarounds,
			multiplyJSUnicodeRegexpZeroStep(a.paths, b.prefixLookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget),
			JSUnicodeRegexpLookaroundCaptureSlotWorkBudget,
		)
		prefixPaths = max(prefixPaths, candidatePaths)
		prefixWork = max(prefixWork, candidateWork)
		prefixLookarounds = max(prefixLookarounds, candidateLookarounds)
	}
	result := jsUnicodeRegexpZeroStepComplexity{
		paths:             paths,
		work:              work,
		lookarounds:       lookarounds,
		hasStep:           a.hasStep || b.hasStep,
		prefixPaths:       prefixPaths,
		prefixWork:        prefixWork,
		prefixLookarounds: prefixLookarounds,
		tailPaths:         max(carriedTailPaths, b.tailPaths),
		tailWork:          max(carriedTailWork, b.tailWork),
		tailLookarounds:   max(carriedTailLookarounds, b.tailLookarounds),
		maxPaths:          max(a.maxPaths, b.maxPaths),
		maxWork:           max(a.maxWork, b.maxWork),
		maxLookarounds:    max(a.maxLookarounds, b.maxLookarounds),
	}
	if a.tailPaths != 0 && b.hasStep {
		// A step in a can be followed by b's zero-step prefix before b charges
		// the next step. Compose the prefix affinely across every continuation
		// exposed by a's tail.
		result.includeJSUnicodeRegexpZeroStepCandidate(
			multiplyJSUnicodeRegexpZeroStep(a.tailPaths, b.prefixPaths, JSUnicodeRegexpZeroStepWorkBudget),
			addJSUnicodeRegexpZeroStep(
				a.tailWork,
				multiplyJSUnicodeRegexpZeroStep(a.tailPaths, b.prefixWork, JSUnicodeRegexpZeroStepWorkBudget),
				JSUnicodeRegexpZeroStepWorkBudget,
			),
			addJSUnicodeRegexpZeroStep(
				a.tailLookarounds,
				multiplyJSUnicodeRegexpZeroStep(a.tailPaths, b.prefixLookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget),
				JSUnicodeRegexpLookaroundCaptureSlotWorkBudget,
			),
		)
	}
	result.includeJSUnicodeRegexpZeroStepMaxima()
	if b.tailPaths != 0 {
		// b is a continuation of a. If b's last step is followed by failure, the
		// matcher resumes a's remaining backtracking paths. Those paths can invoke
		// b again before the next step, even when a itself is wholly zero-step.
		// The maxima accumulated above conservatively close over a's internal
		// retries and b's next-entry prefix; add b's final tail to that complete
		// composed segment so the bound remains usable by an enclosing concat.
		result.includeJSUnicodeRegexpZeroStepCandidate(
			addJSUnicodeRegexpZeroStep(b.tailPaths, result.maxPaths, JSUnicodeRegexpZeroStepWorkBudget),
			addJSUnicodeRegexpZeroStep(b.tailWork, result.maxWork, JSUnicodeRegexpZeroStepWorkBudget),
			addJSUnicodeRegexpZeroStep(b.tailLookarounds, result.maxLookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget),
		)
	}
	return result
}

func composeJSUnicodeRegexpZeroStepTransfer(
	aPaths, aWork, aLookarounds int,
	bPaths, bWork, bLookarounds int,
) (paths, work, lookarounds int) {
	if aPaths == 0 || bPaths == 0 {
		return 0, 0, 0
	}
	paths = multiplyJSUnicodeRegexpZeroStep(aPaths, bPaths, JSUnicodeRegexpZeroStepWorkBudget)
	work = addJSUnicodeRegexpZeroStep(
		aWork,
		multiplyJSUnicodeRegexpZeroStep(aPaths, bWork, JSUnicodeRegexpZeroStepWorkBudget),
		JSUnicodeRegexpZeroStepWorkBudget,
	)
	lookarounds = addJSUnicodeRegexpZeroStep(
		aLookarounds,
		multiplyJSUnicodeRegexpZeroStep(aPaths, bLookarounds, JSUnicodeRegexpLookaroundCaptureSlotWorkBudget),
		JSUnicodeRegexpLookaroundCaptureSlotWorkBudget,
	)
	return paths, work, lookarounds
}

func (c *jsUnicodeRegexpZeroStepComplexity) includeJSUnicodeRegexpZeroStepMaxima() {
	c.maxPaths = max(c.maxPaths, c.paths, c.prefixPaths, c.tailPaths)
	c.maxWork = max(c.maxWork, c.work, c.prefixWork, c.tailWork)
	c.maxLookarounds = max(c.maxLookarounds, c.lookarounds, c.prefixLookarounds, c.tailLookarounds)
}

func (c *jsUnicodeRegexpZeroStepComplexity) includeJSUnicodeRegexpZeroStepPrefix(paths, work, lookarounds int) {
	c.prefixPaths = max(c.prefixPaths, paths)
	c.prefixWork = max(c.prefixWork, work)
	c.prefixLookarounds = max(c.prefixLookarounds, lookarounds)
	c.includeJSUnicodeRegexpZeroStepCandidate(paths, work, lookarounds)
}

func (c *jsUnicodeRegexpZeroStepComplexity) includeJSUnicodeRegexpZeroStepCandidate(paths, work, lookarounds int) {
	c.maxPaths = max(c.maxPaths, paths)
	c.maxWork = max(c.maxWork, work)
	c.maxLookarounds = max(c.maxLookarounds, lookarounds)
}

func addJSUnicodeRegexpZeroStep(a, b int, limit int) int {
	if a > limit-b {
		return limit + 1
	}
	return a + b
}

func multiplyJSUnicodeRegexpZeroStep(a, b int, limit int) int {
	if a != 0 && b > limit/a {
		return limit + 1
	}
	return a * b
}

type jsUnicodeRegexpResourceUsage struct {
	groups          int
	maxDepth        int
	propertyEscapes int
}

func scanJSUnicodeRegexpResourceUsage(pattern string) jsUnicodeRegexpResourceUsage {
	var usage jsUnicodeRegexpResourceUsage
	depth := 0
	inClass := false
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '\\':
			if i+1 < len(pattern) {
				if (pattern[i+1] == 'p' || pattern[i+1] == 'P') && i+2 < len(pattern) && pattern[i+2] == '{' {
					usage.propertyEscapes++
				}
				i++
			}
		case '[':
			if !inClass {
				inClass = true
			}
		case ']':
			if inClass {
				inClass = false
			}
		case '(':
			if !inClass {
				usage.groups++
				depth++
				if depth > usage.maxDepth {
					usage.maxDepth = depth
				}
			}
		case ')':
			if !inClass && depth > 0 {
				depth--
			}
		}
	}
	return usage
}

func validateJSUnicodeRegexpResourceUsage(pattern string) (jsUnicodeRegexpResourceUsage, error) {
	if len(pattern) > JSUnicodeRegexpMaxPatternBytes {
		return jsUnicodeRegexpResourceUsage{}, fmt.Errorf("%w: pattern has %d bytes (limit %d)", ErrJSUnicodeRegexpResourceLimit, len(pattern), JSUnicodeRegexpMaxPatternBytes)
	}
	usage := scanJSUnicodeRegexpResourceUsage(pattern)
	switch {
	case usage.maxDepth > JSUnicodeRegexpMaxNestingDepth:
		return usage, fmt.Errorf("%w: nesting depth is %d (limit %d)", ErrJSUnicodeRegexpResourceLimit, usage.maxDepth, JSUnicodeRegexpMaxNestingDepth)
	case usage.groups > JSUnicodeRegexpMaxGroups:
		return usage, fmt.Errorf("%w: pattern has %d groups (limit %d)", ErrJSUnicodeRegexpResourceLimit, usage.groups, JSUnicodeRegexpMaxGroups)
	case usage.propertyEscapes > JSUnicodeRegexpMaxPropertyEscapes:
		return usage, fmt.Errorf("%w: pattern has %d Unicode property escapes (limit %d)", ErrJSUnicodeRegexpResourceLimit, usage.propertyEscapes, JSUnicodeRegexpMaxPropertyEscapes)
	}
	return usage, nil
}

// CountJSUnicodeRegexpPropertyEscapes returns the resource weight used for
// Unicode-property compilation. Escaped backslashes are ignored, matching the
// compiler preflight rather than conservatively counting raw lookalikes.
func CountJSUnicodeRegexpPropertyEscapes(pattern string) int {
	return scanJSUnicodeRegexpResourceUsage(pattern).propertyEscapes
}

// MatchString reports whether the regexp matches anywhere in input. A budget
// error is returned to let callers install a file-local circuit breaker.
func (re *JSUnicodeRegexp) MatchString(input string) (bool, error) {
	if re == nil || re.regexp == nil {
		return false, nil
	}
	units := jsregexp.ToUnits(input)
	captureSlots := 2 * (re.regexp.NumSubexp() + 1)
	if len(units)+1 > JSUnicodeRegexpCaptureSlotWorkBudget/captureSlots {
		return false, jsregexp.ErrBudget
	}
	match, err := re.regexp.FindSubmatchIndex(context.Background(), units, 0)
	return match != nil, err
}

// LookaroundCaptureSlotWork returns a conservative per-match charge for
// capture slices copied outside the engine's step counter.
func (re *JSUnicodeRegexp) LookaroundCaptureSlotWork() int {
	if re == nil {
		return 0
	}
	return re.lookaroundCaptureSlotWork
}

// ZeroStepWork returns the conservative number of unmetered matcher
// operations in the largest zero-step segment.
func (re *JSUnicodeRegexp) ZeroStepWork() int {
	if re == nil {
		return 0
	}
	return re.zeroStepWork
}

// IsJSUnicodeRegexpBudgetError reports whether a match hit its step budget.
func IsJSUnicodeRegexpBudgetError(err error) bool {
	return errors.Is(err, jsregexp.ErrBudget)
}

type jsRegexpGroupNameOccurrence struct {
	start int
	end   int
	name  string
}

func normalizeJSRegexpGroupNames(pattern string) (string, error) {
	if !utf8.ValidString(pattern) {
		return "", errors.New("invalid UTF-8 in ECMAScript regular expression")
	}

	occurrences, err := scanJSRegexpGroupNames(pattern)
	if err != nil || len(occurrences) == 0 {
		return pattern, err
	}

	names := make(map[string]string, len(occurrences))
	for _, occurrence := range occurrences {
		if _, ok := names[occurrence.name]; !ok {
			names[occurrence.name] = "rslintGroup" + strconv.Itoa(len(names))
		}
	}

	var normalized strings.Builder
	normalized.Grow(len(pattern))
	last := 0
	for _, occurrence := range occurrences {
		normalized.WriteString(pattern[last:occurrence.start])
		normalized.WriteString(names[occurrence.name])
		last = occurrence.end
	}
	normalized.WriteString(pattern[last:])
	return normalized.String(), nil
}

func scanJSRegexpGroupNames(pattern string) ([]jsRegexpGroupNameOccurrence, error) {
	var occurrences []jsRegexpGroupNameOccurrence
	inClass := false
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '\\':
			if i+1 >= len(pattern) {
				// Leave the regexp parser to report an unterminated escape.
				i++
				continue
			}
			if !inClass && pattern[i+1] == 'k' && i+2 < len(pattern) && pattern[i+2] == '<' {
				occurrence, next, err := scanJSRegexpGroupName(pattern, i+3)
				if err != nil {
					return nil, err
				}
				occurrences = append(occurrences, occurrence)
				i = next
				continue
			}
			_, width := utf8.DecodeRuneInString(pattern[i+1:])
			i += 1 + width
		case '[':
			if !inClass {
				inClass = true
			}
			i++
		case ']':
			if inClass {
				inClass = false
			}
			i++
		case '(':
			if !inClass && i+3 < len(pattern) && pattern[i+1] == '?' && pattern[i+2] == '<' && pattern[i+3] != '=' && pattern[i+3] != '!' {
				occurrence, next, err := scanJSRegexpGroupName(pattern, i+3)
				if err != nil {
					return nil, err
				}
				occurrences = append(occurrences, occurrence)
				i = next
				continue
			}
			i++
		default:
			_, width := utf8.DecodeRuneInString(pattern[i:])
			i += width
		}
	}
	return occurrences, nil
}

func scanJSRegexpGroupName(pattern string, start int) (jsRegexpGroupNameOccurrence, int, error) {
	end := strings.IndexByte(pattern[start:], '>')
	if end < 0 {
		return jsRegexpGroupNameOccurrence{}, 0, errors.New("unterminated ECMAScript regular expression group name")
	}
	end += start
	name, err := decodeJSRegexpGroupName(pattern[start:end])
	if err != nil {
		return jsRegexpGroupNameOccurrence{}, 0, err
	}
	return jsRegexpGroupNameOccurrence{start: start, end: end, name: name}, end + 1, nil
}

func decodeJSRegexpGroupName(raw string) (string, error) {
	var decoded strings.Builder
	codePoints := make([]rune, 0, utf8.RuneCountInString(raw))
	for i := 0; i < len(raw); {
		var (
			codePoint rune
			next      int
			err       error
		)
		if raw[i] == '\\' {
			codePoint, next, err = decodeJSRegexpUnicodeEscape(raw, i)
		} else {
			codePoint, next = utf8.DecodeRuneInString(raw[i:])
			next += i
		}
		if err != nil {
			return "", err
		}

		if codePoint >= 0xD800 && codePoint <= 0xDBFF {
			if next >= len(raw) || raw[next] != '\\' {
				return "", errors.New("unpaired high surrogate in ECMAScript regular expression group name")
			}
			low, afterLow, lowErr := decodeJSRegexpUnicodeEscape(raw, next)
			if lowErr != nil || low < 0xDC00 || low > 0xDFFF {
				return "", errors.New("unpaired high surrogate in ECMAScript regular expression group name")
			}
			codePoint = 0x10000 + (codePoint-0xD800)<<10 + low - 0xDC00
			next = afterLow
		} else if codePoint >= 0xDC00 && codePoint <= 0xDFFF {
			return "", errors.New("unpaired low surrogate in ECMAScript regular expression group name")
		}

		codePoints = append(codePoints, codePoint)
		decoded.WriteRune(codePoint)
		i = next
	}

	if len(codePoints) == 0 || !isJSRegexpIdentifierStart(codePoints[0]) {
		return "", errors.New("invalid ECMAScript regular expression group name")
	}
	var nonASCIIValidity map[rune]bool
	for _, codePoint := range codePoints[1:] {
		valid := false
		if codePoint <= 0x7F {
			valid = isJSRegexpIdentifierPart(codePoint)
		} else {
			if nonASCIIValidity == nil {
				nonASCIIValidity = make(map[rune]bool)
			}
			var ok bool
			valid, ok = nonASCIIValidity[codePoint]
			if !ok {
				valid = isJSRegexpIdentifierPart(codePoint)
				nonASCIIValidity[codePoint] = valid
			}
		}
		if !valid {
			return "", errors.New("invalid ECMAScript regular expression group name")
		}
	}
	return decoded.String(), nil
}

func decodeJSRegexpUnicodeEscape(raw string, start int) (rune, int, error) {
	if start+2 > len(raw) || raw[start] != '\\' || raw[start+1] != 'u' {
		return 0, 0, errors.New("invalid escape in ECMAScript regular expression group name")
	}
	if start+2 < len(raw) && raw[start+2] == '{' {
		closeOffset := strings.IndexByte(raw[start+3:], '}')
		if closeOffset < 0 {
			return 0, 0, errors.New("unterminated Unicode escape in ECMAScript regular expression group name")
		}
		end := start + 3 + closeOffset
		hex := raw[start+3 : end]
		if len(hex) == 0 {
			return 0, 0, errors.New("invalid Unicode escape in ECMAScript regular expression group name")
		}
		// ECMAScript HexDigits permits arbitrary leading zeroes. Trim only
		// those zeroes before bounding the numeric value's significant width.
		significant := strings.TrimLeft(hex, "0")
		if significant == "" {
			significant = "0"
		}
		if len(significant) > 6 {
			return 0, 0, errors.New("invalid Unicode escape in ECMAScript regular expression group name")
		}
		value, err := strconv.ParseUint(significant, 16, 32)
		if err != nil || value > utf8.MaxRune || value >= 0xD800 && value <= 0xDFFF {
			return 0, 0, errors.New("invalid Unicode escape in ECMAScript regular expression group name")
		}
		return rune(value), end + 1, nil
	}
	if start+6 > len(raw) {
		return 0, 0, errors.New("short Unicode escape in ECMAScript regular expression group name")
	}
	value, err := strconv.ParseUint(raw[start+2:start+6], 16, 16)
	if err != nil {
		return 0, 0, errors.New("invalid Unicode escape in ECMAScript regular expression group name")
	}
	return rune(value), start + 6, nil
}

var (
	jsRegexpIdentifierStart = sync.OnceValue(func() *jsregexp.Regexp {
		return mustCompileJSRegexpIdentifierProperty(`^\p{ID_Start}$`)
	})
	jsRegexpIdentifierPart = sync.OnceValue(func() *jsregexp.Regexp {
		return mustCompileJSRegexpIdentifierProperty(`^\p{ID_Continue}$`)
	})
)

func mustCompileJSRegexpIdentifierProperty(pattern string) *jsregexp.Regexp {
	re := jsregexp.MustCompile(pattern, "u")
	re.SetStepBudget(JSUnicodeRegexpStepBudget)
	return re
}

func isJSRegexpIdentifierStart(codePoint rune) bool {
	if codePoint == '$' || codePoint == '_' || codePoint >= 'A' && codePoint <= 'Z' || codePoint >= 'a' && codePoint <= 'z' {
		return true
	}
	matched, err := jsRegexpIdentifierStart().Match(context.Background(), string(codePoint))
	return err == nil && matched
}

func isJSRegexpIdentifierPart(codePoint rune) bool {
	if codePoint == '$' || codePoint == '_' || codePoint >= 'A' && codePoint <= 'Z' || codePoint >= 'a' && codePoint <= 'z' ||
		codePoint >= '0' && codePoint <= '9' || codePoint == '\u200C' || codePoint == '\u200D' {
		return true
	}
	matched, err := jsRegexpIdentifierPart().Match(context.Background(), string(codePoint))
	return err == nil && matched
}
