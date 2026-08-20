import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = 'The listener argument should be a function reference.';

const valid = (code: string) => ({ code, filename: 'file.js' });
const invalid = (code: string) => ({
  code,
  filename: 'file.js',
  errors: [{ message }],
});

ruleTester.run('no-invalid-remove-event-listener', null as never, {
  valid: [
    // CallExpression
    valid('new el.removeEventListener("click", () => {})'),
    valid('el.removeEventListener?.("click", () => {})'),
    valid('el.notRemoveEventListener("click", () => {})'),
    valid('el[removeEventListener]("click", () => {})'),
    valid('el["removeEventListener"]("click", () => {})'),

    // Arguments
    valid('el.removeEventListener("click")'),
    valid('el.removeEventListener()'),
    valid('el.removeEventListener(() => {})'),
    valid('el.removeEventListener(...["click", () => {}], () => {})'),
    valid('el.removeEventListener(...args, () => {})'),
    valid('el.removeEventListener(() => {}, "click")'),
    valid('window.removeEventListener("click", bind())'),
    valid('window.removeEventListener("click", handler.notBind())'),
    valid('window.removeEventListener("click", handler[bind]())'),
    valid('window.removeEventListener("click", handler.bind?.())'),
    valid('window.removeEventListener("click", handler?.bind())'),

    valid('window.removeEventListener(handler)'),
    valid(`class MyComponent {
  handler() {}
  disconnectedCallback() {
    this.removeEventListener('click', this.handler);
  }
}`),
    valid('this.removeEventListener("click", getListener())'),
    valid('el.removeEventListener("scroll", handler)'),
    valid('el.removeEventListener("keydown", obj.listener)'),
    valid('removeEventListener("keyup", () => {})'),
    valid('removeEventListener("keydown", function () {})'),
  ],
  invalid: [
    invalid('window.removeEventListener("scroll", handler.bind(abc))'),
    invalid('window.removeEventListener("scroll", this.handler.bind(abc))'),
    invalid('window.removeEventListener("click", () => {})'),
    invalid('window.removeEventListener("keydown", function () {})'),
    // Named function expression and async arrow are still inline functions.
    invalid('el.removeEventListener("click", function handleClick() {})'),
    invalid('el.removeEventListener("click", async () => {})'),
    invalid('el.removeEventListener("click", (e) => { e.preventDefault(); })'),
    invalid('el.removeEventListener("mouseover", fn.bind(abc))'),
    invalid('el?.removeEventListener("mouseover", fn.bind(abc))'),
    invalid('el.removeEventListener("mouseout", function (e) {})'),
    invalid('el?.removeEventListener("mouseout", function (e) {})'),
    invalid('el.removeEventListener("mouseout", function (e) {}, true)'),
    invalid(
      'el.removeEventListener("click", function (e) {}, ...moreArguments)',
    ),
    invalid('el.removeEventListener(() => {}, () => {}, () => {})'),
  ],
});
