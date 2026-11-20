package no_canvas_to_image_test

import (
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
