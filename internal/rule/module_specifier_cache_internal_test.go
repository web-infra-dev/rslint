package rule

import (
	"testing"

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

var esm = ModuleSyntax{ESModule: true}

// TestModuleSpecifierCacheReusesOneFilesCollection locks in the point of the
// cache: a second run over the same unchanged file is answered from the first
// run's collection rather than by reading the file again.
func TestModuleSpecifierCacheReusesOneFilesCollection(t *testing.T) {
	program, _ := specifierCacheProgram(t, specifierCacheFiles(false))
	file := specifierCacheFile(t, program)
	cache := NewModuleSpecifierCache()

	first := cache.get(file, esm)
	second := cache.get(file, esm)

	if len(first) != 1 {
		t.Fatalf("collected %d specifiers, want 1", len(first))
	}
	if &first[0] != &second[0] {
		t.Fatal("the second request collected the file again")
	}
}

// TestModuleSpecifierCacheSeparatesSyntaxes locks in that one file's entry
// answers for the syntaxes it was asked about and no others: a `require` call
// is a reference under commonjs and nothing at all without it.
func TestModuleSpecifierCacheSeparatesSyntaxes(t *testing.T) {
	program, _ := specifierCacheProgram(t, specifierCacheFiles(true))
	file := specifierCacheFile(t, program)
	cache := NewModuleSpecifierCache()

	esmOnly := cache.get(file, esm)
	withCommonJS := cache.get(file, ModuleSyntax{ESModule: true, CommonJS: true})

	if len(esmOnly) != 1 {
		t.Fatalf("collected %d ES module specifiers, want 1", len(esmOnly))
	}
	if len(withCommonJS) != 2 {
		t.Fatalf("collected %d specifiers with commonjs, want 2", len(withCommonJS))
	}
	if reused := cache.get(file, esm); &reused[0] != &esmOnly[0] {
		t.Fatal("the commonjs request displaced the ES module answer")
	}
}

// TestModuleSpecifierCacheSupersedesReplacedFiles locks in what makes the
// cache safe to keep across an editing session without invalidating it: an
// entry answers for one file object, and an editor replaces that object
// whenever the text behind it changes.
func TestModuleSpecifierCacheSupersedesReplacedFiles(t *testing.T) {
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
	cache := NewModuleSpecifierCache()

	before := cache.get(original, ModuleSyntax{ESModule: true, CommonJS: true})
	after := cache.get(edited, ModuleSyntax{ESModule: true, CommonJS: true})

	if len(before) != 1 {
		t.Fatalf("collected %d specifiers before the edit, want 1", len(before))
	}
	if len(after) != 2 {
		t.Fatalf("collected %d specifiers after the edit, want 2", len(after))
	}
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
	cache := NewModuleSpecifierCache()

	before := NewCachedModuleGraph(program, cache).Edges(file, esm)
	after := NewCachedModuleGraph(updated, cache).Edges(file, esm)

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
