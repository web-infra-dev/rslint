package rule

import (
	"runtime"
	"testing"
	"time"
	"weak"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

const specifierCacheImporter = "/specifier-cache-fixture/importer.ts"
const specifierCacheTarget = "/specifier-cache-fixture/target.ts"

var specifierCacheESM = ModuleSyntax{ESModule: true}

// TestSourceFileModuleSpecifierCacheReusesCollection locks in the point of the
// cache: two runs holding the same unchanged SourceFile share its collection.
func TestSourceFileModuleSpecifierCacheReusesCollection(t *testing.T) {
	program, _ := specifierCacheProgram(t, specifierCacheFiles(false))
	file := specifierCacheFile(t, program)

	first := cachedModuleSpecifiers(file, specifierCacheESM)
	second := cachedModuleSpecifiers(file, specifierCacheESM)

	if len(first) != 1 {
		t.Fatalf("collected %d specifiers, want 1", len(first))
	}
	if &first[0] != &second[0] {
		t.Fatal("the second request collected the file again")
	}
}

// TestSourceFileModuleSpecifierCacheSeparatesSyntaxes locks in that one file
// answers for the syntaxes it was asked about and no others: a `require` call
// is a reference under commonjs and nothing at all without it.
func TestSourceFileModuleSpecifierCacheSeparatesSyntaxes(t *testing.T) {
	program, _ := specifierCacheProgram(t, specifierCacheFiles(true))
	file := specifierCacheFile(t, program)

	esmOnly := cachedModuleSpecifiers(file, specifierCacheESM)
	withCommonJS := cachedModuleSpecifiers(file, ModuleSyntax{ESModule: true, CommonJS: true})

	if len(esmOnly) != 1 {
		t.Fatalf("collected %d ES module specifiers, want 1", len(esmOnly))
	}
	if len(withCommonJS) != 2 {
		t.Fatalf("collected %d specifiers with commonjs, want 2", len(withCommonJS))
	}
	if reused := cachedModuleSpecifiers(file, specifierCacheESM); &reused[0] != &esmOnly[0] {
		t.Fatal("the commonjs request displaced the ES module answer")
	}
}

// TestSourceFileModuleSpecifierCacheSeparatesSourceFiles locks in that pointer
// identity, not a path, owns an answer. Two hosts can parse different text at
// the same path and their SourceFiles must never displace or answer for one
// another.
func TestSourceFileModuleSpecifierCacheSeparatesSourceFiles(t *testing.T) {
	originalProgram, _ := specifierCacheProgram(t, specifierCacheFiles(false))
	editedProgram, _ := specifierCacheProgram(t, specifierCacheFiles(true))
	original := specifierCacheFile(t, originalProgram)
	edited := specifierCacheFile(t, editedProgram)
	if original == edited {
		t.Fatal("the fixture programs share one file object; the test proves nothing")
	}
	if original.Path() != edited.Path() {
		t.Fatal("the fixture files disagree on their path; the test proves nothing")
	}
	syntax := ModuleSyntax{ESModule: true, CommonJS: true}
	before := cachedModuleSpecifiers(original, syntax)
	after := cachedModuleSpecifiers(edited, syntax)

	if len(before) != 1 {
		t.Fatalf("collected %d specifiers before the edit, want 1", len(before))
	}
	if len(after) != 2 {
		t.Fatalf("collected %d specifiers after the edit, want 2", len(after))
	}
	if reused := cachedModuleSpecifiers(original, syntax); &reused[0] != &before[0] {
		t.Fatal("the other SourceFile displaced the original answer")
	}
}

// TestSourceFileModuleSpecifierCachePublishesConcurrentCollections locks in
// concurrent editor runs sharing one SourceFile. Misses may redundantly perform
// the pure collection, but every caller must receive the single published
// read-only slice for its requested syntax.
func TestSourceFileModuleSpecifierCachePublishesConcurrentCollections(t *testing.T) {
	program, _ := specifierCacheProgram(t, specifierCacheFiles(true))
	file := specifierCacheFile(t, program)
	syntaxes := [...]ModuleSyntax{
		specifierCacheESM,
		{ESModule: true, CommonJS: true},
	}
	type result struct {
		syntaxIndex int
		collected   []moduleSpecifier
	}
	const callers = 64
	start := make(chan struct{})
	results := make(chan result, callers)
	for i := range callers {
		syntaxIndex := i % len(syntaxes)
		go func() {
			<-start
			results <- result{
				syntaxIndex: syntaxIndex,
				collected:   cachedModuleSpecifiers(file, syntaxes[syntaxIndex]),
			}
		}()
	}
	close(start)

	var published [len(syntaxes)][]moduleSpecifier
	for range callers {
		result := <-results
		wantLength := result.syntaxIndex + 1
		if len(result.collected) != wantLength {
			t.Fatalf("syntax %d collected %d specifiers, want %d", result.syntaxIndex, len(result.collected), wantLength)
		}
		if published[result.syntaxIndex] == nil {
			published[result.syntaxIndex] = result.collected
			continue
		}
		if &result.collected[0] != &published[result.syntaxIndex][0] {
			t.Fatalf("syntax %d published more than one collection", result.syntaxIndex)
		}
	}
}

// TestSourceFileModuleSpecifierCacheReleasesReplacedAndRemovedFiles reproduces
// the incremental lifetime boundary. Removing an import rebuilds the Program,
// replacing the importer and removing its dependency. Their attached answers
// must not make either old AST reachable while the new Program remains live.
func TestSourceFileModuleSpecifierCacheReleasesReplacedAndRemovedFiles(t *testing.T) {
	var replaced weak.Pointer[ast.SourceFile]
	var removed weak.Pointer[ast.SourceFile]
	var updated *compiler.Program

	func() {
		files := specifierCacheFiles(false)
		program, fs := specifierCacheProgram(t, files)
		importer := specifierCacheFile(t, program)
		target := program.GetSourceFile(specifierCacheTarget)
		if target == nil {
			t.Fatal("the fixture target is not in the program")
		}
		replaced = weak.Make(importer)
		removed = weak.Make(target)

		if collected := cachedModuleSpecifiers(importer, specifierCacheESM); len(collected) != 1 {
			t.Fatalf("collected %d importer specifiers, want 1", len(collected))
		}
		cachedModuleSpecifiers(target, specifierCacheESM)

		files[specifierCacheImporter] = "export const used = 1;\n"
		changed := tspath.ToPath(
			specifierCacheImporter,
			program.GetCurrentDirectory(),
			fs.UseCaseSensitiveFileNames(),
		)
		var reused bool
		updated, _, reused = program.UpdateProgram(changed, program.Host(), nil)
		if updated == nil || reused {
			t.Fatalf("the import graph change did not rebuild the program (reused=%v)", reused)
		}
		updated.BindSourceFiles()
		if specifierCacheFile(t, updated) == importer {
			t.Fatal("the rebuilt Program retained the replaced importer")
		}
		if updated.GetSourceFile(specifierCacheTarget) != nil {
			t.Fatal("the rebuilt Program retained the removed dependency")
		}
	}()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		runtime.GC()
		if replaced.Value() == nil && removed.Value() == nil {
			runtime.KeepAlive(updated)
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	runtime.KeepAlive(updated)
	t.Fatalf(
		"attached cache retained old ASTs (replaced=%v, removed=%v)",
		replaced.Value() != nil,
		removed.Value() != nil,
	)
}

// TestCachedModuleGraphResolvesPerProgram locks in the half the cache
// deliberately does not carry. What a specifier names is the Program's answer,
// not the file's: editing an imported file leaves the importer's own text — and
// so its file object, and so its cache entry — untouched, while the file it
// names becomes a different object. A cache that carried resolution would hand
// the rule the source file of a Program that is no longer being linted.
func TestCachedModuleGraphResolvesPerProgram(t *testing.T) {
	files := specifierCacheFiles(false)
	program, fs := specifierCacheProgram(t, files)
	file := specifierCacheFile(t, program)

	files[specifierCacheTarget] = "export const value = 2;\n"
	changed := tspath.ToPath(
		specifierCacheTarget,
		program.GetCurrentDirectory(),
		fs.UseCaseSensitiveFileNames(),
	)
	updated, _, reused := program.UpdateProgram(changed, program.Host(), nil)
	if updated == nil || !reused {
		t.Fatalf("the fixture edit did not update the program in place (reused=%v)", reused)
	}
	updated.BindSourceFiles()
	if specifierCacheFile(t, updated) != file {
		t.Fatal("the update replaced the importer too; the test proves nothing")
	}
	if updated.GetSourceFile(specifierCacheTarget) == program.GetSourceFile(specifierCacheTarget) {
		t.Fatal("the update reused the edited file; the test proves nothing")
	}
	before := NewCachedModuleGraph(program).Edges(file, specifierCacheESM)
	after := NewCachedModuleGraph(updated).Edges(file, specifierCacheESM)

	if len(before) != 1 || len(after) != 1 {
		t.Fatalf("resolved %d and %d edges, want 1 each", len(before), len(after))
	}
	if before[0].Target != program.GetSourceFile(specifierCacheTarget) {
		t.Fatal("the first run resolved outside its own program")
	}
	if after[0].Target != updated.GetSourceFile(specifierCacheTarget) {
		t.Fatal("the second run reused the first run's resolution")
	}
}

// specifierCacheFiles is a two-file project whose importer writes one static
// import, plus a `require` call when withRequire is set. The two shapes stand
// in for one file before and after an edit.
func specifierCacheFiles(withRequire bool) map[string]string {
	importer := "import { value } from './target';\nexport const used = value;\n"
	if withRequire {
		importer += "const also = require('./target');\nexport const second = also;\n"
	}
	return map[string]string{
		specifierCacheImporter: importer,
		specifierCacheTarget:   "export const value = 1;\n",
	}
}

// specifierCacheProgram builds a program over files, which it keeps reading:
// writing to the map and updating the program is how a test edits a file.
func specifierCacheProgram(t *testing.T, files map[string]string) (*compiler.Program, vfs.FS) {
	t.Helper()

	fs := rslint_utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), files)
	host := rslint_utils.CreateCompilerHost("/", fs)
	program, err := rslint_utils.CreateProgramFromOptions(
		true,
		&core.CompilerOptions{AllowJs: core.TSTrue},
		[]string{specifierCacheImporter},
		host,
	)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}
	program.BindSourceFiles()
	return program, fs
}

func specifierCacheFile(t *testing.T, program *compiler.Program) *ast.SourceFile {
	t.Helper()

	file := program.GetSourceFile(specifierCacheImporter)
	if file == nil {
		t.Fatal("the fixture importer is not in the program")
	}
	return file
}
