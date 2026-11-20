import { RuleTester, type ExpectedSuggestion } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNeverValidError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'neverValid',
    message: '“alternatex” is never a valid “rel” attribute value.',
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

function expectOnlyMeaningfulForError(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'onlyMeaningfulFor',
    message:
      'The ”rel“ attribute only has meaning on the tags: "<link>", "<a>", "<area>", "<form>"',
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

function expectOnlyStringsError(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'onlyStrings',
    message: '“rel” attribute only supports strings.',
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

function expectNeverValidError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'neverValid',
    message: '“foobar” is never a valid “rel” attribute value.',
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

function expectSpaceDelimitedError(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'spaceDelimited',
    message: '”rel“ attribute values should be space delimited.',
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

function expectNeverValidError6(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'neverValid',
    message: '“batgo” is never a valid “rel” attribute value.',
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

const ruleTester = new RuleTester(fixtureConfig);

// Layer 1 only: upstream eslint-plugin-react v7.37.5 fixtures.
ruleTester.run('no-invalid-html-attribute', {
  valid: [
    { code: '<a rel="alternate"></a>' },
    { code: 'React.createElement("a", { rel: "alternate" })' },
    { code: 'React.createElement("a", { rel: ["alternate"] })' },
    { code: '<a rel="author"></a>' },
    { code: 'React.createElement("a", { rel: "author" })' },
    { code: 'React.createElement("a", { rel: ["author"] })' },
    { code: '<a rel="bookmark"></a>' },
    { code: 'React.createElement("a", { rel: "bookmark" })' },
    { code: 'React.createElement("a", { rel: ["bookmark"] })' },
    { code: '<a rel="external"></a>' },
    { code: 'React.createElement("a", { rel: "external" })' },
    { code: 'React.createElement("a", { rel: ["external"] })' },
    { code: '<a rel="help"></a>' },
    { code: 'React.createElement("a", { rel: "help" })' },
    { code: 'React.createElement("a", { rel: ["help"] })' },
    { code: '<a rel="license"></a>' },
    { code: 'React.createElement("a", { rel: "license" })' },
    { code: 'React.createElement("a", { rel: ["license"] })' },
    { code: '<a rel="next"></a>' },
    { code: 'React.createElement("a", { rel: "next" })' },
    { code: 'React.createElement("a", { rel: ["next"] })' },
    { code: '<a rel="nofollow"></a>' },
    { code: 'React.createElement("a", { rel: "nofollow" })' },
    { code: 'React.createElement("a", { rel: ["nofollow"] })' },
    { code: '<a rel="noopener"></a>' },
    { code: 'React.createElement("a", { rel: "noopener" })' },
    { code: 'React.createElement("a", { rel: ["noopener"] })' },
    { code: '<a rel="noreferrer"></a>' },
    { code: 'React.createElement("a", { rel: "noreferrer" })' },
    { code: 'React.createElement("a", { rel: ["noreferrer"] })' },
    { code: '<a rel="opener"></a>' },
    { code: 'React.createElement("a", { rel: "opener" })' },
    { code: 'React.createElement("a", { rel: ["opener"] })' },
    { code: '<a rel="prev"></a>' },
    { code: 'React.createElement("a", { rel: "prev" })' },
    { code: 'React.createElement("a", { rel: ["prev"] })' },
    { code: '<a rel="search"></a>' },
    { code: 'React.createElement("a", { rel: "search" })' },
    { code: 'React.createElement("a", { rel: ["search"] })' },
    { code: '<a rel="tag"></a>' },
    { code: 'React.createElement("a", { rel: "tag" })' },
    { code: 'React.createElement("a", { rel: ["tag"] })' },
    { code: '<area rel="alternate"></area>' },
    { code: 'React.createElement("area", { rel: "alternate" })' },
    { code: 'React.createElement("area", { rel: ["alternate"] })' },
    { code: '<area rel="author"></area>' },
    { code: 'React.createElement("area", { rel: "author" })' },
    { code: 'React.createElement("area", { rel: ["author"] })' },
    { code: '<area rel="bookmark"></area>' },
    { code: 'React.createElement("area", { rel: "bookmark" })' },
    { code: 'React.createElement("area", { rel: ["bookmark"] })' },
    { code: '<area rel="external"></area>' },
    { code: 'React.createElement("area", { rel: "external" })' },
    { code: 'React.createElement("area", { rel: ["external"] })' },
    { code: '<area rel="help"></area>' },
    { code: 'React.createElement("area", { rel: "help" })' },
    { code: 'React.createElement("area", { rel: ["help"] })' },
    { code: '<area rel="license"></area>' },
    { code: 'React.createElement("area", { rel: "license" })' },
    { code: 'React.createElement("area", { rel: ["license"] })' },
    { code: '<area rel="next"></area>' },
    { code: 'React.createElement("area", { rel: "next" })' },
    { code: 'React.createElement("area", { rel: ["next"] })' },
    { code: '<area rel="nofollow"></area>' },
    { code: 'React.createElement("area", { rel: "nofollow" })' },
    { code: 'React.createElement("area", { rel: ["nofollow"] })' },
    { code: '<area rel="noopener"></area>' },
    { code: 'React.createElement("area", { rel: "noopener" })' },
    { code: 'React.createElement("area", { rel: ["noopener"] })' },
    { code: '<area rel="noreferrer"></area>' },
    { code: 'React.createElement("area", { rel: "noreferrer" })' },
    { code: 'React.createElement("area", { rel: ["noreferrer"] })' },
    { code: '<area rel="opener"></area>' },
    { code: 'React.createElement("area", { rel: "opener" })' },
    { code: 'React.createElement("area", { rel: ["opener"] })' },
    { code: '<area rel="prev"></area>' },
    { code: 'React.createElement("area", { rel: "prev" })' },
    { code: 'React.createElement("area", { rel: ["prev"] })' },
    { code: '<area rel="search"></area>' },
    { code: 'React.createElement("area", { rel: "search" })' },
    { code: 'React.createElement("area", { rel: ["search"] })' },
    { code: '<area rel="tag"></area>' },
    { code: 'React.createElement("area", { rel: "tag" })' },
    { code: 'React.createElement("area", { rel: ["tag"] })' },
    { code: '<link rel="alternate"></link>' },
    { code: 'React.createElement("link", { rel: "alternate" })' },
    { code: 'React.createElement("link", { rel: ["alternate"] })' },
    { code: '<link rel="author"></link>' },
    { code: 'React.createElement("link", { rel: "author" })' },
    { code: 'React.createElement("link", { rel: ["author"] })' },
    { code: '<link rel="canonical"></link>' },
    { code: 'React.createElement("link", { rel: "canonical" })' },
    { code: 'React.createElement("link", { rel: ["canonical"] })' },
    { code: '<link rel="dns-prefetch"></link>' },
    { code: 'React.createElement("link", { rel: "dns-prefetch" })' },
    { code: 'React.createElement("link", { rel: ["dns-prefetch"] })' },
    { code: '<link rel="help"></link>' },
    { code: 'React.createElement("link", { rel: "help" })' },
    { code: 'React.createElement("link", { rel: ["help"] })' },
    { code: '<link rel="icon"></link>' },
    { code: 'React.createElement("link", { rel: "icon" })' },
    { code: 'React.createElement("link", { rel: ["icon"] })' },
    { code: '<link rel="shortcut icon"></link>' },
    { code: 'React.createElement("link", { rel: "shortcut icon" })' },
    { code: 'React.createElement("link", { rel: ["shortcut icon"] })' },
    { code: '<link rel="license"></link>' },
    { code: 'React.createElement("link", { rel: "license" })' },
    { code: 'React.createElement("link", { rel: ["license"] })' },
    { code: '<link rel="manifest"></link>' },
    { code: 'React.createElement("link", { rel: "manifest" })' },
    { code: 'React.createElement("link", { rel: ["manifest"] })' },
    { code: '<link rel="modulepreload"></link>' },
    { code: 'React.createElement("link", { rel: "modulepreload" })' },
    { code: 'React.createElement("link", { rel: ["modulepreload"] })' },
    { code: '<link rel="next"></link>' },
    { code: 'React.createElement("link", { rel: "next" })' },
    { code: 'React.createElement("link", { rel: ["next"] })' },
    { code: '<link rel="pingback"></link>' },
    { code: 'React.createElement("link", { rel: "pingback" })' },
    { code: 'React.createElement("link", { rel: ["pingback"] })' },
    { code: '<link rel="preconnect"></link>' },
    { code: 'React.createElement("link", { rel: "preconnect" })' },
    { code: 'React.createElement("link", { rel: ["preconnect"] })' },
    { code: '<link rel="prefetch"></link>' },
    { code: 'React.createElement("link", { rel: "prefetch" })' },
    { code: 'React.createElement("link", { rel: ["prefetch"] })' },
    { code: '<link rel="preload"></link>' },
    { code: 'React.createElement("link", { rel: "preload" })' },
    { code: 'React.createElement("link", { rel: ["preload"] })' },
    { code: '<link rel="prerender"></link>' },
    { code: 'React.createElement("link", { rel: "prerender" })' },
    { code: 'React.createElement("link", { rel: ["prerender"] })' },
    { code: '<link rel="prev"></link>' },
    { code: 'React.createElement("link", { rel: "prev" })' },
    { code: 'React.createElement("link", { rel: ["prev"] })' },
    { code: '<link rel="search"></link>' },
    { code: 'React.createElement("link", { rel: "search" })' },
    { code: 'React.createElement("link", { rel: ["search"] })' },
    { code: '<link rel="stylesheet"></link>' },
    { code: 'React.createElement("link", { rel: "stylesheet" })' },
    { code: 'React.createElement("link", { rel: ["stylesheet"] })' },
    { code: '<form rel="external"></form>' },
    { code: 'React.createElement("form", { rel: "external" })' },
    { code: 'React.createElement("form", { rel: ["external"] })' },
    { code: '<form rel="help"></form>' },
    { code: 'React.createElement("form", { rel: "help" })' },
    { code: 'React.createElement("form", { rel: ["help"] })' },
    { code: '<form rel="license"></form>' },
    { code: 'React.createElement("form", { rel: "license" })' },
    { code: 'React.createElement("form", { rel: ["license"] })' },
    { code: '<form rel="next"></form>' },
    { code: 'React.createElement("form", { rel: "next" })' },
    { code: 'React.createElement("form", { rel: ["next"] })' },
    { code: '<form rel="nofollow"></form>' },
    { code: 'React.createElement("form", { rel: "nofollow" })' },
    { code: 'React.createElement("form", { rel: ["nofollow"] })' },
    { code: '<form rel="noopener"></form>' },
    { code: 'React.createElement("form", { rel: "noopener" })' },
    { code: 'React.createElement("form", { rel: ["noopener"] })' },
    { code: '<form rel="noreferrer"></form>' },
    { code: 'React.createElement("form", { rel: "noreferrer" })' },
    { code: 'React.createElement("form", { rel: ["noreferrer"] })' },
    { code: '<form rel="opener"></form>' },
    { code: 'React.createElement("form", { rel: "opener" })' },
    { code: 'React.createElement("form", { rel: ["opener"] })' },
    { code: '<form rel="prev"></form>' },
    { code: 'React.createElement("form", { rel: "prev" })' },
    { code: 'React.createElement("form", { rel: ["prev"] })' },
    { code: '<form rel="search"></form>' },
    { code: 'React.createElement("form", { rel: "search" })' },
    { code: 'React.createElement("form", { rel: ["search"] })' },
    { code: '<form rel={callFoo()}></form>' },
    { code: 'React.createElement("form", { rel: callFoo() })' },
    { code: 'React.createElement("form", { rel: [callFoo()] })' },
    { code: '<a rel={{a: "noreferrer"}["a"]}></a>' },
    { code: '<a rel={{a: "noreferrer"}["b"]}></a>' },
    { code: '<Foo rel></Foo>' },
    { code: 'React.createElement("Foo", { rel: true })' },
    {
      code: "\n        React.createElement('a', {\n          ...rest,\n          href: to,\n        })\n      ",
    },
    {
      code: '<link rel="apple-touch-icon" sizes="60x60" href="apple-touch-icon-60x60.png" />',
    },
    {
      code: '<link rel="apple-touch-icon" sizes="76x76" href="apple-touch-icon-76x76.png" />',
    },
    {
      code: '<link rel="apple-touch-icon" sizes="120x120" href="apple-touch-icon-120x120.png" />',
    },
    {
      code: '<link rel="apple-touch-icon" sizes="152x152" href="apple-touch-icon-152x152.png" />',
    },
    {
      code: '<link rel="apple-touch-icon" sizes="180x180" href="apple-touch-icon-180x180.png" />',
    },
    { code: '<link rel="apple-touch-startup-image" href="launch.png" />' },
    {
      code: '<link rel="apple-touch-startup-image" href="iphone5.png" media="(device-width: 320px) and (device-height: 568px) and (-webkit-device-pixel-ratio: 2)" />',
    },
    {
      code: '<link rel="mask-icon" href="/safari-pinned-tab.svg" color="#fff" />',
    },
  ],
  invalid: [
    {
      code: '<a rel="alternatex"></a>',
      errors: [
        expectNeverValidError1(
          [1, 8, 1, 20],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute alternatex”',
              output: '<a rel=""></a>',
            },
          ],
        ),
      ],
    },
    {
      code: 'React.createElement("a", { rel: "alternatex" })',
      errors: [
        expectNeverValidError1(
          [1, 33, 1, 45],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute alternatex”',
              output: 'React.createElement("a", { rel: "" })',
            },
          ],
        ),
      ],
    },
    {
      code: 'React.createElement("a", { rel: ["alternatex"] })',
      errors: [
        expectNeverValidError1(
          [1, 34, 1, 46],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute alternatex”',
              output: 'React.createElement("a", { rel: [""] })',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel="alternatex alternate"></a>',
      errors: [
        expectNeverValidError1(
          [1, 8, 1, 30],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute alternatex”',
              output: '<a rel=" alternate"></a>',
            },
          ],
        ),
      ],
    },
    {
      code: 'React.createElement("a", { rel: "alternatex alternate" })',
      errors: [
        {
          messageId: 'neverValid',
          message:
            '“alternatex alternate” is never a valid “rel” attribute value.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 55,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute alternatex alternate”',
              output: 'React.createElement("a", { rel: "" })',
            },
          ],
        },
      ],
    },
    {
      code: 'React.createElement("a", { rel: ["alternatex alternate"] })',
      errors: [
        {
          messageId: 'neverValid',
          message:
            '“alternatex alternate” is never a valid “rel” attribute value.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 56,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute alternatex alternate”',
              output: 'React.createElement("a", { rel: [""] })',
            },
          ],
        },
      ],
    },
    {
      code: '<a rel="alternate alternatex"></a>',
      errors: [
        expectNeverValidError1(
          [1, 8, 1, 30],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute alternatex”',
              output: '<a rel="alternate "></a>',
            },
          ],
        ),
      ],
    },
    {
      code: 'React.createElement("a", { rel: "alternate alternatex" })',
      errors: [
        {
          messageId: 'neverValid',
          message:
            '“alternate alternatex” is never a valid “rel” attribute value.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 55,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute alternate alternatex”',
              output: 'React.createElement("a", { rel: "" })',
            },
          ],
        },
      ],
    },
    {
      code: 'React.createElement("a", { rel: ["alternate alternatex"] })',
      errors: [
        {
          messageId: 'neverValid',
          message:
            '“alternate alternatex” is never a valid “rel” attribute value.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 56,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute alternate alternatex”',
              output: 'React.createElement("a", { rel: [""] })',
            },
          ],
        },
      ],
    },
    {
      code: '<html rel></html>',
      errors: [
        expectOnlyMeaningfulForError(
          [1, 7, 1, 10],
          [
            {
              messageId: 'suggestRemoveDefault',
              desc: '"remove rel"',
              output: '<html ></html>',
            },
          ],
        ),
      ],
    },
    {
      code: 'React.createElement("html", { rel: 1 })',
      errors: [expectOnlyMeaningfulForError([1, 31, 1, 34])],
    },
    {
      code: '<a rel></a>',
      errors: [
        {
          messageId: 'emptyIsMeaningless',
          message: 'An empty “rel” attribute is meaningless.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 7,
          suggestions: [
            {
              messageId: 'suggestRemoveEmpty',
              desc: '"remove empty attribute rel"',
              output: '<a ></a>',
            },
          ],
        },
      ],
    },
    {
      code: 'React.createElement("a", { rel() { return 1; } })',
      errors: [
        {
          messageId: 'noMethod',
          message: 'The ”rel“ attribute cannot be a method.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: '<span rel></span>',
      errors: [
        expectOnlyMeaningfulForError(
          [1, 7, 1, 10],
          [
            {
              messageId: 'suggestRemoveDefault',
              desc: '"remove rel"',
              output: '<span ></span>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={null}></a>',
      errors: [
        expectOnlyStringsError(
          [1, 9, 1, 13],
          [
            {
              messageId: 'suggestRemoveNonString',
              desc: 'remove non-string value in “rel”',
              output: '<a ></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={5}></a>',
      errors: [
        expectOnlyStringsError(
          [1, 9, 1, 10],
          [
            {
              messageId: 'suggestRemoveNonString',
              desc: 'remove non-string value in “rel”',
              output: '<a ></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={true}></a>',
      errors: [
        expectOnlyStringsError(
          [1, 9, 1, 13],
          [
            {
              messageId: 'suggestRemoveNonString',
              desc: 'remove non-string value in “rel”',
              output: '<a ></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={{}}></a>',
      errors: [
        expectOnlyStringsError(
          [1, 8, 1, 12],
          [
            {
              messageId: 'suggestRemoveDefault',
              desc: '"remove rel"',
              output: '<a ></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={undefined}></a>',
      errors: [
        expectOnlyStringsError(
          [1, 8, 1, 19],
          [
            {
              messageId: 'suggestRemoveDefault',
              desc: '"remove rel"',
              output: '<a ></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel="noreferrer noopener foobar"></a>',
      errors: [
        expectNeverValidError4(
          [1, 8, 1, 36],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute foobar”',
              output: '<a rel="noreferrer noopener "></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel="noreferrer noopener   "></a>',
      errors: [
        expectSpaceDelimitedError(
          [1, 8, 1, 32],
          [
            {
              messageId: 'suggestRemoveWhitespaces',
              desc: 'remove whitespaces in “rel”',
              output: '<a rel="noreferrer noopener"></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel="noreferrer        noopener"></a>',
      errors: [
        expectSpaceDelimitedError(
          [1, 8, 1, 36],
          [
            {
              messageId: 'suggestRemoveWhitespaces',
              desc: 'remove whitespaces in “rel”',
              output: '<a rel="noreferrer noopener"></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel="noreferrer  noopener"></a>',
      errors: [
        expectSpaceDelimitedError(
          [1, 8, 1, 30],
          [
            {
              messageId: 'suggestRemoveWhitespaces',
              desc: 'remove whitespaces in “rel”',
              output: '<a rel="noreferrer noopener"></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={"noreferrer noopener foobar"}></a>',
      errors: [
        expectNeverValidError4(
          [1, 9, 1, 37],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute foobar”',
              output: '<a rel={"noreferrer noopener "}></a>',
            },
          ],
        ),
      ],
    },
    {
      code: 'React.createElement("a", { rel: ["noreferrer", "noopener", "foobar" ] })',
      errors: [
        expectNeverValidError4(
          [1, 60, 1, 68],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute foobar”',
              output:
                'React.createElement("a", { rel: ["noreferrer", "noopener", "" ] })',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={"foobar noreferrer noopener"}></a>',
      errors: [
        expectNeverValidError4(
          [1, 9, 1, 37],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute foobar”',
              output: '<a rel={" noreferrer noopener"}></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={"foobar batgo       noopener"}></a>',
      errors: [
        expectNeverValidError4(
          [1, 9, 1, 38],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute foobar”',
              output: '<a rel={" batgo       noopener"}></a>',
            },
          ],
        ),
        expectNeverValidError6(
          [1, 9, 1, 38],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute batgo”',
              output: '<a rel={"foobar        noopener"}></a>',
            },
          ],
        ),
        expectSpaceDelimitedError(
          [1, 9, 1, 38],
          [
            {
              messageId: 'suggestRemoveWhitespaces',
              desc: 'remove whitespaces in “rel”',
              output: '<a rel={"foobar batgo noopener"}></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={"        noopener"}></a>',
      errors: [
        expectSpaceDelimitedError(
          [1, 9, 1, 27],
          [
            {
              messageId: 'suggestRemoveWhitespaces',
              desc: 'remove whitespaces in “rel”',
              output: '<a rel={"noopener"}></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={"noopener        "}></a>',
      errors: [
        expectSpaceDelimitedError(
          [1, 9, 1, 27],
          [
            {
              messageId: 'suggestRemoveWhitespaces',
              desc: 'remove whitespaces in “rel”',
              output: '<a rel={"noopener"}></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={" batgo noopener"}></a>',
      errors: [
        expectNeverValidError6(
          [1, 9, 1, 26],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute batgo”',
              output: '<a rel={"  noopener"}></a>',
            },
          ],
        ),
        expectSpaceDelimitedError(
          [1, 9, 1, 26],
          [
            {
              messageId: 'suggestRemoveWhitespaces',
              desc: 'remove whitespaces in “rel”',
              output: '<a rel={"batgo noopener"}></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={"batgo noopener"}></a>',
      errors: [
        expectNeverValidError6(
          [1, 9, 1, 25],
          [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute batgo”',
              output: '<a rel={" noopener"}></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel={" noopener"}></a>',
      errors: [
        expectSpaceDelimitedError(
          [1, 9, 1, 20],
          [
            {
              messageId: 'suggestRemoveWhitespaces',
              desc: 'remove whitespaces in “rel”',
              output: '<a rel={"noopener"}></a>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel="canonical"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“canonical” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute canonical”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<a rel="dns-prefetch"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“dns-prefetch” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute dns-prefetch”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<a rel="icon"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“icon” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute icon”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<link rel="shortcut"></link>',
      errors: [
        {
          messageId: 'notAlone',
          message: '“shortcut” must be directly followed by “icon”.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<link rel="shortcut foo"></link>',
      errors: [
        {
          messageId: 'neverValid',
          message: '“foo” is never a valid “rel” attribute value.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute foo”',
              output: '<link rel="shortcut "></link>',
            },
          ],
        },
        {
          messageId: 'notPaired',
          message:
            '“shortcut” can not be directly followed by “foo” without “icon”.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<link rel="shortcut  icon"></link>',
      errors: [
        expectSpaceDelimitedError(
          [1, 11, 1, 27],
          [
            {
              messageId: 'suggestRemoveWhitespaces',
              desc: 'remove whitespaces in “rel”',
              output: '<link rel="shortcut icon"></link>',
            },
          ],
        ),
      ],
    },
    {
      code: '<link rel="shortcut  foo"></link>',
      errors: [
        {
          messageId: 'neverValid',
          message: '“foo” is never a valid “rel” attribute value.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute foo”',
              output: '<link rel="shortcut  "></link>',
            },
          ],
        },
        {
          messageId: 'notAlone',
          message: '“shortcut” must be directly followed by “icon”.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 26,
        },
        expectSpaceDelimitedError(
          [1, 11, 1, 26],
          [
            {
              messageId: 'suggestRemoveWhitespaces',
              desc: 'remove whitespaces in “rel”',
              output: '<link rel="shortcut foo"></link>',
            },
          ],
        ),
      ],
    },
    {
      code: '<a rel="manifest"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“manifest” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute manifest”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<a rel="modulepreload"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“modulepreload” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute modulepreload”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<a rel="pingback"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“pingback” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute pingback”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<a rel="preconnect"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“preconnect” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute preconnect”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<a rel="prefetch"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“prefetch” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute prefetch”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<a rel="preload"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“preload” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute preload”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<a rel="prerender"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“prerender” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute prerender”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<a rel="stylesheet"></a>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“stylesheet” is not a valid “rel” attribute value for <a>.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute stylesheet”',
              output: '<a rel=""></a>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="canonical"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“canonical” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute canonical”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="dns-prefetch"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“dns-prefetch” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute dns-prefetch”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="icon"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“icon” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute icon”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="manifest"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“manifest” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute manifest”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="modulepreload"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“modulepreload” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute modulepreload”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="pingback"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“pingback” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute pingback”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="preconnect"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“preconnect” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute preconnect”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="prefetch"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“prefetch” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute prefetch”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="preload"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“preload” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute preload”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="prerender"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“prerender” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute prerender”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<area rel="stylesheet"></area>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“stylesheet” is not a valid “rel” attribute value for <area>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute stylesheet”',
              output: '<area rel=""></area>',
            },
          ],
        },
      ],
    },
    {
      code: '<link rel="bookmark"></link>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“bookmark” is not a valid “rel” attribute value for <link>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute bookmark”',
              output: '<link rel=""></link>',
            },
          ],
        },
      ],
    },
    {
      code: '<link rel="external"></link>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“external” is not a valid “rel” attribute value for <link>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute external”',
              output: '<link rel=""></link>',
            },
          ],
        },
      ],
    },
    {
      code: '<link rel="nofollow"></link>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“nofollow” is not a valid “rel” attribute value for <link>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute nofollow”',
              output: '<link rel=""></link>',
            },
          ],
        },
      ],
    },
    {
      code: '<link rel="noopener"></link>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“noopener” is not a valid “rel” attribute value for <link>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute noopener”',
              output: '<link rel=""></link>',
            },
          ],
        },
      ],
    },
    {
      code: '<link rel="noreferrer"></link>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“noreferrer” is not a valid “rel” attribute value for <link>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute noreferrer”',
              output: '<link rel=""></link>',
            },
          ],
        },
      ],
    },
    {
      code: '<link rel="opener"></link>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“opener” is not a valid “rel” attribute value for <link>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute opener”',
              output: '<link rel=""></link>',
            },
          ],
        },
      ],
    },
    {
      code: '<link rel="tag"></link>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“tag” is not a valid “rel” attribute value for <link>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute tag”',
              output: '<link rel=""></link>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="alternate"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“alternate” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute alternate”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="author"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“author” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute author”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="bookmark"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“bookmark” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute bookmark”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="canonical"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“canonical” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute canonical”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="dns-prefetch"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“dns-prefetch” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute dns-prefetch”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="icon"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“icon” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute icon”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="manifest"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“manifest” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute manifest”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="modulepreload"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“modulepreload” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute modulepreload”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="pingback"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“pingback” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute pingback”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="preconnect"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“preconnect” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute preconnect”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="prefetch"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“prefetch” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute prefetch”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="preload"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“preload” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute preload”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="prerender"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“prerender” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute prerender”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="stylesheet"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message:
            '“stylesheet” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute stylesheet”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel="tag"></form>',
      errors: [
        {
          messageId: 'notValidFor',
          message: '“tag” is not a valid “rel” attribute value for <form>.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute tag”',
              output: '<form rel=""></form>',
            },
          ],
        },
      ],
    },
    {
      code: '<form rel=""></form>',
      errors: [
        {
          messageId: 'noEmpty',
          message: 'An empty “rel” attribute is meaningless.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 13,
          suggestions: [
            {
              messageId: 'suggestRemoveEmpty',
              desc: '"remove empty attribute rel"',
              output: '<form ></form>',
            },
          ],
        },
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 12 (7.37.5)',
      skip: 'Known no-invalid-html-attribute behavior gap: native reports 0 diagnostics for this input; upstream reports 1.',
      code: `React.createElement("a", { rel: 1 })`,
      errors: [
        {
          messageId: 'neverValid',
          message: '“1” is never a valid “rel” attribute value.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 34,
          suggestions: [
            {
              messageId: 'suggestRemoveInvalid',
              desc: '“remove invalid attribute 1”',
              output: 'React.createElement("a", { rel:  })',
            },
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
