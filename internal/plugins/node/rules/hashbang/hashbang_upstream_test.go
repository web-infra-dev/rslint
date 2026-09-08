package hashbang

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func hashbangRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "hashbang-fixtures")
	archive := txtarfs.MustParseFile(t, "testdata/upstream.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		tspath.ResolvePath(directory, "tsconfig.json"): `{"compilerOptions":{"allowJs":true,"noEmit":true,"target":"esnext","module":"nodenext","jsx":"preserve"},"include":["**/*"]}`,
	}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(directory, name)] = string(data)
	}
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, files)}
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/hashbang.js
// Upstream tests import the deprecated shebang alias, which shares hashbang's implementation.
// Unnamed ESLint inputs are represented by a virtual file outside any package.
func TestHashbangUpstream(t *testing.T) {
	rule_tester.RunRuleTester(hashbangRoot(t), "tsconfig.json", t, &HashbangRule,
		[]rule_tester.ValidTestCase{
			// string-bin/bin/test.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "string-bin/bin/test.js"},
			// string-bin/lib/test.js
			{Code: "hello();", FileName: "string-bin/lib/test.js"},
			// object-bin/bin/a.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "object-bin/bin/a.js"},
			// object-bin/bin/b.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "object-bin/bin/b.js"},
			// string-bin/bin/test-env-flag.js
			{Code: "#!/usr/bin/env -S node\nhello();", FileName: "string-bin/bin/test.js"},
			// string-bin/bin/test-env-flag-node-flag.js
			{Code: "#!/usr/bin/env -S node --loader tsm\nhello();", FileName: "string-bin/bin/test.js"},
			// string-bin/bin/test-env-ignore-environment.js
			{Code: "#!/usr/bin/env --ignore-environment node\nhello();", FileName: "string-bin/bin/test.js"},
			// string-bin/bin/test-env-flags-node-flag.js
			{Code: "#!/usr/bin/env -i -S node --loader tsm\nhello();", FileName: "string-bin/bin/test.js"},
			// string-bin/bin/test-block-signal.js
			{Code: "#!/usr/bin/env --block-signal=SIGINT -S FOO=bar node --loader tsm\nhello();", FileName: "string-bin/bin/test.js"},
			// object-bin/bin/c.js
			{Code: "hello();", FileName: "object-bin/bin/c.js"},
			// no-bin-field/lib/test.js
			{Code: "hello();", FileName: "no-bin-field/lib/test.js"},
			// <input> with shebang
			{Code: "#!/usr/bin/env node\nhello();", FileName: "input.js"},
			// <input> without shebang
			{Code: "hello();", FileName: "input.js"},
			// convertPath - string-bin/src/bin/test.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "string-bin/src/bin/test.js", Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}},
			// convertPath - string-bin/src/lib/test.js
			{Code: "hello();", FileName: "string-bin/src/lib/test.js", Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}},
			// convertPath - object-bin/src/bin/a.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "object-bin/src/bin/a.js", Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}},
			// convertPath - object-bin/src/bin/b.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "object-bin/src/bin/b.js", Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}},
			// convertPath - object-bin/src/bin/c.js
			{Code: "hello();", FileName: "object-bin/src/bin/c.js", Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}},
			// convertPath - no-bin-field/src/lib/test.js
			{Code: "hello();", FileName: "no-bin-field/src/lib/test.js", Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}},
			// relative path - string-bin/bin/test.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "string-bin/bin/test.js"},
			// relative path - string-bin/lib/test.js
			{Code: "hello();", FileName: "string-bin/lib/test.js"},
			// BOM without newline
			{Code: "\uFEFFhello();", FileName: "string-bin/lib/test.js"},
			// BOM with newline
			{Code: "\uFEFFhello();\n", FileName: "string-bin/lib/test.js"},
			// with windows newline
			{Code: "hello();\r\n", FileName: "string-bin/lib/test.js"},
			// BOM with windows newline
			{Code: "\uFEFFhello();\r\n", FileName: "string-bin/lib/test.js"},
			// blank lines on the top of files.
			{Code: "\n\n\nhello();", FileName: "string-bin/lib/test.js"},
			// Shebang with CLI flags
			{Code: "#!/usr/bin/env node --harmony\nhello();", FileName: "string-bin/bin/test.js"},
			// use node resolution
			{Code: "#!/usr/bin/env node\nhello();", FileName: "object-bin/bin/index.js"},
			// published file cant have shebang
			{Code: "hello();", FileName: "unpublished/published.js", Options: []any{map[string]any{"ignoreUnpublished": true}}},
			// unpublished file can have shebang
			{Code: "#!/usr/bin/env node\nhello();", FileName: "unpublished/unpublished.js", Options: []any{map[string]any{"ignoreUnpublished": true}}},
			// unpublished file can have no shebang
			{Code: "hello();", FileName: "unpublished/unpublished.js", Options: []any{map[string]any{"ignoreUnpublished": true}}},
			// file matching additionalExecutables
			{Code: "#!/usr/bin/env node\nhello();", FileName: "unpublished/something.test.js", Options: []any{map[string]any{"additionalExecutables": []any{"*.test.js"}}}},
			// .ts maps to ts-node
			{Code: "#!/usr/bin/env ts-node\nhello();", FileName: "object-bin/bin/t.ts", Options: []any{map[string]any{"executableMap": map[string]any{".ts": "ts-node"}}}},
			// .ts maps to ts-node
			{Code: "#!/usr/bin/env node\nhello();", FileName: "object-bin/bin/a.js", Options: []any{map[string]any{"executableMap": map[string]any{".ts": "ts-node"}}}},
		}, []rule_tester.InvalidTestCase{
			// bin: string - match - no shebang
			{Code: "hello();", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// bin: string - match - incorrect shebang
			{Code: "#!/usr/bin/node\nhello();", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 16}}},
			// bin: string - no match - with shebang
			{Code: "#!/usr/bin/env node\nhello();", FileName: "string-bin/lib/test.js", Output: []string{"hello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// bin: {a: "./bin/a.js"} - match - no shebang
			{Code: "hello();", FileName: "object-bin/bin/a.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// bin: {b: "./bin/b.js"} - match - no shebang
			{Code: "#!/usr/bin/node\nhello();", FileName: "object-bin/bin/b.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 16}}},
			// bin: {c: "./bin"} - no match - with shebang
			{Code: "#!/usr/bin/env node\nhello();", FileName: "object-bin/bin/c.js", Output: []string{"hello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// bin: undefined - no match - with shebang
			{Code: "#!/usr/bin/env node\nhello();", FileName: "no-bin-field/lib/test.js", Output: []string{"hello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// convertPath in options
			{Code: "hello();", FileName: "string-bin/src/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// convertPath in settings
			{Code: "hello();", FileName: "string-bin/src/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Settings: map[string]any{"node": map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// converted path - string-bin/src/bin/test.js
			{Code: "#!/usr/bin/node\nhello();", FileName: "string-bin/src/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 16}}},
			// converted path - string-bin/src/lib/test.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "string-bin/src/lib/test.js", Output: []string{"hello();"}, Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// converted path - object-bin/src/bin/a.js
			{Code: "hello();", FileName: "object-bin/src/bin/a.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// converted path - object-bin/src/bin/b.js
			{Code: "#!/usr/bin/node\nhello();", FileName: "object-bin/src/bin/b.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 16}}},
			// converted path - object-bin/src/bin/c.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "object-bin/src/bin/c.js", Output: []string{"hello();"}, Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// converted path - no-bin-field/src/lib/test.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "no-bin-field/src/lib/test.js", Output: []string{"hello();"}, Options: []any{map[string]any{"convertPath": map[string]any{"src/**": []any{"^src/(.+)$", "$1"}}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// relative path - string-bin/bin/test.js
			{Code: "hello();", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// relative path - string-bin/lib/test.js
			{Code: "#!/usr/bin/env node\nhello();", FileName: "string-bin/lib/test.js", Output: []string{"hello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// header comments
			{Code: "/* header */\nhello();", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\n/* header */\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 13}}},
			// BOM and line endings
			{Code: "\uFEFFhello();", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// BOM and line endings
			{Code: "hello();\n", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// BOM and line endings
			{Code: "hello();\r\n", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();\r\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// BOM and line endings
			{Code: "\uFEFFhello();\n", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// BOM and line endings
			{Code: "\uFEFFhello();\r\n", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();\r\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// BOM and line endings
			{Code: "#!/usr/bin/env node\r\nhello();", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedLF", Message: "This file must have Unix linebreaks (LF).", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// BOM and line endings
			{Code: "\uFEFF#!/usr/bin/env node\nhello();", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedBOM", Message: "This file must not have Unicode BOM.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// BOM and line endings
			{Code: "\uFEFF#!/usr/bin/env node\r\nhello();", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedBOM", Message: "This file must not have Unicode BOM.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}, {MessageId: "expectedLF", Message: "This file must have Unix linebreaks (LF).", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// Shebang with CLI flags
			{Code: "#!/usr/bin/env node --harmony\nhello();", FileName: "string-bin/lib/test.js", Output: []string{"hello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
			// use node resolution
			{Code: "hello();", FileName: "object-bin/bin/index.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// unpublished file should not have shebang
			{Code: "#!/usr/bin/env node\nhello();", FileName: "unpublished/unpublished.js", Output: []string{"hello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// published file should have shebang
			{Code: "#!/usr/bin/env node\nhello();", FileName: "unpublished/published.js", Output: []string{"hello();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// unpublished file shebang ignored
			{Code: "#!/usr/bin/env node\nhello();", FileName: "unpublished/published.js", Output: []string{"hello();"}, Options: []any{map[string]any{"ignoreUnpublished": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// executable in additionalExecutables without shebang
			{Code: "hello();", FileName: "unpublished/something.test.js", Output: []string{"#!/usr/bin/env node\nhello();"}, Options: []any{map[string]any{"additionalExecutables": []any{"*.test.js"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// file not in additionalExecutables with shebang
			{Code: "#!/usr/bin/env node\nhello();", FileName: "unpublished/not-a-test.js", Output: []string{"hello();"}, Options: []any{map[string]any{"additionalExecutables": []any{"*.test.js"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			// .ts maps to ts-node
			{Code: "hello();", FileName: "object-bin/bin/t.ts", Output: []string{"#!/usr/bin/env ts-node\nhello();"}, Options: []any{map[string]any{"executableMap": map[string]any{".ts": "ts-node"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env ts-node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			// .ts maps to ts-node
			{Code: "#!/usr/bin/env node\nhello();", FileName: "object-bin/bin/t.ts", Output: []string{"#!/usr/bin/env ts-node\nhello();"}, Options: []any{map[string]any{"executableMap": map[string]any{".ts": "ts-node"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env ts-node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
		})
}

// All six JavaScript examples from the pinned rule documentation, with the
// described BOM/CRLF applied to the source rather than included in a comment.
func TestHashbangDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(hashbangRoot(t), "tsconfig.json", t, &HashbangRule,
		[]rule_tester.ValidTestCase{
			{Code: "#!/usr/bin/env node\nconsole.log(\"hello\");", FileName: "string-bin/bin/test.js"},
			{Code: "console.log(\"hello\");", FileName: "string-bin/lib/test.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "console.log(\"hello\");", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nconsole.log(\"hello\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbangNode", Message: "This file needs shebang \"#!/usr/bin/env node\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 22}}},
			{Code: "\uFEFF#!/usr/bin/env node\nconsole.log(\"hello\");", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nconsole.log(\"hello\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedBOM", Message: "This file must not have Unicode BOM.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			{Code: "#!/usr/bin/env node\r\nconsole.log(\"hello\");", FileName: "string-bin/bin/test.js", Output: []string{"#!/usr/bin/env node\nconsole.log(\"hello\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedLF", Message: "This file must have Unix linebreaks (LF).", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
			{Code: "#!/usr/bin/env node\nconsole.log(\"hello\");", FileName: "string-bin/lib/test.js", Output: []string{"console.log(\"hello\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedHashbang", Message: "This file needs no shebang.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}}},
		})
}
