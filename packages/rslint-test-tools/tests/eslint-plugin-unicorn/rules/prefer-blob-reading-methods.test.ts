import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string) => ({ code, filename: 'file.mjs' });
const invalid = (code: string, method: string, replacement: string) => ({
  code,
  filename: 'file.mjs',
  errors: [
    {
      messageId: 'error',
      message: `Prefer \`Blob#${replacement}()\` over \`FileReader#${method}(blob)\`.`,
      data: { method, replacement },
    },
  ],
});

ruleTester.run('prefer-blob-reading-methods', null as never, {
  valid: [
    valid('blob.arrayBuffer()'),
    valid('blob.text()'),
    valid('new Response(blob).arrayBuffer()'),
    valid('new Response(blob).text()'),
    valid('fileReader.readAsDataURL(blob)'),
    valid('fileReader.readAsBinaryString(blob)'),
    valid('fileReader.readAsText(blob, "ascii")'),
    valid('fileReader.readAsArrayBuffer(blob, extraArg)'),
    valid('fileReader?.readAsArrayBuffer(blob)'),
    valid('fileReader.readAsText?.(blob)'),
  ],
  invalid: [
    invalid(
      'fileReader.readAsArrayBuffer(blob)',
      'readAsArrayBuffer',
      'arrayBuffer',
    ),
    invalid('fileReader.readAsText(blob)', 'readAsText', 'text'),
  ],
});
