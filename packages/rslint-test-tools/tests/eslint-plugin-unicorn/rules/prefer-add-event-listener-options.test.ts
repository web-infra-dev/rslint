import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string) => ({ code, filename: 'file.js' });
const invalid = (code: string, value: 'true' | 'false') => ({
  code,
  filename: 'file.js',
  errors: [
    {
      message: `Prefer \`{capture: ${value}}\` over \`${value}\`.`,
    },
  ],
});

ruleTester.run('prefer-add-event-listener-options', null as never, {
  valid: [
    valid('window.addEventListener("click", listener)'),
    valid('window.addEventListener("click", listener, {capture: true})'),
    valid('window.addEventListener("click", listener, {capture: false})'),
    valid('window.addEventListener("click", listener, {passive: true})'),
    valid('window.addEventListener("click", listener, {once: true})'),
    valid('window.addEventListener("click", listener, {signal})'),
    valid('window.addEventListener("click", listener, options)'),
    valid('window.addEventListener("click", listener, capture)'),
    valid('window.addEventListener("click", listener, Boolean(value))'),
    valid(
      'window.addEventListener("click", listener, condition ? true : false)',
    ),
    valid('window.addEventListener("click", listener, undefined)'),
    valid('window.addEventListener("click", listener, null)'),
    valid('window["addEventListener"]("click", listener, true)'),
    valid('window?.addEventListener("click", listener, true)'),
    valid('window.addEventListener?.("click", listener, true)'),
    valid('window.addEventListener("click", ...arguments_, true)'),
  ],
  invalid: [
    invalid('window.addEventListener("click", listener, true)', 'true'),
    invalid('window.addEventListener("click", listener, false)', 'false'),
    invalid('window.addEventListener("click", listener, (true))', 'true'),
    invalid('window.addEventListener("click", () => {}, true)', 'true'),
    invalid('window.addEventListener("click", function () {}, false)', 'false'),
    invalid('document.body.addEventListener("click", listener, true)', 'true'),
    invalid('(window).addEventListener("click", listener, false)', 'false'),
    invalid(
      'window.addEventListener("click", listener, /* useCapture */ true)',
      'true',
    ),
    invalid(
      'window.addEventListener("click", listener, true /* useCapture */)',
      'true',
    ),
    invalid(
      'window.addEventListener(\n\t"click",\n\tlistener,\n\ttrue\n)',
      'true',
    ),
  ],
});
