import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-test-prefixes', {} as never, {
  valid: [
    { code: 'describe("foo", function () {})' },
    { code: 'it("foo", function () {})' },
    { code: 'it.concurrent("foo", function () {})' },
    { code: 'test("foo", function () {})' },
    { code: 'test.concurrent("foo", function () {})' },
    { code: 'describe.only("foo", function () {})' },
    { code: 'it.only("foo", function () {})' },
    { code: 'it.each()("foo", function () {})' },
    { code: 'it.each``("foo", function () {})' },
    { code: 'test.only("foo", function () {})' },
    { code: 'test.each()("foo", function () {})' },
    { code: 'test.each``("foo", function () {})' },
    { code: 'describe.skip("foo", function () {})' },
    { code: 'it.skip("foo", function () {})' },
    { code: 'test.skip("foo", function () {})' },
    { code: 'foo()' },
    { code: '[1,2,3].forEach()' },
  ],
  invalid: [
    {
      code: 'fdescribe("foo", function () {})',
      output: 'describe.only("foo", function () {})',
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'describe.only' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'xdescribe.each([])("foo", function () {})',
      output: 'describe.skip.each([])("foo", function () {})',
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'describe.skip.each' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'fit("foo", function () {})',
      output: 'it.only("foo", function () {})',
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'it.only' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'xdescribe("foo", function () {})',
      output: 'describe.skip("foo", function () {})',
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'describe.skip' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'xit("foo", function () {})',
      output: 'it.skip("foo", function () {})',
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'it.skip' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'xtest("foo", function () {})',
      output: 'test.skip("foo", function () {})',
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'test.skip' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'xit.each``("foo", function () {})',
      output: 'it.skip.each``("foo", function () {})',
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'it.skip.each' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'xtest.each``("foo", function () {})',
      output: 'test.skip.each``("foo", function () {})',
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'test.skip.each' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'xit.each([])("foo", function () {})',
      output: 'it.skip.each([])("foo", function () {})',
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'it.skip.each' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: 'xtest.each([])("foo", function () {})',
      output: 'test.skip.each([])("foo", function () {})',
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'test.skip.each' },
          column: 1,
          line: 1,
        },
      ],
    },
    {
      code: `
        import { xit } from '@jest/globals';

        xit("foo", function () {})
      `,
      output: `
        import { xit } from '@jest/globals';

        it.skip("foo", function () {})
      `,
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'it.skip' },
          column: 1,
          line: 3,
        },
      ],
    },
    {
      code: `
        import { xit as skipThis } from '@jest/globals';

        skipThis("foo", function () {})
      `,
      output: `
        import { xit as skipThis } from '@jest/globals';

        it.skip("foo", function () {})
      `,
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'it.skip' },
          column: 1,
          line: 3,
        },
      ],
    },
    {
      code: `
        import { fit as onlyThis } from '@jest/globals';

        onlyThis("foo", function () {})
      `,
      output: `
        import { fit as onlyThis } from '@jest/globals';

        it.only("foo", function () {})
      `,
      errors: [
        {
          messageId: 'usePreferredName',
          data: { preferredNodeName: 'it.only' },
          column: 1,
          line: 3,
        },
      ],
    },
  ],
});
