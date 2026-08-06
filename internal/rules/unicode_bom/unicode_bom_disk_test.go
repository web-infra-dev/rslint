package unicode_bom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
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
		{name: "no mark", bytes: source},
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
	host := utils.CreateCompilerHost(filepath.Dir(filePath), fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Module:       core.ModuleKindESNext,
		SkipLibCheck: core.TSTrue,
		Target:       core.ScriptTargetESNext,
	}, []string{filePath}, host)
	assert.NilError(t, err)

	ruleRan := false
	var diagnostics []rule.RuleDiagnostic
	linter.RunLinterInProgram(
		program,
		nil,
		nil,
		utils.ExcludePaths,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			if sourceFile.FileName() != filePath {
				return nil
			}
			return []linter.ConfiguredRule{{
				Name:     UnicodeBomRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					ruleRan = true
					return UnicodeBomRule.Run(ctx, []any{"never"})
				},
			}}
		},
		false,
		func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
		nil,
		nil,
	)
	assert.Assert(t, ruleRan, "the rule did not run for %s", filePath)

	return diagnostics
}

// utf16LE encodes ASCII source as little-endian UTF-16, without a mark — the
// caller prepends one.
func utf16LE(source string) string {
	encoded := make([]byte, 0, len(source)*2)
	for _, r := range source {
		encoded = append(encoded, byte(r), byte(r>>8))
	}
	return string(encoded)
}
