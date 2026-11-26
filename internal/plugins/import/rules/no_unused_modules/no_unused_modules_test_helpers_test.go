package no_unused_modules

import (
	"sync"
	"testing"
	"testing/fstest"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// fixtureFiles holds testdata/no_unused_modules.txtar below /testdata.
var fixtureFiles = sync.OnceValues(func() (fstest.MapFS, error) {
	archive, err := txtarfs.ParseFile("testdata/no_unused_modules.txtar")
	if err != nil {
		return nil, err
	}
	names, err := archive.FileNames("")
	if err != nil {
		return nil, err
	}
	files := make(fstest.MapFS, len(names))
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			return nil, err
		}
		files["testdata/"+name] = &fstest.MapFile{Data: data, Mode: 0o644}
	}
	return files, nil
})

const tsconfig = "testdata/tsconfig.json"

// fixturePath mirrors upstream's testFilePath for tests/files.
func fixturePath(relative string) string {
	return tspath.ResolvePath("/testdata", relative)
}

// fixtureRoot returns the archived fixtures with files replaced or added.
// Upstream keeps one module-level index across test cases, so a case often
// depends on what an earlier case linted into another file; overlays state
// that earlier file text explicitly instead.
func fixtureRoot(t *testing.T, overlays map[string]string) rule_tester.Root {
	t.Helper()
	files, err := fixtureFiles()
	if err != nil {
		t.Fatal(err)
	}
	base := bundled.WrapFS(iovfs.From(files, true))
	if len(overlays) == 0 {
		return rule_tester.Root{Dir: "/", FS: base}
	}
	virtual := make(map[string]string, len(overlays))
	for name, text := range overlays {
		virtual[tspath.ResolvePath("/", name)] = text
	}
	return rule_tester.Root{Dir: "/", FS: utils.NewOverlayVFS(base, virtual)}
}

// runGroup runs one upstream ruleTester.run block against the fixtures plus
// overlays.
func runGroup(t *testing.T, name string, overlays map[string]string, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		rule_tester.RunRuleTester(fixtureRoot(t, overlays), tsconfig, t, &NoUnusedModulesRule, valid, invalid)
	})
}

func unusedError(name string, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unused",
		Message:   "exported declaration '" + name + "' not used within other modules",
		Line:      line,
		Column:    column,
	}
}

func unusedRange(name string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	err := unusedError(name, line, column)
	err.EndLine = endLine
	err.EndColumn = endColumn
	return err
}

func noExportsError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noExports",
		Message:   "No exports found",
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}
