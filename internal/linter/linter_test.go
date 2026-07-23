package linter

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// noopRule returns a rule that reports on every identifier (for testing file filtering).
func noopRule() []ConfiguredRule {
	return []ConfiguredRule{
		{
			Name:     "test-rule",
			Severity: rule.SeverityWarning,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return rule.RuleListeners{
					ast.KindIdentifier: func(node *ast.Node) {
						ctx.ReportNode(node, rule.RuleMessage{Id: "test", Description: "test"})
					},
				}
			},
		},
	}
}

// createTestProgramWithFiles creates a TS program in a temp directory with the given files.
// Returns the program and a map of short filename -> normalized absolute path.
func createTestProgramWithFiles(t *testing.T, sourceFiles map[string]string) (*compiler.Program, map[string]string) {
	t.Helper()

	tmpDir := t.TempDir()

	includes := make([]string, 0, len(sourceFiles))
	normalizedPaths := make(map[string]string, len(sourceFiles))
	for name, content := range sourceFiles {
		filePath := filepath.Join(tmpDir, name)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", name, err)
		}
		includes = append(includes, "./"+name)
		normalizedPaths[name] = tspath.NormalizePath(filePath)
	}

	includeJSON := `"` + strings.Join(includes, `","`) + `"`
	tsconfig := `{"include":[` + includeJSON + `]}`
	if err := os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(tsconfig), 0644); err != nil {
		t.Fatalf("Failed to write tsconfig: %v", err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}

	return program, normalizedPaths
}

// tmpDirPath returns the normalized directory path for a file in normalizedPaths.
func tmpDirPath(t *testing.T, normalizedPaths map[string]string, fileName string) string {
	t.Helper()
	return tspath.GetDirectoryPath(normalizedPaths[fileName])
}

func TestRunLinter_ExecutedRules(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	result, err := runLinterPositional([]*compiler.Program{program}, true, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{
				{Name: "rule-a", Severity: rule.SeverityWarning, Run: func(ctx rule.RuleContext) rule.RuleListeners { return nil }},
				{Name: "rule-b", Severity: rule.SeverityWarning, Run: func(ctx rule.RuleContext) rule.RuleListeners { return nil }},
			}
		},
		false, func(d rule.RuleDiagnostic) {}, nil, nil,
	)

	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if _, ok := result.ExecutedRules["rule-a"]; !ok {
		t.Error("Expected rule-a in ExecutedRules")
	}
	if _, ok := result.ExecutedRules["rule-b"]; !ok {
		t.Error("Expected rule-b in ExecutedRules")
	}
	if len(result.ExecutedRules) != 2 {
		t.Errorf("Expected 2 executed rules, got %d", len(result.ExecutedRules))
	}
}

func TestRunLinter_DoesNotExecutePluginPlaceholderInNativePass(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	pluginRunCalled := false

	result, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{paths["a.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:               "community/example",
				IsEslintPluginRule: true,
				Run: func(rule.RuleContext) rule.RuleListeners {
					pluginRunCalled = true
					return nil
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if pluginRunCalled {
		t.Fatal("Node-dispatched plugin placeholder executed in the native pass")
	}
	if _, ok := result.ExecutedRules["community/example"]; !ok {
		t.Fatal("plugin rule should still contribute to the run's rule count")
	}
	if result.LintedFileCount != 1 {
		t.Fatalf("plugin-only target should still count as linted, got %d", result.LintedFileCount)
	}
}

func TestRunLinter_GlobalDeclarationMetadata(t *testing.T) {
	source := "#!/usr/bin/env node\n" +
		"/*global configOn:off, inlineOn, repeated:off */\n" +
		"/*global repeated, inlineOn:off */"
	program, paths := createTestProgramWithFiles(t, map[string]string{"globals.ts": source})
	configGlobals := map[string]bool{"configOn": true, "configOff": false}

	var captured *rule.RuleContext
	result, err := runLinterPositional([]*compiler.Program{program}, true, []string{paths["globals.ts"]}, nil, utils.ExcludePaths,
		func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     "capture-globals",
				Globals:  configGlobals,
				Severity: rule.SeverityWarning,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					captured = &ctx
					return nil
				},
			}}
		},
		false, func(rule.RuleDiagnostic) {}, nil, nil,
	)
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if result.LintedFileCount != 1 || captured == nil {
		t.Fatalf("captured context = %v, linted files = %d; want one", captured != nil, result.LintedFileCount)
	}

	if !reflect.DeepEqual(captured.ConfigGlobals, configGlobals) {
		t.Errorf("ConfigGlobals = %#v, want %#v", captured.ConfigGlobals, configGlobals)
	}
	wantGlobals := map[string]bool{
		"configOn": false, "configOff": false, "inlineOn": false, "repeated": true,
	}
	if !reflect.DeepEqual(captured.Globals, wantGlobals) {
		t.Errorf("Globals = %#v, want %#v", captured.Globals, wantGlobals)
	}

	wantInline := []struct {
		name      string
		declared  bool
		positions []int
	}{
		{name: "configOn", declared: false, positions: []int{strings.Index(source, "configOn")}},
		{name: "inlineOn", declared: false, positions: []int{strings.Index(source, "inlineOn"), strings.LastIndex(source, "inlineOn")}},
		{name: "repeated", declared: true, positions: []int{strings.Index(source, "repeated"), strings.LastIndex(source, "repeated")}},
	}
	if len(captured.InlineGlobals) != len(wantInline) {
		t.Fatalf("InlineGlobals has %d entries, want %d: %#v", len(captured.InlineGlobals), len(wantInline), captured.InlineGlobals)
	}
	for i, want := range wantInline {
		got := captured.InlineGlobals[i]
		if got.Name != want.name || got.Declared != want.declared {
			t.Errorf("InlineGlobals[%d] = (%q, %v), want (%q, %v)", i, got.Name, got.Declared, want.name, want.declared)
		}
		if len(got.NameRanges) != len(want.positions) {
			t.Fatalf("InlineGlobals[%d].NameRanges has %d entries, want %d", i, len(got.NameRanges), len(want.positions))
		}
		for rangeIndex, textRange := range got.NameRanges {
			wantPosition := want.positions[rangeIndex]
			if textRange.Pos() != wantPosition || source[textRange.Pos():textRange.End()] != want.name {
				t.Errorf("InlineGlobals[%d].NameRanges[%d] = %d:%d (%q), want %d:%d (%q)", i, rangeIndex, textRange.Pos(), textRange.End(), source[textRange.Pos():textRange.End()], wantPosition, wantPosition+len(want.name), want.name)
			}
		}
	}
}

func TestRunLinter_ExecutedRulesPerFile(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	// Different files get different rules — ExecutedRules should be the union.
	result, err := runLinterPositional([]*compiler.Program{program}, true, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule {
			if sf.FileName() == paths["a.ts"] {
				return []ConfiguredRule{
					{Name: "only-a", Severity: rule.SeverityWarning, Run: func(ctx rule.RuleContext) rule.RuleListeners { return nil }},
				}
			}
			return []ConfiguredRule{
				{Name: "only-b", Severity: rule.SeverityWarning, Run: func(ctx rule.RuleContext) rule.RuleListeners { return nil }},
			}
		},
		false, func(d rule.RuleDiagnostic) {}, nil, nil,
	)

	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if len(result.ExecutedRules) != 2 {
		t.Errorf("Expected 2 executed rules (union), got %d", len(result.ExecutedRules))
	}
	if _, ok := result.ExecutedRules["only-a"]; !ok {
		t.Error("Expected only-a in ExecutedRules")
	}
	if _, ok := result.ExecutedRules["only-b"]; !ok {
		t.Error("Expected only-b in ExecutedRules")
	}
}

func TestRunLinter_ExecutedRulesAcrossPrograms(t *testing.T) {
	programA, pathsA := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})
	programB, pathsB := createTestProgramWithFiles(t, map[string]string{
		"b.ts": "const b = 2;",
	})

	configuredRule := func(name string) ConfiguredRule {
		return ConfiguredRule{
			Name:     name,
			Severity: rule.SeverityWarning,
			Run:      func(rule.RuleContext) rule.RuleListeners { return nil },
		}
	}
	result, err := RunLinter(RunLinterOptions{
		Programs:    []*compiler.Program{programA, programB},
		TargetFiles: [][]string{{pathsA["a.ts"]}, {pathsB["b.ts"]}},
		GetRulesForFile: func(file *ast.SourceFile) []ConfiguredRule {
			if file.FileName() == pathsA["a.ts"] {
				return []ConfiguredRule{configuredRule("shared"), configuredRule("only-a")}
			}
			return []ConfiguredRule{configuredRule("shared"), configuredRule("only-b")}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if result.LintedFileCount != 2 {
		t.Fatalf("LintedFileCount = %d, want 2", result.LintedFileCount)
	}
	want := map[string]struct{}{
		"shared": {},
		"only-a": {},
		"only-b": {},
	}
	if !reflect.DeepEqual(result.ExecutedRules, want) {
		t.Fatalf("ExecutedRules = %v, want %v", result.ExecutedRules, want)
	}
}

func TestRunLinter_ExecutedRulesEmpty(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	// No rules returned → ExecutedRules should be empty.
	result, err := runLinterPositional([]*compiler.Program{program}, true, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []ConfiguredRule { return nil },
		false, func(d rule.RuleDiagnostic) {}, nil, nil,
	)

	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if len(result.ExecutedRules) != 0 {
		t.Errorf("Expected 0 executed rules, got %d", len(result.ExecutedRules))
	}
	if result.ExecutedRules == nil {
		t.Error("ExecutedRules should be a writable, non-nil empty map")
	}
}

func TestListenerRegistryResetReleasesAndReusesListeners(t *testing.T) {
	registry := newListenerRegistry()
	var calls []string

	registry.add(ast.KindIdentifier, func(*ast.Node) { calls = append(calls, "first") })
	registry.add(ast.KindIdentifier, func(*ast.Node) { calls = append(calls, "second") })
	registry.add(rule.ListenerOnExit(ast.KindIdentifier), func(*ast.Node) { calls = append(calls, "exit") })

	if len(registry.activeKinds) != 2 {
		t.Fatalf("activeKinds has %d entries, want 2", len(registry.activeKinds))
	}
	identifierCapacity := cap(registry.byKind[ast.KindIdentifier])
	registry.reset()

	if len(registry.activeKinds) != 0 {
		t.Fatalf("activeKinds has %d entries after reset, want 0", len(registry.activeKinds))
	}
	for kind, listeners := range registry.byKind {
		if len(listeners) != 0 {
			t.Fatalf("listeners for kind %d have length %d after reset, want 0", kind, len(listeners))
		}
		for index, listener := range listeners[:cap(listeners)] {
			if listener != nil {
				t.Fatalf("listener slot %d for kind %d retained a closure after reset", index, kind)
			}
		}
	}

	registry.add(ast.KindIdentifier, func(*ast.Node) { calls = append(calls, "replacement") })
	if cap(registry.byKind[ast.KindIdentifier]) != identifierCapacity {
		t.Fatalf("identifier listener capacity = %d after reuse, want %d", cap(registry.byKind[ast.KindIdentifier]), identifierCapacity)
	}
	for _, listener := range registry.listeners(ast.KindIdentifier) {
		listener(nil)
	}
	if !reflect.DeepEqual(calls, []string{"replacement"}) {
		t.Fatalf("calls after registry reuse = %v, want only the replacement listener", calls)
	}
}

func TestListenerRegistryIsolationAndRuleOrderAcrossFiles(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const alpha = 1;",
		"b.ts": "const beta = 2;",
	})

	configuredListenerRule := func(kind ast.Kind, name string, severity rule.DiagnosticSeverity) ConfiguredRule {
		return ConfiguredRule{
			Name:     name,
			Severity: severity,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return rule.RuleListeners{
					kind: func(node *ast.Node) {
						ctx.ReportNode(node, rule.RuleMessage{Id: name, Description: name})
					},
				}
			},
		}
	}

	var diagnostics []rule.RuleDiagnostic
	_, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{paths["a.ts"], paths["b.ts"]}},
		GetRulesForFile: func(sourceFile *ast.SourceFile) []ConfiguredRule {
			if sourceFile.FileName() == paths["a.ts"] {
				return []ConfiguredRule{
					configuredListenerRule(ast.KindIdentifier, "a-first", rule.SeverityWarning),
					configuredListenerRule(ast.KindIdentifier, "a-second", rule.SeverityError),
				}
			}
			return []ConfiguredRule{
				configuredListenerRule(ast.KindNumericLiteral, "b-first", rule.SeverityError),
				configuredListenerRule(ast.KindNumericLiteral, "b-second", rule.SeverityWarning),
			}
		},
		OnDiagnostic: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if len(diagnostics) != 4 {
		t.Fatalf("got %d diagnostics, want exactly 4 with no listener leaking between files: %#v", len(diagnostics), diagnostics)
	}

	byFile := make(map[string][]rule.RuleDiagnostic, 2)
	for _, diagnostic := range diagnostics {
		byFile[diagnostic.FilePath] = append(byFile[diagnostic.FilePath], diagnostic)
	}
	for _, testCase := range []struct {
		path       string
		ruleNames  []string
		severities []rule.DiagnosticSeverity
	}{
		{path: paths["a.ts"], ruleNames: []string{"a-first", "a-second"}, severities: []rule.DiagnosticSeverity{rule.SeverityWarning, rule.SeverityError}},
		{path: paths["b.ts"], ruleNames: []string{"b-first", "b-second"}, severities: []rule.DiagnosticSeverity{rule.SeverityError, rule.SeverityWarning}},
	} {
		fileDiagnostics := byFile[testCase.path]
		if len(fileDiagnostics) != 2 {
			t.Fatalf("diagnostics for %s = %#v, want 2", testCase.path, fileDiagnostics)
		}
		for index, diagnostic := range fileDiagnostics {
			if diagnostic.RuleName != testCase.ruleNames[index] || diagnostic.Message.Id != testCase.ruleNames[index] || diagnostic.Severity != testCase.severities[index] {
				t.Errorf("diagnostic %d for %s = (%q, %q, %v), want (%q, %q, %v)", index, testCase.path, diagnostic.RuleName, diagnostic.Message.Id, diagnostic.Severity, testCase.ruleNames[index], testCase.ruleNames[index], testCase.severities[index])
			}
		}
	}
}

func TestRuleContextReporterPreservesDiagnosticSemantics(t *testing.T) {
	const source = "// rslint-disable-next-line reporter-semantics\nconst blocked = 0;\n  const target = 1;\n"
	program, paths := createTestProgramWithFiles(t, map[string]string{"reporter.ts": source})
	sourceFile := program.GetSourceFile(paths["reporter.ts"])
	if sourceFile == nil || sourceFile.Statements == nil || len(sourceFile.Statements.Nodes) != 2 {
		t.Fatal("reporter fixture did not parse into two statements")
	}

	blockedNode := sourceFile.Statements.Nodes[0]
	targetNode := sourceFile.Statements.Nodes[1]
	nodeRange := utils.TrimNodeTextRange(sourceFile, targetNode)
	explicitRange := core.NewTextRange(nodeRange.Pos()+len("const "), nodeRange.Pos()+len("const target"))
	fix := rule.RuleFix{Text: "replacement", Range: explicitRange}
	suggestion := rule.RuleSuggestion{
		Message:  rule.RuleMessage{Id: "suggestion", Description: "suggest replacement"},
		FixesArr: []rule.RuleFix{fix},
	}
	emptyFixes := make([]rule.RuleFix, 0)
	emptySuggestions := make([]rule.RuleSuggestion, 0)
	message := func(id string) rule.RuleMessage {
		return rule.RuleMessage{Id: id, Description: "description-" + id, Data: map[string]string{"id": id}}
	}

	var diagnostics []rule.RuleDiagnostic
	result, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{paths["reporter.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     "reporter-semantics",
				Severity: rule.SeverityWarning,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					ctx.ReportNode(blockedNode, message("disabled"))
					ctx.ReportRange(explicitRange, message("range"))
					ctx.ReportRangeWithFixes(explicitRange, message("range-fixes"), fix)
					ctx.ReportRangeWithSuggestions(explicitRange, message("range-suggestions"), suggestion)
					ctx.ReportRangeWithFixesAndSuggestions(explicitRange, message("range-combined"), []rule.RuleFix{fix}, []rule.RuleSuggestion{suggestion})
					ctx.ReportRangeWithFixesAndSuggestions(explicitRange, message("range-empty-combined"), emptyFixes, emptySuggestions)
					ctx.ReportNode(targetNode, message("node"))
					ctx.ReportNodeWithFixes(targetNode, message("node-empty-fixes"))
					ctx.ReportNodeWithSuggestions(targetNode, message("node-empty-suggestions"))
					ctx.ReportNodeWithFixesAndSuggestions(targetNode, message("node-empty-combined"), nil, nil)
					return nil
				},
			}}
		},
		OnDiagnostic: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
	if result.LintedFileCount != 1 {
		t.Fatalf("LintedFileCount = %d, want 1", result.LintedFileCount)
	}

	type expectation struct {
		id             string
		textRange      core.TextRange
		hasFixes       bool
		fixes          []rule.RuleFix
		hasSuggestions bool
		suggestions    []rule.RuleSuggestion
	}
	expected := []expectation{
		{id: "range", textRange: explicitRange},
		{id: "range-fixes", textRange: explicitRange, hasFixes: true, fixes: []rule.RuleFix{fix}},
		{id: "range-suggestions", textRange: explicitRange, hasSuggestions: true, suggestions: []rule.RuleSuggestion{suggestion}},
		{id: "range-combined", textRange: explicitRange, hasFixes: true, fixes: []rule.RuleFix{fix}, hasSuggestions: true, suggestions: []rule.RuleSuggestion{suggestion}},
		{id: "range-empty-combined", textRange: explicitRange, hasFixes: true, fixes: []rule.RuleFix{}, hasSuggestions: true, suggestions: []rule.RuleSuggestion{}},
		{id: "node", textRange: nodeRange},
		{id: "node-empty-fixes", textRange: nodeRange, hasFixes: true},
		{id: "node-empty-suggestions", textRange: nodeRange, hasSuggestions: true},
		{id: "node-empty-combined", textRange: nodeRange, hasFixes: true, hasSuggestions: true},
	}
	if len(diagnostics) != len(expected) {
		t.Fatalf("got %d diagnostics, want %d (disabled report must be suppressed)", len(diagnostics), len(expected))
	}
	for index, want := range expected {
		got := diagnostics[index]
		if got.RuleName != "reporter-semantics" || got.Severity != rule.SeverityWarning {
			t.Errorf("diagnostic %d metadata = (%q, %v), want (reporter-semantics, warning)", index, got.RuleName, got.Severity)
		}
		if got.SourceFile != sourceFile || got.FilePath != paths["reporter.ts"] {
			t.Errorf("diagnostic %d source = (%v, %q), want fixture source and %q", index, got.SourceFile == sourceFile, got.FilePath, paths["reporter.ts"])
		}
		if got.Range != want.textRange {
			t.Errorf("diagnostic %d range = %v, want %v", index, got.Range, want.textRange)
		}
		if !reflect.DeepEqual(got.Message, message(want.id)) {
			t.Errorf("diagnostic %d message = %#v, want %#v", index, got.Message, message(want.id))
		}
		if (got.FixesPtr != nil) != want.hasFixes {
			t.Errorf("diagnostic %d FixesPtr presence = %v, want %v", index, got.FixesPtr != nil, want.hasFixes)
		} else if want.hasFixes && !reflect.DeepEqual(*got.FixesPtr, want.fixes) {
			t.Errorf("diagnostic %d fixes = %#v, want %#v", index, *got.FixesPtr, want.fixes)
		}
		if (got.Suggestions != nil) != want.hasSuggestions {
			t.Errorf("diagnostic %d Suggestions presence = %v, want %v", index, got.Suggestions != nil, want.hasSuggestions)
		} else if want.hasSuggestions && !reflect.DeepEqual(*got.Suggestions, want.suggestions) {
			t.Errorf("diagnostic %d suggestions = %#v, want %#v", index, *got.Suggestions, want.suggestions)
		}
		if got.Origin != rule.DiagnosticOriginLint || got.PreFormatted {
			t.Errorf("diagnostic %d origin/preformatted = (%v, %v), want lint/false", index, got.Origin, got.PreFormatted)
		}
	}
}
