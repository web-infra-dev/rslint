package no_useless_path_segments_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	target "github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_useless_path_segments"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/embedfs"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func pathRoot(t *testing.T) embedfs.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = "/no-useless-path-segments"
	files := map[string]string{}
	for _, fixture := range []string{"upstream", "extras"} {
		archive := txtarfs.MustParseFile(t, "testdata/"+fixture+".txtar")
		names, err := archive.FileNames("")
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			data, err := archive.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			files[tspath.ResolvePath(root.Dir, name)] = string(data)
		}
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	return root
}

// The upstream rule deliberately has no message ID. This helper asserts the
// entire literal range (UTF-16 columns), exact message and every autofix pass.
func invalidPath(code, literal, imported, proposed, output string, options any, followingOutputs ...string) rule_tester.InvalidTestCase {
	start := strings.Index(code, literal)
	if start < 0 {
		panic("missing test literal")
	}
	prefix := code[:start]
	line := strings.Count(prefix, "\n") + 1
	column := ecmascript.StringCodeUnitCount(prefix[strings.LastIndex(prefix, "\n")+1:]) + 1
	endLine := line + strings.Count(literal, "\n")
	endColumn := column + ecmascript.StringCodeUnitCount(literal)
	if endLine != line {
		endColumn = ecmascript.StringCodeUnitCount(literal[strings.LastIndex(literal, "\n")+1:]) + 1
	}
	return rule_tester.InvalidTestCase{
		Code: code, FileName: "files/foo.js", Options: options, Output: append([]string{output}, followingOutputs...),
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "", Message: `Useless path segments for "` + imported + `", should be "` + proposed + `"`,
			Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
		}},
	}
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-useless-path-segments.js
// Upstream repeats the same cases under node/webpack labels, but never passes
// the label into settings. Both runs therefore use the default Node resolver.
// Babel's dynamic import cases run with the native parser here.
func TestNoUselessPathSegmentsUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "require(\"./../files/malformed.js\")", FileName: "files/foo.js", Options: nil},
		{Code: "import \"./malformed.js\"", FileName: "files/foo.js", Options: nil},
		{Code: "import \"./test-module\"", FileName: "files/foo.js", Options: nil},
		{Code: "import \"./bar/\"", FileName: "files/foo.js", Options: nil},
		{Code: "import \".\"", FileName: "files/foo.js", Options: nil},
		{Code: "import \"..\"", FileName: "files/foo.js", Options: nil},
		{Code: "import fs from \"fs\"", FileName: "files/foo.js", Options: nil},
		{Code: "import \"../index\"", FileName: "files/foo.js", Options: nil},
		{Code: "import \"../my-custom-index\"", FileName: "files/foo.js", Options: map[string]any{"noUselessIndex": true}},
		{Code: "import \"./bar.js\"", FileName: "files/foo.js", Options: map[string]any{"noUselessIndex": true}},
		{Code: "import \"./bar\"", FileName: "files/foo.js", Options: map[string]any{"noUselessIndex": true}},
		{Code: "import \"./bar/\"", FileName: "files/foo.js", Options: map[string]any{"noUselessIndex": true}},
		{Code: "import \"./malformed.js\"", FileName: "files/foo.js", Options: map[string]any{"noUselessIndex": true}},
		{Code: "import \"./malformed\"", FileName: "files/foo.js", Options: map[string]any{"noUselessIndex": true}},
		{Code: "import \"./importType\"", FileName: "files/foo.js", Options: map[string]any{"noUselessIndex": true}},
		{Code: "import(\".\")", FileName: "files/foo.js", Options: nil},
		{Code: "import(\"..\")", FileName: "files/foo.js", Options: nil},
		{Code: "import(\"fs\").then(function(fs) {})", FileName: "files/foo.js", Options: nil},
	}
	invalid := []rule_tester.InvalidTestCase{
		// CommonJS
		invalidPath("require(\"./../files/malformed.js\")", "\"./../files/malformed.js\"", "./../files/malformed.js", "../files/malformed.js", "require(\"../files/malformed.js\")", map[string]any{"commonjs": true}, "require(\"./malformed.js\")"),
		invalidPath("require(\"./../files/malformed\")", "\"./../files/malformed\"", "./../files/malformed", "../files/malformed", "require(\"../files/malformed\")", map[string]any{"commonjs": true}, "require(\"./malformed\")"),
		invalidPath("require(\"../files/malformed.js\")", "\"../files/malformed.js\"", "../files/malformed.js", "./malformed.js", "require(\"./malformed.js\")", map[string]any{"commonjs": true}),
		invalidPath("require(\"../files/malformed\")", "\"../files/malformed\"", "../files/malformed", "./malformed", "require(\"./malformed\")", map[string]any{"commonjs": true}),
		invalidPath("require(\"./test-module/\")", "\"./test-module/\"", "./test-module/", "./test-module", "require(\"./test-module\")", map[string]any{"commonjs": true}),
		invalidPath("require(\"./\")", "\"./\"", "./", ".", "require(\".\")", map[string]any{"commonjs": true}),
		invalidPath("require(\"../\")", "\"../\"", "../", "..", "require(\"..\")", map[string]any{"commonjs": true}),
		invalidPath("require(\"./deep//a\")", "\"./deep//a\"", "./deep//a", "./deep/a", "require(\"./deep/a\")", map[string]any{"commonjs": true}),
		// CommonJS with noUselessIndex
		invalidPath("require(\"./bar/index.js\")", "\"./bar/index.js\"", "./bar/index.js", "./bar/", "require(\"./bar/\")", map[string]any{"commonjs": true, "noUselessIndex": true}),
		invalidPath("require(\"./bar/index\")", "\"./bar/index\"", "./bar/index", "./bar/", "require(\"./bar/\")", map[string]any{"commonjs": true, "noUselessIndex": true}),
		invalidPath("require(\"./importPath/\")", "\"./importPath/\"", "./importPath/", "./importPath", "require(\"./importPath\")", map[string]any{"commonjs": true, "noUselessIndex": true}),
		invalidPath("require(\"./importPath/index.js\")", "\"./importPath/index.js\"", "./importPath/index.js", "./importPath", "require(\"./importPath\")", map[string]any{"commonjs": true, "noUselessIndex": true}),
		invalidPath("require(\"./importType/index\")", "\"./importType/index\"", "./importType/index", "./importType", "require(\"./importType\")", map[string]any{"commonjs": true, "noUselessIndex": true}),
		invalidPath("require(\"./index\")", "\"./index\"", "./index", ".", "require(\".\")", map[string]any{"commonjs": true, "noUselessIndex": true}),
		invalidPath("require(\"../index\")", "\"../index\"", "../index", "..", "require(\"..\")", map[string]any{"commonjs": true, "noUselessIndex": true}),
		invalidPath("require(\"../index.js\")", "\"../index.js\"", "../index.js", "..", "require(\"..\")", map[string]any{"commonjs": true, "noUselessIndex": true}),
		// ES modules
		invalidPath("import \"./../files/malformed.js\"", "\"./../files/malformed.js\"", "./../files/malformed.js", "../files/malformed.js", "import \"../files/malformed.js\"", nil, "import \"./malformed.js\""),
		invalidPath("import \"./../files/malformed\"", "\"./../files/malformed\"", "./../files/malformed", "../files/malformed", "import \"../files/malformed\"", nil, "import \"./malformed\""),
		invalidPath("import \"../files/malformed.js\"", "\"../files/malformed.js\"", "../files/malformed.js", "./malformed.js", "import \"./malformed.js\"", nil),
		invalidPath("import \"../files/malformed\"", "\"../files/malformed\"", "../files/malformed", "./malformed", "import \"./malformed\"", nil),
		invalidPath("import \"./test-module/\"", "\"./test-module/\"", "./test-module/", "./test-module", "import \"./test-module\"", nil),
		invalidPath("import \"./\"", "\"./\"", "./", ".", "import \".\"", nil),
		invalidPath("import \"../\"", "\"../\"", "../", "..", "import \"..\"", nil),
		invalidPath("import \"./deep//a\"", "\"./deep//a\"", "./deep//a", "./deep/a", "import \"./deep/a\"", nil),
		// ES modules with noUselessIndex
		invalidPath("import \"./bar/index.js\"", "\"./bar/index.js\"", "./bar/index.js", "./bar/", "import \"./bar/\"", map[string]any{"noUselessIndex": true}),
		invalidPath("import \"./bar/index\"", "\"./bar/index\"", "./bar/index", "./bar/", "import \"./bar/\"", map[string]any{"noUselessIndex": true}),
		invalidPath("import \"./importPath/\"", "\"./importPath/\"", "./importPath/", "./importPath", "import \"./importPath\"", map[string]any{"noUselessIndex": true}),
		invalidPath("import \"./importPath/index.js\"", "\"./importPath/index.js\"", "./importPath/index.js", "./importPath", "import \"./importPath\"", map[string]any{"noUselessIndex": true}),
		invalidPath("import \"./importPath/index\"", "\"./importPath/index\"", "./importPath/index", "./importPath", "import \"./importPath\"", map[string]any{"noUselessIndex": true}),
		invalidPath("import \"./index\"", "\"./index\"", "./index", ".", "import \".\"", map[string]any{"noUselessIndex": true}),
		invalidPath("import \"../index\"", "\"../index\"", "../index", "..", "import \"..\"", map[string]any{"noUselessIndex": true}),
		invalidPath("import \"../index.js\"", "\"../index.js\"", "../index.js", "..", "import \"..\"", map[string]any{"noUselessIndex": true}),
		// Dynamic imports
		invalidPath("import(\"./\")", "\"./\"", "./", ".", "import(\".\")", nil),
		invalidPath("import(\"../\")", "\"../\"", "../", "..", "import(\"..\")", nil),
		invalidPath("import(\"./deep//a\")", "\"./deep//a\"", "./deep//a", "./deep/a", "import(\"./deep/a\")", nil),
	}
	rule_tester.RunRuleTester(pathRoot(t), "tsconfig.json", t, &target.NoUselessPathSegmentsRule, valid, invalid)
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-useless-path-segments.md
// The doc's /index examples require the option described later in that page.
func TestNoUselessPathSegmentsDocumentation(t *testing.T) {
	valid := []rule_tester.ValidTestCase{}
	for _, code := range []string{`import "./header.js";`, `import "./pages";`, `import "./pages/about";`, `import ".";`, `import "..";`, `import fs from "fs";`} {
		valid = append(valid, rule_tester.ValidTestCase{Code: code, FileName: "my-project/app.js"})
	}
	invalid := []rule_tester.InvalidTestCase{}
	for _, pair := range [][2]string{
		{"./../my-project/pages/about.js", "../my-project/pages/about.js"},
		{"./../my-project/pages/about", "../my-project/pages/about"},
		{"../my-project/pages/about.js", "./pages/about.js"},
		{"../my-project/pages/about", "./pages/about"},
		{"./pages//about", "./pages/about"}, {"./pages/", "./pages"},
		{"./pages/index", "./pages"}, {"./pages/index.js", "./pages"},
		{"./helpers/index", "./helpers/"},
	} {
		literal := `"` + pair[0] + `"`
		item := invalidPath("import "+literal+";", literal, pair[0], pair[1], `import "`+pair[1]+`";`, map[string]any{"noUselessIndex": true})
		if strings.HasPrefix(pair[0], "./../my-project/") {
			item.Output = append(item.Output, `import "./`+strings.TrimPrefix(pair[1], "../my-project/")+`";`)
		}
		item.FileName = "my-project/app.js"
		invalid = append(invalid, item)
	}
	rule_tester.RunRuleTester(pathRoot(t), "tsconfig.json", t, &target.NoUselessPathSegmentsRule, valid, invalid)
}
