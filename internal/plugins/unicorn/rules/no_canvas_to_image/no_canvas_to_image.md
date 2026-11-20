# no-canvas-to-image

## Rule details

Prefer drawing a canvas directly with `drawImage()` when transferring its image
to another canvas. This rule reports converting a canvas through
`loadImage(canvas.toDataURL())` before a `drawImage()` call, and passing
`getImageData()` results directly to `putImageData()`.

The first argument can be a direct expression or a `const` initializer whose
binding is used once after its declaration. Awaited expressions are supported.
For the data URL pattern, the receiver must be an identifier or a non-computed
property whose name contains `canvas`, ignoring case.

The rule uses syntax and local references to recognize these patterns. It allows
pixel data that is accessed or modified before drawing.

Examples of **incorrect** code for this rule

```javascript
const image = await loadImage(canvas.toDataURL());
context.drawImage(image, 0, 0);

const imageData = sourceContext.getImageData(0, 0, width, height);
targetContext.putImageData(imageData, 0, 0);
```

Examples of **correct** code for this rule

```javascript
context.drawImage(canvas, 0, 0);
targetContext.drawImage(sourceCanvas, 0, 0);

const imageData = context.getImageData(0, 0, width, height);
imageData.data[0] = 0;
context.putImageData(imageData, 0, 0);
```

## Options

This rule has no options.

## References

- [Unicorn documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/docs/rules/no-canvas-to-image.md)
- [Upstream implementation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v80.0.0/rules/no-canvas-to-image.js)
