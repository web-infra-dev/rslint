// Ported from eslint-plugin-unicorn v77.0.0 (MIT).
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/no-empty-file.js
package no_empty_file_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_empty_file"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func emptyFileError(line, column, endLine, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "no-empty-file", Message: "Empty files are not allowed.",
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}}
}

// Complete upstream snapshot group, retaining unsupported cases as explained skips.
func TestNoEmptyFileUpstreamSnapshot(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_empty_file.NoEmptyFileRule,
		[]rule_tester.ValidTestCase{
			{Code: "const x = 0;", FileName: "example.js"},
			{Code: ";; const x = 0;", FileName: "example.js"},
			{Code: "{{{;;const x = 0;}}}", FileName: "example.js"},
			{Code: "'use strict';\nconst x = 0;", FileName: "example.js"},
			{Code: ";;'use strict';", FileName: "example.js"},
			{Code: "{'use strict';}", FileName: "example.js"},
			{Code: "(\"use strict\")", FileName: "example.js"},
			{Code: "`use strict`", FileName: "example.js"},
			{Code: "({})", FileName: "example.js"},
			{Code: "#!/usr/bin/env node\nconsole.log('done');", FileName: "example.js"},
			{Code: "false", FileName: "example.js"},
			{Code: "(\"\")", FileName: "example.js"},
			{Code: "NaN", FileName: "example.js"},
			{Code: "undefined", FileName: "example.js"},
			{Code: "null", FileName: "example.js"},
			{Code: "[]", FileName: "example.js"},
			{Code: "(() => {})()", FileName: "example.js"},
			{Code: "/// <reference types=\"example\" />", FileName: "example.d.ts"},
			{Code: "/// <reference types=\"example\" />", FileName: "example.ts"},
			// Skipped: rslint does not support the vue parser.
			{Code: "<template><div/></template>", FileName: "example.vue", Skip: true},
			// Skipped: rslint does not support the vue parser.
			{Code: "<template><div/></template>\n<script></script>", FileName: "example.vue", Skip: true},
			// Skipped: rslint does not support the html parser.
			{Code: "<div>Hello</div>", FileName: "example.html", Skip: true},
			// Skipped: rslint does not support the html parser.
			{Code: "<!DOCTYPE html>", FileName: "example.html", Skip: true},
			// Skipped: rslint does not support the html parser.
			{Code: "<!-- comment -->", FileName: "example.html", Options: map[string]any{"allowComments": true}, Skip: true},
			// Skipped: rslint does not support the css language.
			{Code: "a { color: red; }", FileName: "example.css", Skip: true},
			// Skipped: rslint does not support the css language.
			{Code: "/* comment */", FileName: "example.css", Options: map[string]any{"allowComments": true}, Skip: true},
			// Skipped: rslint does not support the markdown language.
			{Code: "# Title", FileName: "example.md", Skip: true},
			// Skipped: rslint does not support the markdown language.
			{Code: "<!-- a --> text <!-- b -->", FileName: "example.md", Skip: true},
			// Skipped: rslint does not support the markdown language.
			{Code: "<!-- comment -->", FileName: "example.md", Options: map[string]any{"allowComments": true}, Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "key: value", FileName: "example.yaml", Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "&anchor value", FileName: "example.yaml", Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "---\n---\nkey: value", FileName: "example.yaml", Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "# comment", FileName: "example.yaml", Options: map[string]any{"allowComments": true}, Skip: true},
			{Code: "// comment", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "/* comment */", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "/**\n * @typedef {object} Foo\n * @property {string} bar\n */", FileName: "example.js", Options: map[string]any{"allowComments": true}},
			{Code: "// No need to write tests here.", FileName: "example.test.ts", Options: map[string]any{"allowComments": true}},
			// Skipped: rslint does not support the plain-text parser.
			{Code: "# Logs\nnode_modules\n", FileName: ".gitignore", Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "key = \"value\"", FileName: "example.toml", Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "[table]", FileName: "example.toml", Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "[[table]]", FileName: "example.toml", Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "items = []", FileName: "example.toml", Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "table = {}", FileName: "example.toml", Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "# comment", FileName: "example.toml", Options: map[string]any{"allowComments": true}, Skip: true},
		}, []rule_tester.InvalidTestCase{
			{Code: "", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 1)},
			{Code: "\uFEFF", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 1)},
			{Code: " ", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 2)},
			{Code: "\t", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 2)},
			{Code: "\n", FileName: "example.js", Errors: emptyFileError(1, 1, 2, 1)},
			{Code: "\r", FileName: "example.js", Errors: emptyFileError(1, 1, 2, 1)},
			{Code: "\r\n", FileName: "example.js", Errors: emptyFileError(1, 1, 2, 1)},
			{Code: "", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 1)},
			{Code: "// comment", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 11)},
			{Code: "/* comment */", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 14)},
			{Code: "#!/usr/bin/env node", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 20)},
			{Code: "'use asm';", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 11)},
			{Code: "'use strict';", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 14)},
			{Code: "\"use strict\"", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 13)},
			{Code: "\"\"", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 3)},
			{Code: ";", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 2)},
			{Code: ";;", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 3)},
			{Code: "{}", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 3)},
			{Code: "{;;}", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 5)},
			{Code: "{{}}", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 5)},
			{Code: "", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 1, 1)},
			{Code: " ", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 1, 2)},
			{Code: "#!/usr/bin/env node", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 1, 20)},
			{Code: "#!/usr/bin/env node\n// comment", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 2, 11)},
			{Code: "; // comment", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 1, 13)},
			{Code: "'use strict'; // comment", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 1, 25)},
			{Code: "{/* comment */}", FileName: "example.js", Options: map[string]any{"allowComments": true}, Errors: emptyFileError(1, 1, 1, 16)},
			{Code: "{}", FileName: "example.mjs", Errors: emptyFileError(1, 1, 1, 3)},
			// Skipped: rslint does not support the mixed-case extensions (TypeScript file discovery is case-sensitive).
			{Code: "{}", FileName: "example.cJs", Skip: true},
			{Code: "{}", FileName: "example.ts", Errors: emptyFileError(1, 1, 1, 3)},
			{Code: "{}", FileName: "example.tsx", Errors: emptyFileError(1, 1, 1, 3)},
			{Code: "{}", FileName: "example.jsx", Errors: emptyFileError(1, 1, 1, 3)},
			// Skipped: rslint does not support the mixed-case extensions (TypeScript file discovery is case-sensitive).
			{Code: "{}", FileName: "example.MTS", Skip: true},
			{Code: "{}", FileName: "example.cts", Errors: emptyFileError(1, 1, 1, 3)},
			// Skipped: rslint does not support the non-JS/TS extension.
			{Code: "", FileName: "example.md", Skip: true},
			// Skipped: rslint does not support the non-JS/TS extension.
			{Code: "", FileName: "example.vue", Skip: true},
			// Skipped: rslint does not support the non-JS/TS extension.
			{Code: "", FileName: "example.svelte", Skip: true},
			// Skipped: rslint does not support the non-JS/TS extension.
			{Code: "", FileName: "example.astro", Skip: true},
			// Skipped: rslint does not support the non-JS/TS extension.
			{Code: "", FileName: "example.css", Skip: true},
			// Skipped: rslint does not support the non-JS/TS extension.
			{Code: "", FileName: "example.txt", Skip: true},
			// Skipped: rslint does not support the vue parser.
			{Code: "", FileName: "example.vue", Skip: true},
			// Skipped: rslint does not support the html parser.
			{Code: "", FileName: "example.html", Skip: true},
			// Skipped: rslint does not support the html parser.
			{Code: "   \n\t ", FileName: "example.html", Skip: true},
			// Skipped: rslint does not support the html parser.
			{Code: "<!-- comment -->", FileName: "example.html", Skip: true},
			// Skipped: rslint does not support the css language.
			{Code: "", FileName: "example.css", Skip: true},
			// Skipped: rslint does not support the css language.
			{Code: "   \n\t ", FileName: "example.css", Skip: true},
			// Skipped: rslint does not support the css language.
			{Code: "/* comment */", FileName: "example.css", Skip: true},
			// Skipped: rslint does not support the markdown language.
			{Code: "", FileName: "example.md", Skip: true},
			// Skipped: rslint does not support the markdown language.
			{Code: "   \n\t ", FileName: "example.md", Skip: true},
			// Skipped: rslint does not support the markdown language.
			{Code: "<!-- comment -->", FileName: "example.md", Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "", FileName: "example.yaml", Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "   \n\t ", FileName: "example.yaml", Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "# comment", FileName: "example.yaml", Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "%YAML 1.2\n# comment", FileName: "example.yaml", Options: map[string]any{"allowComments": true}, Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "&anchor", FileName: "example.yaml", Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "&anchor\n# comment", FileName: "example.yaml", Options: map[string]any{"allowComments": true}, Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "!tag\n# comment", FileName: "example.yaml", Options: map[string]any{"allowComments": true}, Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "---\n# comment", FileName: "example.yaml", Options: map[string]any{"allowComments": true}, Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "...\n# comment", FileName: "example.yaml", Options: map[string]any{"allowComments": true}, Skip: true},
			// Skipped: rslint does not support the yaml language.
			{Code: "---\n---", FileName: "example.yaml", Skip: true},
			// Skipped: rslint does not support the plain-text parser.
			{Code: "", FileName: ".gitignore", Skip: true},
			// Skipped: rslint does not support the plain-text parser.
			{Code: "   \n\t ", FileName: ".gitignore", Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "", FileName: "example.toml", Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "   \n\t ", FileName: "example.toml", Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "# comment", FileName: "example.toml", Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "", FileName: "example.toml", Options: map[string]any{"allowComments": true}, Skip: true},
			// Skipped: rslint does not support the toml language.
			{Code: "   \n\t ", FileName: "example.toml", Options: map[string]any{"allowComments": true}, Skip: true},
		})
}

// Regression for upstream issue #2175.
func TestNoEmptyFileUpstreamTypescript(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_empty_file.NoEmptyFileRule,
		[]rule_tester.ValidTestCase{
			{Code: "(() => {})();", FileName: "example.ts"},
		}, []rule_tester.InvalidTestCase{
			{Code: "\"\";", FileName: "example.ts", Errors: emptyFileError(1, 1, 1, 4)},
			{Code: "\"use strict\";", FileName: "example.ts", Errors: emptyFileError(1, 1, 1, 14)},
		})
}

// Additional documentation examples; the remaining examples are already in the snapshot group.
func TestNoEmptyFileUpstreamDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_empty_file.NoEmptyFileRule,
		[]rule_tester.ValidTestCase{
			{Code: ";;\nconst x = 0;", FileName: "example.js"},
			{Code: "{\n\tconst x = 0;\n}", FileName: "example.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "// Comment", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 11)},
			{Code: "/* Comment */", FileName: "example.js", Errors: emptyFileError(1, 1, 1, 14)},
			{Code: "{\n}", FileName: "example.js", Errors: emptyFileError(1, 1, 2, 2)},
		})
}

func TestNoEmptyFileUpstreamProcessor(t *testing.T) {
	// Upstream preprocesses "Physical file" in document.txt into an empty block.js.
	// rslint does not implement ESLint processors or physical/virtual filename pairs.
	t.Skip("unsupported upstream processor case: empty block.js extracted from document.txt")
}
