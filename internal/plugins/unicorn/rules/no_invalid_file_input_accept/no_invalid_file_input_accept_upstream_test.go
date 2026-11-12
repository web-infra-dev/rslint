package no_invalid_file_input_accept_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_invalid_file_input_accept"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	invalidAcceptPrefix   = "Invalid file input `accept` value: "
	staticAcceptMessage   = "The file input `accept` value must be static."
	missingAcceptMessage  = invalidAcceptPrefix + "The `accept` attribute must have a value."
	emptyAcceptMessage    = invalidAcceptPrefix + "Empty entries are not allowed."
	wildcardAcceptMessage = invalidAcceptPrefix + "Wildcard MIME types must be `audio/*`, `image/*`, or `video/*`."
)

func normalizeMessage(value string) string {
	return "Prefer `" + value + "` as the file input `accept` value."
}

func acceptCase(code string, column, endColumn int, id, message string, output ...string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code, Tsx: true, Output: output,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "no-invalid-file-input-accept/" + id, Message: message,
			Line: 1, Column: column, EndLine: 1, EndColumn: endColumn,
		}},
	}
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-invalid-file-input-accept.js
// The documentation's incorrect JSX examples are in the snapshot group below.
func TestNoInvalidFileInputAcceptUpstreamJSX(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&no_invalid_file_input_accept.NoInvalidFileInputAcceptRule,
		[]rule_tester.ValidTestCase{
			// jsx.snapshot valid cases.
			{Code: `<input type="file" accept="image/png" />`, Tsx: true},
			{Code: `<input type="file" accept=".png" />`, Tsx: true},
			{Code: `<input type="file" accept="image/png, .png" />`, Tsx: true},
			{Code: `<input type="file" accept=".tar.gz" />`, Tsx: true},
			{Code: `<input type="file" accept="audio/*" />`, Tsx: true},
			{Code: `<input type="file" accept="video/*" />`, Tsx: true},
			{Code: `<input type="file" accept="image/*" />`, Tsx: true},
			{Code: `<input type="file" accept="audio/*, video/*" />`, Tsx: true},
			{Code: `<input type="FILE" accept="image/png" />`, Tsx: true},
			{Code: `<input type={"file"} accept="image/png" />`, Tsx: true},
			{Code: `const type = "file"; <input type={type} accept="image/png" />`, Tsx: true},
			{Code: `const accept = "image/png"; <input type="file" accept={accept} />`, Tsx: true},
			{Code: `<input accept="image/png" />`, Tsx: true},
			{Code: `<input type="text" accept="image/png" />`, Tsx: true},
			{Code: `<input type={type} accept="image/png" />`, Tsx: true},
			{Code: `<Input type="file" accept="image/jpg" />`, Tsx: true},
			{Code: `<input type="file" Accept="image/jpg" />`, Tsx: true},
			{Code: `<input Type="file" accept="image/jpg" />`, Tsx: true},
			{Code: `<Foo.input type="file" accept="image/jpg" />`, Tsx: true},
			{Code: `<input type="file" xlink:accept="image/jpg" />`, Tsx: true},
			// Documentation example. Its other correct JSX examples are above.
			{Code: `<input type="file" accept="image/jpeg" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// jsx.snapshot invalid cases.
			acceptCase(`const modes = new Set(["foo"]); modes.clear(); <input type="file" accept={modes.size ? "bad" : "image/jpeg"} />`,
				74, 109, "static", staticAcceptMessage),
			acceptCase(`const modes = new Set(["foo"]); modes.clear(); <input type="file" accept={(modes.size && "image/jpeg") || value} />`,
				74, 113, "static", staticAcceptMessage),
			acceptCase(`const object = {value: true}; Object.defineProperty(object, "value", {get() { return false; }}); <input type="file" accept={object.value ? "image/jpeg" : value} />`,
				124, 161, "static", staticAcceptMessage),
			acceptCase(`<input type="file" accept />`,
				20, 26, "invalid", missingAcceptMessage),
			acceptCase(`<input type="file" accept={allowedTypes} />`,
				27, 41, "static", staticAcceptMessage),
			acceptCase("<input type=\"file\" accept={`image/${type}`} />",
				27, 44, "static", staticAcceptMessage),
			acceptCase(`<input type="file" accept="image/jpg" />`,
				27, 38, "normalize", normalizeMessage("image/jpeg"), `<input type="file" accept="image/jpeg" />`),
			acceptCase(`<input type="file" accept="image/svg" />`,
				27, 38, "normalize", normalizeMessage("image/svg+xml"), `<input type="file" accept="image/svg+xml" />`),
			acceptCase(`<input type="file" accept="png" />`,
				27, 32, "normalize", normalizeMessage(".png"), `<input type="file" accept=".png" />`),
			acceptCase(`<input type="file" accept=".PNG" />`,
				27, 33, "normalize", normalizeMessage(".png"), `<input type="file" accept=".png" />`),
			acceptCase(`<input type="file" accept=".png, .PNG" />`,
				27, 39, "normalize", normalizeMessage(".png"), `<input type="file" accept=".png" />`),
			acceptCase(`<input type="file" accept="" />`,
				27, 29, "invalid", emptyAcceptMessage),
			acceptCase(`<input type="file" accept="IMAGE/*" />`,
				27, 36, "normalize", normalizeMessage("image/*"), `<input type="file" accept="image/*" />`),
			acceptCase(`<input type="file" accept="TEXT/*" />`,
				27, 35, "invalid", wildcardAcceptMessage),
			acceptCase(`<input type="file" accept="*/plain" />`,
				27, 36, "invalid", wildcardAcceptMessage),
			acceptCase(`<input type="file" accept="not-real-mime-type/*" />`,
				27, 49, "invalid", wildcardAcceptMessage),
			acceptCase(`<input type="file" accept="image/png," />`,
				27, 39, "invalid", emptyAcceptMessage),
			acceptCase(`<input type="file" accept="image/png,, .png" />`,
				27, 45, "invalid", emptyAcceptMessage),
			acceptCase(`<input type="file" accept="IMAGE/PNG,.PNG, image/png" />`,
				27, 54, "normalize", normalizeMessage("image/png, .png"), `<input type="file" accept="image/png, .png" />`),
			acceptCase(`<input type="file" accept="image/png,.png" />`,
				27, 43, "normalize", normalizeMessage("image/png, .png"), `<input type="file" accept="image/png, .png" />`),
			acceptCase(`<input type="file" accept="image/png; charset=utf-8" />`,
				27, 53, "normalize", normalizeMessage("image/png"), `<input type="file" accept="image/png" />`),
			acceptCase(`<input type="file" accept={"IMAGE/PNG"} />`,
				27, 40, "normalize", normalizeMessage("image/png")),
			acceptCase(`<input type="file" accept=<div /> />`,
				27, 34, "static", staticAcceptMessage),

			// jsx cases with explicit autofix expectations.
			acceptCase(`<input type="file" accept="image/x-icon" />`,
				27, 41, "normalize", normalizeMessage("image/vnd.microsoft.icon"), `<input type="file" accept="image/vnd.microsoft.icon" />`),
			acceptCase(`<input type="file" accept="application/x-rar-compressed" />`,
				27, 57, "normalize", normalizeMessage("application/vnd.rar"), `<input type="file" accept="application/vnd.rar" />`),
			acceptCase(`<input type="file" accept="application/x-zip-compressed" />`,
				27, 57, "normalize", normalizeMessage("application/zip"), `<input type="file" accept="application/zip" />`),
			acceptCase(`<input type="file" accept="image/x-icon, image/vnd.microsoft.icon, application/x-zip-compressed" />`,
				27, 97, "normalize", normalizeMessage("image/vnd.microsoft.icon, application/zip"), `<input type="file" accept="image/vnd.microsoft.icon, application/zip" />`),
			acceptCase(`<input type="file" accept="image/jpg; charset=utf-8" />`,
				27, 53, "normalize", normalizeMessage("image/jpeg"), `<input type="file" accept="image/jpeg" />`),
			acceptCase(`<input type="file" accept="image/svg; charset=utf-8" />`,
				27, 53, "normalize", normalizeMessage("image/svg+xml"), `<input type="file" accept="image/svg+xml" />`),
		},
	)
}

// These fixtures require html-eslint's HTML AST, including its template parser
// mode and ESLint language plugin. The native rule runner receives JS/TS ASTs.
func TestNoInvalidFileInputAcceptUpstreamHTML(t *testing.T) {
	for _, group := range []struct {
		name  string
		cases []string
	}{
		{"html.snapshot", []string{
			// Valid.
			`<input type="file" accept="image/png">`,
			`<input type="file" accept=".png">`,
			`<input type="file" accept="image/png, .png">`,
			`<input type="file" accept=".tar.gz">`,
			`<input type="file" accept="audio/*">`,
			`<input type="file" accept="video/*">`,
			`<input type="file" accept="image/*">`,
			`<input type="FILE" accept="image/png">`,
			`<input accept="image/png">`,
			`<input type="text" accept="image/png">`,
			`<input type="file" accept="image/jpg"`,
			`<input type="file"`,
			`<textarea>Pick an image: <input type="file" accept="image/jpg"></textarea>`,
			`<title><input type="file" accept="image/jpg"></title>`,
			`<input type=file accept=image/png>`,
			`<input type=file accept=image/png >`,
			`<input type=file accept=image/png />`,
			`<input type=file accept=image/png disabled>`,
			`<input type=file accept=image/*>`,
			`<input accept=image/png type=file>`,
			`<input accept=image/png disabled type=file>`,
			`<input accept=image/jpg title="file type=file disabled">`,
			`<input type=file data-url=https://example.com accept=image/png>`,
			`<input data-url=/upload type=file accept=image/png>`,
			`<input type=file accept="image&#x2F;png">`,
			`<input type=file accept=image&#x2F;png>`,
			`<input type=file accept="image&#47;jpg">`,
			`<input type=file accept="image/png&#44;.png">`,
			`<input type="file">`,
			`<input type="file" x accept>`,
			// Invalid.
			`<input type="file" accept>`,
			`<input type="file" accept="${allowedTypes}">`,
			`<input type="file" accept="image/jpg">`,
			`<input type="file" accept="image/svg">`,
			`<input type="file" accept="png">`,
			`<input type="file" accept="IMAGE/*">`,
			`<input type="file" accept="TEXT/*">`,
			`<input type="file" accept="*/plain">`,
			`<input type="file" accept="not-real-mime-type/*">`,
			`<input type="file" accept="image/png,">`,
			`<input type="file" accept="IMAGE/PNG,.PNG, image/png">`,
			`<input type="file" accept="image/png,.png">`,
			`<input type="file" accept="image/png; charset=utf-8">`,
			`<input / type="file" accept="image/jpg">`,
			`<input type="file" accept/>`,
			`<input type=file a/b accept=image/jpg>`,
			`<input type="file" / accept>`,
		}},
		{"htmlTemplates.snapshot", []string{
			// Valid.
			`<input type="file" accept="video/mp4">`,
			// Invalid.
			`<input type="file" accept="{{allowedTypes}}">`,
		}},
		{"htmlPlugin", []string{
			// Valid.
			`<input type="file" accept="application/pdf">`,
			// Invalid.
			`<input type="file" accept="IMAGE/JPEG">`,
		}},
		{"html", []string{
			// Invalid.
			`<input type="file" accept="image/x-icon">`,
			`<input type="file" accept="application/x-rar-compressed">`,
			`<input type="file" accept="application/x-zip-compressed">`,
			`<input type="file" accept="image/x-icon, image/vnd.microsoft.icon, application/x-zip-compressed">`,
			`<input type="file" accept="image/jpg; charset=utf-8">`,
			`<input type="file" accept="image/svg; charset=utf-8">`,
			`<input type = "file" accept="png">`,
			`<input type=file accept=png>`,
			`<input type=file accept=png />`,
			`<input type=file accept=png disabled>`,
			`<input disabled accept=png type=file>`,
			`<input type=file accept=.PNG />`,
			`<input type=file accept=png/>`,
			`<input type=file accept=.png/>`,
			`<input type=file accept=image/png/>`,
			`<input type=file accept=image/*/>`,
			`<input type=file accept=png,jpeg>`,
			`<input type=file accept = "png,jpeg">`,
			`<input type=file accept=IMAGE/PNG>`,
			`<input type=file accept=image/jpg>`,
			`<input type=file accept=IMAGE/*>`,
			`<input type=file accept=image/png,>`,
			`<input accept=image/jpg type=file>`,
			`<input accept=IMAGE/PNG type=file>`,
			`<input type=file data-url=https://example.com accept=png>`,
			`<input data-url=/upload type=file accept=image/jpg>`,
		}},
		{"documentation", []string{
			// The incorrect example is also in html.snapshot.
			`<input type="file" accept="image/jpeg">`,
		}},
	} {
		t.Run(group.name, func(t *testing.T) {
			for _, code := range group.cases {
				t.Run(code, func(t *testing.T) {
					t.Skip("requires html-eslint HTML nodes")
				})
			}
		})
	}
}
