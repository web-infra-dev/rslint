package no_canvas_to_image_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_canvas_to_image"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCanvasToImageExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_canvas_to_image.NoCanvasToImageRule,
		[]rule_tester.ValidTestCase{
			{Code: "const data = source?.context.getImageData(0, 0, 1, 1); context.putImageData(data, 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(loadImage(source?.canvas.toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "context.putImageData(source?.context.getImageData(0, 0, 1, 1), 0, 0);", FileName: "file.js"},
			{Code: "const image = loadImage(canvas.toDataURL()); const alias = image; context.drawImage(image, 0, 0);", FileName: "file.js"},
			{Code: "const image = loadImage(canvas.toDataURL()); image = other; context.drawImage(image, 0, 0);", FileName: "file.js"},
			{Code: "const {image} = loadImage(canvas.toDataURL()); context.drawImage(image, 0, 0);", FileName: "file.js"},
			{Code: "const image = loadImage(canvas.toDataURL()); function draw(image) { context.drawImage(image, 0, 0); }", FileName: "file.js"},
			{Code: "const image = loadImage(canvas.toDataURL()); context.drawImage(image, 0, 0); context.drawImage(image, 1, 1);", FileName: "file.js"},
			{Code: "context.drawImage(loadImage?.(canvas.toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(loader.loadImage(canvas.toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(loadImage(), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(loadImage(source[\"canvas\"].toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(loadImage((source?.canvas).toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "context[\"drawImage\"](loadImage(canvas.toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(); context.putImageData();", FileName: "file.js"},
			{Code: "context.putImageData((context.getImageData(0, 0, 1, 1) as ImageData), 0, 0);", FileName: "file.ts"},
			{Code: "const image = loadImage(canvas.toDataURL()); type Image = typeof image; context.drawImage(image, 0, 0);", FileName: "file.ts"},
			{Code: "const image = loadImage(canvas.toDataURL()); const alias = image; context.drawImage(alias, 0, 0);", FileName: "file.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "/* 😀 */ context.drawImage(loadImage(canvas.toDataURL()), 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 57}}},
			{Code: "context.drawImage((await (loadImage((canvas.toDataURL())))), 0, 0);", FileName: "file.mjs", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 59}}},
			{Code: "context.drawImage(loadImage(source.CANVAS.toDataURL(\"image/png\"), extra), 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 73}}},
			{Code: "context.putImageData(await await (other.getImageData(...args)), 0, 0);", FileName: "file.mjs", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 63}}},
			{Code: "const image = await await loadImage(canvas.toDataURL()); function paint() { context.drawImage(image, 0, 0); }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 95, EndLine: 1, EndColumn: 100}}},
			{Code: "const image = loadImage(canvas.toDataURL()); { const image = other; use(image); } context.drawImage(image, 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 101, EndLine: 1, EndColumn: 106}}},
			{Code: "const image = loadImage(canvas.toDataURL()); context.drawImage(image, 0, 0);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 64, EndLine: 1, EndColumn: 69}}},
			{Code: "// 😀 canvas\ncontext.drawImage(loadImage(canvas.toDataURL()), 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 2, Column: 19, EndLine: 2, EndColumn: 48}}},
			{Code: "const imageData = context.getImageData(0, 0, 1, 1); context.putImageData((imageData), 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 75, EndLine: 1, EndColumn: 84}}},
		})
}

// JSDoc type references are synthesized by tsgo; authored TS types and runtime
// expressions inside JSDoc casts still count as uses in Unicorn v80.0.0.
func TestNoCanvasToImageJSDocReferences(t *testing.T) {
	cases := []struct {
		code, filename string
		start, end     int
	}{
		{"const image = loadImage(canvas.toDataURL()); /** @type {typeof image} */ (other); context.drawImage(image, 0, 0);", "file.js", 101, 106},
		{"const image = loadImage(canvas.toDataURL()); /** @satisfies {typeof image} */ (other); context.drawImage(image, 0, 0);", "file.js", 106, 111},
		{"const image = loadImage(canvas.toDataURL()); /** @type {any} */ (image); context.drawImage(image, 0, 0);", "file.js", 0, 0},
		{"const image = loadImage(canvas.toDataURL()); other as typeof image; context.drawImage(image, 0, 0);", "file.ts", 0, 0},
		{"const image = loadImage(canvas.toDataURL()); other satisfies typeof image; context.drawImage(image, 0, 0);", "file.ts", 0, 0},
		{"const imageData = context.getImageData(0, 0, 1, 1); /** @type {typeof imageData} */ (other); context.putImageData(imageData, 0, 0);", "file.js", 115, 124},
		{"const imageData = context.getImageData(0, 0, 1, 1); /** @satisfies {typeof imageData} */ (other); context.putImageData(imageData, 0, 0);", "file.js", 120, 129},
		{"const imageData = context.getImageData(0, 0, 1, 1); /** @type {any} */ (imageData); context.putImageData(imageData, 0, 0);", "file.js", 0, 0},
		{"const imageData = context.getImageData(0, 0, 1, 1); other as typeof imageData; context.putImageData(imageData, 0, 0);", "file.ts", 0, 0},
		{"const imageData = context.getImageData(0, 0, 1, 1); other satisfies typeof imageData; context.putImageData(imageData, 0, 0);", "file.ts", 0, 0},
	}
	for _, test := range cases {
		t.Run(test.code, func(t *testing.T) {
			raw, source, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(test.code, test.filename, "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			sourceOnly, err := lintprogram.NewFromBoundSources(raw, []*ast.SourceFile{source})
			if err != nil {
				t.Fatal(err)
			}
			for mode, program := range map[string]*lintprogram.Program{"source-only": sourceOnly, "project-backed": lintprogram.NewFromCompiler(raw)} {
				t.Run(mode, func(t *testing.T) {
					var diagnostics []rule.RuleDiagnostic
					linter.LintSingleFile(linter.LintSingleFileOptions{
						Program: program, File: source.FileName(),
						GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
							return []rule.ConfiguredRule{{
								Name: no_canvas_to_image.NoCanvasToImageRule.Name, Severity: rule.SeverityError,
								Run: func(ctx rule.RuleContext) rule.RuleListeners {
									return no_canvas_to_image.NoCanvasToImageRule.Run(ctx, nil)
								},
							}}
						},
						Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
					})
					want := 0
					if test.start != 0 {
						want = 1
					}
					if len(diagnostics) != want {
						t.Fatalf("got %d diagnostics, want %d", len(diagnostics), want)
					}
					if want == 0 {
						return
					}
					d := diagnostics[0]
					if d.Message.Id != "no-canvas-to-image" || d.Message.Description != "Prefer drawing the canvas directly with `drawImage()`." {
						t.Fatalf("unexpected message: %+v", d.Message)
					}
					line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(source, d.Range.Pos())
					endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(source, d.Range.End())
					if line != 0 || endLine != 0 || int(column)+1 != test.start || int(endColumn)+1 != test.end {
						t.Fatalf("unexpected diagnostic range: %+v", d.Range)
					}
					if d.FixesPtr != nil || d.Suggestions != nil {
						t.Fatal("unexpected edits")
					}
				})
			}
		})
	}
}
