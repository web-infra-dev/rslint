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
// normalized tsconfig path and the declaration order within each config. A
// shared path has one Program instance but may have multiple associations.
type programConfigOrders map[string]int

// lintProgramSet is the unique set of real tsconfig-backed Programs used by a
// lint run. ConfigOrders is parallel to Programs. Synthetic fallback Programs
// are appended only while binding a lint target plan and have no config entry.
type lintProgramSet struct {
	Programs     []*compiler.Program
	ConfigOrders []programConfigOrders
}

func exactFilesystemPathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
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
	return exactFilesystemPathID(authoritativeFilesystemPath(filePath, fsys))
}

// configPathForLintTarget returns the target path used for files/ignores
// matching. Program source names are deliberately excluded: adding or removing
// a file from a TypeScript Program must not change its lint configuration.
func configPathForLintTarget(target resolvedLintTarget, fsys vfs.FS) string {
	matchPath, _ := rslintconfig.ResolveConfigPathSpaceWithCanonical(
		target.Path,
		target.CanonicalPath,
		target.OwnerConfigDir,
		fsys,
	)
	return matchPath
}

func storeSourcePathMapping(mapping map[string]string, sourcePath string, canonicalSourcePath string, value string) {
	if mapping == nil {
		return
	}
	normalizedSource := tspath.NormalizePath(sourcePath)
	mapping[normalizedSource] = value
	if canonicalSourcePath != "" {
		mapping[exactFilesystemPathID(canonicalSourcePath)] = value
	}
}

// createProgramSetForConfigs builds each normalized tsconfig path once while
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
			tsconfigPath = tspath.NormalizePath(tsconfigPath)
			tsconfigID := exactFilesystemPathID(tsconfigPath)
			if programIndex, ok := programByTsconfig[tsconfigID]; ok {
				if _, alreadyAssociated := set.ConfigOrders[programIndex][configDir]; !alreadyAssociated {
					set.ConfigOrders[programIndex][configDir] = order
				}
				continue
			}

			// Relative paths in a tsconfig are resolved from the declared path,
			// including when that path is a file symlink. This matches tsc/tsgo;
			// realpath is only a source-identity fallback during target binding.
			programCwd := tspath.GetDirectoryPath(tsconfigPath)
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

func configWithGitignore(
	rslintConfig rslintconfig.RslintConfig,
	configDir string,
	fsys vfs.FS,
) rslintconfig.RslintConfig {
	gitGlobs := rslintconfig.ReadGitignoreAsGlobs(
		configDir,
		fsys,
		rslintconfig.ExtractConfigIgnores(rslintConfig),
	)
	if len(gitGlobs) == 0 {
		return rslintConfig
	}
	return append(rslintconfig.RslintConfig{{Ignores: gitGlobs}}, rslintConfig...)
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

func configsForLintTargetPlan(
	configMap map[string]rslintconfig.RslintConfig,
	plan lintTargetPlan,
) map[string]rslintconfig.RslintConfig {
	if len(configMap) == 0 || len(plan.Targets) == 0 {
		return nil
	}
	active := make(map[string]rslintconfig.RslintConfig)
	for _, target := range plan.Targets {
		if entries, ok := configMap[target.OwnerConfigDir]; ok {
			active[target.OwnerConfigDir] = entries
		}
	}
	return active
}

func resolveLintTargetPlan(
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
	configTargetScopes map[string]rslintconfig.LintDiscoveryScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) (lintTargetPlan, error) {
	type targetWithOwner struct {
		path          string
		canonicalPath string
		owner         string
	}
	var targetFiles []targetWithOwner
	if configMap != nil {
		for _, target := range discoverLintFilesMultiConfig(configMap, configTargetScopes, fsys, allowFiles, allowDirs, singleThreaded) {
			targetFiles = append(targetFiles, targetWithOwner{
				path:          target.Path,
				canonicalPath: target.CanonicalPath,
				owner:         target.ConfigDirectory,
			})
		}
	} else {
		for _, target := range rslintconfig.DiscoverLintTargets(rslintConfig, currentDirectory, fsys, allowFiles, allowDirs, singleThreaded) {
			targetFiles = append(targetFiles, targetWithOwner{
				path:          target.Path,
				canonicalPath: target.CanonicalPath,
				owner:         currentDirectory,
			})
		}
	}

	plan := lintTargetPlan{Targets: make([]resolvedLintTarget, 0, len(targetFiles))}
	seenCanonical := make(map[string]resolvedLintTarget, len(targetFiles))
	for _, discovered := range targetFiles {
		targetPath := discovered.path
		ownerConfigDir := discovered.owner
		canonicalPath := tspath.NormalizePath(targetPath)
		if discovered.canonicalPath != "" {
			canonicalPath = tspath.NormalizePath(discovered.canonicalPath)
		}
		if discovered.canonicalPath == "" && fsys != nil {
			if realPath := fsys.Realpath(canonicalPath); realPath != "" {
				canonicalPath = tspath.NormalizePath(realPath)
			}
		}
		canonicalKey := exactFilesystemPathID(canonicalPath)
		target := resolvedLintTarget{
			Path:           tspath.NormalizePath(targetPath),
			CanonicalPath:  canonicalPath,
			OwnerConfigDir: tspath.NormalizePath(ownerConfigDir),
		}
		if existing, exists := seenCanonical[canonicalKey]; exists {
			if canonicalFilesystemPathID(existing.OwnerConfigDir, fsys) != canonicalFilesystemPathID(target.OwnerConfigDir, fsys) {
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

func preferredCallerTargetPaths(plan lintTargetPlan) map[string]string {
	if len(plan.Targets) == 0 {
		return nil
	}
	preferred := make(map[string]string, len(plan.Targets))
	for _, target := range plan.Targets {
		canonicalID := exactFilesystemPathID(target.CanonicalPath)
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

func exactProgramSourceFile(program *compiler.Program, targetPath string) *ast.SourceFile {
	if program == nil || targetPath == "" {
		return nil
	}
	targetPath = tspath.NormalizePath(targetPath)
	sourceFile := program.GetSourceFile(targetPath)
	if sourceFile == nil || exactFilesystemPathID(sourceFile.FileName()) != exactFilesystemPathID(targetPath) {
		return nil
	}
	return sourceFile
}

// programFileIndex joins lint targets to Program sources by exact physical
// path. It resolves each unique source path once and retains Program membership
// so project declaration order remains authoritative during lookup.
type programFileIndex struct {
	programs         []*compiler.Program
	fsys             vfs.FS
	singleThreaded   bool
	built            bool
	sourcesByProgram []map[string]*ast.SourceFile
}

func newProgramFileIndex(programs []*compiler.Program, fsys vfs.FS, singleThreaded bool) *programFileIndex {
	return &programFileIndex{
		programs:       programs,
		fsys:           fsys,
		singleThreaded: singleThreaded,
	}
}

func (index *programFileIndex) sourceFile(programIndex int, canonicalTarget string) *ast.SourceFile {
	if index == nil || index.fsys == nil || canonicalTarget == "" ||
		programIndex < 0 || programIndex >= len(index.programs) {
		return nil
	}
	if !index.built {
		index.build()
	}
	return index.sourcesByProgram[programIndex][exactFilesystemPathID(canonicalTarget)]
}

type programSourceMembership struct {
	programIndex int
	sourceIndex  int
	sourceFile   *ast.SourceFile
}

func (index *programFileIndex) build() {
	index.built = true
	index.sourcesByProgram = make([]map[string]*ast.SourceFile, len(index.programs))

	sourceIndexByPath := make(map[string]int)
	var sourcePaths []string
	var memberships []programSourceMembership
	for programIndex, program := range index.programs {
		if program == nil {
			continue
		}
		for _, sourceFile := range program.GetSourceFiles() {
			sourcePath := tspath.NormalizePath(sourceFile.FileName())
			sourcePathID := exactFilesystemPathID(sourcePath)
			sourceIndex, exists := sourceIndexByPath[sourcePathID]
			if !exists {
				sourceIndex = len(sourcePaths)
				sourceIndexByPath[sourcePathID] = sourceIndex
				sourcePaths = append(sourcePaths, sourcePath)
			}
			memberships = append(memberships, programSourceMembership{
				programIndex: programIndex,
				sourceIndex:  sourceIndex,
				sourceFile:   sourceFile,
			})
		}
	}

	canonicalIDs := make([]string, len(sourcePaths))
	work := core.NewWorkGroup(index.singleThreaded)
	for i := range sourcePaths {
		work.Queue(func() {
			canonicalIDs[i] = canonicalFilesystemPathID(sourcePaths[i], index.fsys)
		})
	}
	work.RunAndWait()

	for _, membership := range memberships {
		sources := index.sourcesByProgram[membership.programIndex]
		if sources == nil {
			sources = make(map[string]*ast.SourceFile)
			index.sourcesByProgram[membership.programIndex] = sources
		}
		canonicalID := canonicalIDs[membership.sourceIndex]
		existing := sources[canonicalID]
		if existing == nil || membership.sourceFile.FileName() < existing.FileName() {
			sources[canonicalID] = membership.sourceFile
		}
	}
}

func groupFallbackTargets(
	gaps []resolvedLintTarget,
	currentDirectory string,
	useCaseSensitive bool,
) [][]resolvedLintTarget {
	if len(gaps) == 0 {
		return nil
	}

	groups := make([][]resolvedLintTarget, 0, 1)
	keysByGroup := make([]map[tspath.Path]struct{}, 0, 1)
	for _, gap := range gaps {
		key := tspath.ToPath(gap.Path, currentDirectory, useCaseSensitive)
		groupIndex := -1
		for i, keys := range keysByGroup {
			if _, exists := keys[key]; !exists {
				groupIndex = i
				break
			}
		}
		if groupIndex == -1 {
			groupIndex = len(groups)
			groups = append(groups, nil)
			keysByGroup = append(keysByGroup, make(map[tspath.Path]struct{}))
		}
		groups[groupIndex] = append(groups[groupIndex], gap)
		keysByGroup[groupIndex][key] = struct{}{}
	}
	return groups
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
	programFiles := newProgramFileIndex(set.Programs, fsys, singleThreaded)
	for _, target := range plan.Targets {
		programIndexes, cached := programIndexesByConfig[target.OwnerConfigDir]
		if !cached {
			programIndexes = orderedProgramIndexesForConfig(set, target.OwnerConfigDir)
			programIndexesByConfig[target.OwnerConfigDir] = programIndexes
		}
		bound := false
		for _, programIndex := range programIndexes {
			sourceFile := exactProgramSourceFile(set.Programs[programIndex], target.Path)
			if sourceFile == nil {
				sourceFile = programFiles.sourceFile(programIndex, target.CanonicalPath)
			}
			if sourceFile == nil {
				continue
			}
			sourcePath := sourceFile.FileName()
			binding.TargetsByProgram[programIndex] = append(binding.TargetsByProgram[programIndex], sourcePath)
			storeSourcePathMapping(binding.OwnerConfigDirBySourcePath, sourcePath, target.CanonicalPath, target.OwnerConfigDir)
			storeSourcePathMapping(binding.ConfigPathBySourcePath, sourcePath, target.CanonicalPath, configPathForLintTarget(target, fsys))
			binding.TypeInfoFiles[sourcePath] = struct{}{}
			binding.TypeInfoFiles[target.Path] = struct{}{}
			binding.TypeInfoFiles[target.CanonicalPath] = struct{}{}
			if tspath.NormalizePath(sourcePath) != target.Path {
				storeSourcePathMapping(binding.TargetPathBySourcePath, sourcePath, target.CanonicalPath, target.Path)
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
		useCaseSensitive := true
		if fsys != nil {
			useCaseSensitive = fsys.UseCaseSensitiveFileNames()
		}
		for _, fallbackTargets := range groupFallbackTargets(gaps, currentDirectory, useCaseSensitive) {
			fallbackFiles := make([]string, 0, len(fallbackTargets))
			for _, gap := range fallbackTargets {
				fallbackFiles = append(fallbackFiles, gap.Path)
			}
			fallback, err := createFallbackProgram(fallbackFiles, singleThreaded, currentDirectory, fsys, parseCache)
			if err != nil {
				return lintTargetBinding{}, err
			}
			if fallback == nil {
				return lintTargetBinding{}, fmt.Errorf("create fallback Program for %d lint target(s): no Program returned", len(fallbackTargets))
			}
			fallbackIndex := len(binding.Programs)
			binding.Programs = append(binding.Programs, fallback)
			binding.TargetsByProgram = append(binding.TargetsByProgram, nil)
			for _, gap := range fallbackTargets {
				sourceFile := exactProgramSourceFile(fallback, gap.Path)
				if sourceFile == nil {
					return lintTargetBinding{}, fmt.Errorf("fallback Program did not contain lint target %q", gap.Path)
				}
				sourcePath := sourceFile.FileName()
				binding.TargetsByProgram[fallbackIndex] = append(binding.TargetsByProgram[fallbackIndex], sourcePath)
				storeSourcePathMapping(binding.OwnerConfigDirBySourcePath, sourcePath, gap.CanonicalPath, gap.OwnerConfigDir)
				storeSourcePathMapping(binding.ConfigPathBySourcePath, sourcePath, gap.CanonicalPath, configPathForLintTarget(gap, fsys))
				if tspath.NormalizePath(sourcePath) != gap.Path {
					storeSourcePathMapping(binding.TargetPathBySourcePath, sourcePath, gap.CanonicalPath, gap.Path)
				}
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
	configTargetScopes map[string]rslintconfig.LintDiscoveryScope,
	fs vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []rslintconfig.DiscoveredLintTarget {
	return rslintconfig.DiscoverLintTargetsMultiConfig(configMap, configTargetScopes, fs, allowFiles, allowDirs, singleThreaded)
}

// buildTypeCheckSkipMask returns a parallel-to-programs []bool marking which
// programs must be excluded from the type-check phase. A program is skipped
// when it was NOT built from a real tsconfig on disk — i.e. its CompilerOptions
// carry no ConfigFilePath. That covers the non-project-backed fallback Program used for
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
		ctx := context.Background()
		for _, target := range targetsByProgram[i] {
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
