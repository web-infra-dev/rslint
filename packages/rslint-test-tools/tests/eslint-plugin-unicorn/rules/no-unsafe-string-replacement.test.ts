import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const replacementMessage = (method: 'replace' | 'replaceAll') =>
  'Do not use a non-literal replacement value with `String#' + method + '()`.';

const invalid = (
  code: string,
  method: 'replace' | 'replaceAll' = 'replace',
) => ({
  code,
  filename: 'file.js',
  errors: [{ message: replacementMessage(method) }],
});

const valid = (code: string) => ({ code, filename: 'file.js' });

ruleTester.run('no-unsafe-string-replacement', null as never, {
  valid: [
    valid('template.replace("{url}", "https://example.com")'),
    valid('template.replace("{url}", `https://example.com`)'),
    valid('template.replace("{url}", String.raw`https://example.com`)'),
    valid('template.replace("{url}", () => htmlEscape(url))'),
    valid('template.replace("{url}", function () { return htmlEscape(url); })'),
    {
      code: 'template.replace("{url}", "https://example.com" as string)',
      filename: 'file.ts',
    },
    {
      code: 'template.replace("{url}", "https://example.com" satisfies string)',
      filename: 'file.ts',
    },
    {
      code: 'template.replace("{url}", "https://example.com"!)',
      filename: 'file.ts',
    },
    {
      code: 'template.replace("{url}", <string>"https://example.com")',
      filename: 'file.ts',
    },
    valid('template.replaceAll("{url}", "https://example.com")'),
    valid('template.replaceAll("{url}", `https://example.com`)'),
    valid(
      'string.replaceAll(/(?<symbol>`|\\$(?={))/g, String.raw`\\$<symbol>`)',
    ),
    valid('template.replaceAll("{url}", () => htmlEscape(url))'),
    valid(
      'template.replaceAll("{url}", function () { return htmlEscape(url); })',
    ),
    valid('template.replace("{url}", "$` onerror=alert(1) ")'),
    valid('template.replace("{url}")'),
    valid('template.replace("{url}", replacement, extraArgument)'),
    valid('template.replace(...argumentsArray)'),
    valid('template.replace("{url}", ...replacement)'),
    valid('template[replace]("{url}", replacement)'),
    valid('template["replace"]("{url}", replacement)'),
    valid('template.notReplace("{url}", replacement)'),
    valid('replace("{url}", replacement)'),
    valid('const router = useRouter(); router.replace(pathname, {locale});'),
    valid(
      'const router = useRouter(); const options = {locale}; router.replace(pathname, options);',
    ),
    valid('router.replace(pathname, {locale: nextLocale});'),
    {
      code: 'router.replace(pathname, {locale: nextLocale} as RouterOptions);',
      filename: 'file.ts',
    },
    {
      code: 'const options = {locale: nextLocale}; router.replace(pathname, options as RouterOptions);',
      filename: 'file.ts',
    },
    {
      code: 'declare const router: {replace(href: string, options: object): void}; router.replace(pathname, {locale});',
      filename: 'file.ts',
    },
    {
      code: 'function foo(object: {replaceAll(a: string, b: object): void}) { object.replaceAll("{url}", {}); }',
      filename: 'file.ts',
    },
    {
      code: 'function foo(value: number) { value.replace("{url}", replacement); }',
      filename: 'file.ts',
    },
    {
      code: `declare const pathname: string;
declare const options: unknown;
declare function useRouter(): {replace(href: string, options: unknown): void};
useRouter().replace(pathname, options);`,
      filename: 'file.ts',
    },
  ],
  invalid: [
    invalid('template.replace("{url}", htmlEscape(url))'),
    invalid('template.replaceAll("{url}", htmlEscape(url))', 'replaceAll'),
    invalid('template.replace("{url}", replacement)'),
    invalid('template.replace("{url}", options.replacement)'),
    invalid('template.replace("{url}", options?.replacement)'),
    invalid('template.replace("{url}", String(url))'),
    invalid('template.replace("{url}", String.raw`${url}`)'),
    invalid('template.replace("{url}", css`safe string`)'),
    invalid(
      'const String = {raw: () => replacement}; template.replace("{url}", String.raw`ignored`)',
    ),
    {
      ...invalid('template.replace("{url}", htmlEscape(url) as string)'),
      filename: 'file.ts',
    },
    invalid('template.replace("{url}", `${url}`)'),
    invalid('template.replace("{url}", url ? htmlEscape(url) : "")'),
    invalid('template.replace("{url}", {toString() { return url; }})'),
    invalid('template.replace("{url}", {toString: () => url})'),
    invalid('template.replace("{url}", {valueOf: () => url})'),
    invalid(
      'template.replace("{url}", {__proto__: {toString() { return url; }}})',
    ),
    invalid('template.replace("{url}", [url])'),
    invalid('template.replace("{url}", 1)'),
    invalid('template.replace("{url}", (htmlEscape(url), url))'),
    invalid('template.replaceAll("{url}", String(++count))', 'replaceAll'),
    invalid('template?.replace("{url}", replacement)'),
    invalid('template.replace?.("{url}", replacement)'),
    invalid(
      'template.replace(\n\t"{url}",\n\t/* comment */ htmlEscape(url)\n)',
    ),
  ],
});
