package unicode_bom

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/testutil"
	"github.com/web-infra-dev/rslint/internal/utils"
	"gotest.tools/v3/assert"
)

// TestUnicodeBomOnDisk lints real files through the real file system. Every
// other test in this package hands the rule its source as caller-supplied
// content; a file read off disk reaches the rule by a different route, and the
// answer has to be the same one.
func TestUnicodeBomOnDisk(t *testing.T) {
	t.Parallel()

	const source = "var a = 123;\n"

	for _, tc := range []struct {
		name       string
		bytes      string
		wantText   string
		wantReport bool
	}{
		{name: "utf-8 mark", bytes: "\uFEFF" + source, wantReport: true},
		// A UTF-16 file is decoded to UTF-8 and loses its mark on the way, so
		// it counts as marked just like a UTF-8 one.
		{name: "utf-16 mark", bytes: "\xFF\xFE" + utf16LE(source), wantReport: true},
		{name: "utf-16 big-endian mark", bytes: "\xFE\xFF" + utf16BE(source), wantReport: true},
		// The raw UTF-16 byte count equals the decoded UTF-8 byte count here.
		// Non-ASCII text must therefore stay on the authoritative header path.
		{name: "equal-length utf-16 mark", bytes: "\xFF\xFE" + utf16LE("日本"), wantText: "日本", wantReport: true},
		{name: "no mark", bytes: source},
		{name: "no mark with non-ascii text", bytes: "const 日本 = 1;\n"},
		// Real-user: eslint#6580 / #4878 — files written by Windows editors
		// where the mark sits ahead of a shebang.
		{
			name:       "mark ahead of a shebang",
			bytes:      "\uFEFF#!/usr/bin/env node\n" + source,
			wantText:   "#!/usr/bin/env node\n" + source,
			wantReport: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filePath := writeFixture(t, tc.bytes)
			diagnostics := lintOnDisk(t, filePath)

			if !tc.wantReport {
				assert.Equal(t, 0, len(diagnostics), "expected no diagnostic")
				return
			}

			assert.Equal(t, 1, len(diagnostics))
			assert.Equal(t, "Unexpected Unicode BOM (Byte Order Mark).", diagnostics[0].Message.Description)
			wantText := tc.wantText
			if wantText == "" {
				wantText = source
			}
			assert.Equal(t, wantText, diagnostics[0].SourceFile.Text(),
				"the mark never reaches the source text")
		})
	}
}

// TestUnicodeBomUnmarkedDiskFastPath proves that an unchanged, unmarked ASCII
// disk file is decided from metadata without asking the BOM source to reopen
// the file. The ordinary diagnostics tests continue to lock in the result.
func TestUnicodeBomUnmarkedDiskFastPath(t *testing.T) {
	t.Parallel()

	filePath := writeFixture(t, "var a = 123;\n")
	fs := &countingBOMFileSystem{FS: bundled.WrapFS(cachedvfs.From(osvfs.FS()))}

	assert.Equal(t, 0, len(lintFile(t, filePath, fs)))
	assert.Equal(t, int64(0), fs.calls.Load(), "the fast path must not reread the file header")
}

// TestUnicodeBomFastPathDefersToOverlay ensures a virtual source remains
// authoritative even when it shadows a real disk file whose metadata would
// otherwise qualify for the shortcut.
func TestUnicodeBomFastPathDefersToOverlay(t *testing.T) {
	t.Parallel()

	const source = "var a = 123;\n"
	filePath := writeFixture(t, source)
	fs := utils.NewOverlayVFS(
		bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		map[string]string{filePath: utils.BOM + source},
	)

	diagnostics := lintFile(t, filePath, fs)
	assert.Equal(t, 1, len(diagnostics))
	assert.Equal(t, "unexpected", diagnostics[0].Message.Id)
}

// TestUnicodeBomFastPathRejectsStaleStat covers the BOM-only rewrite used by
// multi-pass fixing. cachedvfs can still hold the marked file's old size; that
// metadata must not make the newly unmarked source look marked again.
func TestUnicodeBomFastPathRejectsStaleStat(t *testing.T) {
	t.Parallel()

	const source = "var a = 123;\n"
	filePath := writeFixture(t, utils.BOM+source)
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	assert.Assert(t, fs.Stat(filePath) != nil, "failed to prime the VFS stat cache")
	assert.NilError(t, os.WriteFile(filePath, []byte(source), 0o644))

	assert.Equal(t, 0, len(lintFile(t, filePath, fs)),
		"stale marked-file metadata must fall back to the current header")
}

// TestUnicodeBomFixRewritesFile walks a marked file through a full lint, fix
// and write cycle the way the CLI does — putting the mark back in front of the
// parsed text before fixing, because that is what the file holds — then lints
// the written bytes again.
func TestUnicodeBomFixRewritesFile(t *testing.T) {
	t.Parallel()

	const source = "var a = 123;\n"
	filePath := writeFixture(t, utils.BOM+source)

	diagnostics := lintOnDisk(t, filePath)
	assert.Equal(t, 1, len(diagnostics))

	original := utils.BOM + diagnostics[0].SourceFile.Text()
	fixed, _, wasFixed := linter.ApplyRuleFixes(original, diagnostics)
	assert.Assert(t, wasFixed, "the diagnostic must carry an applicable fix")
	assert.NilError(t, os.WriteFile(filePath, []byte(fixed), 0o644))

	written, err := os.ReadFile(filePath)
	assert.NilError(t, err)
	assert.Equal(t, source, string(written), "the written bytes must have lost the mark")

	assert.Equal(t, 0, len(lintOnDisk(t, filePath)),
		"the file must be clean once its own fix has been written back")
}

// writeFixture puts bytes on disk verbatim, byte order mark included, and
// returns the normalized path a Program will know the file by.
func writeFixture(t *testing.T, bytes string) string {
	t.Helper()

	filePath := tspath.NormalizePath(filepath.Join(t.TempDir(), "file.ts"))
	assert.NilError(t, os.WriteFile(filePath, []byte(bytes), 0o644))
	return filePath
}

// lintOnDisk runs the rule over a real file with no overlay in the way.
func lintOnDisk(t *testing.T, filePath string) []rule.RuleDiagnostic {
	t.Helper()

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	return lintFile(t, filePath, fs)
}

func lintFile(t *testing.T, filePath string, fs vfs.FS) []rule.RuleDiagnostic {
	t.Helper()

	host := utils.CreateCompilerHost(filepath.Dir(filePath), fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Module:       core.ModuleKindESNext,
		SkipLibCheck: core.TSTrue,
		Target:       core.ScriptTargetESNext,
	}, []string{filePath}, host)
	assert.NilError(t, err)

	ruleRan := false
	var diagnostics []rule.RuleDiagnostic
	testutil.LintProgram(t, testutil.LintProgramOptions{
		Program:                lintprogram.NewFromCompiler(program),
		ExcludedPathSubstrings: testutil.DefaultExcludedPathSubstrings,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
			if sourceFile.FileName() != filePath {
				return nil
			}
			return []rule.ConfiguredRule{{
				Name:     UnicodeBomRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					ruleRan = true
					return UnicodeBomRule.Run(ctx, []any{"never"})
				},
			}}
		},
		OnDiagnostic: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})
	assert.Assert(t, ruleRan, "the rule did not run for %s", filePath)

	return diagnostics
}

type countingBOMFileSystem struct {
	vfs.FS
	calls atomic.Int64
}

func (fs *countingBOMFileSystem) SourceHasBOM(path string) bool {
	fs.calls.Add(1)
	return utils.SourceHasBOM(fs.FS, path)
}

// utf16LE encodes BMP source as little-endian UTF-16, without a mark — the
// caller prepends one.
func utf16LE(source string) string {
	encoded := make([]byte, 0, len(source)*2)
	for _, r := range source {
		encoded = append(encoded, byte(r), byte(r>>8))
	}
	return string(encoded)
}

func utf16BE(source string) string {
	encoded := make([]byte, 0, len(source)*2)
	for _, r := range source {
		encoded = append(encoded, byte(r>>8), byte(r))
	}
	return string(encoded)
}
