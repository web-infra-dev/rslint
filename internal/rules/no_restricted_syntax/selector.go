package no_restricted_syntax

import (
	"strconv"

	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// selector models the ESLint / esquery selector grammar evaluated against an
// estree.Tree.
type selector interface {
	isSelector()
}

// subjectSelector records esquery's leading `!` marker. For ESLint rule
// listeners the marker only changes matching direction for sibling and
// adjacent combinators; all other matchers treat it as transparent.
type subjectSelector struct {
	Inner selector
}

func (subjectSelector) isSelector() {}

// identifierSelector matches a node by its ESTree TypeID. Wildcard == true is
// `*`. Unknown type names carry TypeID == estree.UnknownType and Wildcard == false.
type identifierSelector struct {
	Name     string
	TypeID   estree.TypeID
	Wildcard bool
}

func (identifierSelector) isSelector() {}

// compoundSelector represents two sequence atoms that must both match the same
// node (e.g. `[kind="const"]VariableDeclaration` or `.callee#Identifier`).
type compoundSelector struct {
	Left  selector
	Right selector
}

func (compoundSelector) isSelector() {}

type pathSpecial uint8

const (
	pathNormal pathSpecial = iota
	pathType
	pathParent
	pathLength
)

// pathSegment is a pre-compiled segment of an attribute or class path.
type pathSegment struct {
	raw     string
	fieldID estree.FieldID
	index   int // canonical non-negative array index, or -1
	special pathSpecial
}

func compilePathSegment(raw string) pathSegment {
	seg := pathSegment{
		raw:   raw,
		index: canonicalArrayIndex(raw),
	}
	switch raw {
	case "type":
		seg.special = pathType
	case "parent":
		seg.special = pathParent
	case "length":
		seg.special = pathLength
	default:
		seg.fieldID, _ = estree.FieldIDByName(raw)
	}
	return seg
}

func compilePath(raw []string) []pathSegment {
	out := make([]pathSegment, len(raw))
	for i, s := range raw {
		out[i] = compilePathSegment(s)
	}
	return out
}

func canonicalArrayIndex(s string) int {
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return -1
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return -1
	}
	return n
}

// classSelector wraps another selector and additionally requires that the
// matched node sits at a specific named field path on its parent (e.g.
// `Literal.key` requires the Literal to be the key of a Property).
type classSelector struct {
	Inner selector
	Path  []pathSegment
}

func (classSelector) isSelector() {}

// attrOp lists the comparison operators allowed inside an attribute filter.
type attrOp int

const (
	attrPresent attrOp = iota
	attrEqual
	attrNotEqual
	attrLess
	attrGreater
	attrLessOrEqual
	attrGreaterOrEqual
)

// attrValueKind discriminates the right-hand side of an attribute comparison.
type attrValueKind int

const (
	attrValueNone   attrValueKind = iota // presence
	attrValueString                      // "x" or 'x'
	attrValueNumber                      // 42
	attrValueRegex                       // /pattern/flags
	attrValueIdent                       // bareword literal (e.g. undefined)
	attrValueType                        // type(string), type(number), ...
)

type attrValue struct {
	Kind          attrValueKind
	Str           string
	Num           float64
	numStr        string // precomputed ecmascript.NumberToString(Num)
	Regex         string
	Flags         string
	Ident         string
	compiledRegex *esregexp.RegExp
	regexPrefix   string
}

// attrSelector matches a node whose attribute (a dotted path inside the
// node's ESTree fields) satisfies a comparison. With Op == attrPresent the
// rule checks that the path resolves to a non-null, non-undefined value.
type attrSelector struct {
	Inner selector
	Path  []pathSegment
	Op    attrOp
	Value attrValue
}

func (attrSelector) isSelector() {}

// combinatorKind enumerates the relations between two selectors in a
// compound expression.
type combinatorKind int

const (
	combDescendant combinatorKind = iota // " " — A B
	combChild                            // > — A > B
	combAdjacent                         // + — A + B (next sibling)
	combSibling                          // ~ — A ~ B (general sibling)
)

// combinatorSelector pairs Right with a combinator and an ancestor / sibling
// constraint Left. memoID is non-zero when per-walk memoization is assigned
// for combDescendant or combSibling (which reserves memoID and memoID+1).
type combinatorSelector struct {
	Kind   combinatorKind
	Left   selector
	Right  selector
	memoID uint32
}

func (combinatorSelector) isSelector() {}

// relativeSelector is the leading-combinator form accepted inside :has(),
// for example `:has(> Identifier)`. esquery models its left side as the exact
// node on which :has() is being evaluated.
type relativeSelector struct {
	Kind  combinatorKind
	Inner selector
}

func (relativeSelector) isSelector() {}

// pseudoSelector covers the `:is(...)`, `:matches(...)`, `:not(...)`,
// `:has(...)`, `:nth-child(N)`, `:nth-last-child(N)`, `:first-child`,
// `:last-child`, and node-class (`:statement`, `:expression`, `:declaration`,
// `:function`, `:pattern`) forms.
type pseudoSelector struct {
	Name   string
	Args   []selector
	N      int // for :nth-child / :nth-last-child
	memoID uint32
}

func (pseudoSelector) isSelector() {}

// unionSelector represents the comma-separated alternative form
// (`A, B, C` matches anything that matches any of the listed selectors).
type unionSelector struct {
	Selectors []selector
}

func (unionSelector) isSelector() {}
