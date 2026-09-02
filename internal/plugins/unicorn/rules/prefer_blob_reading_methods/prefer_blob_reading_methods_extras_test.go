// TestPreferBlobReadingMethodsExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case identifies the upstream
// branch, Dimension 4 row, or real-user scenario it covers.
package prefer_blob_reading_methods_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_blob_reading_methods"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func extrasInvalid(code, fileName string, methods ...string) rule_tester.InvalidTestCase {
	seen := map[string]int{}
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(methods))
	for _, method := range methods {
		errors = append(errors, methodError(code, method, seen[method]))
		seen[method]++
	}
	return rule_tester.InvalidTestCase{Code: code, FileName: fileName, Errors: errors}
}

func TestPreferBlobReadingMethodsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_blob_reading_methods.PreferBlobReadingMethodsRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream isMethodCall() argument-count branches.
			{Code: `fileReader.readAsText()`, FileName: "file.mjs"},
			{Code: `fileReader.readAsArrayBuffer(blob, extra)`, FileName: "file.mjs"},

			// Locks in upstream isMethodCall() default allowSpreadElement: false.
			{Code: `fileReader.readAsText(...blobs)`, FileName: "file.mjs"},

			// ---- Dimension 4: element-access key forms are excluded ----
			{Code: `fileReader['readAsText'](blob)`, FileName: "file.mjs"},
			{Code: "fileReader[`readAsText`](blob)", FileName: "file.mjs"},
			{Code: `fileReader[0](blob)`, FileName: "file.mjs"},
			{Code: `fileReader[Symbol.readAsText](blob)`, FileName: "file.mjs"},

			// ---- Dimension 4: optional calls and members are excluded ----
			{Code: `(fileReader.readAsText)?.(blob)`, FileName: "file.mjs"},
			{Code: `(fileReader?.readAsText)(blob)`, FileName: "file.mjs"},
			{Code: `/** @type {any} */ (fileReader?.readAsText)(blob)`, FileName: "file.js"},

			// ---- Dimension 4: authored TypeScript callee wrappers stay visible ----
			{Code: `(fileReader.readAsText as any)(blob)`, FileName: "file.ts"},
			{Code: `(fileReader.readAsArrayBuffer satisfies any)(blob)`, FileName: "file.ts"},

			// Locks in upstream listener/method-name rejection branches.
			{Code: `new fileReader.readAsText(blob)`, FileName: "file.mjs"},
			{Code: `fileReader.readAsText`, FileName: "file.mjs"},
			{Code: `fileReader.readAsDataURL(blob)`, FileName: "file.mjs"},

			// ---- Dimension 4: private member names are excluded ----
			{Code: `class Reader { #readAsText(blob) {} read(blob) { this.#readAsText(blob); } }`, FileName: "file.mjs"},

			// ---- Real-user: file preview flows may still use readAsDataURL ----
			{Code: `const reader = new FileReader(); reader.readAsDataURL(file);`, FileName: "file.mjs"},

			// N/A: declaration/container forms; the rule only targets call expressions.
			// N/A: declaration key forms; the rule inspects only a call's member name.
			// N/A: ancestor walks and body-absent declarations; the rule performs no ancestor traversal.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver and callee wrappers ----
			extrasInvalid(`(fileReader).readAsArrayBuffer(blob)`, "file.mjs", "readAsArrayBuffer"),
			extrasInvalid(`((fileReader)).readAsText(blob)`, "file.mjs", "readAsText"),
			extrasInvalid(`(fileReader.readAsText)(blob)`, "file.mjs", "readAsText"),

			// ---- Dimension 4: JavaScript JSDoc callee wrappers are transparent ----
			extrasInvalid(`/** @type {any} */ (fileReader.readAsText)(blob)`, "file.js", "readAsText"),
			extrasInvalid(
				`/** @satisfies {any} */ (fileReader.readAsArrayBuffer)(blob)`,
				"file.js",
				"readAsArrayBuffer",
			),

			// ---- Dimension 4: TypeScript receiver wrappers ----
			extrasInvalid(`fileReader!.readAsText(blob)`, "file.ts", "readAsText"),
			extrasInvalid(`(fileReader as FileReader).readAsArrayBuffer(blob)`, "file.ts", "readAsArrayBuffer"),
			extrasInvalid(`(fileReader satisfies FileReader).readAsText(blob)`, "file.ts", "readAsText"),

			// Locks in upstream isMethodCall() handling of type arguments.
			extrasInvalid(`fileReader.readAsText<string>(blob)`, "file.ts", "readAsText"),

			// ---- Dimension 4: same-kind nesting reports both calls ----
			extrasInvalid(
				`fileReader.readAsText(other.readAsArrayBuffer(blob))`,
				"file.mjs",
				"readAsText",
				"readAsArrayBuffer",
			),

			// Locks in comments and multiline whitespace around the call boundary.
			extrasInvalid("fileReader\n\t.readAsText /* keep */ (\n\t\tblob\n\t)", "file.mjs", "readAsText"),

			// ---- Real-user: issue #1269 Promise-wrapped FileReader flow ----
			extrasInvalid(
				"const arrayBuffer = await new Promise((resolve, reject) => {\n\tconst reader = new FileReader();\n\treader.onload = () => resolve(reader.result);\n\treader.onerror = () => reject(reader.error);\n\treader.readAsArrayBuffer(blob);\n});",
				"file.mjs",
				"readAsArrayBuffer",
			),

			// ---- Real-user: FileReader calls commonly appear in input handlers ----
			extrasInvalid(
				"input.addEventListener('change', () => {\n\tconst reader = new FileReader();\n\treader.readAsText(input.files[0]);\n});",
				"file.mjs",
				"readAsText",
			),
		},
	)
}
