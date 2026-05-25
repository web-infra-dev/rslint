package no_restricted_syntax

// selector models the subset of the ESLint / esquery selector grammar that the
// rule needs in order to evaluate user-supplied patterns. Each variant is a
// node in the parsed selector tree. The matcher in selector_matcher.go walks
// it against tsgo AST nodes.
type selector interface {
	isSelector()
}

// identifierSelector matches a node by its (ESTree) type name. Name == "*"
// is the wildcard. ESTree type names map to one or more tsgo ast.Kind values
// via estreeKindMap in selector_mapping.go.
type identifierSelector struct {
	Name string
}

func (identifierSelector) isSelector() {}

// classSelector wraps another selector and additionally requires that the
// matched node sits at a specific named field on its parent (e.g.
// `Literal.key` requires the Literal to be the key of a Property).
type classSelector struct {
	Inner selector
	Class string
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
	attrValueBool                        // true / false
	attrValueRegex                       // /pattern/flags
	attrValueIdent                       // bareword (e.g. type(undefined))
	attrValueNull                        // null literal
)

type attrValue struct {
	Kind  attrValueKind
	Str   string
	Num   float64
	Bool  bool
	Regex string
	Flags string
	Ident string
}

// attrSelector matches a node whose attribute (a dotted path inside the
// node's logical fields) satisfies a comparison. With Op == attrPresent the
// rule only checks that the value resolves to truthy / non-nil.
type attrSelector struct {
	Inner selector
	Path  []string
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
// constraint Left.
type combinatorSelector struct {
	Kind  combinatorKind
	Left  selector
	Right selector
}

func (combinatorSelector) isSelector() {}

// pseudoSelector covers the `:is(...)`, `:matches(...)`, `:not(...)`,
// `:has(...)`, `:nth-child(N)`, `:nth-last-child(N)`, `:first-child`,
// `:last-child` forms.
type pseudoSelector struct {
	Name string
	Args []selector
	N    int // for :nth-child / :nth-last-child
}

func (pseudoSelector) isSelector() {}

// unionSelector represents the comma-separated alternative form
// (`A, B, C` matches anything that matches any of the listed selectors).
type unionSelector struct {
	Selectors []selector
}

func (unionSelector) isSelector() {}
