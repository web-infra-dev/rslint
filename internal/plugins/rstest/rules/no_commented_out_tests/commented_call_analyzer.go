package no_commented_out_tests

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

var rstestRootCandidate = regexp.MustCompile(`(?m)^[\t ]*[;(]*[\t ]*(test|it|describe)\b`)

type rstestAPIKind uint8

const (
	rstestAPIInvalid rstestAPIKind = iota
	rstestAPITest
	rstestAPIDescribe
)

type rstestAPIPhase uint8

const (
	rstestPhaseInvalid rstestAPIPhase = iota
	rstestPhaseAPI
	rstestPhaseConditionalFactory
	rstestPhaseParameterizedFactory
	rstestPhaseExtendFactory
	rstestPhaseRegistrar
	rstestPhaseRegistered
)

type rstestExpression struct {
	kind       rstestAPIKind
	phase      rstestAPIPhase
	rootOffset int
	extendable bool
}

func invalidRstestExpression() rstestExpression {
	return rstestExpression{}
}

func unwrapRstestExpression(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		default:
			return node
		}
	}
	return nil
}

func staticRstestMemberName(node *ast.Node) (string, *ast.Node, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		name := access.Name()
		if name == nil {
			return "", nil, false
		}
		return name.Text(), access.Expression, true
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		argument := unwrapRstestExpression(access.ArgumentExpression)
		if argument == nil {
			return "", nil, false
		}
		switch argument.Kind {
		case ast.KindStringLiteral:
			return argument.AsStringLiteral().Text, access.Expression, true
		case ast.KindNoSubstitutionTemplateLiteral:
			return argument.AsNoSubstitutionTemplateLiteral().Text, access.Expression, true
		default:
			return "", nil, false
		}
	default:
		return "", nil, false
	}
}

func applyRstestMember(receiver rstestExpression, member string) rstestExpression {
	if receiver.phase != rstestPhaseAPI {
		return invalidRstestExpression()
	}

	switch member {
	case "only", "skip", "todo", "concurrent", "sequential":
		receiver.extendable = false
		return receiver
	case "fails":
		if receiver.kind != rstestAPITest {
			return invalidRstestExpression()
		}
		receiver.extendable = false
		return receiver
	case "runIf", "skipIf":
		receiver.phase = rstestPhaseConditionalFactory
		receiver.extendable = false
		return receiver
	case "each", "for":
		receiver.phase = rstestPhaseParameterizedFactory
		receiver.extendable = false
		return receiver
	case "extend":
		if receiver.kind != rstestAPITest || !receiver.extendable {
			return invalidRstestExpression()
		}
		receiver.phase = rstestPhaseExtendFactory
		return receiver
	default:
		return invalidRstestExpression()
	}
}

func evaluateRstestExpression(node *ast.Node) rstestExpression {
	node = unwrapRstestExpression(node)
	if node == nil {
		return invalidRstestExpression()
	}

	switch node.Kind {
	case ast.KindIdentifier:
		name := node.AsIdentifier().Text
		switch name {
		case "test", "it":
			return rstestExpression{
				kind:       rstestAPITest,
				phase:      rstestPhaseAPI,
				rootOffset: node.Loc.Pos(),
				extendable: true,
			}
		case "describe":
			return rstestExpression{
				kind:       rstestAPIDescribe,
				phase:      rstestPhaseAPI,
				rootOffset: node.Loc.Pos(),
			}
		default:
			return invalidRstestExpression()
		}

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		member, receiverNode, ok := staticRstestMemberName(node)
		if !ok {
			return invalidRstestExpression()
		}
		return applyRstestMember(evaluateRstestExpression(receiverNode), member)

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		expression := evaluateRstestExpression(call.Expression)
		switch expression.phase {
		case rstestPhaseAPI, rstestPhaseRegistrar:
			if call.TypeArguments != nil {
				return invalidRstestExpression()
			}
			expression.phase = rstestPhaseRegistered
			expression.extendable = false
			return expression
		case rstestPhaseConditionalFactory:
			if call.TypeArguments != nil {
				return invalidRstestExpression()
			}
			expression.phase = rstestPhaseAPI
			expression.extendable = false
			return expression
		case rstestPhaseParameterizedFactory:
			expression.phase = rstestPhaseRegistrar
			return expression
		case rstestPhaseExtendFactory:
			expression.phase = rstestPhaseAPI
			expression.extendable = true
			return expression
		default:
			return invalidRstestExpression()
		}

	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		expression := evaluateRstestExpression(tagged.Tag)
		if expression.phase != rstestPhaseParameterizedFactory {
			return invalidRstestExpression()
		}
		expression.phase = rstestPhaseRegistrar
		return expression
	default:
		return invalidRstestExpression()
	}
}

func rootStartsLogicalLine(text string, offset int) bool {
	if offset < 0 || offset > len(text) {
		return false
	}
	lineStart := strings.LastIndexByte(text[:offset], '\n') + 1
	prefix := strings.TrimSpace(text[lineStart:offset])
	for _, char := range prefix {
		if char != '(' && char != ';' {
			return false
		}
	}
	return true
}

func insideMarkdownFence(text string, offset int) bool {
	if offset < 0 || offset > len(text) {
		return false
	}
	inFence := false
	for _, line := range strings.Split(text[:offset], "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
		}
	}
	return inFence
}

func registrationExpressionStatement(node *ast.Node) *ast.Node {
	for node != nil && node.Parent != nil {
		parent := node.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindTypeAssertionExpression,
			ast.KindNonNullExpression,
			ast.KindSatisfiesExpression:
			node = parent
		case ast.KindExpressionStatement:
			return parent
		default:
			return nil
		}
	}
	return nil
}

func registrationHasCleanLineEnd(text string, end int) bool {
	if end < 0 || end > len(text) {
		return false
	}
	lineEnd := len(text)
	if newline := strings.IndexByte(text[end:], '\n'); newline >= 0 {
		lineEnd = end + newline
	}
	trailing := strings.TrimSpace(text[end:lineEnd])
	return trailing == "" ||
		strings.HasPrefix(trailing, ";") ||
		strings.HasPrefix(trailing, "//") ||
		strings.HasPrefix(trailing, "/*")
}

func hasDiagnosticInRange(diagnostics []*ast.Diagnostic, start, end int) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic == nil {
			continue
		}
		// Parser diagnostics for missing tokens have zero-width ranges. Treat
		// their position as intersecting the statement so parser-recovered
		// prose such as `test (see docs)` is not reported as code.
		if diagnostic.Pos() <= end && diagnostic.End() >= start {
			return true
		}
	}
	return false
}

func findRegistrationInCandidate(text string, expectedRootOffset int, scriptKind core.ScriptKind) (int, bool) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/commented-rstest",
		Path:     "/commented-rstest",
	}, text, scriptKind)

	bestOffset := len(text) + 1
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			expression := evaluateRstestExpression(node)
			statement := registrationExpressionStatement(node)
			if expression.phase == rstestPhaseRegistered &&
				expression.rootOffset >= 0 &&
				expression.rootOffset <= expectedRootOffset &&
				!strings.Contains(text[expression.rootOffset:expectedRootOffset], "\n") &&
				expression.rootOffset < bestOffset &&
				statement != nil &&
				!hasDiagnosticInRange(sourceFile.Diagnostics(), statement.Pos(), statement.End()) &&
				registrationHasCleanLineEnd(text, node.End()) {
				bestOffset = expectedRootOffset
			}
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)

	if bestOffset > len(text) {
		return 0, false
	}
	return bestOffset, true
}

func findCommentedRstestRegistrations(text string, scriptKind core.ScriptKind) []int {
	matches := rstestRootCandidate.FindAllStringSubmatchIndex(text, -1)
	offsets := make([]int, 0, len(matches))
	seen := make(map[int]struct{}, len(matches))
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		candidateStart := match[0]
		rootOffset := match[2]
		if insideMarkdownFence(text, rootOffset) {
			continue
		}
		relativeRoot := rootOffset - candidateStart
		foundOffset, found := findRegistrationInCandidate(text[candidateStart:], relativeRoot, scriptKind)
		if !found {
			continue
		}
		absoluteOffset := candidateStart + foundOffset
		if !rootStartsLogicalLine(text, absoluteOffset) {
			continue
		}
		if _, exists := seen[absoluteOffset]; exists {
			continue
		}
		seen[absoluteOffset] = struct{}{}
		offsets = append(offsets, absoluteOffset)
	}
	return offsets
}
