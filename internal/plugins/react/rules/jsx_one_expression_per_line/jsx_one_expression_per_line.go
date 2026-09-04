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
	trimmed := strings.TrimRight(source, " ")
	return len(trimmed) < len(source) && !strings.HasSuffix(trimmed, "\n") && !strings.HasSuffix(trimmed, "\r")
}

func isJSXText(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindJsxText || node.Kind == ast.KindJsxTextAllWhiteSpaces)
}

func isWhitespaceOnly(source string) bool {
	return ecmascript.IsBlank(source)
}

// hasLeadingNewline mirrors the upstream `/^\s*\n/` test. The match count in
// the original rule is the length of a non-global match, so this deliberately
// returns a boolean rather than counting every line terminator.
func hasLeadingNewline(source string) bool {
	foundLineFeed := false
	for _, r := range source {
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			return foundLineFeed
		}
		if r == '\n' {
			foundLineFeed = true
		}
	}
	return foundLineFeed
}

// hasTrailingNewline mirrors the upstream `/\n\s*$/` test.
func hasTrailingNewline(source string) bool {
	foundLineFeed := false
	for _, r := range reverseRunes(source) {
		if !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			return foundLineFeed
		}
		if r == '\n' {
			foundLineFeed = true
		}
	}
	return foundLineFeed
}

func reverseRunes(source string) []rune {
	runes := []rune(source)
	for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
		runes[left], runes[right] = runes[right], runes[left]
	}
	return runes
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
	child            *ast.Node
	descriptor       string
	firstChild       bool
	leadingSpace     bool
	leadingSpaceFrom []*ast.Node
	leadingNewline   bool
	trailingSpace    bool
	trailingNewline  bool
}

func hasSuppressedJSXPredecessor(details *fixDetails, detailsByChild map[*ast.Node]*fixDetails) bool {
	for _, previous := range details.leadingSpaceFrom {
		previousDetails := detailsByChild[previous]
		if reactutil.IsJsxLike(previous) && previousDetails != nil &&
			!previousDetails.firstChild && !shouldKeepLeadingSpace(previousDetails, detailsByChild) {
			return true
		}
		if isJSXText(previous) && previousDetails != nil &&
			hasSuppressedJSXPredecessor(previousDetails, detailsByChild) {
			return true
		}
	}
	return false
}

func shouldKeepLeadingSpace(details *fixDetails, detailsByChild map[*ast.Node]*fixDetails) bool {
	for _, previous := range details.leadingSpaceFrom {
		previousDetails := detailsByChild[previous]
		if isJSXText(previous) && previousDetails != nil &&
			hasSuppressedJSXPredecessor(previousDetails, detailsByChild) {
			return false
		}
		if !isJSXText(previous) || previousDetails == nil || !previousDetails.firstChild {
			return true
		}
	}
	return false
}

func shouldKeepTrailingTextSpace(details *fixDetails, detailsByChild map[*ast.Node]*fixDetails) bool {
	if details.firstChild || details.trailingSpace {
		return false
	}
	return !hasSuppressedJSXPredecessor(details, detailsByChild)
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
	openingStartLine := scanner.ComputeLineOfPosition(lineMap, opening.Pos())
	openingEndLine := scanner.ComputeLineOfPosition(lineMap, opening.End())
	closingStartLine := scanner.ComputeLineOfPosition(lineMap, closing.Pos())
	closingEndLine := scanner.ComputeLineOfPosition(lineMap, closing.End())

	if len(children) == 1 {
		child := children[0]
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
				if (isJSXText(previous) && strings.HasSuffix(previousRaw, " ")) ||
					(isJSXText(child) && strings.HasPrefix(childRaw, " ")) ||
					isSpaceBetween(previous, child) {
					details.leadingSpace = true
					details.leadingSpaceFrom = append(details.leadingSpaceFrom, previous)
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
	for _, details := range detailsByChild {
		detailsList = append(detailsList, details)
	}
	sort.Slice(detailsList, func(i, j int) bool {
		return detailsList[i].child.Pos() < detailsList[j].child.Pos()
	})
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
			rawSource := nodeSource(source, child)
			fixedSource := fixSource(rawSource)
			if isJSXText(child) && hasTrailingTextSpace(rawSource) && shouldKeepTrailingTextSpace(&currentDetails, detailsByChild) {
				fixedSource += " "
			}
			leadingSpace := ""
			if currentDetails.leadingSpace && shouldKeepLeadingSpace(&currentDetails, detailsByChild) {
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
	Name:   "react/jsx-one-expression-per-line",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		return rule.RuleListeners{
			ast.KindJsxElement:  func(node *ast.Node) { handleJSX(ctx, node, opts) },
			ast.KindJsxFragment: func(node *ast.Node) { handleJSX(ctx, node, opts) },
		}
	},
}
