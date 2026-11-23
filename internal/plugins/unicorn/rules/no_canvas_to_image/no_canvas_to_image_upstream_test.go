package no_canvas_to_image_test

// Ported from eslint-plugin-unicorn v80.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/test/no-canvas-to-image.js

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_canvas_to_image"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCanvasToImageUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_canvas_to_image.NoCanvasToImageRule,
		[]rule_tester.ValidTestCase{
			{Code: "context.drawImage(canvas, 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(image, 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(await image, 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(await loadImage(url), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(await loadImage(...canvas.toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(await loadImage(canvas.toBlob()), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(await loadImage(canvas[\"toDataURL\"]()), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(await loadImage(canvas.toDataURL?.()), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(await loadImage(canvas?.toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage?.(await loadImage(canvas.toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "context?.drawImage(await loadImage(canvas.toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(await loadImage(chart.toDataURL()), 0, 0);", FileName: "file.js"},
			{Code: "const image = await loadImage(chart.toDataURL()); context.drawImage(image, 0, 0);", FileName: "file.js"},
			{Code: "context.putImageData(imageData, 0, 0);", FileName: "file.js"},
			{Code: "context.putImageData(await imageData, 0, 0);", FileName: "file.js"},
			{Code: "context.putImageData(...context.getImageData(0, 0, width, height));", FileName: "file.js"},
			{Code: "context.putImageData(context[\"getImageData\"](0, 0, width, height), 0, 0);", FileName: "file.js"},
			{Code: "context.putImageData(context.getImageData?.(0, 0, width, height), 0, 0);", FileName: "file.js"},
			{Code: "context.putImageData(context?.getImageData(0, 0, width, height), 0, 0);", FileName: "file.js"},
			{Code: "context.putImageData?.(context.getImageData(0, 0, width, height), 0, 0);", FileName: "file.js"},
			{Code: "context?.putImageData(context.getImageData(0, 0, width, height), 0, 0);", FileName: "file.js"},
			{Code: "const url = canvas.toDataURL();", FileName: "file.js"},
			{Code: "const image = await loadImage(url); context.drawImage(image, 0, 0);", FileName: "file.js"},
			{Code: "const image = await loadImage(canvas.toBlob()); context.drawImage(image, 0, 0);", FileName: "file.js"},
			{Code: "let image = await loadImage(canvas.toDataURL());\ncontext.drawImage(image, 0, 0);", FileName: "file.js"},
			{Code: "const image = await loadImage(canvas.toDataURL());\nconsole.log(image);\ncontext.drawImage(image, 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(image, 0, 0);\nconst image = await loadImage(canvas.toDataURL());", FileName: "file.js"},
			{Code: "const imageData = context.getImageData(0, 0, width, height);\nimageData.data[0] = 0;\ncontext.putImageData(imageData, 0, 0);", FileName: "file.js"},
			{Code: "let imageData = context.getImageData(0, 0, width, height);\notherContext.putImageData(imageData, 0, 0);", FileName: "file.js"},
			{Code: "const imageData = context.getImageData(0, 0, width, height);\nconsole.log(imageData);\notherContext.putImageData(imageData, 0, 0);", FileName: "file.js"},
			{Code: "otherContext.putImageData(imageData, 0, 0);\nconst imageData = context.getImageData(0, 0, width, height);", FileName: "file.js"},
			{Code: "context.drawImage(...images, 0, 0);", FileName: "file.js"},
			{Code: "context.drawImage(canvas, 0, 0);", FileName: "file.js"},
			{Code: "targetContext.drawImage(sourceCanvas, 0, 0);", FileName: "file.js"},
			{Code: "const imageData = context.getImageData(0, 0, width, height);\nimageData.data[0] = 0;\ncontext.putImageData(imageData, 0, 0);", FileName: "file.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "context.drawImage(loadImage(canvas.toDataURL()), 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 48}}},
			{Code: "context.drawImage(await loadImage(canvas.toDataURL()), 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 54}}},
			{Code: "context.drawImage(await loadImage(sourceCanvas.toDataURL()), 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 60}}},
			{Code: "context.drawImage(await loadImage(this.canvas.toDataURL()), 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 59}}},
			{Code: "const image = await loadImage(canvas.toDataURL());\ncontext.drawImage(image, 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 2, Column: 19, EndLine: 2, EndColumn: 24}}},
			{Code: "const image = loadImage(canvas.toDataURL('image/jpeg'));\ncontext.drawImage(await image, 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 2, Column: 19, EndLine: 2, EndColumn: 30}}},
			{Code: "context.putImageData(context.getImageData(0, 0, width, height), 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 63}}},
			{Code: "context2.putImageData(context1.getImageData(0, 0, width, height), 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 65}}},
			{Code: "const imageData = context1.getImageData(0, 0, width, height);\ncontext2.putImageData(imageData, 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 2, Column: 23, EndLine: 2, EndColumn: 32}}},
			{Code: "const image = await loadImage(canvas.toDataURL());\ncontext.drawImage(image, 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 2, Column: 19, EndLine: 2, EndColumn: 24}}},
			{Code: "const imageData = sourceContext.getImageData(0, 0, width, height);\ntargetContext.putImageData(imageData, 0, 0);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-canvas-to-image", Message: "Prefer drawing the canvas directly with `drawImage()`.", Line: 2, Column: 28, EndLine: 2, EndColumn: 37}}},
		})
}
