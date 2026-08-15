package loader

import (
	"context"
	"fmt"

	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func sessionForTest(context *buildContext) *Session {
	return &Session{context: context}
}

func buildProjectsForConfigs(
	configs map[string]rslintconfig.RslintConfig,
	singleThreaded bool,
	context *buildContext,
) (ProjectSet, error) {
	return sessionForTest(context).BuildProjects(configs, singleThreaded)
}

func buildProjectsForConfig(
	configDirectory string,
	config rslintconfig.RslintConfig,
	singleThreaded bool,
	context *buildContext,
) (ProjectSet, error) {
	return sessionForTest(context).BuildProject(configDirectory, config, singleThreaded)
}

func executeProjectPlanForTest(
	plan projectPlan,
	singleThreaded bool,
	context *buildContext,
) (ProjectSet, error) {
	return sessionForTest(context).executeProjectPlan(plan, singleThreaded)
}

func resolveTargetPlanForTest(
	configMap map[string]rslintconfig.RslintConfig,
	config rslintconfig.RslintConfig,
	currentDirectory string,
	scopes map[string]rslintconfig.LintDiscoveryScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) (rslintconfig.LintTargetPlan, error) {
	return rslintconfig.ResolveLintTargetPlan(
		configMap,
		config,
		currentDirectory,
		scopes,
		fsys,
		allowFiles,
		allowDirs,
		singleThreaded,
	)
}

func activeConfigsForTest(
	configs map[string]rslintconfig.RslintConfig,
	plan rslintconfig.LintTargetPlan,
) map[string]rslintconfig.RslintConfig {
	return plan.ActiveConfigs(configs)
}

func preferredCallerPathsForTest(plan rslintconfig.LintTargetPlan) map[string]string {
	return plan.PreferredCallerPaths()
}

func loadAPIForTest(
	projects ProjectSet,
	plan rslintconfig.LintTargetPlan,
	currentDirectory string,
	context *buildContext,
	singleThreaded bool,
) (LoadResult, error) {
	return sessionForTest(context).LoadAPI(projects, plan, currentDirectory, singleThreaded)
}

func createCompatibilityProgramForTest(
	rootFileNames []string,
	singleThreaded bool,
	currentDirectory string,
	context *buildContext,
) (*compiler.Program, error) {
	program, err := context.createCompatibilityProgram(
		singleThreaded,
		currentDirectory,
		sourceOnlyCompilerOptions(),
		rootFileNames,
	)
	if err != nil {
		return nil, fmt.Errorf("create compatibility Program for %d lint target(s): %w", len(rootFileNames), err)
	}
	return program, nil
}

func buildRootProgramsForTest(
	groups [][]rslintconfig.DiscoveredLintTarget,
	currentDirectory string,
	context *buildContext,
	singleThreaded bool,
) ([]*lintprogram.Program, []rule.RuleDiagnostic, error) {
	result := LoadResult{}
	if err := sessionForTest(context).appendRootPrograms(&result, groups, currentDirectory, singleThreaded); err != nil {
		return nil, nil, err
	}
	diagnostics := collectTargetSyntacticDiagnostics(result.Programs, result.TargetsByProgram, false, false)
	return result.Programs, diagnostics, nil
}

func collectTargetSyntacticDiagnostics(
	programs []*lintprogram.Program,
	targetsByProgram [][]string,
	typeCheck bool,
	typeCheckOnly bool,
) []rule.RuleDiagnostic {
	type diagnosticKey struct {
		path string
		code int32
		pos  int
		end  int
	}
	seen := make(map[diagnosticKey]struct{})
	var diagnostics []rule.RuleDiagnostic
	for index, sourceProgram := range programs {
		if index >= len(targetsByProgram) {
			continue
		}
		coveredByTypeCheck := typeCheck && sourceProgram.CanProvideProgramDiagnostics()
		for _, target := range targetsByProgram[index] {
			file := sourceProgram.GetSourceFile(target)
			if file == nil {
				continue
			}
			for _, diagnostic := range sourceProgram.SyntacticDiagnostics(context.Background(), file) {
				if coveredByTypeCheck || typeCheckOnly {
					continue
				}
				loc := diagnostic.Loc()
				key := diagnosticKey{file.FileName(), diagnostic.Code(), loc.Pos(), loc.End()}
				if _, exists := seen[key]; exists {
					continue
				}
				seen[key] = struct{}{}
				diagnostics = append(diagnostics, rule.RuleDiagnostic{
					RuleName:     fmt.Sprintf("TypeScript(TS%d)", diagnostic.Code()),
					SourceFile:   file,
					FilePath:     file.FileName(),
					Range:        loc,
					Message:      rule.RuleMessage{Description: diagnostic.String()},
					Severity:     rule.SeverityError,
					Origin:       rule.DiagnosticOriginTypeScript,
					PreFormatted: true,
				})
			}
		}
	}
	return diagnostics
}

func remapDiagnosticTargetPaths(diagnostics []rule.RuleDiagnostic, mapping map[string]string) {
	for index := range diagnostics {
		if target := mapping[diagnostics[index].FilePath]; target != "" {
			diagnostics[index].FilePath = target
		}
	}
}

func deduplicateTypeScriptDiagnostics(
	diagnostics []rule.RuleDiagnostic,
	fsys vfs.FS,
	preferred ...map[string]string,
) []rule.RuleDiagnostic {
	type diagnosticKey struct {
		path string
		code string
		pos  int
		end  int
		text string
	}
	var preferredPaths map[string]string
	if len(preferred) > 0 {
		preferredPaths = preferred[0]
	}
	seen := make(map[diagnosticKey]int)
	result := make([]rule.RuleDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if diagnostic.Origin != rule.DiagnosticOriginTypeScript {
			result = append(result, diagnostic)
			continue
		}
		canonical := canonicalPathID(diagnostic.FilePath, fsys)
		key := diagnosticKey{
			path: canonical,
			code: diagnostic.RuleName,
			pos:  diagnostic.Range.Pos(),
			end:  diagnostic.Range.End(),
			text: diagnostic.Message.Description,
		}
		if existingIndex, exists := seen[key]; exists {
			preferredPath := preferredPaths[canonical]
			if preferredPath != "" && exactPathID(diagnostic.FilePath) == exactPathID(preferredPath) {
				result[existingIndex] = diagnostic
			}
			continue
		}
		seen[key] = len(result)
		result = append(result, diagnostic)
	}
	return result
}

type lintConfigResolverOptions struct {
	Config                     rslintconfig.RslintConfig
	CurrentDirectory           string
	ConfigPathBySourcePath     map[string]string
	OwnerConfigDirBySourcePath map[string]string
	FS                         vfs.FS
}

type testLintConfigResolver struct {
	resolver   *rslintconfig.FileConfigResolver
	pathByFile map[string]string
}

func newLintConfigResolver(opts lintConfigResolverOptions) *testLintConfigResolver {
	return &testLintConfigResolver{
		resolver:   rslintconfig.NewFileConfigResolver(opts.Config, authoritativePath(opts.CurrentDirectory, opts.FS), false),
		pathByFile: opts.ConfigPathBySourcePath,
	}
}

func (resolver *testLintConfigResolver) EnabledRulesForFile(fileName string) []rule.ConfiguredRule {
	if mapped := resolver.pathByFile[fileName]; mapped != "" {
		fileName = mapped
	}
	rules, _ := resolver.resolver.EnabledRulesForFile(fileName)
	return rules
}

func configuredRuleNameSet(rules []rule.ConfiguredRule) map[string]struct{} {
	result := make(map[string]struct{}, len(rules))
	for _, configured := range rules {
		result[configured.Name] = struct{}{}
	}
	return result
}
