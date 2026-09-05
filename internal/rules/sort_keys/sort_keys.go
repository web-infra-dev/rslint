package sort_keys

import (
	_ "embed"
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed sort_keys.schema.json
var schemaJSON []byte

type Options struct {
	Order                    string
	CaseSensitive            bool
	Natural                  bool
	MinKeys                  int
	AllowLineSeparatedGroups bool
	IgnoreComputedKeys       bool
}

func parseOptions(options []any) Options {
	opts := Options{
		Order:                    "asc",
		CaseSensitive:            true,
		Natural:                  false,
		MinKeys:                  2,
		AllowLineSeparatedGroups: false,
		IgnoreComputedKeys:       false,
	}
	if len(options) > 0 {
		if order, ok := options[0].(string); ok {
			opts.Order = order
		}
	}
	if len(options) > 1 {
		optsMap, _ := options[1].(map[string]any)
		if v, ok := optsMap["caseSensitive"].(bool); ok {
			opts.CaseSensitive = v
		}
		if v, ok := optsMap["natural"].(bool); ok {
			opts.Natural = v
		}
		if v, ok := utils.CoerceIntegral(optsMap["minKeys"]); ok {
			opts.MinKeys = v
		}
		if v, ok := optsMap["allowLineSeparatedGroups"].(bool); ok {
			opts.AllowLineSeparatedGroups = v
		}
		if v, ok := optsMap["ignoreComputedKeys"].(bool); ok {
			opts.IgnoreComputedKeys = v
		}
	}
	return opts
}

// propertyStaticName mirrors upstream's getPropertyName: a statically known
// name (literal key, or non-computed identifier key) wins; otherwise, a
// computed key whose expression is a bare identifier falls back to that
// identifier's own name (not its runtime value) — matching upstream's
// `node.key.name` fallback, which reads the Identifier node's `name` field
// regardless of whether the key is computed.
func propertyStaticName(nameNode *ast.Node) (string, bool) {
	if name, ok := utils.GetStaticPropertyName(nameNode); ok {
		return name, true
	}
	if nameNode.Kind == ast.KindComputedPropertyName {
		expr := ast.SkipParentheses(nameNode.AsComputedPropertyName().Expression)
		if expr.Kind == ast.KindIdentifier {
			return expr.AsIdentifier().Text, true
		}
	}
	return "", false
}

// keyReportNode returns the node whose range diagnostics should be reported
// against: the computed key's inner expression (unwrapped of parens), or the
// key node itself when not computed.
func keyReportNode(nameNode *ast.Node) *ast.Node {
	if nameNode.Kind == ast.KindComputedPropertyName {
		return ast.SkipParentheses(nameNode.AsComputedPropertyName().Expression)
	}
	return nameNode
}

func isValidOrder(prevName, thisName string, opts Options) bool {
	a, b := prevName, thisName
	if !opts.CaseSensitive {
		// Upstream lowercases with String.prototype.toLowerCase, whose full
		// case conversion differs from Go's simple one; see
		// [ecmascript.StringToLowerCase].
		a = ecmascript.StringToLowerCase(a)
		b = ecmascript.StringToLowerCase(b)
	}
	var cmp int
	if opts.Natural {
		cmp = utils.NaturalCompare(a, b)
	} else {
		// Upstream asks `a <= b`, which reads both strings by UTF-16 code unit;
		// see [ecmascript.CompareStrings].
		cmp = ecmascript.CompareStrings(a, b)
	}
	if opts.Order == "desc" {
		cmp = -cmp
	}
	return cmp <= 0
}

func buildMessage(opts Options, thisName, prevName string) rule.RuleMessage {
	natural := ""
	if opts.Natural {
		natural = "natural "
	}
	insensitive := ""
	if !opts.CaseSensitive {
		insensitive = "insensitive "
	}
	return rule.RuleMessage{
		Id: "sortKeys",
		Description: fmt.Sprintf(
			"Expected object keys to be in %s%s%sending order. '%s' should be before '%s'.",
			natural, insensitive, opts.Order, thisName, prevName,
		),
	}
}

// containsBlankLine reports whether s (a run of source text between two
// adjacent tokens, with any comment spans already excluded by the caller)
// contains a blank line: two or more line terminators separated only by
// whitespace. Upstream asks the same question of the line numbers its parser
// hands it, and that parser counts a line by every ECMAScript line terminator —
// a lone carriage return and the two Unicode separators among them — so the
// scan here reads the same set rather than `\n` alone.
func containsBlankLine(s string) bool {
	terminators := 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case ecmascript.IsLineTerminator(r):
			// A carriage return and the line feed behind it end one line.
			if r == '\r' && i+size < len(s) && s[i+size] == '\n' {
				size++
			}
			terminators++
			if terminators >= 2 {
				return true
			}
		case ecmascript.IsWhiteSpace(r):
			// Whitespace: neither ends a line nor breaks a run of them.
		default:
			terminators = 0
		}
		i += size
	}
	return false
}

// hasBlankLineBetween reports whether there is a blank line anywhere in
// [start, end), mirroring upstream's token-pair scan (comments included,
// via sourceCode.getTokensBetween(..., {includeComments: true})) without
// walking every token: only comment spans can themselves contain a blank
// line without one being present in the gap, so it's enough to check the
// text outside of comment ranges plus each inter-comment gap. comments is
// ctx.Comments.All() — the whole file's comments, source-ordered — searched
// with a position binary search rather than a fresh scanner pass per call
// (scanning forward from an arbitrary mid-token position, such as a
// property's End() sitting right before its trailing comma, does not
// reliably discover comments further ahead in the gap).
func hasBlankLineBetween(sf *ast.SourceFile, comments []*ast.CommentRange, start, end int) bool {
	if start >= end {
		return false
	}
	text := sf.Text()
	idx := sort.Search(len(comments), func(i int) bool { return comments[i].End() > start })
	pos := start
	for ; idx < len(comments) && comments[idx].Pos() < end; idx++ {
		c := comments[idx]
		if containsBlankLine(text[pos:c.Pos()]) {
			return true
		}
		if c.End() > pos {
			pos = c.End()
		}
	}
	return containsBlankLine(text[pos:end])
}

// objectState tracks sort progress for one ObjectLiteralExpression. Pushed on
// enter and popped on exit so nested object literals get independent state,
// exactly like upstream's stack of { upper, prevNode, prevName, ... } frames.
type objectState struct {
	isPattern     bool
	prevNode      *ast.Node
	prevName      string
	hasPrevName   bool
	prevBlankLine bool
	numKeys       int
}

// isObjectLiteralMember reports whether node is a direct member of an
// ObjectLiteralExpression. PropertyAssignment / ShorthandPropertyAssignment
// only ever appear there, but MethodDeclaration / GetAccessor / SetAccessor
// are shared with class bodies, so the parent must be checked before treating
// a visit as an object-literal property (mirrors upstream's
// `node.parent.type === "ObjectExpression"` / "ObjectPattern" guards, which
// exist for the same reason ESTree's single Property node type needs them).
func isObjectLiteralMember(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression
}

// https://eslint.org/docs/latest/rules/sort-keys
var SortKeysRule = rule.Rule{
	Name:   "sort-keys",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		var stack []*objectState

		handleProperty := func(prop *ast.Node) {
			if !isObjectLiteralMember(prop) || len(stack) == 0 {
				return
			}
			st := stack[len(stack)-1]
			// A destructuring assignment target is an ObjectPattern of its own
			// in ESTree, which upstream returns early on; tsgo spells it with
			// the same ObjectLiteralExpression it uses for a value, so the
			// distinction is drawn here instead.
			if st.isPattern {
				return
			}

			nameNode := prop.Name()
			if nameNode == nil {
				return
			}

			computed := nameNode.Kind == ast.KindComputedPropertyName
			if opts.IgnoreComputedKeys && computed {
				st.hasPrevName = false
				return
			}

			oldPrevName, oldHasPrevName := st.prevName, st.hasPrevName
			thisName, thisOk := propertyStaticName(nameNode)

			// Nothing below reads a blank line unless the option asking about
			// one is on, and the scan for one has the whole file's comment
			// list to materialize first, so it waits to be asked.
			isBlankLineBetweenNodes := false
			if opts.AllowLineSeparatedGroups {
				isBlankLineBetweenNodes = st.prevBlankLine
				if st.prevNode != nil && hasBlankLineBetween(ctx.SourceFile, ctx.Comments.All(), st.prevNode.End(), utils.TrimNodeTextRange(ctx.SourceFile, prop).Pos()) {
					isBlankLineBetweenNodes = true
				}
				st.prevNode = prop
			}

			if thisOk {
				st.prevName, st.hasPrevName = thisName, true
			}

			if isBlankLineBetweenNodes {
				st.prevBlankLine = !thisOk
				return
			}

			if !oldHasPrevName || !thisOk || st.numKeys < opts.MinKeys {
				return
			}

			if !isValidOrder(oldPrevName, thisName, opts) {
				ctx.ReportNode(keyReportNode(nameNode), buildMessage(opts, thisName, oldPrevName))
			}
		}

		return rule.RuleListeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				numKeys := 0
				if objLit := node.AsObjectLiteralExpression(); objLit != nil && objLit.Properties != nil {
					numKeys = len(objLit.Properties.Nodes)
				}
				stack = append(stack, &objectState{numKeys: numKeys, isPattern: ast.IsAssignmentTarget(node)})
			},
			rule.ListenerOnExit(ast.KindObjectLiteralExpression): func(node *ast.Node) {
				stack = stack[:len(stack)-1]
			},
			ast.KindSpreadAssignment: func(node *ast.Node) {
				if isObjectLiteralMember(node) && len(stack) > 0 {
					stack[len(stack)-1].hasPrevName = false
				}
			},
			ast.KindPropertyAssignment:          handleProperty,
			ast.KindShorthandPropertyAssignment: handleProperty,
			ast.KindMethodDeclaration:           handleProperty,
			ast.KindGetAccessor:                 handleProperty,
			ast.KindSetAccessor:                 handleProperty,
		}
	},
}
