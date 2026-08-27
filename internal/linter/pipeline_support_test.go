package linter

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func pipelineTestProgram(t *testing.T, root string, fileName string, text string) *program.Program {
	t.Helper()
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{fileName: text})
	result, err := program.NewFromRoots(program.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(root, fs),
		CompilerOptions: &core.CompilerOptions{},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func pipelineTestGeneration(
	t *testing.T,
	root string,
	fileName string,
	text string,
	configuredRules []rule.ConfiguredRule,
	pluginConfig *EslintPluginFileConfig,
) Generation {
	t.Helper()
	result := Generation{
		Native: NativeGeneration{
			Programs:         []*program.Program{pipelineTestProgram(t, root, fileName, text)},
			TargetsByProgram: [][]string{{fileName}},
			SingleThreaded:   true,
			Cwd:              root,
			RulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return configuredRules
			},
		},
		Target: TargetProjection{
			Path: func(string) string { return fileName },
			ReadText: func(_ string, source ast.SourceFileLike) (string, error) {
				return source.Text(), nil
			},
		},
	}
	if pluginConfig != nil {
		result.Plugin = &PluginGeneration{
			ConfigForFile: func(string) EslintPluginFileConfig { return *pluginConfig },
		}
	}
	return result
}

func pipelineTestProvider(generation Generation, release ReleaseFunc) GenerationProvider {
	return GenerationProviderFunc(func(context.Context, SourceSnapshot) (Generation, ReleaseFunc, error) {
		return generation, release, nil
	})
}

func runPipelineWithParallelRuleResolver(
	t *testing.T,
	resolver func(*ast.SourceFile) []rule.ConfiguredRule,
) (recovered any, releases int32) {
	t.Helper()
	root := tspath.NormalizePath(t.TempDir())
	firstPath := tspath.ResolvePath(root, "first.ts")
	secondPath := tspath.ResolvePath(root, "second.ts")
	generation := Generation{Native: NativeGeneration{
		Programs: []*program.Program{
			pipelineTestProgram(t, root, firstPath, "const first = 1;"),
			pipelineTestProgram(t, root, secondPath, "const second = 2;"),
		},
		TargetsByProgram: [][]string{{firstPath}, {secondPath}},
		RulesForFile:     resolver,
	}}
	var releaseCount atomic.Int32
	func() {
		defer func() { recovered = recover() }()
		_, _ = RunPipeline(context.Background(), NewLintRequest(
			pipelineTestProvider(generation, func() { releaseCount.Add(1) }),
			ObservationPolicy{},
			nil,
		))
	}()
	return recovered, releaseCount.Load()
}

type pipelineProgressiveDiagnostics struct {
	baseline  []rule.RuleDiagnostic
	parentCtx context.Context
	run       DeferredPluginRun
	onPublish func()
	onSubmit  func()
}

func (p *pipelineProgressiveDiagnostics) PublishBaseline(
	_ context.Context,
	diagnostics []rule.RuleDiagnostic,
) {
	if p.onPublish != nil {
		p.onPublish()
	}
	p.baseline = append([]rule.RuleDiagnostic(nil), diagnostics...)
}

func (p *pipelineProgressiveDiagnostics) Submit(parentCtx context.Context, run DeferredPluginRun) {
	if p.onSubmit != nil {
		p.onSubmit()
	}
	p.parentCtx = parentCtx
	p.run = run
}

type pipelineAutofixProvider struct {
	t            *testing.T
	root         string
	fileName     string
	initial      string
	next         map[string]string
	acquisitions int
	observed     []string
	rules        func(string) []rule.ConfiguredRule
}

func (p *pipelineAutofixProvider) AcquireGeneration(
	_ context.Context,
	snapshot SourceSnapshot,
) (Generation, ReleaseFunc, error) {
	p.acquisitions++
	current := p.initial
	files := snapshot.Files()
	if len(files) > 0 {
		if len(files) != 1 || files[0].Path != p.fileName {
			return Generation{}, nil, errors.New("unexpected source snapshot")
		}
		current = files[0].Text
	}
	p.observed = append(p.observed, current)
	configuredRules := []rule.ConfiguredRule{}
	if next, ok := p.next[current]; ok {
		before := current
		configuredRules = append(configuredRules, rule.ConfiguredRule{
			Name:     "native/fix",
			Severity: rule.SeverityError,
			Run: func(ruleCtx rule.RuleContext) rule.RuleListeners {
				rangeToFix := core.NewTextRange(0, len(before))
				ruleCtx.ReportRangeWithFixes(
					rangeToFix,
					rule.RuleMessage{Description: "fix"},
					rule.RuleFix{Range: rangeToFix, Text: next},
				)
				return nil
			},
		})
	}
	if p.rules != nil {
		configuredRules = append(configuredRules, p.rules(current)...)
	}
	var pluginConfig *EslintPluginFileConfig
	for _, configured := range configuredRules {
		if configured.IsEslintPluginRule {
			pluginConfig = &EslintPluginFileConfig{}
			break
		}
	}
	return pipelineTestGeneration(p.t, p.root, p.fileName, current, configuredRules, pluginConfig), nil, nil
}

type pipelineFinalChangeRecorder struct {
	commits   int
	committed [][]FileChange
	commit    func(context.Context, []FileChange) (CommitResult, error)
}

func (r *pipelineFinalChangeRecorder) CommitFinalChanges(
	ctx context.Context,
	changes []FileChange,
) (CommitResult, error) {
	r.commits++
	r.committed = append(r.committed, cloneFileChanges(changes))
	if r.commit != nil {
		return r.commit(ctx, changes)
	}
	paths := make([]string, len(changes))
	for index, change := range changes {
		paths[index] = change.Path
	}
	return CommitResult{ConfirmedPaths: paths}, nil
}
