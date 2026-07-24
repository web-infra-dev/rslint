package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// EditDemand is a native diagnostic consumer's demand for optional edit
// artifacts. It controls materialization only: diagnostic detection, message,
// range, and severity must not depend on it.
type EditDemand uint8

const EditDemandNone EditDemand = 0

const (
	EditDemandAutofix EditDemand = 1 << iota
	EditDemandSuggestion

	EditDemandAll = EditDemandAutofix | EditDemandSuggestion
)

// IsValid reports whether the demand contains only supported edit kinds.
// Zero is valid and means diagnostics without optional edit artifacts.
func (d EditDemand) IsValid() bool {
	return d&^EditDemandAll == 0
}

// DiagnosticConsumer owns the reporting callback and the optional artifacts
// it consumes. The linter copies one immutable consumer into each native lint
// task; the demand is not a Program property.
type DiagnosticConsumer struct {
	Demand EditDemand
	Report func(RuleDiagnostic)
}

type RuleContext struct {
	SourceFile *ast.SourceFile
	Settings   map[string]interface{}
	// ConfigGlobals contains only globals from the effective
	// `languageOptions.globals` configuration, before inline comments are
	// applied. A false value is an explicit "off" setting.
	ConfigGlobals map[string]bool
	// InlineGlobals contains `/* global */` declaration metadata in first-name
	// source order. Rules can use its exact name ranges without scanning source
	// text again. Treat the slice and its nested ranges as read-only.
	InlineGlobals []InlineGlobal
	// Globals is the fully resolved set of declared global names for this
	// file — config `languageOptions.globals` merged with inline
	// `/* global */` comments, resolved once per file by the linter. Rules
	// should read this instead of parsing comments or config themselves. Nil
	// only when neither source mentions any globals; a name maps to false when
	// its final setting is "off".
	Globals map[string]bool
	// Comments lazily provides every comment in SourceFile, in source order.
	// Rules should call Comments.All instead of walking the token tree with
	// utils.ForEachComment. The first consumer computes the list and every
	// later consumer for this file reuses it.
	Comments *CommentStore
	// Refs lazily provides the per-file identifier-reference index. Rules
	// that need "all references to this declared symbol" should query this
	// instead of walking the AST and calling TypeChecker.GetSymbolAtLocation
	// per identifier. Keys are binder symbols (node.Symbol()); see RefStore.
	// Nil when no program is available.
	Refs           *RefStore
	Program        *compiler.Program
	TypeChecker    *checker.Checker
	DisableManager *DisableManager
	reporter       ruleContextReporter
}

// ruleContextReporter is immutable after Rule.Run starts. Keeping only the
// rule-specific metadata here avoids allocating a family of bound reporting
// closures for every rule context.
type ruleContextReporter struct {
	ruleName string
	severity DiagnosticSeverity
	consumer DiagnosticConsumer
}

// WithReporter returns a context configured to report diagnostics for one
// rule with every edit artifact enabled. Direct rule tests and compatibility
// callers use this method; the native linter uses WithDiagnosticConsumer.
func (ctx RuleContext) WithReporter(
	ruleName string,
	severity DiagnosticSeverity,
	onDiagnostic func(RuleDiagnostic),
) RuleContext {
	return ctx.WithDiagnosticConsumer(ruleName, severity, DiagnosticConsumer{
		Demand: EditDemandAll,
		Report: onDiagnostic,
	})
}

// WithDiagnosticConsumer binds a native diagnostic consumer to one rule
// context. A zero demand is valid and requests diagnostics without edits.
func (ctx RuleContext) WithDiagnosticConsumer(
	ruleName string,
	severity DiagnosticSeverity,
	consumer DiagnosticConsumer,
) RuleContext {
	if ctx.SourceFile == nil {
		panic("rule: reporter requires a source file")
	}
	if consumer.Report == nil {
		panic("rule: reporter requires a diagnostic handler")
	}
	if !consumer.Demand.IsValid() {
		panic("rule: invalid edit demand")
	}
	ctx.reporter = ruleContextReporter{
		ruleName: ruleName,
		severity: severity,
		consumer: consumer,
	}
	return ctx
}

func (ctx *RuleContext) requireReporter() {
	if ctx.reporter.consumer.Report == nil {
		panic("rule: uninitialized RuleContext reporter")
	}
}

// shouldReportRange assumes requireReporter has already run.
func (ctx *RuleContext) shouldReportRange(textRange core.TextRange) bool {
	reporter := &ctx.reporter
	return !ctx.DisableManager.IsRuleDisabled(reporter.ruleName, textRange.Pos())
}

func (ctx *RuleContext) emitRange(textRange core.TextRange, msg RuleMessage, fixes *[]RuleFix, suggestions *[]RuleSuggestion) {
	reporter := &ctx.reporter
	reporter.consumer.Report(RuleDiagnostic{
		RuleName:    reporter.ruleName,
		Range:       textRange,
		Message:     msg,
		FixesPtr:    fixes,
		Suggestions: suggestions,
		SourceFile:  ctx.SourceFile,
		FilePath:    ctx.SourceFile.FileName(),
		Severity:    reporter.severity,
	})
}

// reportRange assumes requireReporter has already run. Keeping validation in
// the public methods makes node reports fail consistently before touching a
// nil SourceFile while paying only one reporter check per diagnostic.
func (ctx *RuleContext) reportRange(textRange core.TextRange, msg RuleMessage, fixes *[]RuleFix, suggestions *[]RuleSuggestion) {
	if !ctx.shouldReportRange(textRange) {
		return
	}
	if ctx.reporter.consumer.Demand&EditDemandAutofix == 0 {
		fixes = nil
	}
	if ctx.reporter.consumer.Demand&EditDemandSuggestion == 0 {
		suggestions = nil
	}
	ctx.emitRange(textRange, msg, fixes, suggestions)
}

func (ctx *RuleContext) reportRangeWithDeferredFixes(textRange core.TextRange, msg RuleMessage, build func() []RuleFix) {
	if !ctx.shouldReportRange(textRange) {
		return
	}

	var fixes *[]RuleFix
	if ctx.reporter.consumer.Demand&EditDemandAutofix != 0 {
		built := build()
		if len(built) > 0 {
			fixes = &built
		}
	}
	ctx.emitRange(textRange, msg, fixes, nil)
}

func (ctx *RuleContext) reportRangeWithDeferredSuggestions(textRange core.TextRange, msg RuleMessage, build func() []RuleSuggestion) {
	if !ctx.shouldReportRange(textRange) {
		return
	}

	var suggestions *[]RuleSuggestion
	if ctx.reporter.consumer.Demand&EditDemandSuggestion != 0 {
		built := build()
		if len(built) > 0 {
			suggestions = &built
		}
	}
	ctx.emitRange(textRange, msg, nil, suggestions)
}

func (ctx *RuleContext) reportRangeWithDeferredFixesAndSuggestions(
	textRange core.TextRange,
	msg RuleMessage,
	buildFixes func() []RuleFix,
	buildSuggestions func() []RuleSuggestion,
) {
	if !ctx.shouldReportRange(textRange) {
		return
	}

	var fixes *[]RuleFix
	if ctx.reporter.consumer.Demand&EditDemandAutofix != 0 {
		built := buildFixes()
		if len(built) > 0 {
			fixes = &built
		}
	}

	var suggestions *[]RuleSuggestion
	if ctx.reporter.consumer.Demand&EditDemandSuggestion != 0 {
		built := buildSuggestions()
		if len(built) > 0 {
			suggestions = &built
		}
	}
	ctx.emitRange(textRange, msg, fixes, suggestions)
}

func (ctx *RuleContext) ReportRange(textRange core.TextRange, msg RuleMessage) {
	ctx.requireReporter()
	ctx.reportRange(textRange, msg, nil, nil)
}

func (ctx *RuleContext) ReportRangeWithFixes(textRange core.TextRange, msg RuleMessage, fixes ...RuleFix) {
	ctx.requireReporter()
	ctx.reportRange(textRange, msg, &fixes, nil)
}

func (ctx *RuleContext) ReportRangeWithSuggestions(textRange core.TextRange, msg RuleMessage, suggestions ...RuleSuggestion) {
	ctx.requireReporter()
	ctx.reportRange(textRange, msg, nil, &suggestions)
}

// ReportRangeWithDeferredFixes reports a diagnostic and constructs its
// autofixes only when the consumer requests them and the diagnostic is not
// suppressed. build runs synchronously and is never retained; an empty result
// means the diagnostic is not autofixable.
func (ctx *RuleContext) ReportRangeWithDeferredFixes(textRange core.TextRange, msg RuleMessage, build func() []RuleFix) {
	ctx.requireReporter()
	if build == nil {
		panic("rule: deferred fixes require a builder")
	}
	ctx.reportRangeWithDeferredFixes(textRange, msg, build)
}

// ReportRangeWithDeferredSuggestions is the suggestion counterpart of
// ReportRangeWithDeferredFixes. An empty result attaches no suggestions.
func (ctx *RuleContext) ReportRangeWithDeferredSuggestions(textRange core.TextRange, msg RuleMessage, build func() []RuleSuggestion) {
	ctx.requireReporter()
	if build == nil {
		panic("rule: deferred suggestions require a builder")
	}
	ctx.reportRangeWithDeferredSuggestions(textRange, msg, build)
}

// ReportRangeWithDeferredFixesAndSuggestions reports one diagnostic whose
// autofixes and suggestions are constructed independently when their matching
// artifact categories are requested. Each requested builder runs synchronously
// after suppression and is never retained; empty results attach no artifact.
func (ctx *RuleContext) ReportRangeWithDeferredFixesAndSuggestions(
	textRange core.TextRange,
	msg RuleMessage,
	buildFixes func() []RuleFix,
	buildSuggestions func() []RuleSuggestion,
) {
	ctx.requireReporter()
	if buildFixes == nil {
		panic("rule: deferred fixes require a builder")
	}
	if buildSuggestions == nil {
		panic("rule: deferred suggestions require a builder")
	}
	ctx.reportRangeWithDeferredFixesAndSuggestions(textRange, msg, buildFixes, buildSuggestions)
}

func (ctx *RuleContext) ReportNode(node *ast.Node, msg RuleMessage) {
	ctx.requireReporter()
	ctx.reportRange(utils.TrimNodeTextRange(ctx.SourceFile, node), msg, nil, nil)
}

func (ctx *RuleContext) ReportNodeWithFixes(node *ast.Node, msg RuleMessage, fixes ...RuleFix) {
	ctx.requireReporter()
	ctx.reportRange(utils.TrimNodeTextRange(ctx.SourceFile, node), msg, &fixes, nil)
}

func (ctx *RuleContext) ReportNodeWithSuggestions(node *ast.Node, msg RuleMessage, suggestions ...RuleSuggestion) {
	ctx.requireReporter()
	ctx.reportRange(utils.TrimNodeTextRange(ctx.SourceFile, node), msg, nil, &suggestions)
}

// ReportNodeWithDeferredFixes is the node-keyed twin of
// ReportRangeWithDeferredFixes.
func (ctx *RuleContext) ReportNodeWithDeferredFixes(node *ast.Node, msg RuleMessage, build func() []RuleFix) {
	ctx.requireReporter()
	if build == nil {
		panic("rule: deferred fixes require a builder")
	}
	ctx.reportRangeWithDeferredFixes(utils.TrimNodeTextRange(ctx.SourceFile, node), msg, build)
}

// ReportNodeWithDeferredSuggestions is the node-keyed twin of
// ReportRangeWithDeferredSuggestions.
func (ctx *RuleContext) ReportNodeWithDeferredSuggestions(node *ast.Node, msg RuleMessage, build func() []RuleSuggestion) {
	ctx.requireReporter()
	if build == nil {
		panic("rule: deferred suggestions require a builder")
	}
	ctx.reportRangeWithDeferredSuggestions(utils.TrimNodeTextRange(ctx.SourceFile, node), msg, build)
}

// ReportNodeWithDeferredFixesAndSuggestions is the node-keyed twin of
// ReportRangeWithDeferredFixesAndSuggestions.
func (ctx *RuleContext) ReportNodeWithDeferredFixesAndSuggestions(
	node *ast.Node,
	msg RuleMessage,
	buildFixes func() []RuleFix,
	buildSuggestions func() []RuleSuggestion,
) {
	ctx.requireReporter()
	if buildFixes == nil {
		panic("rule: deferred fixes require a builder")
	}
	if buildSuggestions == nil {
		panic("rule: deferred suggestions require a builder")
	}
	ctx.reportRangeWithDeferredFixesAndSuggestions(
		utils.TrimNodeTextRange(ctx.SourceFile, node),
		msg,
		buildFixes,
		buildSuggestions,
	)
}

// ReportNodeWithFixesAndSuggestions emits a single diagnostic carrying both
// an autofix and one or more suggestions. It is used by rules that promote a
// suggestion to an autofix while keeping the suggestion available.
func (ctx *RuleContext) ReportNodeWithFixesAndSuggestions(node *ast.Node, msg RuleMessage, fixes []RuleFix, suggestions []RuleSuggestion) {
	ctx.requireReporter()
	ctx.reportRange(utils.TrimNodeTextRange(ctx.SourceFile, node), msg, &fixes, &suggestions)
}

// ReportRangeWithFixesAndSuggestions is the range-keyed twin of
// ReportNodeWithFixesAndSuggestions.
func (ctx *RuleContext) ReportRangeWithFixesAndSuggestions(textRange core.TextRange, msg RuleMessage, fixes []RuleFix, suggestions []RuleSuggestion) {
	ctx.requireReporter()
	ctx.reportRange(textRange, msg, &fixes, &suggestions)
}

func ReportNodeWithFixesOrSuggestions(ctx RuleContext, node *ast.Node, fix bool, msg RuleMessage, suggestionMsg RuleMessage, fixes ...RuleFix) {
	if fix {
		ctx.ReportNodeWithFixes(node, msg, fixes...)
	} else {
		ctx.ReportNodeWithSuggestions(node, msg, RuleSuggestion{
			Message:  suggestionMsg,
			FixesArr: fixes,
		})
	}
}
