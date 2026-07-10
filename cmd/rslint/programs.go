package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// programConfigOrders records which rslint configs declared one Program's
// canonical tsconfig and the declaration order within each config. A shared
// tsconfig has one Program instance but may have multiple config associations.
type programConfigOrders map[string]int

// lintProgramSet is the unique set of real tsconfig-backed Programs used by a
// lint run. ConfigOrders is parallel to Programs. Synthetic fallback Programs
// are appended only while binding a lint target plan and have no config entry.
type lintProgramSet struct {
	Programs     []*compiler.Program
	ConfigOrders []programConfigOrders
}

func filesystemPathID(filePath string, fsys vfs.FS) string {
	useCaseSensitive := true
	if fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", useCaseSensitive))
}

func authoritativeFilesystemPath(filePath string, fsys vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
	}
	return filePath
}

func canonicalFilesystemPathID(filePath string, fsys vfs.FS) string {
	return filesystemPathID(authoritativeFilesystemPath(filePath, fsys), fsys)
}

func relativePathUnderRoot(filePath string, root string, useCaseSensitive bool) (string, bool) {
	filePath = tspath.NormalizePath(filePath)
	root = tspath.NormalizePath(root)
	compareOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          root,
		UseCaseSensitiveFileNames: useCaseSensitive,
	}
	if tspath.ComparePaths(filePath, root, compareOptions) == 0 {
		return "", true
	}
	if !tspath.StartsWithDirectory(filePath, root, useCaseSensitive) {
		return "", false
	}
	return tspath.GetRelativePathFromDirectory(root, filePath, compareOptions), true
}

// configPathForBoundSource returns the path used for files/ignores matching.
// Program source aliases are authoritative over the caller's display path. The
// result is rooted in the physical config directory so a symlinked config root
// and its real path share one stable matching space.
func configPathForBoundSource(sourcePath string, target resolvedLintTarget, fsys vfs.FS) string {
	useCaseSensitive := true
	if fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	ownerRoot := tspath.NormalizePath(target.OwnerConfigDir)
	authoritativeRoot := authoritativeFilesystemPath(ownerRoot, fsys)

	if relative, ok := relativePathUnderRoot(sourcePath, ownerRoot, useCaseSensitive); ok {
		return tspath.ResolvePath(authoritativeRoot, relative)
	}
	canonicalSource := authoritativeFilesystemPath(sourcePath, fsys)
	if relative, ok := relativePathUnderRoot(canonicalSource, authoritativeRoot, useCaseSensitive); ok {
		return tspath.ResolvePath(authoritativeRoot, relative)
	}
	if relative, ok := relativePathUnderRoot(target.Path, ownerRoot, useCaseSensitive); ok {
		return tspath.ResolvePath(authoritativeRoot, relative)
	}
	if relative, ok := relativePathUnderRoot(target.CanonicalPath, authoritativeRoot, useCaseSensitive); ok {
		return tspath.ResolvePath(authoritativeRoot, relative)
	}
	return tspath.NormalizePath(sourcePath)
}

func storeSourcePathMapping(mapping map[string]string, sourcePath string, value string, fsys vfs.FS) {
	if mapping == nil {
		return
	}
	normalizedSource := tspath.NormalizePath(sourcePath)
	mapping[normalizedSource] = value
	mapping[filesystemPathID(normalizedSource, fsys)] = value
}

// createProgramSetForConfigs builds each canonical tsconfig once while
// retaining every config that declared it. Config roots are processed in a
// stable order; target binding later uses the per-config project order rather
// than map iteration or Program construction order.
func createProgramSetForConfigs(
	configMap map[string]rslintconfig.RslintConfig,
	singleThreaded bool,
	fsys vfs.FS,
	parseCache *utils.ParseCache,
) (lintProgramSet, error) {
	if len(configMap) == 0 {
		return lintProgramSet{}, nil
	}

	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	set := lintProgramSet{}
	programByTsconfig := make(map[string]int)
	for _, configDir := range configDirs {
		entries := configMap[configDir]
		tsConfigs, err := rslintconfig.ResolveTsConfigPaths(entries, configDir, fsys)
		if err != nil {
			return lintProgramSet{}, fmt.Errorf("resolve tsconfigs for %q: %w", configDir, err)
		}

		for order, tsconfigPath := range tsConfigs {
			tsconfigID := canonicalFilesystemPathID(tsconfigPath, fsys)
			if programIndex, ok := programByTsconfig[tsconfigID]; ok {
				if _, alreadyAssociated := set.ConfigOrders[programIndex][configDir]; !alreadyAssociated {
					set.ConfigOrders[programIndex][configDir] = order
				}
				continue
			}

			// Build shared Programs from the tsconfig's own directory so the
			// result does not depend on which rslint config happened to declare
			// the tsconfig first.
			programCwd := tspath.GetDirectoryPath(tspath.NormalizePath(tsconfigPath))
			host := utils.WithParseCache(utils.CreateCompilerHost(programCwd, fsys), parseCache)
			program, err := utils.CreateProgramLenient(singleThreaded, fsys, programCwd, tsconfigPath, host)
			if err != nil {
				return lintProgramSet{}, fmt.Errorf("create TypeScript Program from %q: %w", tsconfigPath, err)
			}

			programByTsconfig[tsconfigID] = len(set.Programs)
			set.Programs = append(set.Programs, program)
			set.ConfigOrders = append(set.ConfigOrders, programConfigOrders{configDir: order})
		}
	}

	return set, nil
}

func createProgramSetForConfig(
	configDir string,
	entries rslintconfig.RslintConfig,
	singleThreaded bool,
	fsys vfs.FS,
	parseCache *utils.ParseCache,
) (lintProgramSet, error) {
	return createProgramSetForConfigs(
		map[string]rslintconfig.RslintConfig{configDir: entries},
		singleThreaded,
		fsys,
		parseCache,
	)
}

// parallelGitignoreAndPrograms reads gitignore state and builds the Program
// registry for a single rslint config (single-config and legacy JSON paths).
//
// When singleThreaded is true, both run sequentially in the calling goroutine
// — honoring the user's --singleThreaded flag (no concurrency at all).
// Otherwise the two are dispatched as parallel goroutines: they have no data
// dependency, since Program creation only reads
// entry.LanguageOptions.ParserOptions.Project (see
// LoadTsConfigsFromRslintConfig), never entry.Ignores. Calling it before vs.
// after gitignore globs are prepended is equivalent for TS Program creation.
//
// The returned config is the gitignore-augmented config (gitignore globs
// prepended when non-empty), suitable for downstream target discovery /
// GetConfigForFile.
func parallelGitignoreAndPrograms(
	rslintConfig rslintconfig.RslintConfig,
	configDir string,
	fsys vfs.FS,
	singleThreaded bool,
	parseCache *utils.ParseCache,
) (rslintconfig.RslintConfig, lintProgramSet, error) {
	configIgnores := rslintconfig.ExtractConfigIgnores(rslintConfig)

	var (
		gitGlobs   []string
		programs   lintProgramSet
		programErr error
	)
	// gitignore reading and program creation are independent
	// (Program creation only reads parserOptions.project, never Ignores),
	// so run them on the shared WorkGroup — which honors --singleThreaded the
	// same way the lint and type-check phases do.
	wg := core.NewWorkGroup(singleThreaded)
	wg.Queue(func() {
		gitGlobs = rslintconfig.ReadGitignoreAsGlobs(configDir, fsys, configIgnores)
	})
	wg.Queue(func() {
		programs, programErr = createProgramSetForConfig(configDir, rslintConfig, singleThreaded, fsys, parseCache)
	})
	wg.RunAndWait()

	if programErr != nil {
		return rslintConfig, lintProgramSet{}, programErr
	}
	if len(gitGlobs) > 0 {
		rslintConfig = append(
			rslintconfig.RslintConfig{{Ignores: gitGlobs}},
			rslintConfig...,
		)
	}
	return rslintConfig, programs, nil
}

// createFallbackProgram creates a Program for selected lint targets not
// included in any existing Program. It uses minimal compiler options sufficient
// for AST parsing (no type checking).
func createFallbackProgram(
	gapFiles []string,
	singleThreaded bool,
	configDir string,
	fsys vfs.FS,
	parseCache *utils.ParseCache,
) (*compiler.Program, error) {
	host := utils.WithParseCache(utils.CreateCompilerHost(configDir, fsys), parseCache)
	program, err := utils.CreateProgramFromOptionsLenient(singleThreaded, &core.CompilerOptions{
		Target:    core.ScriptTargetESNext,
		Module:    core.ModuleKindESNext,
		Jsx:       core.JsxEmitPreserve,
		AllowJs:   core.TSTrue,
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, gapFiles, host)
	if err != nil {
		return nil, fmt.Errorf("create fallback Program for %d lint target(s): %w", len(gapFiles), err)
	}
	return program, nil
}

type resolvedLintTarget struct {
	Path           string
	CanonicalPath  string
	OwnerConfigDir string
}

type lintTargetPlan struct {
	Targets []resolvedLintTarget
}

func resolveLintTargetPlan(
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
	configTargetFiles map[string][]string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) (lintTargetPlan, error) {
	type targetWithOwner struct {
		path  string
		owner string
	}
	var targetFiles []targetWithOwner
	if configMap != nil {
		for _, target := range discoverLintFilesMultiConfig(configMap, configTargetFiles, fsys, allowFiles, allowDirs, singleThreaded) {
			targetFiles = append(targetFiles, targetWithOwner{path: target.Path, owner: target.ConfigDirectory})
		}
	} else {
		for _, targetPath := range rslintconfig.DiscoverLintFiles(rslintConfig, currentDirectory, fsys, allowFiles, allowDirs, singleThreaded) {
			targetFiles = append(targetFiles, targetWithOwner{path: targetPath, owner: currentDirectory})
		}
	}

	plan := lintTargetPlan{Targets: make([]resolvedLintTarget, 0, len(targetFiles))}
	seenCanonical := make(map[string]resolvedLintTarget, len(targetFiles))
	for _, discovered := range targetFiles {
		targetPath := discovered.path
		ownerConfigDir := discovered.owner
		canonicalPath := tspath.NormalizePath(targetPath)
		if fsys != nil {
			if realPath := fsys.Realpath(canonicalPath); realPath != "" {
				canonicalPath = tspath.NormalizePath(realPath)
			}
		}
		canonicalKey := canonicalPath
		if fsys != nil {
			canonicalKey = string(tspath.ToPath(canonicalPath, "", fsys.UseCaseSensitiveFileNames()))
		}
		target := resolvedLintTarget{
			Path:           tspath.NormalizePath(targetPath),
			CanonicalPath:  canonicalPath,
			OwnerConfigDir: tspath.NormalizePath(ownerConfigDir),
		}
		if existing, exists := seenCanonical[canonicalKey]; exists {
			if filesystemPathID(existing.OwnerConfigDir, fsys) != filesystemPathID(target.OwnerConfigDir, fsys) {
				return lintTargetPlan{}, fmt.Errorf(
					"lint target aliases %q and %q resolve to the same file but are governed by different configs %q and %q",
					existing.Path,
					target.Path,
					existing.OwnerConfigDir,
					target.OwnerConfigDir,
				)
			}
			continue
		}
		seenCanonical[canonicalKey] = target
		plan.Targets = append(plan.Targets, target)
	}
	return plan, nil
}

func preferredCallerTargetPaths(plan lintTargetPlan, fsys vfs.FS) map[string]string {
	if len(plan.Targets) == 0 {
		return nil
	}
	preferred := make(map[string]string, len(plan.Targets))
	for _, target := range plan.Targets {
		canonicalID := canonicalFilesystemPathID(target.CanonicalPath, fsys)
		if _, exists := preferred[canonicalID]; !exists {
			preferred[canonicalID] = target.Path
		}
	}
	return preferred
}

type lintTargetBinding struct {
	Programs                   []*compiler.Program
	TypeInfoFiles              map[string]struct{}
	GapFiles                   []string
	TargetsByProgram           [][]string
	TargetPathBySourcePath     map[string]string
	ConfigPathBySourcePath     map[string]string
	OwnerConfigDirBySourcePath map[string]string
}

func aliasPathUnderRoot(path string, canonicalRoot string, lexicalRoot string, useCaseSensitive bool) string {
	if path == "" || canonicalRoot == "" || lexicalRoot == "" {
		return ""
	}
	relative, ok := relativePathUnderRoot(path, canonicalRoot, useCaseSensitive)
	if !ok {
		return ""
	}
	return tspath.ResolvePath(lexicalRoot, relative)
}

type programTargetLookup struct {
	program             *compiler.Program
	fsys                vfs.FS
	canonicalIndexBuilt bool
	canonicalSources    map[string]*ast.SourceFile
}

func (lookup *programTargetLookup) canonicalSourceFile(canonicalPath string) *ast.SourceFile {
	if lookup.fsys == nil || canonicalPath == "" {
		return nil
	}
	if !lookup.canonicalIndexBuilt {
		lookup.canonicalIndexBuilt = true
		lookup.canonicalSources = make(map[string]*ast.SourceFile)
		for _, sourceFile := range lookup.program.GetSourceFiles() {
			canonicalID := canonicalFilesystemPathID(sourceFile.FileName(), lookup.fsys)
			existing := lookup.canonicalSources[canonicalID]
			if existing == nil || sourceFile.FileName() < existing.FileName() {
				lookup.canonicalSources[canonicalID] = sourceFile
			}
		}
	}
	return lookup.canonicalSources[canonicalFilesystemPathID(canonicalPath, lookup.fsys)]
}

// sourceFileForTarget first performs bounded target-driven lookups. Only when
// every lexical/root alias misses does it build one lazy canonical source index
// for the Program, covering file-level symlinks without imposing graph-wide
// realpath work on normal projects.
func (lookup *programTargetLookup) sourceFileForTarget(target resolvedLintTarget) *ast.SourceFile {
	program := lookup.program
	fsys := lookup.fsys
	if program == nil {
		return nil
	}
	// Normal projects hit the lexical target immediately. Keep realpath work
	// entirely off this hot path; aliases are only derived after both direct
	// target forms miss.
	targetPath := tspath.NormalizePath(target.Path)
	if sourceFile := program.GetSourceFile(targetPath); sourceFile != nil {
		return sourceFile
	}
	canonicalPath := ""
	if target.CanonicalPath != "" {
		canonicalPath = tspath.NormalizePath(target.CanonicalPath)
	}
	if canonicalPath != "" && canonicalPath != targetPath {
		if sourceFile := program.GetSourceFile(canonicalPath); sourceFile != nil {
			return sourceFile
		}
	}
	if fsys == nil {
		return nil
	}
	if canonicalPath == "" {
		canonicalPath = targetPath
	}
	if canonicalPath == "" {
		return nil
	}

	useCaseSensitive := program.UseCaseSensitiveFileNames()
	lookupAlias := func(candidate string) *ast.SourceFile {
		if candidate == "" {
			return nil
		}
		candidate = tspath.NormalizePath(candidate)
		compareOptions := tspath.ComparePathsOptions{UseCaseSensitiveFileNames: useCaseSensitive}
		if tspath.ComparePaths(candidate, targetPath, compareOptions) == 0 ||
			tspath.ComparePaths(candidate, canonicalPath, compareOptions) == 0 {
			return nil
		}
		return program.GetSourceFile(candidate)
	}

	ownerRoot := tspath.NormalizePath(target.OwnerConfigDir)
	ownerCanonical := ownerRoot
	if realPath := fsys.Realpath(ownerRoot); realPath != "" {
		ownerCanonical = tspath.NormalizePath(realPath)
	}
	ownerAlias := aliasPathUnderRoot(canonicalPath, ownerCanonical, ownerRoot, useCaseSensitive)
	if sourceFile := lookupAlias(ownerAlias); sourceFile != nil {
		return sourceFile
	}

	programRoot := tspath.NormalizePath(program.GetCurrentDirectory())
	programCanonical := programRoot
	if realPath := fsys.Realpath(programRoot); realPath != "" {
		programCanonical = tspath.NormalizePath(realPath)
	}
	programAlias := aliasPathUnderRoot(canonicalPath, programCanonical, programRoot, useCaseSensitive)
	compareOptions := tspath.ComparePathsOptions{UseCaseSensitiveFileNames: useCaseSensitive}
	if tspath.ComparePaths(programAlias, ownerAlias, compareOptions) != 0 {
		if sourceFile := lookupAlias(programAlias); sourceFile != nil {
			return sourceFile
		}
	}
	return lookup.canonicalSourceFile(canonicalPath)
}

func orderedProgramIndexesForConfig(set lintProgramSet, configDir string) []int {
	indexes := make([]int, 0, len(set.Programs))
	for i := range set.Programs {
		if i < len(set.ConfigOrders) {
			if _, ok := set.ConfigOrders[i][configDir]; ok {
				indexes = append(indexes, i)
			}
		}
	}
	sort.SliceStable(indexes, func(i, j int) bool {
		left := set.ConfigOrders[indexes[i]][configDir]
		right := set.ConfigOrders[indexes[j]][configDir]
		if left != right {
			return left < right
		}
		return indexes[i] < indexes[j]
	})
	return indexes
}

// bindLintTargetPlan binds every stable target to a Program from its governing
// config. Calling this for each fix pass recomputes gap status from the current
// import graph instead of retaining an initial gap classification.
func bindLintTargetPlan(
	set lintProgramSet,
	plan lintTargetPlan,
	currentDirectory string,
	fsys vfs.FS,
	parseCache *utils.ParseCache,
	singleThreaded bool,
) (lintTargetBinding, error) {
	binding := lintTargetBinding{
		Programs:                   append([]*compiler.Program(nil), set.Programs...),
		TargetsByProgram:           make([][]string, len(set.Programs)),
		TypeInfoFiles:              make(map[string]struct{}),
		TargetPathBySourcePath:     make(map[string]string),
		ConfigPathBySourcePath:     make(map[string]string),
		OwnerConfigDirBySourcePath: make(map[string]string),
	}

	var gaps []resolvedLintTarget
	programIndexesByConfig := make(map[string][]int)
	programLookups := make([]programTargetLookup, len(set.Programs))
	for i, program := range set.Programs {
		programLookups[i] = programTargetLookup{program: program, fsys: fsys}
	}
	for _, target := range plan.Targets {
		programIndexes, cached := programIndexesByConfig[target.OwnerConfigDir]
		if !cached {
			programIndexes = orderedProgramIndexesForConfig(set, target.OwnerConfigDir)
			programIndexesByConfig[target.OwnerConfigDir] = programIndexes
		}
		bound := false
		for _, programIndex := range programIndexes {
			sourceFile := programLookups[programIndex].sourceFileForTarget(target)
			if sourceFile == nil {
				continue
			}
			sourcePath := sourceFile.FileName()
			binding.TargetsByProgram[programIndex] = append(binding.TargetsByProgram[programIndex], sourcePath)
			storeSourcePathMapping(binding.OwnerConfigDirBySourcePath, sourcePath, target.OwnerConfigDir, fsys)
			storeSourcePathMapping(binding.ConfigPathBySourcePath, sourcePath, configPathForBoundSource(sourcePath, target, fsys), fsys)
			binding.TypeInfoFiles[sourcePath] = struct{}{}
			binding.TypeInfoFiles[target.Path] = struct{}{}
			binding.TypeInfoFiles[target.CanonicalPath] = struct{}{}
			if tspath.NormalizePath(sourcePath) != target.Path {
				storeSourcePathMapping(binding.TargetPathBySourcePath, sourcePath, target.Path, fsys)
			}
			bound = true
			break
		}
		if !bound {
			gaps = append(gaps, target)
			binding.GapFiles = append(binding.GapFiles, target.Path)
		}
	}

	if len(gaps) > 0 {
		fallbackFiles := make([]string, 0, len(gaps))
		for _, gap := range gaps {
			fallbackFiles = append(fallbackFiles, gap.Path)
		}
		fallback, err := createFallbackProgram(fallbackFiles, singleThreaded, currentDirectory, fsys, parseCache)
		if err != nil {
			return lintTargetBinding{}, err
		}
		if fallback == nil {
			return lintTargetBinding{}, fmt.Errorf("create fallback Program for %d lint target(s): no Program returned", len(gaps))
		}
		fallbackIndex := len(binding.Programs)
		binding.Programs = append(binding.Programs, fallback)
		binding.TargetsByProgram = append(binding.TargetsByProgram, nil)
		fallbackLookup := programTargetLookup{program: fallback, fsys: fsys}
		for _, gap := range gaps {
			sourceFile := fallbackLookup.sourceFileForTarget(gap)
			if sourceFile == nil {
				return lintTargetBinding{}, fmt.Errorf("fallback Program did not contain lint target %q", gap.Path)
			}
			sourcePath := sourceFile.FileName()
			binding.TargetsByProgram[fallbackIndex] = append(binding.TargetsByProgram[fallbackIndex], sourcePath)
			storeSourcePathMapping(binding.OwnerConfigDirBySourcePath, sourcePath, gap.OwnerConfigDir, fsys)
			storeSourcePathMapping(binding.ConfigPathBySourcePath, sourcePath, configPathForBoundSource(sourcePath, gap, fsys), fsys)
			if tspath.NormalizePath(sourcePath) != gap.Path {
				storeSourcePathMapping(binding.TargetPathBySourcePath, sourcePath, gap.Path, fsys)
			}
		}
	}

	for i := range binding.TargetsByProgram {
		sort.Strings(binding.TargetsByProgram[i])
	}
	if len(binding.GapFiles) == 0 {
		binding.TypeInfoFiles = nil
	}
	if len(binding.TargetPathBySourcePath) == 0 {
		binding.TargetPathBySourcePath = nil
	}
	return binding, nil
}

func discoverLintFilesMultiConfig(
	configMap map[string]rslintconfig.RslintConfig,
	configTargetFiles map[string][]string,
	fs vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []rslintconfig.DiscoveredLintTarget {
	var scopes map[string]rslintconfig.LintDiscoveryScope
	if len(configTargetFiles) > 0 {
		scopes = make(map[string]rslintconfig.LintDiscoveryScope, len(configTargetFiles))
		for configDir, targetFiles := range configTargetFiles {
			scopes[configDir] = rslintconfig.LintDiscoveryScope{Files: targetFiles}
		}
	}
	return rslintconfig.DiscoverLintTargetsMultiConfig(configMap, scopes, fs, allowFiles, allowDirs, singleThreaded)
}

// buildTypeCheckSkipMask returns a parallel-to-programs []bool marking which
// programs must be excluded from the type-check phase. A program is skipped
// when it was NOT built from a real tsconfig on disk — i.e. its CompilerOptions
// carry no ConfigFilePath. That covers the AST-only fallback Program used for
// selected files absent from every tsconfig-backed Program, including projects
// with no tsconfig at all.
//
// Their options are synthesized, not the user's tsconfig, so semantic
// diagnostics there are unreliable and must not be surfaced. This mirrors the
// type-checking boundary: only Programs backed by parserOptions.project or the
// auto-detected tsconfig.json participate in program-level --type-check
// diagnostics. Programs built via utils.CreateProgram ->
// GetParsedCommandLineOfConfigFile carry a non-empty ConfigFilePath.
//
// Returns nil when every program is tsconfig-backed, so callers don't allocate
// an all-false slice. Deriving from the programs keeps the CLI initial build
// and the --fix rebuild path consistent by construction.
func buildTypeCheckSkipMask(programs []*compiler.Program) []bool {
	var mask []bool
	for i, prog := range programs {
		opts := prog.Options()
		if opts == nil || opts.ConfigFilePath == "" {
			if mask == nil {
				mask = make([]bool, len(programs))
			}
			mask[i] = true
		}
	}
	return mask
}

type syntacticDiagnosticKey struct {
	path string
	code int32
	pos  int
	end  int
}

func collectTargetSyntacticDiagnostics(
	programs []*compiler.Program,
	targetsByProgram [][]string,
	skipTypeCheck []bool,
	typeCheck bool,
	typeCheckOnly bool,
) ([]rule.RuleDiagnostic, map[string]struct{}) {
	syntaxErrorFiles := make(map[string]struct{})
	if len(programs) == 0 || len(targetsByProgram) == 0 {
		return nil, syntaxErrorFiles
	}

	seen := make(map[syntacticDiagnosticKey]struct{})
	var diagnostics []rule.RuleDiagnostic
	for i, program := range programs {
		// When --type-check runs, tsconfig-backed Programs surface syntactic
		// diagnostics through the type-check phase. We still inspect every target
		// here so the lint-rule phase can skip malformed files, matching ESLint.
		coveredByTypeCheck := typeCheck && (i >= len(skipTypeCheck) || !skipTypeCheck[i])
		if i >= len(targetsByProgram) || len(targetsByProgram[i]) == 0 {
			continue
		}

		targets := make(map[string]struct{}, len(targetsByProgram[i]))
		for _, target := range targetsByProgram[i] {
			targets[target] = struct{}{}
		}

		ctx := context.Background()
		for target := range targets {
			file := program.GetSourceFile(target)
			if file == nil {
				continue
			}
			for _, diagnostic := range program.GetSyntacticDiagnostics(ctx, file) {
				syntaxErrorFiles[file.FileName()] = struct{}{}
				if coveredByTypeCheck || typeCheckOnly {
					continue
				}
				loc := diagnostic.Loc()
				key := syntacticDiagnosticKey{
					path: file.FileName(),
					code: diagnostic.Code(),
					pos:  loc.Pos(),
					end:  loc.End(),
				}
				if _, ok := seen[key]; ok {
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
					PreFormatted: true,
				})
			}
		}
	}
	return diagnostics, syntaxErrorFiles
}
