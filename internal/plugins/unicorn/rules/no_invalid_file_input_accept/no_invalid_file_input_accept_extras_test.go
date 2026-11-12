package no_invalid_file_input_accept_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_invalid_file_input_accept"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func TestNoInvalidFileInputAcceptExtras(t *testing.T) {
	const extensionMessage = invalidAcceptPrefix +
		"File extensions must start with `.` and cannot contain whitespace, commas, slashes, wildcards, or template markers."
	const mimeMessage = invalidAcceptPrefix + "MIME types must use valid `type/subtype` syntax."
	multiline := acceptCase("<input data-note=\"😀\" type=\"file\" accept=\"\uFEFF .İΣ,\n  image/png\u00A0\" />",
		42, 14, "normalize", normalizeMessage(".i\u0307ς, image/png"),
		"<input data-note=\"😀\" type=\"file\" accept=\".i\u0307ς, image/png\" />")
	multiline.Errors[0].EndLine = 2

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&no_invalid_file_input_accept.NoInvalidFileInputAcceptRule,
		[]rule_tester.ValidTestCase{
			{Code: `<input type="file" />`, Tsx: true},
			{Code: `<input type accept="bad" />`, Tsx: true},
			{Code: `<input type={42} accept="bad" />`, Tsx: true},
			{Code: `<input type=" file " accept="bad" />`, Tsx: true},
			// The reference represents a supplementary character in the rendered type.
			{Code: `<input type="f&#x10069;le" accept="bad" />`, Tsx: true},
			{Code: `<svg:input type="file" accept="bad" />`, Tsx: true},
			{Code: `<input {...{type: "file"}} accept="bad" />`, Tsx: true},
			// Upstream uses the first explicit attribute with a matching name.
			{Code: `<input type="text" type="file" accept="bad" />`, Tsx: true},
			{Code: `<input type="file" accept=".png" accept="bad" />`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			acceptCase(`<input {...props} accept='IMAGE/JPG' type={"FILE"}></input>`,
				26, 37, "normalize", normalizeMessage("image/jpeg"), `<input {...props} accept='image/jpeg' type={"FILE"}></input>`),
			acceptCase(`<input type="file" accept={42} />`, 27, 31, "static", staticAcceptMessage),
			acceptCase(`<input type="file" accept={""} />`, 27, 31, "invalid", emptyAcceptMessage),
			// JSX attribute text keeps a backslash literal. The JS expression cooks its escape.
			acceptCase(`<input type="file" accept = 'P\u004EG' />`,
				29, 39, "normalize", normalizeMessage(`.p\u004eg`), `<input type="file" accept = '.p\u004eg' />`),
			acceptCase(`<input type="file" accept={"P\u004EG"} />`,
				27, 39, "normalize", normalizeMessage(".png")),
			acceptCase("const kind = `FI${\"LE\"}`; <input type={(kind)} accept={(\"IMAGE/\" + \"PNG\")} />",
				55, 75, "normalize", normalizeMessage("image/png")),
			acceptCase(`<input type={("file" as string)!} accept={("IMAGE/PNG" satisfies string)} />`,
				42, 74, "normalize", normalizeMessage("image/png")),
			acceptCase(`<input type="file" accept={Object.freeze("IMAGE/PNG")} />`,
				27, 55, "normalize", normalizeMessage("image/png")),
			acceptCase(`<input type="file" accept={(change(), ".png")} />`, 27, 47, "static", staticAcceptMessage),

			// An invalid token prevents applying partial normalization.
			acceptCase(`<input type="file" accept="image/jpg, .tar/gz" />`, 27, 47, "invalid", extensionMessage),
			acceptCase("<input type=\"file\" accept=\".two\u00A0words\" />", 27, 39, "invalid", extensionMessage),
			acceptCase(`<input type="file" accept="." />`, 27, 30, "invalid", extensionMessage),
			// MIME parsing uses HTTP whitespace and token characters.
			acceptCase("<input type=\"file\" accept=\"image/png\u00A0; charset=utf-8\" />", 27, 54, "invalid", mimeMessage),
			acceptCase(`<input type="file" accept="image /png" />`, 27, 39, "invalid", mimeMessage),
			acceptCase(`<input type="file" accept="image/{png}" />`, 27, 40, "invalid", mimeMessage),
			acceptCase(`<input type="file" accept="IMAGE/PNG ; charset=utf-8; charset=ascii" />`,
				27, 69, "normalize", normalizeMessage("image/png"), `<input type="file" accept="image/png" />`),
			acceptCase(`<input type="file" accept="image/png;profile=*" />`, 27, 48, "invalid", wildcardAcceptMessage),
			multiline,

			acceptCase(`<input type="fi&#108;e" accept="&#46;PNG" />`,
				32, 42, "normalize", normalizeMessage(".png"), `<input type="fi&#108;e" accept=".png" />`),
			// Re-encode decoded delimiters and character-reference starts in the fix.
			acceptCase(`<input type="file" accept="&quot;PNG" />`,
				27, 38, "normalize", normalizeMessage(`."png`), `<input type="file" accept=".&quot;png" />`),
			acceptCase(`<input type='file' accept='&apos;PNG' />`,
				27, 38, "normalize", normalizeMessage(".'png"), `<input type='file' accept='.&apos;png' />`),
			acceptCase(`<input type="file" accept="&amp;AMP;PNG" />`,
				27, 41, "normalize", normalizeMessage(".&amp;png"), `<input type="file" accept=".&amp;amp;png" />`),
			acceptCase(`<input type="file" accept="&#0;PNG" />`,
				27, 36, "normalize", normalizeMessage(".\x00png"), `<input type="file" accept=".&#0;png" />`),
			// Report the rendered value. A source edit could change the value in JSX readers
			// that interpret these numeric references differently.
			acceptCase(`<input type="file" accept="&#x1F600;PNG" />`,
				27, 41, "normalize", normalizeMessage(".😀png")),
			acceptCase(`<input type="file" accept="&#00000000046;PNG" />`,
				27, 46, "normalize", normalizeMessage(".png")),
			acceptCase(`<input type="file" accept="&#xD800;PNG" />`,
				27, 40, "normalize", normalizeMessage("."+ecmascript.StringFromCodeUnits([]uint16{0xD800})+"png")),
		},
	)
}

func TestNoInvalidFileInputAcceptEditDemand(t *testing.T) {
	const source = `<><input type="file" accept="image/jpg" /><input type="file" accept={"IMAGE/PNG"} /></>`
	root := fixtures.GetRootDir()
	file := tspath.ResolvePath(root.Dir, "case.tsx")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{file: source})
	compiler, err := utils.CreateProgramLenient(true, fs, root.Dir, "tsconfig.json", utils.CreateCompilerHost(root.Dir, fs))
	if err != nil {
		t.Fatal(err)
	}
	lintRule := no_invalid_file_input_accept.NoInvalidFileInputAcceptRule
	options := linter.LintSingleFileOptions{
		Program: program.NewFromCompiler(compiler),
		File:    file,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name: lintRule.Name, Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners { return lintRule.Run(ctx, nil) },
			}}
		},
	}
	var all []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		options.Consumer = rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}}
		linter.LintSingleFile(options)
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: got %d diagnostics, want 2", demand, len(diagnostics))
		}
		if demand == rule.EditDemandAll {
			all = diagnostics
			if len(all[0].Fixes()) != 1 || all[1].FixesPtr != nil || all[0].Suggestions != nil || all[1].Suggestions != nil {
				t.Fatalf("expected one autofix for the quoted attribute: %#v", all)
			}
		}
		for i, got := range diagnostics {
			want := all[i]
			if demand&rule.EditDemandAutofix == 0 {
				want.FixesPtr = nil
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d, diagnostic %d: got %#v, want %#v", demand, i, got, want)
			}
		}
	}
}
