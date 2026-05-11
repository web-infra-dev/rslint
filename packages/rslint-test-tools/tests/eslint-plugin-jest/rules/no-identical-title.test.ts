import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-identical-title', {} as never, {
  valid: [
    { code: 'it(); it();' },
    { code: 'describe(); describe();' },
    { code: 'describe("foo", () => {}); it("foo", () => {});' },
    {
      code: `
      describe("foo", () => {
        it("works", () => {});
      });
    `,
    },
    {
      code: `
      it('one', () => {});
      it('two', () => {});
    `,
    },
    {
      code: `
      describe('foo', () => {});
      describe('foe', () => {});
    `,
    },
    {
      code: `
      it(\`one\`, () => {});
      it(\`two\`, () => {});
    `,
    },
    {
      code: `
      describe(\`foo\`, () => {});
      describe(\`foe\`, () => {});
    `,
    },
    {
      code: `
      describe('foo', () => {
        test('this', () => {});
        test('that', () => {});
      });
    `,
    },
    {
      code: `
      test.concurrent('this', () => {});
      test.concurrent('that', () => {});
    `,
    },
    {
      code: `
      test.concurrent('this', () => {});
      test.only.concurrent('that', () => {});
    `,
    },
    {
      code: `
      test.only.concurrent('this', () => {});
      test.concurrent('that', () => {});
    `,
    },
    {
      code: `
      test.only.concurrent('this', () => {});
      test.only.concurrent('that', () => {});
    `,
    },
    {
      code: `
      test.only('this', () => {});
      test.only('that', () => {});
    `,
    },
    {
      code: `
      describe('foo', () => {
        it('works', () => {});

        describe('foe', () => {
          it('works', () => {});
        });
      });
    `,
    },
    {
      code: `
      describe('foo', () => {
        describe('foe', () => {
          it('works', () => {});
        });

        it('works', () => {});
      });
    `,
    },
    { code: "describe('foo', () => describe('foe', () => {}));" },
    {
      code: `
      describe('foo', () => {
        describe('foe', () => {});
      });

      describe('foe', () => {});
    `,
    },
    { code: 'test("number" + n, function() {});' },
    {
      code: 'test("number" + n, function() {}); test("number" + n, function() {});',
    },
    { code: 'it(`${n}`, function() {});' },
    { code: 'it(`${n}`, function() {}); it(`${n}`, function() {});' },
    {
      code: `
      describe('a class named ' + myClass.name, () => {
        describe('#myMethod', () => {});
      });

      describe('something else', () => {});
    `,
    },
    {
      code: `
      describe('my class', () => {
        describe('#myMethod', () => {});
        describe('a class named ' + myClass.name, () => {});
      });
    `,
    },
    {
      code: `
      describe("foo", () => {
        it(\`ignores $\{someVar} with the same title\`, () => {});
        it(\`ignores $\{someVar} with the same title\`, () => {});
      });
    `.replace(/\\\{/u, '{'),
    },
    {
      code: `
      const test = { content: () => "foo" };
      test.content(\`something that is not from jest\`, () => {});
      test.content(\`something that is not from jest\`, () => {});
    `,
    },
    {
      code: `
      const describe = { content: () => "foo" };
      describe.content(\`something that is not from jest\`, () => {});
      describe.content(\`something that is not from jest\`, () => {});
    `,
    },
    {
      code: `
      describe.each\`
        description
        $\{'b'}
      \`('$description', () => {});

      describe.each\`
        description
        $\{'a'}
      \`('$description', () => {});
    `,
    },
    {
      code: `
      describe('top level', () => {
        describe.each\`\`('nested each', () => {
          describe.each\`\`('nested nested each', () => {});
        });

        describe('nested', () => {});
      });
    `,
    },
    {
      code: `
      describe.each\`\`('my title', value => {});
      describe.each\`\`('my title', value => {});
      describe.each([])('my title', value => {});
      describe.each([])('my title', value => {});
    `,
    },
    {
      code: `
      describe.each([])('when the value is %s', value => {});
      describe.each([])('when the value is %s', value => {});
    `,
    },
  ],
  invalid: [
    {
      code: `
        describe('foo', () => {
          it('works', () => {});
          it('works', () => {});
        });
      `,
      errors: [{ messageId: 'multipleTestTitle', column: 6, line: 3 }],
    },
    {
      code: `
        it('works', () => {});
        it('works', () => {});
      `,
      errors: [{ messageId: 'multipleTestTitle', column: 4, line: 2 }],
    },
    {
      code: `
        test.only('this', () => {});
        test('this', () => {});
      `,
      errors: [{ messageId: 'multipleTestTitle', column: 6, line: 2 }],
    },
    {
      code: `
        xtest('this', () => {});
        test('this', () => {});
      `,
      errors: [{ messageId: 'multipleTestTitle', column: 6, line: 2 }],
    },
    {
      code: `
        test.only('this', () => {});
        test.only('this', () => {});
      `,
      errors: [{ messageId: 'multipleTestTitle', column: 11, line: 2 }],
    },
    {
      code: `
        test.concurrent('this', () => {});
        test.concurrent('this', () => {});
      `,
      errors: [{ messageId: 'multipleTestTitle', column: 17, line: 2 }],
    },
    {
      code: `
        test.only('this', () => {});
        test.concurrent('this', () => {});
      `,
      errors: [{ messageId: 'multipleTestTitle', column: 17, line: 2 }],
    },
    {
      code: `
        describe('foo', () => {});
        describe('foo', () => {});
      `,
      errors: [{ messageId: 'multipleDescribeTitle', column: 10, line: 2 }],
    },
    {
      code: `
        describe('foo', () => {});
        xdescribe('foo', () => {});
      `,
      errors: [{ messageId: 'multipleDescribeTitle', column: 11, line: 2 }],
    },
    {
      code: `
        fdescribe('foo', () => {});
        describe('foo', () => {});
      `,
      errors: [{ messageId: 'multipleDescribeTitle', column: 10, line: 2 }],
    },
    {
      code: `
        describe('foo', () => {
          describe('foe', () => {});
        });
        describe('foo', () => {});
      `,
      errors: [{ messageId: 'multipleDescribeTitle', column: 10, line: 4 }],
    },
    {
      code: `
        describe("foo", () => {
          it(\`catches backticks with the same title\`, () => {});
          it(\`catches backticks with the same title\`, () => {});
        });
      `,
      errors: [{ messageId: 'multipleTestTitle', column: 6, line: 3 }],
    },
    // TODO: Add this test case when we support global aliases
    // {
    //   code: `
    //     context('foo', () => {
    //       describe('foe', () => {});
    //     });
    //     describe('foo', () => {});
    //   `,
    //   errors: [{ messageId: 'multipleDescribeTitle', column: 10, line: 4 }],
    //   settings: { jest: { globalAliases: { describe: ['context'] } } },
    // },
  ],
});
