import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

const parserOptions = {
  ecmaVersion: 2018,
  sourceType: 'module',
  ecmaFeatures: {
    jsx: true,
  },
};

function expectMoveToNewLineError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'moveToNewLine',
    message: '`{"foo"}` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMoveToNewLineError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'moveToNewLine',
    message: '`{"bar"}` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMoveToNewLineError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'moveToNewLine',
    message: '`Bar` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMoveToNewLineError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'moveToNewLine',
    message: "`{ I18n.t('baz') }` must be placed on a new line",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMoveToNewLineError5(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'moveToNewLine',
    message: '`Baz` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMoveToNewLineError6(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'moveToNewLine',
    message: '`input` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMoveToNewLineError7(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'moveToNewLine',
    message: '`Foo` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMoveToNewLineError8(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'moveToNewLine',
    message: '` baz        ` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMoveToNewLineError9(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'moveToNewLine',
    message: '`{          foo        }` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMoveToNewLineError10(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'moveToNewLine',
    message: '`button` must be placed on a new line',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-one-expression-per-line', {
  valid: [
    {
      code: '<App />',
    },
    {
      code: `
\t\t\t\t<AllTabs>
\t\t\t\t\tFail
\t\t\t\t</AllTabs>
      `,
    },
    {
      code: `
\t\t\t\t<TagsWithTabs>
          Fail
\t\t\t\t</TagsWithTabs>
      `,
    },
    {
      code: `
        <ClosedTagWithTabs>
          Fail
\t\t\t\t</ClosedTagWithTabs>
      `,
    },
    {
      code: `
\t\t\t\t<OpenTagWithTabs>
          OK
        </OpenTagWithTabs>
      `,
    },
    {
      code: `
        <TextWithTabs>
\t\t\t\t\t\tOK
        </TextWithTabs>
      `,
    },
    {
      code: `
        <AllSpaces>
          OK
        </AllSpaces>
      `,
    },
    {
      code: '<App></App>',
    },
    {
      code: '<App foo="bar" />',
    },
    {
      code: `
        <App>
          <Foo />
        </App>
      `,
    },
    {
      code: `
        <App>
          <Foo />
          <Bar />
        </App>
      `,
    },
    {
      code: `
        <App>
          <Foo></Foo>
        </App>
      `,
    },
    {
      code: `
        <App>
          foo bar baz  whatever
        </App>
      `,
    },
    {
      code: `
        <App>
          <Foo>
          </Foo>
        </App>
      `,
    },
    {
      code: `
        <App
          foo="bar"
        >
        <Foo />
        </App>
      `,
    },
    {
      code: `
        <
        App
        >
          <
            Foo
          />
        </
        App
        >
      `,
    },
    {
      code: '<App>foo</App>',
      options: [{ allow: 'literal' }],
    },
    {
      code: '<App>123</App>',
      options: [{ allow: 'literal' }],
    },
    {
      code: '<App>foo</App>',
      options: [{ allow: 'single-child' }],
    },
    {
      code: '<App>{"foo"}</App>',
      options: [{ allow: 'single-child' }],
    },
    {
      code: '<App>123</App>',
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: '<App>foo</App>',
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: '<App>{"foo"}</App>',
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: '<App>{<Bar />}</App>',
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: '<App>{foo && <Bar />}</App>',
      options: [{ allow: 'single-child' }],
    },
    {
      code: '<App><Foo /></App>',
      options: [{ allow: 'single-child' }],
    },
    {
      code: '<></>',
    },
    {
      code: `
        <>
          <Foo />
        </>
      `,
      // TODO: FIXME: remove no-ts-old and fix
    },
    {
      code: `
        <>
          <Foo />
          <Bar />
        </>
      `,
      // TODO: FIXME: remove no-ts-old and fix
    },
    {
      code: '<App>Hello {name}</App>',
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
        <App>
          Hello {name} there!
        </App>`,
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
        <App>
          Hello {<Bar />} there!
        </App>`,
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
        <App>
          Hello {(<Bar />)} there!
        </App>`,
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
        <App>
          Hello {(() => <Bar />)()} there!
        </App>`,
      options: [{ allow: 'non-jsx' }],
    },
  ],

  invalid: [
    {
      code: `
        <App>{"foo"}</App>
      `,
      output: `
        <App>
{"foo"}
</App>
      `,
      errors: [expectMoveToNewLineError1([2, 14, 2, 21])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>foo</App>
      `,
      output: `
        <App>
foo
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`foo` must be placed on a new line',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 17,
        },
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          foo {"bar"}
        </div>
      `,
      output: `
        <div>
          foo${' ' /* intentional trailing space */}
{' '}
{"bar"}
        </div>
      `,
      errors: [expectMoveToNewLineError2([3, 15, 3, 22])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          {"foo"} bar
        </div>
      `,
      output: `
        <div>
          {"foo"}
{' '}
bar
</div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '` bar        ` must be placed on a new line',
          line: 3,
          column: 18,
          endLine: 4,
          endColumn: 9,
        },
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
          <Foo /><Bar />
        </App>
      `,
      output: `
        <App>
          <Foo />
<Bar />
        </App>
      `,
      errors: [expectMoveToNewLineError3([3, 18, 3, 25])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          <span />foo
        </div>
      `,
      output: `
        <div>
          <span />
foo
</div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`foo        ` must be placed on a new line',
          line: 3,
          column: 19,
          endLine: 4,
          endColumn: 9,
        },
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          <span />{"foo"}
        </div>
      `,
      output: `
        <div>
          <span />
{"foo"}
        </div>
      `,
      errors: [expectMoveToNewLineError1([3, 19, 3, 26])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          {"foo"} { I18n.t('baz') }
        </div>
      `,
      output: `
        <div>
          {"foo"}${' ' /* intentional trailing space */}
{' '}
{ I18n.t('baz') }
        </div>
      `,
      errors: [expectMoveToNewLineError4([3, 19, 3, 36])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <Text style={styles.foo}>{ bar } <Text/> { I18n.t('baz') }</Text>
      `,
      output: `
        <Text style={styles.foo}>
{ bar }${' ' /* intentional trailing space */}
{' '}
<Text/>${' ' /* intentional trailing space */}
{' '}
{ I18n.t('baz') }
</Text>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`{ bar }` must be placed on a new line',
          line: 2,
          column: 34,
          endLine: 2,
          endColumn: 41,
        },
        {
          messageId: 'moveToNewLine',
          message: '`Text` must be placed on a new line',
          line: 2,
          column: 42,
          endLine: 2,
          endColumn: 49,
        },
        expectMoveToNewLineError4([2, 50, 2, 67]),
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <Text style={styles.foo}> <Bar/> <Baz/></Text>
      `,
      output: `
        <Text style={styles.foo}>${' ' /* intentional trailing space */}
{' '}
<Bar/>${' ' /* intentional trailing space */}
{' '}
<Baz/>
</Text>
      `,
      errors: [
        expectMoveToNewLineError3([2, 35, 2, 41]),
        expectMoveToNewLineError5([2, 42, 2, 48]),
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <Text style={styles.foo}> <Bar/> <Baz/> <Bunk/> <Bruno/> </Text>
      `,
      output: `
        <Text style={styles.foo}>${' ' /* intentional trailing space */}
{' '}
<Bar/>${' ' /* intentional trailing space */}
{' '}
<Baz/>${' ' /* intentional trailing space */}
{' '}
<Bunk/>${' ' /* intentional trailing space */}
{' '}
<Bruno/>
{' '}
 </Text>
      `,
      errors: [
        expectMoveToNewLineError3([2, 35, 2, 41]),
        expectMoveToNewLineError5([2, 42, 2, 48]),
        {
          messageId: 'moveToNewLine',
          message: '`Bunk` must be placed on a new line',
          line: 2,
          column: 49,
          endLine: 2,
          endColumn: 56,
        },
        {
          messageId: 'moveToNewLine',
          message: '`Bruno` must be placed on a new line',
          line: 2,
          column: 57,
          endLine: 2,
          endColumn: 65,
        },
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <Text style={styles.foo}> <Bar /></Text>
      `,
      output: `
        <Text style={styles.foo}>${' ' /* intentional trailing space */}
{' '}
<Bar />
</Text>
      `,
      errors: [expectMoveToNewLineError3([2, 35, 2, 42])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <Text style={styles.foo}> <Bar />
        </Text>
      `,
      output: `
        <Text style={styles.foo}>${' ' /* intentional trailing space */}
{' '}
<Bar />
        </Text>
      `,
      errors: [expectMoveToNewLineError3([2, 35, 2, 42])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <Text style={styles.foo}>
          <Bar /> <Baz />
        </Text>
      `,
      output: `
        <Text style={styles.foo}>
          <Bar />${' ' /* intentional trailing space */}
{' '}
<Baz />
        </Text>
      `,
      errors: [expectMoveToNewLineError5([3, 19, 3, 26])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <Text style={styles.foo}>
          <Bar /> <Baz />
        </Text>
      `,
      output: `
        <Text style={styles.foo}>
          <Bar />${' ' /* intentional trailing space */}
{' '}
<Baz />
        </Text>
      `,
      errors: [expectMoveToNewLineError5([3, 19, 3, 26])],
      options: [{ allow: 'non-jsx' }],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <Text style={styles.foo}>
          { bar } { I18n.t('baz') }
        </Text>
      `,
      output: `
        <Text style={styles.foo}>
          { bar }${' ' /* intentional trailing space */}
{' '}
{ I18n.t('baz') }
        </Text>
      `,
      errors: [expectMoveToNewLineError4([3, 19, 3, 36])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          foo<input />
        </div>
      `,
      output: `
        <div>
          foo
<input />
        </div>
      `,
      errors: [expectMoveToNewLineError6([3, 14, 3, 23])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          {"foo"}<span />
        </div>
      `,
      output: `
        <div>
          {"foo"}
<span />
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`span` must be placed on a new line',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 26,
        },
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          foo <input />
        </div>
      `,
      output: `
        <div>
          foo${' ' /* intentional trailing space */}
{' '}
<input />
        </div>
      `,
      errors: [expectMoveToNewLineError6([3, 15, 3, 24])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          <input /> foo
        </div>
      `,
      output: `
        <div>
          <input />
{' '}
foo
</div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '` foo        ` must be placed on a new line',
          line: 3,
          column: 20,
          endLine: 4,
          endColumn: 9,
        },
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          <span /> <input />
        </div>
      `,
      output: `
        <div>
          <span />${' ' /* intentional trailing space */}
{' '}
<input />
        </div>
      `,
      errors: [expectMoveToNewLineError6([3, 20, 3, 29])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          <span />
        {' '}<input />
        </div>
      `,
      output: `
        <div>
          <span />
        {' '}
<input />
        </div>
      `,
      errors: [expectMoveToNewLineError6([4, 14, 4, 23])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          {"foo"} <input />
        </div>
      `,
      output: `
        <div>
          {"foo"}${' ' /* intentional trailing space */}
{' '}
<input />
        </div>
      `,
      errors: [expectMoveToNewLineError6([3, 19, 3, 28])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
          <input /> {"foo"}
        </div>
      `,
      output: `
        <div>
          <input />${' ' /* intentional trailing space */}
{' '}
{"foo"}
        </div>
      `,
      errors: [expectMoveToNewLineError1([3, 21, 3, 28])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
          <Foo></Foo><Bar></Bar>
        </App>
      `,
      output: `
        <App>
          <Foo></Foo>
<Bar></Bar>
        </App>
      `,
      errors: [expectMoveToNewLineError3([3, 22, 3, 33])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
        <Foo></Foo></App>
      `,
      output: `
        <App>
        <Foo></Foo>
</App>
      `,
      errors: [expectMoveToNewLineError7([3, 9, 3, 20])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App><Foo />
        </App>
      `,
      output: `
        <App>
<Foo />
        </App>
      `,
      errors: [expectMoveToNewLineError7([2, 14, 2, 21])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
        <Foo/></App>
      `,
      output: `
        <App>
        <Foo/>
</App>
      `,
      errors: [expectMoveToNewLineError7([3, 9, 3, 15])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App><Foo
        />
        </App>
      `,
      output: `
        <App>
<Foo
        />
        </App>
      `,
      errors: [expectMoveToNewLineError7([2, 14, 3, 11])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App
        >
        <Foo /></App>
      `,
      output: `
        <App
        >
        <Foo />
</App>
      `,
      errors: [expectMoveToNewLineError7([4, 9, 4, 16])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App
        >
        <Foo
        /></App>
      `,
      output: `
        <App
        >
        <Foo
        />
</App>
      `,
      errors: [expectMoveToNewLineError7([4, 9, 5, 11])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App
        ><Foo />
        </App>
      `,
      output: `
        <App
        >
<Foo />
        </App>
      `,
      errors: [expectMoveToNewLineError7([3, 10, 3, 17])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
          <Foo></Foo
        ></App>
      `,
      output: `
        <App>
          <Foo></Foo
        >
</App>
      `,
      errors: [expectMoveToNewLineError7([3, 11, 4, 10])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
          <Foo></
        Foo></App>
      `,
      output: `
        <App>
          <Foo></
        Foo>
</App>
      `,
      errors: [expectMoveToNewLineError7([3, 11, 4, 13])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
          <Foo></
        Foo><Bar />
        </App>
      `,
      output: `
        <App>
          <Foo></
        Foo>
<Bar />
        </App>
      `,
      errors: [expectMoveToNewLineError3([4, 13, 4, 20])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
          <Foo>
            <Bar /></Foo>
        </App>
      `,
      output: `
        <App>
          <Foo>
            <Bar />
</Foo>
        </App>
      `,
      errors: [expectMoveToNewLineError3([4, 13, 4, 20])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
          <Foo>
            <Bar> baz </Bar>
          </Foo>
        </App>
      `,
      output: `
        <App>
          <Foo>
            <Bar>
{' '}
baz
{' '}
</Bar>
          </Foo>
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '` baz ` must be placed on a new line',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 23,
        },
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      // Would be nice to handle in one pass, but multipass works fine.
      code: `
        <App>
          foo {"bar"} baz
        </App>
      `,
      // Assert the complete output after all autofix rounds.
      output: `
        <App>
          foo\x20
{' '}
{"bar"}
{' '}
baz
</App>
      `,
      errors: [
        expectMoveToNewLineError2([3, 15, 3, 22]),
        expectMoveToNewLineError8([3, 22, 4, 9]),
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      // Would be nice to handle in one pass, but multipass works fine.
      code: `
        <App>
          foo {"bar"}
        </App>
      `,
      output: `
        <App>
          foo${' ' /* intentional trailing space */}
{' '}
{"bar"}
        </App>
      `,
      errors: [expectMoveToNewLineError2([3, 15, 3, 22])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      // Would be nice to handle in one pass, but multipass works fine.
      code: `
        <App>
          foo
        {' '}
        {"bar"} baz
        </App>
      `,
      output: `
        <App>
          foo
        {' '}
        {"bar"}
{' '}
baz
</App>
      `,
      errors: [expectMoveToNewLineError8([5, 16, 6, 9])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      // Would be nice to handle in one pass, but multipass works fine.
      code: `
        <App>

          foo {"bar"} baz

        </App>
      `,
      // Assert the complete output after all autofix rounds.
      output: `
        <App>

          foo\x20
{' '}
{"bar"}
{' '}
baz

</App>
      `,
      errors: [
        expectMoveToNewLineError2([4, 15, 4, 22]),
        expectMoveToNewLineError8([4, 22, 6, 9]),
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      // Would be nice to handle in one pass, but multipass works fine.
      code: `
        <App>

          foo
        {' '}
        {"bar"} baz

        </App>
      `,
      output: `
        <App>

          foo
        {' '}
        {"bar"}
{' '}
baz

</App>
      `,
      errors: [expectMoveToNewLineError8([6, 16, 8, 9])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>{
          foo
        }</App>
      `,
      output: `
        <App>
{
          foo
        }
</App>
      `,
      errors: [expectMoveToNewLineError9([2, 14, 4, 10])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App> {
          foo
        } </App>
      `,
      output: `
        <App>${' ' /* intentional trailing space */}
{' '}
{
          foo
        }
{' '}
 </App>
      `,
      errors: [expectMoveToNewLineError9([2, 15, 4, 10])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
        {' '}
        {
          foo
        } </App>
      `,
      output: `
        <App>
        {' '}
        {
          foo
        }
{' '}
 </App>
      `,
      errors: [expectMoveToNewLineError9([4, 9, 6, 10])],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App><Foo /></App>
      `,
      output: `
        <App>
<Foo />
</App>
      `,
      options: [{ allow: 'none' }],
      errors: [expectMoveToNewLineError7([2, 14, 2, 21])],
    },
    {
      code: `
        <App>foo</App>
      `,
      output: `
        <App>
foo
</App>
      `,
      options: [{ allow: 'none' }],
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`foo` must be placed on a new line',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
        <App>{"foo"}</App>
      `,
      output: `
        <App>
{"foo"}
</App>
      `,
      options: [{ allow: 'none' }],
      errors: [expectMoveToNewLineError1([2, 14, 2, 21])],
    },
    {
      code: `
        <App>foo
        </App>
      `,
      output: `
        <App>
foo
</App>
      `,
      options: [{ allow: 'literal' }],
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`foo        ` must be placed on a new line',
          line: 2,
          column: 14,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
        <App><Foo /></App>
      `,
      output: `
        <App>
<Foo />
</App>
      `,
      options: [{ allow: 'literal' }],
      errors: [expectMoveToNewLineError7([2, 14, 2, 21])],
    },
    {
      code: `
        <App><Foo /></App>
      `,
      output: `
        <App>
<Foo />
</App>
      `,
      options: [{ allow: 'non-jsx' }],
      errors: [expectMoveToNewLineError7([2, 14, 2, 21])],
    },
    {
      code: `
        <App
          foo="1"
          bar="2"
        >baz</App>
      `,
      options: [{ allow: 'literal' }],
      output: `
        <App
          foo="1"
          bar="2"
        >
baz
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`baz` must be placed on a new line',
          line: 5,
          column: 10,
          endLine: 5,
          endColumn: 13,
        },
      ],
    },
    {
      code: `
        <App>foo
        bar
        </App>
      `,
      options: [{ allow: 'literal' }],
      output: `
        <App>
foo
        bar
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`foo        bar        ` must be placed on a new line',
          line: 2,
          column: 14,
          endLine: 4,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
        <>{"foo"}</>
      `,
      output: `
        <>
{"foo"}
</>
      `,
      errors: [expectMoveToNewLineError1([2, 11, 2, 18])],
      // TODO: FIXME: remove no-ts-old and fix
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <App>
          <Foo /><></>
        </App>
      `,
      output: `
        <App>
          <Foo />
<></>
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`<></>` must be placed on a new line',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 23,
        },
      ],

      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <
        ><Foo />
        </>
      `,
      output: `
        <
        >
<Foo />
        </>
      `,
      errors: [expectMoveToNewLineError7([3, 10, 3, 17])],

      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <div>
        <MyComponent>a</MyComponent>
        <MyOther>{a}</MyOther>
        </div>
      `,
      output: `
        <div>
        <MyComponent>
a
</MyComponent>
        <MyOther>
{a}
</MyOther>
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`a` must be placed on a new line',
          line: 3,
          column: 22,
          endLine: 3,
          endColumn: 23,
        },
        {
          messageId: 'moveToNewLine',
          message: '`{a}` must be placed on a new line',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 21,
        },
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      // TODO: handle in a single pass
      code: `
        const IndexPage = () => (
          <h1>{"Hi people"}<button/></h1>
        );
      `,
      // Assert the complete output after all autofix rounds.
      output: `
        const IndexPage = () => (
          <h1>
{"Hi people"}
<button/>
</h1>
        );
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`{"Hi people"}` must be placed on a new line',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 28,
        },
        expectMoveToNewLineError10([3, 28, 3, 37]),
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        const IndexPage = () => (
          <h1>
{"Hi people"}<button/></h1>
        );
      `,
      output: `
        const IndexPage = () => (
          <h1>
{"Hi people"}
<button/>
</h1>
        );
      `,
      errors: [expectMoveToNewLineError10([4, 14, 4, 23])],
      languageOptions: { parserOptions: parserOptions },
    },
    // TODO: handle in a single pass (see above)
    {
      code: `
        <Layout>
        <p>Welcome to your new Gatsby site.</p>
        <p>Now go build something great.</p>
        <h1>Hi people<button/></h1>
        </Layout>
      `,
      // Assert the complete output after all autofix rounds.
      output: `
        <Layout>
        <p>
Welcome to your new Gatsby site.
</p>
        <p>
Now go build something great.
</p>
        <h1>
Hi people
<button/>
</h1>
        </Layout>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message:
            '`Welcome to your new Gatsby site.` must be placed on a new line',
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 44,
        },
        {
          messageId: 'moveToNewLine',
          message:
            '`Now go build something great.` must be placed on a new line',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 41,
        },
        {
          messageId: 'moveToNewLine',
          message: '`Hi people` must be placed on a new line',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 22,
        },
        expectMoveToNewLineError10([5, 22, 5, 31]),
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <Layout>
        <p>
Welcome to your new Gatsby site.
</p>
        <p>
Now go build something great.
</p>
        <h1>
Hi people<button/></h1>
        </Layout>
      `,
      output: `
        <Layout>
        <p>
Welcome to your new Gatsby site.
</p>
        <p>
Now go build something great.
</p>
        <h1>
Hi people
<button/>
</h1>
        </Layout>
      `,
      errors: [expectMoveToNewLineError10([10, 10, 10, 19])],
      languageOptions: { parserOptions: parserOptions },
    },
    // TODO: handle in a single pass
    {
      code: `
        <Layout>
          <div style={{ maxWidth: \`300px\`, marginBottom: \`1.45rem\` }}>
            <Image />
          </div><Link to="/page-2/">Go to page 2</Link>
        </Layout>
      `,
      // Assert the complete output after all autofix rounds.
      output: `
        <Layout>
          <div style={{ maxWidth: \`300px\`, marginBottom: \`1.45rem\` }}>
            <Image />
          </div>
<Link to="/page-2/">
Go to page 2
</Link>
        </Layout>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`Link` must be placed on a new line',
          line: 5,
          column: 17,
          endLine: 5,
          endColumn: 56,
        },
        {
          messageId: 'moveToNewLine',
          message: '`Go to page 2` must be placed on a new line',
          line: 5,
          column: 37,
          endLine: 5,
          endColumn: 49,
        },
      ],
      languageOptions: { parserOptions: parserOptions },
    },
    {
      code: `
        <Layout>
          <div style={{ maxWidth: \`300px\`, marginBottom: \`1.45rem\` }}>
            <Image />
          </div>
<Link to="/page-2/">Go to page 2</Link>
        </Layout>
      `,
      output: `
        <Layout>
          <div style={{ maxWidth: \`300px\`, marginBottom: \`1.45rem\` }}>
            <Image />
          </div>
<Link to="/page-2/">
Go to page 2
</Link>
        </Layout>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          message: '`Go to page 2` must be placed on a new line',
          line: 6,
          column: 21,
          endLine: 6,
          endColumn: 33,
        },
      ],
      languageOptions: { parserOptions: parserOptions },
    },
  ],
});
