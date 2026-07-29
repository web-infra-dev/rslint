package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestFilterNonTypeAwareRules(t *testing.T) {
	rules := []linter.ConfiguredRule{
		{Name: "syntax-rule", RequiresTypeInfo: false},
		{Name: "type-rule", RequiresTypeInfo: true},
		{Name: "another-syntax", RequiresTypeInfo: false},
	}

	filtered := linter.FilterNonTypeAwareRules(rules)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(filtered))
	}
	if filtered[0].Name != "syntax-rule" {
		t.Errorf("expected syntax-rule, got %s", filtered[0].Name)
	}
	if filtered[1].Name != "another-syntax" {
		t.Errorf("expected another-syntax, got %s", filtered[1].Name)
	}
}

func TestFilterNonTypeAwareRules_AllTypeAware(t *testing.T) {
	rules := []linter.ConfiguredRule{
		{Name: "type-rule-1", RequiresTypeInfo: true},
		{Name: "type-rule-2", RequiresTypeInfo: true},
	}

	filtered := linter.FilterNonTypeAwareRules(rules)

	if len(filtered) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(filtered))
	}
}

func TestFilterNonTypeAwareRules_NoneTypeAware(t *testing.T) {
	rules := []linter.ConfiguredRule{
		{Name: "rule-a", RequiresTypeInfo: false},
		{Name: "rule-b", RequiresTypeInfo: false},
	}

	filtered := linter.FilterNonTypeAwareRules(rules)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(filtered))
	}
}

func TestFilterNonTypeAwareRules_Empty(t *testing.T) {
	filtered := linter.FilterNonTypeAwareRules(nil)
	if len(filtered) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(filtered))
	}
}

// Verify RequiresTypeInfo propagates through CreateRule.
func TestRequiresTypeInfo_Propagation(t *testing.T) {
	r := rule.CreateRule(rule.Rule{
		Name:             "test-rule",
		RequiresTypeInfo: true,
		Run:              func(ctx rule.RuleContext, options []any) rule.RuleListeners { return nil },
	})

	if r.Name != "@typescript-eslint/test-rule" {
		t.Errorf("unexpected name: %s", r.Name)
	}
	if !r.RequiresTypeInfo {
		t.Error("RequiresTypeInfo should be true after CreateRule")
	}
}

// createTestProgram creates a Program from temp files for testing utils.CollectProgramFiles.
func createTestProgram(t *testing.T, files map[string]string) *compiler.Program {
	t.Helper()
	tmpDir := t.TempDir()

	includes := make([]string, 0, len(files))
	for name, content := range files {
		fp := filepath.Join(tmpDir, name)
		os.MkdirAll(filepath.Dir(fp), 0755)
		os.WriteFile(fp, []byte(content), 0644)
		includes = append(includes, "./"+name)
	}

	tsconfig := `{"include":["` + includes[0] + `"]}`
	if len(includes) > 1 {
		tsconfig = `{"include":["**/*.ts"]}`
	}
	os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(tsconfig), 0644)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}
	return program
}

func TestBuildProgramFileSet_ContainsSourceFiles(t *testing.T) {
	program := createTestProgram(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	fileSet := utils.CollectProgramFiles([]*compiler.Program{program}, bundled.WrapFS(cachedvfs.From(osvfs.FS())), false)

	// Should contain user source files
	found := 0
	for k := range fileSet {
		if filepath.Ext(k) == ".ts" && !strings.Contains(k, "bundled:") {
			found++
		}
	}
	if found < 2 {
		t.Errorf("Expected at least 2 .ts files in fileSet, got %d", found)
	}
}

func TestBuildProgramFileSet_RealpathKeyForSymlinks(t *testing.T) {
	// On macOS, /tmp is a symlink to /private/tmp.
	// Verify that utils.CollectProgramFiles adds the resolved path as an alternate key.
	tmpDir := t.TempDir()
	realTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Skip("EvalSymlinks not available")
	}
	if tspath.NormalizePath(realTmpDir) == tspath.NormalizePath(tmpDir) {
		t.Skip("No symlink divergence on this system")
	}

	// Create a file and program using the UNRESOLVED path
	os.WriteFile(filepath.Join(tmpDir, "test.ts"), []byte("const x = 1;"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(`{"include":["**/*.ts"]}`), 0644)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("Failed to create program: %v", err)
	}

	fileSet := utils.CollectProgramFiles([]*compiler.Program{program}, bundled.WrapFS(cachedvfs.From(osvfs.FS())), false)

	// The file should be findable via BOTH the original and resolved paths
	resolvedFilePath := tspath.NormalizePath(filepath.Join(realTmpDir, "test.ts"))
	if _, exists := fileSet[resolvedFilePath]; !exists {
		// Dump keys for debugging
		for k := range fileSet {
			if !strings.Contains(k, "bundled:") {
				t.Logf("fileSet key: %s", k)
			}
		}
		t.Errorf("Expected fileSet to contain resolved path %s", resolvedFilePath)
	}
}

type bindingIndexTestFS struct {
	vfs.FS
	mu            sync.Mutex
	files         map[string]string
	entries       map[string]vfs.Entries
	realPaths     map[string]string
	calls         map[string]int
	caseSensitive bool
}

func (fsys *bindingIndexTestFS) UseCaseSensitiveFileNames() bool {
	return fsys.caseSensitive
}

func (fsys *bindingIndexTestFS) FileExists(filePath string) bool {
	_, ok := fsys.files[tspath.NormalizePath(filePath)]
	return ok
}

func (fsys *bindingIndexTestFS) ReadFile(filePath string) (string, bool) {
	content, ok := fsys.files[tspath.NormalizePath(filePath)]
	return content, ok
}

func (fsys *bindingIndexTestFS) GetAccessibleEntries(directoryPath string) vfs.Entries {
	if entries, ok := fsys.entries[tspath.NormalizePath(directoryPath)]; ok {
		return entries
	}
	return fsys.FS.GetAccessibleEntries(directoryPath)
}

func (fsys *bindingIndexTestFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	fsys.mu.Lock()
	fsys.calls[filePath]++
	fsys.mu.Unlock()
	if realPath := fsys.realPaths[filePath]; realPath != "" {
		return realPath
	}
	return filePath
}

func (fsys *bindingIndexTestFS) resetCalls() {
	fsys.mu.Lock()
	clear(fsys.calls)
	fsys.mu.Unlock()
}

func (fsys *bindingIndexTestFS) callCount(filePath string) int {
	fsys.mu.Lock()
	defer fsys.mu.Unlock()
	return fsys.calls[tspath.NormalizePath(filePath)]
}

func newBindingIndexTestFS(files []string, realPaths map[string]string) *bindingIndexTestFS {
	contents := make(map[string]string, len(files))
	for _, filePath := range files {
		contents[tspath.NormalizePath(filePath)] = "export {};\n"
	}
	return &bindingIndexTestFS{
		FS:            osvfs.FS(),
		files:         contents,
		entries:       make(map[string]vfs.Entries),
		realPaths:     realPaths,
		calls:         make(map[string]int),
		caseSensitive: true,
	}
}

func createBindingIndexTestProgram(t *testing.T, fsys vfs.FS, rootFiles ...string) *compiler.Program {
	t.Helper()
	currentDirectory := "/"
	if len(rootFiles) > 0 {
		currentDirectory = tspath.GetDirectoryPath(rootFiles[0])
	}
	program, err := utils.CreateProgramFromOptionsLenient(
		true,
		&core.CompilerOptions{
			NoLib:     core.TSTrue,
			NoResolve: core.TSTrue,
		},
		rootFiles,
		utils.CreateCompilerHost(currentDirectory, fsys),
	)
	if err != nil {
		t.Fatalf("CreateProgramFromOptionsLenient: %v", err)
	}
	return program
}

func TestBindLintTargetPlan_PreservesExactAndProjectOrder(t *testing.T) {
	const (
		configDir  = "/repo"
		targetPath = "/repo/target.ts"
		aliasPath  = "/aliases/target.ts"
	)
	fsys := newBindingIndexTestFS(
		[]string{targetPath, aliasPath},
		map[string]string{aliasPath: targetPath},
	)

	t.Run("earlier alias beats later exact", func(t *testing.T) {
		aliasProgram := createBindingIndexTestProgram(t, fsys, aliasPath)
		exactProgram := createBindingIndexTestProgram(t, fsys, targetPath)
		set := lintProgramSet{
			Programs: []*compiler.Program{aliasProgram, exactProgram},
			ConfigOrders: []programConfigOrders{
				{configDir: 0},
				{configDir: 1},
			},
		}
		plan := lintTargetPlan{Targets: []resolvedLintTarget{{
			Path:           targetPath,
			CanonicalPath:  targetPath,
			OwnerConfigDir: configDir,
		}}}

		binding, err := bindLintTargetPlan(set, plan, configDir, fsys, utils.NewParseCache(), true)
		if err != nil {
			t.Fatalf("bindLintTargetPlan: %v", err)
		}
		if len(binding.TargetsByProgram[0]) != 1 || binding.TargetsByProgram[0][0] != aliasPath {
			t.Fatalf("earlier alias Program must win, got %v", binding.TargetsByProgram)
		}
		if len(binding.TargetsByProgram[1]) != 0 {
			t.Fatalf("later exact Program must remain unused, got %v", binding.TargetsByProgram)
		}
	})

	t.Run("exact beats alias within one Program", func(t *testing.T) {
		program := createBindingIndexTestProgram(t, fsys, aliasPath, targetPath)
		set := lintProgramSet{
			Programs:     []*compiler.Program{program},
			ConfigOrders: []programConfigOrders{{configDir: 0}},
		}
		plan := lintTargetPlan{Targets: []resolvedLintTarget{{
			Path:           targetPath,
			CanonicalPath:  targetPath,
			OwnerConfigDir: configDir,
		}}}
		fsys.resetCalls()

		binding, err := bindLintTargetPlan(set, plan, configDir, fsys, utils.NewParseCache(), true)
		if err != nil {
			t.Fatalf("bindLintTargetPlan: %v", err)
		}
		if got := binding.TargetsByProgram[0]; len(got) != 1 || got[0] != targetPath {
			t.Fatalf("exact source must beat its alias, got %v", got)
		}
		if calls := fsys.callCount(aliasPath); calls != 0 {
			t.Fatalf("exact hit unexpectedly built the alias index: alias realpath calls=%d", calls)
		}
	})
}

func TestProgramFileIndex_IsTargetAwareAndLazyPerGoverningGroup(t *testing.T) {
	const (
		targetPath     = "/physical/target.ts"
		targetAlias    = "/caller/target.ts"
		sourceAlias    = "/sources/target.ts"
		otherSource    = "/sources/dependency.ts"
		otherCanonical = "/physical/dependency.ts"
	)
	fsys := newBindingIndexTestFS(
		[]string{targetPath, sourceAlias, otherSource},
		map[string]string{
			sourceAlias: sourceAlias,
			otherSource: otherCanonical,
		},
	)
	targetProgram := createBindingIndexTestProgram(t, fsys, targetPath, otherSource)
	aliasProgram := createBindingIndexTestProgram(t, fsys, sourceAlias)
	fsys.resetCalls()

	index := newProgramFileIndex(
		[]*compiler.Program{targetProgram, aliasProgram},
		[]resolvedLintTarget{{
			Path:          targetAlias,
			CanonicalPath: targetPath,
		}},
		fsys,
		true,
	)
	sourceFile := index.sourceFile([]int{0}, 0, targetPath)
	if sourceFile == nil || sourceFile.FileName() != targetPath {
		t.Fatalf("target-backed source lookup returned %v", sourceFile)
	}
	if calls := fsys.callCount(targetPath); calls != 0 {
		t.Fatalf("known target identity unexpectedly called Realpath %d time(s)", calls)
	}
	if !index.builtByProgram[0] || index.builtByProgram[1] {
		t.Fatalf("only the requested Program may be built: %v", index.builtByProgram)
	}
	if sources := index.sourcesByProgram[0]; len(sources) != 1 {
		t.Fatalf("per-Program index retained non-target identities: %v", sources)
	}
}

func TestProgramFileIndex_BuildsGoverningGroupInOneBatch(t *testing.T) {
	const (
		firstAlias   = "/sources/first.ts"
		firstTarget  = "/physical/first.ts"
		secondAlias  = "/sources/second.ts"
		secondTarget = "/physical/second.ts"
	)
	fsys := newBindingIndexTestFS(
		[]string{firstAlias, secondAlias},
		map[string]string{
			firstAlias:  firstTarget,
			secondAlias: secondTarget,
		},
	)
	first := createBindingIndexTestProgram(t, fsys, firstAlias)
	second := createBindingIndexTestProgram(t, fsys, secondAlias)
	fsys.resetCalls()

	index := newProgramFileIndex(
		[]*compiler.Program{first, second},
		[]resolvedLintTarget{
			{Path: firstTarget, CanonicalPath: firstTarget},
			{Path: secondTarget, CanonicalPath: secondTarget},
		},
		fsys,
		false,
	)
	if sourceFile := index.sourceFile([]int{0, 1}, 0, firstTarget); sourceFile == nil ||
		sourceFile.FileName() != firstAlias {
		t.Fatalf("first Program lookup returned %v", sourceFile)
	}
	if !index.builtByProgram[0] || !index.builtByProgram[1] {
		t.Fatalf("governing Program group was not built together: %v", index.builtByProgram)
	}
	if sourceFile := index.sourceFile([]int{0, 1}, 1, secondTarget); sourceFile == nil ||
		sourceFile.FileName() != secondAlias {
		t.Fatalf("second Program lookup returned %v", sourceFile)
	}
	if calls := fsys.callCount(firstAlias); calls != 1 {
		t.Fatalf("first source resolved %d times, want 1", calls)
	}
	if calls := fsys.callCount(secondAlias); calls != 1 {
		t.Fatalf("second source resolved %d times, want 1", calls)
	}
}

func TestProgramFileIndex_CanonicalizesRegularFilesByDirectory(t *testing.T) {
	const (
		sourceDirectory = "/aliases"
		firstSource     = "/aliases/first.ts"
		firstTarget     = "/physical/first.ts"
		secondSource    = "/aliases/second.ts"
		secondTarget    = "/physical/second.ts"
	)
	fsys := newBindingIndexTestFS([]string{firstSource, secondSource}, map[string]string{
		sourceDirectory: "/physical",
		firstSource:     firstTarget,
		secondSource:    secondTarget,
	})
	fsys.entries[sourceDirectory] = vfs.Entries{
		Files:    []string{"first.ts", "second.ts"},
		Symlinks: map[string]struct{}{},
	}
	program := createBindingIndexTestProgram(t, fsys, firstSource, secondSource)
	fsys.resetCalls()

	index := newProgramFileIndex(
		[]*compiler.Program{program},
		[]resolvedLintTarget{
			{Path: firstTarget, CanonicalPath: firstTarget},
			{Path: secondTarget, CanonicalPath: secondTarget},
		},
		fsys,
		false,
	)
	for targetPath, sourcePath := range map[string]string{
		firstTarget:  firstSource,
		secondTarget: secondSource,
	} {
		sourceFile := index.sourceFile([]int{0}, 0, targetPath)
		if sourceFile == nil || sourceFile.FileName() != sourcePath {
			t.Fatalf("directory-derived canonical lookup for %q returned %v", targetPath, sourceFile)
		}
	}
	if calls := fsys.callCount(sourceDirectory); calls != 1 {
		t.Fatalf("source directory resolved %d times, want 1", calls)
	}
	for _, sourcePath := range []string{firstSource, secondSource} {
		if calls := fsys.callCount(sourcePath); calls != 0 {
			t.Fatalf("regular source %q unexpectedly resolved per file %d time(s)", sourcePath, calls)
		}
	}
}

func TestProgramFileIndex_FallsBackForUncertainFileIdentity(t *testing.T) {
	const (
		sourceDirectory = "/aliases"
		sourcePath      = "/aliases/target.ts"
		otherSource     = "/aliases/other.ts"
		targetPath      = "/physical/target.ts"
	)
	tests := []struct {
		name       string
		ignoreCase bool
		entries    vfs.Entries
	}{
		{
			name: "file symlink",
			entries: vfs.Entries{
				Files:    []string{"other.ts", "target.ts"},
				Symlinks: map[string]struct{}{"target.ts": {}},
			},
		},
		{
			name: "metadata unavailable",
			entries: vfs.Entries{
				Files: []string{"other.ts", "target.ts"},
			},
		},
		{
			name: "entry unavailable",
			entries: vfs.Entries{
				Files:    []string{"other.ts"},
				Symlinks: map[string]struct{}{},
			},
		},
		{
			name:       "ambiguous case",
			ignoreCase: true,
			entries: vfs.Entries{
				Files:    []string{"other.ts", "Target.ts", "target.ts"},
				Symlinks: map[string]struct{}{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsys := newBindingIndexTestFS([]string{sourcePath, otherSource}, map[string]string{
				sourceDirectory: "/physical",
				sourcePath:      targetPath,
				otherSource:     "/physical/other.ts",
			})
			fsys.caseSensitive = !test.ignoreCase
			fsys.entries[sourceDirectory] = test.entries
			program := createBindingIndexTestProgram(t, fsys, sourcePath, otherSource)
			fsys.resetCalls()

			index := newProgramFileIndex(
				[]*compiler.Program{program},
				[]resolvedLintTarget{{Path: targetPath, CanonicalPath: targetPath}},
				fsys,
				true,
			)
			sourceFile := index.sourceFile([]int{0}, 0, targetPath)
			if sourceFile == nil || sourceFile.FileName() != sourcePath {
				t.Fatalf("fallback canonical lookup returned %v", sourceFile)
			}
			if calls := fsys.callCount(sourcePath); calls != 1 {
				t.Fatalf("uncertain source identity resolved %d times, want 1", calls)
			}
		})
	}
}

func TestProgramFileIndex_UsesPerFileIdentityForSingletonDirectory(t *testing.T) {
	const (
		sourceDirectory = "/aliases"
		sourcePath      = "/aliases/target.ts"
		targetPath      = "/physical/target.ts"
	)
	fsys := newBindingIndexTestFS([]string{sourcePath}, map[string]string{
		sourceDirectory: "/physical",
		sourcePath:      targetPath,
	})
	fsys.entries[sourceDirectory] = vfs.Entries{
		Files:    []string{"target.ts"},
		Symlinks: map[string]struct{}{},
	}
	program := createBindingIndexTestProgram(t, fsys, sourcePath)
	fsys.resetCalls()

	index := newProgramFileIndex(
		[]*compiler.Program{program},
		[]resolvedLintTarget{{Path: targetPath, CanonicalPath: targetPath}},
		fsys,
		false,
	)
	sourceFile := index.sourceFile([]int{0}, 0, targetPath)
	if sourceFile == nil || sourceFile.FileName() != sourcePath {
		t.Fatalf("singleton canonical lookup returned %v", sourceFile)
	}
	if calls := fsys.callCount(sourceDirectory); calls != 0 {
		t.Fatalf("singleton source unexpectedly resolved its directory %d time(s)", calls)
	}
	if calls := fsys.callCount(sourcePath); calls != 1 {
		t.Fatalf("singleton source resolved per file %d time(s), want 1", calls)
	}
}

func TestProgramFileIndex_UsesFilesystemCasingForDirectoryIdentity(t *testing.T) {
	const (
		sourceDirectory = "C:/repo"
		sourcePath      = "C:/repo/target.ts"
		otherSource     = "C:/repo/other.ts"
		targetPath      = "C:/Physical/Target.ts"
		otherTarget     = "C:/Physical/Other.ts"
	)
	fsys := newBindingIndexTestFS([]string{sourcePath, otherSource}, map[string]string{
		sourceDirectory: "C:/Physical",
	})
	fsys.caseSensitive = false
	fsys.entries[sourceDirectory] = vfs.Entries{
		Files:    []string{"Other.ts", "Target.ts"},
		Symlinks: map[string]struct{}{},
	}
	program := createBindingIndexTestProgram(t, fsys, sourcePath, otherSource)
	fsys.resetCalls()

	index := newProgramFileIndex(
		[]*compiler.Program{program},
		[]resolvedLintTarget{
			{Path: targetPath, CanonicalPath: targetPath},
			{Path: otherTarget, CanonicalPath: otherTarget},
		},
		fsys,
		true,
	)
	sourceFile := index.sourceFile([]int{0}, 0, targetPath)
	if sourceFile == nil || sourceFile.FileName() != sourcePath {
		t.Fatalf("case-corrected canonical lookup returned %v", sourceFile)
	}
	for _, path := range []string{sourcePath, otherSource} {
		if calls := fsys.callCount(path); calls != 0 {
			t.Fatalf("case-corrected regular source %q unexpectedly used per-file Realpath %d time(s)", path, calls)
		}
	}
}

func TestProgramFileIndex_ResolvesSharedUnknownSourceOnce(t *testing.T) {
	const (
		sourceAlias = "/sources/target.ts"
		targetPath  = "/physical/target.ts"
	)
	fsys := newBindingIndexTestFS(
		[]string{sourceAlias},
		map[string]string{sourceAlias: targetPath},
	)
	first := createBindingIndexTestProgram(t, fsys, sourceAlias)
	second := createBindingIndexTestProgram(t, fsys, sourceAlias)
	fsys.resetCalls()

	index := newProgramFileIndex(
		[]*compiler.Program{first, second},
		[]resolvedLintTarget{{Path: targetPath, CanonicalPath: targetPath}},
		fsys,
		true,
	)
	for programIndex := range index.programs {
		sourceFile := index.sourceFile([]int{programIndex}, programIndex, targetPath)
		if sourceFile == nil || sourceFile.FileName() != sourceAlias {
			t.Fatalf("Program %d lookup returned %v", programIndex, sourceFile)
		}
	}
	if calls := fsys.callCount(sourceAlias); calls != 1 {
		t.Fatalf("shared source identity resolved %d times, want 1", calls)
	}
}

func TestProgramFileIndex_PreservesLexicalAliasTieBreak(t *testing.T) {
	const targetPath = "/physical/target.ts"
	aliases := []string{"/aliases/z.ts", "/aliases/a.ts"}
	fsys := newBindingIndexTestFS(
		aliases,
		map[string]string{
			aliases[0]: targetPath,
			aliases[1]: targetPath,
		},
	)
	program := createBindingIndexTestProgram(t, fsys, aliases...)
	fsys.resetCalls()

	index := newProgramFileIndex(
		[]*compiler.Program{program},
		[]resolvedLintTarget{{Path: targetPath, CanonicalPath: targetPath}},
		fsys,
		false,
	)
	sourceFile := index.sourceFile([]int{0}, 0, targetPath)
	if sourceFile == nil || sourceFile.FileName() != aliases[1] {
		t.Fatalf("canonical alias tie-break returned %v, want %q", sourceFile, aliases[1])
	}
}

func TestProgramFileIndex_UsesTspathIdentityAcrossFilesystemRoots(t *testing.T) {
	tests := []struct {
		name       string
		sourcePath string
		targetPath string
	}{
		{
			name:       "drive",
			sourcePath: "C:/Repo/Alias.ts",
			targetPath: "C:/Physical/Alias.ts",
		},
		{
			name:       "unc",
			sourcePath: "//server/share/repo/Alias.ts",
			targetPath: "//server/share/physical/Alias.ts",
		},
		{
			name:       "drive to unc reparse target",
			sourcePath: "C:/Repo/Alias.ts",
			targetPath: "//server/share/physical/Alias.ts",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourceDirectory := tspath.GetDirectoryPath(test.sourcePath)
			targetDirectory := tspath.GetDirectoryPath(test.targetPath)
			otherSource := tspath.CombinePaths(sourceDirectory, "Other.ts")
			fsys := newBindingIndexTestFS(
				[]string{test.sourcePath, otherSource},
				map[string]string{sourceDirectory: targetDirectory},
			)
			fsys.entries[sourceDirectory] = vfs.Entries{
				Files:    []string{"Alias.ts", "Other.ts"},
				Symlinks: map[string]struct{}{},
			}
			program := createBindingIndexTestProgram(t, fsys, test.sourcePath, otherSource)
			fsys.resetCalls()

			index := newProgramFileIndex(
				[]*compiler.Program{program},
				[]resolvedLintTarget{{Path: test.targetPath, CanonicalPath: test.targetPath}},
				fsys,
				true,
			)
			sourceFile := index.sourceFile([]int{0}, 0, test.targetPath)
			if sourceFile == nil || sourceFile.FileName() != test.sourcePath {
				t.Fatalf("lookup returned %v, want %q", sourceFile, test.sourcePath)
			}
			if calls := fsys.callCount(sourceDirectory); calls != 1 {
				t.Fatalf("source directory resolved %d times, want 1", calls)
			}
			for _, sourcePath := range []string{test.sourcePath, otherSource} {
				if calls := fsys.callCount(sourcePath); calls != 0 {
					t.Fatalf("regular source %q unexpectedly used per-file Realpath %d time(s)", sourcePath, calls)
				}
			}
		})
	}
}
