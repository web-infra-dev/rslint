// Package padding provides the framework-neutral engine used by test-framework
// rules that require blank lines around selected statement kinds.
package padding

import (
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// StatementType classifies a statement for padding configuration. Framework
// wrappers supply the source names that map to each non-wildcard type.
type StatementType uint8

const (
	statementUnknown StatementType = iota
	StatementAny
	StatementAfterAll
	StatementAfterEach
	StatementBeforeAll
	StatementBeforeEach
	StatementDescribe
	StatementExpect
	StatementTest
)

// PaddingType controls the required spacing for one adjacent statement pair.
type PaddingType uint8

const (
	// PaddingAny accepts the pair without inspecting its spacing. Put a more
	// specific PaddingAny config after a broad PaddingAlways config to exempt
	// pairs such as consecutive expect statements.
	PaddingAny PaddingType = iota
	// PaddingAlways requires at least one blank line between the statements.
	PaddingAlways
)

// StatementTypes is a set of statement classifications matched by a Config.
type StatementTypes []StatementType

// Types constructs a statement-type set for a Config.
func Types(types ...StatementType) StatementTypes {
	return types
}

// Config describes the padding requirement for one adjacent statement pair.
// Configs are matched from last to first, following ESLint's
// padding-line-between-statements precedence.
type Config struct {
	Padding  PaddingType
	Previous StatementTypes
	Next     StatementTypes
}

// StatementNames maps the first identifier token of an expression statement
// to its framework-specific classification. The engine intentionally follows
// the syntax-only contract of the upstream Jest/Vitest padding rules; it does
// not resolve bindings or imports.
type StatementNames map[string]StatementType

var missingPaddingMessage = rule.RuleMessage{
	Id:          "missingPadding",
	Description: "Expected blank line before this statement.",
}

// Definition describes one framework-specific padding rule.
type Definition struct {
	Name string
	// Family groups rules that must not report the same statement boundary
	// twice. The lowest-priority matching rule owns the boundary.
	Family string
	// Priority orders rules within a family. Atomic rules should keep the zero
	// value; aggregate rules should use a larger value so a specific enabled
	// rule owns an overlapping diagnostic.
	Priority int
	Message  rule.RuleMessage
	Names    StatementNames
	Configs  []Config
}

// NewRule creates a no-options padding rule. It copies names and configs so
// package-level tables owned by framework wrappers cannot mutate a live rule.
func NewRule(definition Definition) rule.Rule {
	definition.Names = cloneNames(definition.Names)
	definition.Configs = cloneConfigs(definition.Configs)

	return rule.Rule{
		Name:   definition.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			if definition.Family == "" || len(definition.Names) == 0 || len(definition.Configs) == 0 {
				return nil
			}
			state := rule.CachedByFile(ctx, paddingFileCacheKey{}, func() *engine {
				return &engine{sourceFile: ctx.SourceFile}
			})
			state.requests = append(state.requests, ruleRequest{
				ctx:      ctx,
				name:     definition.Name,
				family:   definition.Family,
				priority: definition.Priority,
				message:  definition.Message,
				names:    definition.Names,
				configs:  definition.Configs,
			})
			if state.listenerRegistered {
				return nil
			}
			state.listenerRegistered = true
			return rule.RuleListeners{
				ast.KindBlock: func(node *ast.Node) {
					if node.Parent != nil && node.Parent.Kind == ast.KindClassStaticBlockDeclaration {
						return
					}
					state.addStatementList(node.Statements())
				},
				ast.KindCaseClause: func(node *ast.Node) {
					state.addStatementList(node.Statements())
				},
				ast.KindDefaultClause: func(node *ast.Node) {
					state.addStatementList(node.Statements())
				},
				rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
					if state.sourceFile != nil && state.sourceFile.Statements != nil {
						state.addStatementList(state.sourceFile.Statements.Nodes)
					}
					state.run()
				},
			}
		},
	}
}

func cloneNames(names StatementNames) StatementNames {
	cloned := make(StatementNames, len(names))
	for name, statementType := range names {
		cloned[name] = statementType
	}
	return cloned
}

func cloneConfigs(configs []Config) []Config {
	cloned := make([]Config, len(configs))
	for i, config := range configs {
		cloned[i] = Config{
			Padding:  config.Padding,
			Previous: slices.Clone(config.Previous),
			Next:     slices.Clone(config.Next),
		}
	}
	return cloned
}

type paddingFileCacheKey struct{}

type ruleRequest struct {
	ctx      rule.RuleContext
	name     string
	family   string
	priority int
	message  rule.RuleMessage
	names    StatementNames
	configs  []Config
}

type engine struct {
	sourceFile         *ast.SourceFile
	requests           []ruleRequest
	pairs              []statementPair
	listenerRegistered bool
}

type statementInfo struct {
	node      *ast.Node
	firstName string
}

type statementPair struct {
	previous statementInfo
	next     statementInfo
}

func (state *engine) run() {
	if state.sourceFile == nil {
		return
	}

	sort.SliceStable(state.pairs, func(i, j int) bool {
		return utils.TrimNodeTextRange(state.sourceFile, state.pairs[i].next.node).Pos() <
			utils.TrimNodeTextRange(state.sourceFile, state.pairs[j].next.node).Pos()
	})
	sort.SliceStable(state.requests, func(i, j int) bool {
		left, right := state.requests[i], state.requests[j]
		if left.family != right.family {
			return strings.Compare(left.family, right.family) < 0
		}
		if left.priority != right.priority {
			return left.priority < right.priority
		}
		return strings.Compare(left.name, right.name) < 0
	})
	for _, pair := range state.pairs {
		reportedFamilies := make(map[string]struct{})
		for _, request := range state.requests {
			if _, reported := reportedFamilies[request.family]; reported {
				continue
			}
			previousType := request.names[pair.previous.firstName]
			nextType := request.names[pair.next.firstName]
			paddingType, matched := matchConfig(previousType, nextType, request.configs)
			if !matched || paddingType == PaddingAny ||
				hasPaddingLine(request.ctx, pair.previous.node, pair.next.node) {
				continue
			}
			nextStart := utils.TrimNodeTextRange(state.sourceFile, pair.next.node).Pos()
			if request.ctx.DisableManager != nil &&
				request.ctx.DisableManager.IsRuleDisabled(request.name, nextStart) {
				continue
			}
			message := request.message
			if message.Id == "" {
				message = missingPaddingMessage
			}
			request.ctx.ReportNodeWithDeferredFixes(pair.next.node, message, func() []rule.RuleFix {
				position, text := paddingInsertion(request.ctx, pair.previous.node, pair.next.node)
				return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(position, position), text)}
			})
			reportedFamilies[request.family] = struct{}{}
		}
	}
}

func (state *engine) addStatementList(statements []*ast.Node) {
	for i := 1; i < len(statements); i++ {
		state.pairs = append(state.pairs, statementPair{
			previous: statementInfo{
				node:      statements[i-1],
				firstName: firstTokenName(state.sourceFile, statements[i-1]),
			},
			next: statementInfo{
				node:      statements[i],
				firstName: firstTokenName(state.sourceFile, statements[i]),
			},
		})
	}
}

func matchConfig(previous, next StatementType, configs []Config) (PaddingType, bool) {
	for i := len(configs) - 1; i >= 0; i-- {
		config := configs[i]
		if matches(previous, config.Previous) && matches(next, config.Next) {
			return config.Padding, true
		}
	}
	return PaddingAny, false
}

func matches(statementType StatementType, candidates StatementTypes) bool {
	return slices.Contains(candidates, StatementAny) || slices.Contains(candidates, statementType)
}

func firstTokenName(sourceFile *ast.SourceFile, statement *ast.Node) string {
	statement = unwrapLabels(statement)
	if statement == nil || statement.Kind != ast.KindExpressionStatement {
		return ""
	}
	expression := statement.AsExpressionStatement().Expression
	expression = utils.ESTreeRuntimeExpression(expression)
	if expression != nil && expression.Kind == ast.KindAwaitExpression {
		expression = utils.ESTreeRuntimeExpression(expression.AsAwaitExpression().Expression)
	}
	if expression == nil {
		return ""
	}
	token, ok := utils.TokenAtOrAfter(sourceFile, utils.TrimNodeTextRange(sourceFile, expression).Pos())
	if !ok || token.Kind != ast.KindIdentifier {
		return ""
	}
	lexicalScanner := scanner.NewScanner()
	lexicalScanner.SetText(token.Text)
	if lexicalScanner.Scan() != ast.KindIdentifier {
		return ""
	}
	return lexicalScanner.TokenValue()
}

func unwrapLabels(statement *ast.Node) *ast.Node {
	for statement != nil && statement.Kind == ast.KindLabeledStatement {
		statement = statement.AsLabeledStatement().Statement
	}
	return statement
}

func hasPaddingLine(ctx rule.RuleContext, previous, next *ast.Node) bool {
	previousToken, ok := actualLastToken(ctx.SourceFile, previous)
	if !ok {
		return false
	}
	nextStart := utils.TrimNodeTextRange(ctx.SourceFile, next).Pos()
	previousEnd := previousToken.End
	for _, comment := range commentsBetween(ctx, previousEnd, nextStart) {
		if linesApart(ctx.SourceFile, previousEnd, comment.Pos()) >= 2 {
			return true
		}
		previousEnd = comment.End()
	}
	return linesApart(ctx.SourceFile, previousEnd, nextStart) >= 2
}

func paddingInsertion(ctx rule.RuleContext, previous, next *ast.Node) (int, string) {
	previousToken, ok := actualLastToken(ctx.SourceFile, previous)
	if !ok {
		position := utils.TrimNodeTextRange(ctx.SourceFile, previous).End()
		return position, "\n"
	}

	nextStart := utils.TrimNodeTextRange(ctx.SourceFile, next).Pos()
	insertAfter := previousToken.End
	nextEntityStart := nextStart
	for _, comment := range commentsBetween(ctx, previousToken.End, nextStart) {
		if linesApart(ctx.SourceFile, insertAfter, comment.Pos()) != 0 {
			nextEntityStart = comment.Pos()
			break
		}
		insertAfter = comment.End()
	}

	if linesApart(ctx.SourceFile, insertAfter, nextEntityStart) == 0 {
		return insertAfter, "\n\n"
	}
	return insertAfter, "\n"
}

func commentsBetween(ctx rule.RuleContext, start, end int) []*ast.CommentRange {
	if ctx.Comments == nil {
		return nil
	}
	return utils.CommentsInSpan(ctx.Comments.All(), start, end)
}

func linesApart(sourceFile *ast.SourceFile, left, right int) int {
	lineMap := sourceFile.ECMALineMap()
	return scanner.ComputeLineOfPosition(lineMap, right) - scanner.ComputeLineOfPosition(lineMap, left)
}

func actualLastToken(sourceFile *ast.SourceFile, statement *ast.Node) (utils.SourceToken, bool) {
	statementRange := utils.TrimNodeTextRange(sourceFile, statement)
	last, ok := utils.TokenBeforePosition(sourceFile, statementRange.End())
	if !ok || last.Kind != ast.KindSemicolonToken {
		return last, ok
	}

	previous, hasPrevious := utils.TokenBeforePosition(sourceFile, last.Start)
	next, hasNext := utils.TokenAtOrAfter(sourceFile, last.End)
	if hasPrevious && hasNext && previous.Start >= statementRange.Pos() &&
		linesApart(sourceFile, previous.End, last.Start) != 0 &&
		linesApart(sourceFile, last.End, next.Start) == 0 {
		return previous, true
	}
	return last, true
}
