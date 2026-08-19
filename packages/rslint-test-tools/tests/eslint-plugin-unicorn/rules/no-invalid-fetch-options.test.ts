import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const invalidBodyMessage = (method: string) =>
  `"body" is not allowed when method is "${method}".`;
const valid = (code: string) => ({ code, filename: 'file.js' });
const invalid = (code: string, method: string) => ({
  code,
  filename: 'file.js',
  errors: [
    {
      messageId: 'no-invalid-fetch-options',
      message: invalidBodyMessage(method),
    },
  ],
});

ruleTester.run('no-invalid-fetch-options', null as never, {
  valid: [
    valid('fetch(url, {method: "POST", body})'),
    valid('new Request(url, {method: "POST", body})'),
    valid('fetch(url, {})'),
    valid('new Request(url, {})'),
    valid('fetch(url)'),
    valid('new Request(url)'),
    valid('fetch(url, {method: "UNKNOWN", body})'),
    valid('new Request(url, {method: "UNKNOWN", body})'),
    valid('fetch(url, {body: undefined})'),
    valid('new Request(url, {body: undefined})'),
    valid('fetch(url, {body: null})'),
    valid('new Request(url, {body: null})'),
    valid('fetch(url, {body: void 0})'),
    valid('new Request(url, {method: "GET", body: void 0})'),
    valid('fetch(url, {...options, body})'),
    valid('new Request(url, {...options, body})'),
    valid('new fetch(url, {body})'),
    valid('Request(url, {body})'),
    valid('not_fetch(url, {body})'),
    valid('new not_Request(url, {body})'),
    valid('fetch({body}, url)'),
    valid('new Request({body}, url)'),
    valid('fetch(url, {[body]: "foo=bar"})'),
    valid('new Request(url, {[body]: "foo=bar"})'),
    valid(`fetch(url, {
	body: 'foo=bar',
	body: undefined,
});`),
    valid(`new Request(url, {
	body: 'foo=bar',
	body: undefined,
});`),
    valid(`fetch(url, {
	method: 'HEAD',
	body: 'foo=bar',
	method: 'post',
});`),
    valid(`new Request(url, {
	method: 'HEAD',
	body: 'foo=bar',
	method: 'post',
});`),
  ],
  invalid: [
    invalid('fetch(url, {body})', 'GET'),
    invalid('new Request(url, {body})', 'GET'),
    invalid('fetch(url, {method: "GET", body})', 'GET'),
    invalid('new Request(url, {method: "GET", body})', 'GET'),
    invalid('fetch(url, {method: "HEAD", body})', 'HEAD'),
    invalid('new Request(url, {method: "HEAD", body})', 'HEAD'),
    invalid('fetch(url, {method: "head", body})', 'HEAD'),
    invalid('new Request(url, {method: "head", body})', 'HEAD'),
    invalid(
      'const method = "head"; new Request(url, {method, body: "foo=bar"})',
      'HEAD',
    ),
    invalid(
      'const method = "head"; fetch(url, {method, body: "foo=bar"})',
      'HEAD',
    ),
    invalid('fetch(url, {body}, extraArgument)', 'GET'),
    invalid('new Request(url, {body}, extraArgument)', 'GET'),
    invalid(
      `fetch(url, {
	body: undefined,
	body: 'foo=bar',
});`,
      'GET',
    ),
    invalid(
      `new Request(url, {
	body: undefined,
	body: 'foo=bar',
});`,
      'GET',
    ),
    invalid(
      `fetch(url, {
	method: 'post',
	body: 'foo=bar',
	method: 'HEAD',
});`,
      'HEAD',
    ),
    invalid(
      `new Request(url, {
	method: 'post',
	body: 'foo=bar',
	method: 'HEAD',
});`,
      'HEAD',
    ),
    invalid('fetch(url, {method: "get".toUpperCase(), body})', 'GET'),
    invalid(
      'new Request(url, {method: String.fromCharCode(72, 69, 65, 68), body})',
      'HEAD',
    ),
    invalid('fetch(url, {method: Array.of("GET")[0], body})', 'GET'),
    invalid('fetch(url, {method: "xGETy".slice(1, 4), body})', 'GET'),
    invalid('fetch(url, {method: "xHEADy".substring(1, 5), body})', 'HEAD'),
    invalid(
      'const S = String; fetch(url, {method: S.fromCharCode(71, 69, 84), body})',
      'GET',
    ),
    invalid(
      'const A = Array; fetch(url, {method: A.of("HEAD")[0], body})',
      'HEAD',
    ),
  ],
});
