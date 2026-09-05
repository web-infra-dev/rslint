package jsx_one_expression_per_line

import (
	_ "embed"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed jsx_one_expression_per_line.schema.json
var schemaJSON []byte

const (
	ruleName         = "react/jsx-one-expression-per-line"
	allowNone        = "none"
	allowLiteral     = "literal"
	allowSingleChild = "single-child"
	allowNonJSX      = "non-jsx"
)

type ruleOptions struct {
	allow string
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{allow: allowNone}
	if len(options) == 0 {
		return opts
	}
	if m, ok := options[0].(map[string]any); ok {
		if value, ok := m["allow"].(string); ok {
			opts.allow = value
		}
	}
	return opts
}

func nodeSource(source string, node *ast.Node) string {
	if node == nil {
		return ""
	}
	start, end := node.Pos(), node.End()
	if start < 0 || end < start || end > len(source) {
		return ""
	}
	return source[start:end]
}

func fixSource(source string) string {
	return strings.TrimRight(strings.TrimLeft(source, " "), " ")
}

func hasTrailingTextSpace(source string) bool {
	return strings.HasSuffix(source, " ") && !hasTrailingNewline(source)
}

func isJSXText(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindJsxText || node.Kind == ast.KindJsxTextAllWhiteSpaces)
}

func isWhitespaceOnly(source string) bool {
	return ecmascript.IsBlank(source)
}

// hasLeadingNewline mirrors the upstream `/^\s*\n/g` test. Its anchored
// match contributes one entry even when it spans several line terminators.
func hasLeadingNewline(source string) bool {
	end := ecmascript.SkipLeadingWhitespace(source, 0, len(source))
	return strings.Contains(source[:end], "\n")
}

// hasTrailingNewline mirrors the upstream `/\n\s*$/` test.
func hasTrailingNewline(source string) bool {
	start := ecmascript.SkipTrailingWhitespace(source, 0, len(source))
	return strings.Contains(source[start:], "\n")
}

func childDescriptor(source string, child *ast.Node) string {
	if child == nil {
		return ""
	}
	var tagName *ast.Node
	switch child.Kind {
	case ast.KindJsxElement:
		opening := child.AsJsxElement().OpeningElement
		if opening != nil {
			tagName = opening.AsJsxOpeningElement().TagName
		}
	case ast.KindJsxSelfClosingElement:
		tagName = child.AsJsxSelfClosingElement().TagName
	}
	if descriptor := reactutil.GetJsxElementTypeString(tagName); descriptor != "" {
		return descriptor
	}
	return strings.ReplaceAll(nodeSource(source, child), "\n", "")
}

func jsxParts(node *ast.Node) (opening, closing *ast.Node) {
	if node == nil {
		return nil, nil
	}
	switch node.Kind {
	case ast.KindJsxElement:
		element := node.AsJsxElement()
		return element.OpeningElement, element.ClosingElement
	case ast.KindJsxFragment:
		fragment := node.AsJsxFragment()
		return fragment.OpeningFragment, fragment.ClosingFragment
	default:
		return nil, nil
	}
}

type childLines struct {
	start int
	end   int
}

func getChildLines(source string, lineMap []core.TextPos, child *ast.Node) (childLines, bool) {
	if child == nil {
		return childLines{}, false
	}
	raw := nodeSource(source, child)
	if isJSXText(child) && isWhitespaceOnly(raw) {
		return childLines{}, false
	}
	lines := childLines{
		start: scanner.ComputeLineOfPosition(lineMap, child.Pos()),
		end:   scanner.ComputeLineOfPosition(lineMap, child.End()),
	}
	if isJSXText(child) {
		if hasLeadingNewline(raw) {
			lines.start++
		}
		if hasTrailingNewline(raw) {
			lines.end--
		}
	}
	return lines, true
}

func isSpaceBetween(previous, current *ast.Node) bool {
	if previous == nil || current == nil || previous.End() > current.Pos() {
		return false
	}
	return previous.End() < current.Pos()
}

type fixDetails struct {
	child *ast.Node

	descriptor      string
	firstChild      bool
	leadingSpace    bool
	leadingNewline  bool
	trailingSpace   bool
	trailingNewline bool

	// leadingSpaceOwn and leadingSpaceGap record the two sources of a leading
	// space that survive the fix: a space the child's own text starts with, and
	// a gap between the child and its predecessor. leadingSpaceFromText records
	// the third source, a preceding text sibling whose raw source ends with a
	// space; that one only survives when the sibling keeps that space.
	leadingSpaceOwn      bool
	leadingSpaceGap      bool
	leadingSpaceFromText []*ast.Node
}

// acceptedFixes mirrors which reports ESLint's fixer applies in its first pass.
// SourceCodeFixer walks the reports in source order and skips any whose range
// starts where the previously applied range ended, so among children that abut
// each other only every other fix lands per pass. eslint-plugin-react relies on
// that: the skipped reports are re-created against the already rewritten source
// in a later pass, which is what decides whether a raw trailing space is
// trimmed away or left in place. Rslint applies abutting fixes in a single
// pass, so the rule has to resolve those decisions up front.
func acceptedFixes(detailsList []*fixDetails) map[*ast.Node]bool {
	accepted := make(map[*ast.Node]bool, len(detailsList))
	lastEnd := -1
	for _, details := range detailsList {
		if details == nil || details.child == nil {
			continue
		}
		if details.child.Pos() > lastEnd {
			accepted[details.child] = true
			lastEnd = details.child.End()
		}
	}
	return accepted
}

// shouldKeepTrailingTextSpace reports whether a text child's raw trailing space
// survives. When the child's own fix lands in the first pass the space is
// trimmed with the rest of the source. When it is skipped, the following
// sibling's fix inserts a line break behind the space first, so the space is no
// longer at the end of the node and every later pass preserves it.
func shouldKeepTrailingTextSpace(details *fixDetails, accepted map[*ast.Node]bool) bool {
	if details == nil {
		// A child the rule never reports keeps its source verbatim.
		return true
	}
	if details.firstChild || details.trailingSpace {
		return false
	}
	return !accepted[details.child]
}

// shouldKeepLeadingSpace reports whether the `{' '}` marker for a leading space
// is still warranted once the surrounding fixes have been applied. A space the
// child's own text starts with, or a gap between the two children, is always
// still there. A space that only comes from a preceding text sibling is there
// only while that sibling keeps it.
func shouldKeepLeadingSpace(details *fixDetails, detailsByChild map[*ast.Node]*fixDetails, accepted map[*ast.Node]bool) bool {
	if details.leadingSpaceOwn || details.leadingSpaceGap {
		return true
	}
	for _, previous := range details.leadingSpaceFromText {
		if shouldKeepTrailingTextSpace(detailsByChild[previous], accepted) {
			return true
		}
	}
	return false
}

func handleJSX(ctx rule.RuleContext, node *ast.Node, options ruleOptions) {
	children := reactutil.GetJsxChildren(node)
	if len(children) == 0 {
		return
	}

	if options.allow == allowNonJSX {
		hasJSXChild := false
		for _, child := range children {
			if reactutil.IsJsxLike(child) {
				hasJSXChild = true
				break
			}
		}
		if !hasJSXChild {
			return
		}
	}

	opening, closing := jsxParts(node)
	if opening == nil || closing == nil {
		return
	}
	lineMap := ctx.SourceFile.ECMALineMap()
	source := ctx.SourceFile.Text()
	openingEndLine := scanner.ComputeLineOfPosition(lineMap, opening.End())
	closingStartLine := scanner.ComputeLineOfPosition(lineMap, closing.Pos())

	if len(children) == 1 && (options.allow == allowSingleChild || options.allow == allowLiteral) {
		child := children[0]
		openingStartLine := scanner.ComputeLineOfPosition(lineMap, scanner.SkipTrivia(source, opening.Pos()))
		closingEndLine := scanner.ComputeLineOfPosition(lineMap, closing.End())
		childStartLine := scanner.ComputeLineOfPosition(lineMap, child.Pos())
		childEndLine := scanner.ComputeLineOfPosition(lineMap, child.End())
		if openingStartLine == openingEndLine &&
			openingEndLine == closingStartLine &&
			closingStartLine == closingEndLine &&
			closingEndLine == childStartLine &&
			childStartLine == childEndLine {
			if options.allow == allowSingleChild ||
				(options.allow == allowLiteral && isJSXText(child)) {
				return
			}
		}
	}

	childrenByLine := make(map[int][]*ast.Node)
	for _, child := range children {
		lines, ok := getChildLines(source, lineMap, child)
		if !ok {
			continue
		}
		childrenByLine[lines.start] = append(childrenByLine[lines.start], child)
		if lines.start != lines.end {
			childrenByLine[lines.end] = append(childrenByLine[lines.end], child)
		}
	}

	lineNumbers := make([]int, 0, len(childrenByLine))
	for line := range childrenByLine {
		lineNumbers = append(lineNumbers, line)
	}
	sort.Ints(lineNumbers)

	detailsByChild := make(map[*ast.Node]*fixDetails)
	for _, line := range lineNumbers {
		lineChildren := childrenByLine[line]
		for i, child := range lineChildren {
			var previous, next *ast.Node
			if i == 0 {
				if line == openingEndLine {
					previous = opening
				}
			} else {
				previous = lineChildren[i-1]
			}
			if i == len(lineChildren)-1 && line == closingStartLine {
				next = closing
			}
			if previous == nil && next == nil {
				continue
			}

			details := detailsByChild[child]
			if details == nil {
				details = &fixDetails{
					child:      child,
					descriptor: childDescriptor(source, child),
				}
				detailsByChild[child] = details
			}
			if previous != nil {
				if previous == opening {
					details.firstChild = true
				}
				previousRaw := nodeSource(source, previous)
				childRaw := nodeSource(source, child)
				if isJSXText(previous) && strings.HasSuffix(previousRaw, " ") {
					details.leadingSpace = true
					details.leadingSpaceFromText = append(details.leadingSpaceFromText, previous)
				}
				if isJSXText(child) && strings.HasPrefix(childRaw, " ") {
					details.leadingSpace = true
					details.leadingSpaceOwn = true
				}
				if isSpaceBetween(previous, child) {
					details.leadingSpace = true
					details.leadingSpaceGap = true
				}
				details.leadingNewline = true
			}
			if next != nil {
				nextRaw := nodeSource(source, next)
				childRaw := nodeSource(source, child)
				details.trailingSpace = details.trailingSpace ||
					(isJSXText(next) && strings.HasPrefix(nextRaw, " ")) ||
					(isJSXText(child) && strings.HasSuffix(childRaw, " ")) ||
					isSpaceBetween(child, next)
				details.trailingNewline = true
			}
		}
	}

	detailsList := make([]*fixDetails, 0, len(detailsByChild))
	// Children are already in source order. Only reports that survive disable
	// directives can participate in ESLint's adjacent-fix selection.
	for _, child := range children {
		details := detailsByChild[child]
		if details == nil {
			continue
		}
		if ctx.DisableManager.IsRuleDisabled(ruleName, child.Pos()) {
			delete(detailsByChild, child)
			continue
		}
		detailsList = append(detailsList, details)
	}
	var accepted map[*ast.Node]bool
	for _, details := range detailsList {
		if details == nil || details.child == nil {
			continue
		}
		currentDetails := *details
		message := rule.RuleMessage{
			Id:          "moveToNewLine",
			Description: "`" + currentDetails.descriptor + "` must be placed on a new line",
			Data:        map[string]string{"descriptor": currentDetails.descriptor},
		}
		child := currentDetails.child
		ctx.ReportRangeWithDeferredFixes(core.NewTextRange(child.Pos(), child.End()), message, func() []rule.RuleFix {
			if accepted == nil {
				accepted = acceptedFixes(detailsList)
			}
			rawSource := nodeSource(source, child)
			fixedSource := fixSource(rawSource)
			if isJSXText(child) && hasTrailingTextSpace(rawSource) && shouldKeepTrailingTextSpace(&currentDetails, accepted) {
				// Preserve the whole trailing run when a following fix protects it.
				fixedSource = strings.TrimLeft(rawSource, " ")
			}
			leadingSpace := ""
			if currentDetails.leadingSpace && shouldKeepLeadingSpace(&currentDetails, detailsByChild, accepted) {
				leadingSpace = "\n{' '}"
			}
			trailingSpace := ""
			if currentDetails.trailingSpace {
				trailingSpace = "{' '}\n"
			}
			leadingNewline := ""
			if currentDetails.leadingNewline {
				leadingNewline = "\n"
			}
			trailingNewline := ""
			if currentDetails.trailingNewline {
				trailingNewline = "\n"
			}
			return []rule.RuleFix{{
				Text:  leadingSpace + leadingNewline + fixedSource + trailingNewline + trailingSpace,
				Range: core.NewTextRange(child.Pos(), child.End()),
			}}
		})
	}
}

var JsxOneExpressionPerLineRule = rule.Rule{
	Name:   ruleName,
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		return rule.RuleListeners{
			ast.KindJsxElement:  func(node *ast.Node) { handleJSX(ctx, node, opts) },
			ast.KindJsxFragment: func(node *ast.Node) { handleJSX(ctx, node, opts) },
		}
	},
}
