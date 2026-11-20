// Ported from eslint-plugin-unicorn v80.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/test/no-canvas-to-image.js
import { RuleTester } from '@rslint/rule-tester';

const tester = new RuleTester({ plugin: 'unicorn' });
tester.run('no-canvas-to-image', {
  valid: [
    'context.drawImage(canvas, 0, 0);',
    'context.drawImage(image, 0, 0);',
    'context.drawImage(await image, 0, 0);',
    'context.drawImage(await loadImage(url), 0, 0);',
    'context.drawImage(await loadImage(...canvas.toDataURL()), 0, 0);',
    'context.drawImage(await loadImage(canvas.toBlob()), 0, 0);',
    'context.drawImage(await loadImage(canvas["toDataURL"]()), 0, 0);',
    'context.drawImage(await loadImage(canvas.toDataURL?.()), 0, 0);',
    'context.drawImage(await loadImage(canvas?.toDataURL()), 0, 0);',
    'context.drawImage?.(await loadImage(canvas.toDataURL()), 0, 0);',
    'context?.drawImage(await loadImage(canvas.toDataURL()), 0, 0);',
    'context.drawImage(await loadImage(chart.toDataURL()), 0, 0);',
    'const image = await loadImage(chart.toDataURL()); context.drawImage(image, 0, 0);',
    'context.putImageData(imageData, 0, 0);',
    'context.putImageData(await imageData, 0, 0);',
    'context.putImageData(...context.getImageData(0, 0, width, height));',
    'context.putImageData(context["getImageData"](0, 0, width, height), 0, 0);',
    'context.putImageData(context.getImageData?.(0, 0, width, height), 0, 0);',
    'context.putImageData(context?.getImageData(0, 0, width, height), 0, 0);',
    'context.putImageData?.(context.getImageData(0, 0, width, height), 0, 0);',
    'context?.putImageData(context.getImageData(0, 0, width, height), 0, 0);',
    'const url = canvas.toDataURL();',
    'const image = await loadImage(url); context.drawImage(image, 0, 0);',
    'const image = await loadImage(canvas.toBlob()); context.drawImage(image, 0, 0);',
    'let image = await loadImage(canvas.toDataURL());\ncontext.drawImage(image, 0, 0);',
    'const image = await loadImage(canvas.toDataURL());\nconsole.log(image);\ncontext.drawImage(image, 0, 0);',
    'context.drawImage(image, 0, 0);\nconst image = await loadImage(canvas.toDataURL());',
    'const imageData = context.getImageData(0, 0, width, height);\nimageData.data[0] = 0;\ncontext.putImageData(imageData, 0, 0);',
    'let imageData = context.getImageData(0, 0, width, height);\notherContext.putImageData(imageData, 0, 0);',
    'const imageData = context.getImageData(0, 0, width, height);\nconsole.log(imageData);\notherContext.putImageData(imageData, 0, 0);',
    'otherContext.putImageData(imageData, 0, 0);\nconst imageData = context.getImageData(0, 0, width, height);',
    'context.drawImage(...images, 0, 0);',
    'context.drawImage(canvas, 0, 0);',
    'targetContext.drawImage(sourceCanvas, 0, 0);',
    'const imageData = context.getImageData(0, 0, width, height);\nimageData.data[0] = 0;\ncontext.putImageData(imageData, 0, 0);',
  ],
  invalid: [
    {
      code: 'context.drawImage(loadImage(canvas.toDataURL()), 0, 0);',
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: 'context.drawImage(await loadImage(canvas.toDataURL()), 0, 0);',
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'context.drawImage(await loadImage(sourceCanvas.toDataURL()), 0, 0);',
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: 'context.drawImage(await loadImage(this.canvas.toDataURL()), 0, 0);',
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
    {
      code: 'const image = await loadImage(canvas.toDataURL());\ncontext.drawImage(image, 0, 0);',
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      code: "const image = loadImage(canvas.toDataURL('image/jpeg'));\ncontext.drawImage(await image, 0, 0);",
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'context.putImageData(context.getImageData(0, 0, width, height), 0, 0);',
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
    {
      code: 'context2.putImageData(context1.getImageData(0, 0, width, height), 0, 0);',
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 65,
        },
      ],
    },
    {
      code: 'const imageData = context1.getImageData(0, 0, width, height);\ncontext2.putImageData(imageData, 0, 0);',
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'const image = await loadImage(canvas.toDataURL());\ncontext.drawImage(image, 0, 0);',
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'const imageData = sourceContext.getImageData(0, 0, width, height);\ntargetContext.putImageData(imageData, 0, 0);',
      errors: [
        {
          messageId: 'no-canvas-to-image',
          message: 'Prefer drawing the canvas directly with `drawImage()`.',
          line: 2,
          column: 28,
          endLine: 2,
          endColumn: 37,
        },
      ],
    },
  ],
});
