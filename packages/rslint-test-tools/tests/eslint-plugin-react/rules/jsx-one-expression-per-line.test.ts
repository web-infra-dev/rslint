import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
import { diagnostic } from '../expectations';

const parserOptions = {
  ecmaVersion: 2018,
  sourceType: 'module',
  ecmaFeatures: {
    jsx: true,
  },
};

function expectMoveToNewLineError1(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'moveToNewLine',
    '`{"foo"}` must be placed on a new line',
    range,
  );
}

function expectMoveToNewLineError2(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'moveToNewLine',
    '`{"bar"}` must be placed on a new line',
    range,
  );
}

function expectMoveToNewLineError3(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'moveToNewLine',
    '`Bar` must be placed on a new line',
    range,
  );
}

function expectMoveToNewLineError4(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'moveToNewLine',
    "`{ I18n.t('baz') }` must be placed on a new line",
    range,
  );
}

function expectMoveToNewLineError5(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'moveToNewLine',
    '`Baz` must be placed on a new line',
    range,
  );
}

function expectMoveToNewLineError6(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'moveToNewLine',
    '`input` must be placed on a new line',
    range,
  );
}

function expectMoveToNewLineError7(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'moveToNewLine',
    '`Foo` must be placed on a new line',
    range,
  );
}

function expectMoveToNewLineError8(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'moveToNewLine',
    '` baz        ` must be placed on a new line',
    range,
  );
}

function expectMoveToNewLineError9(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'moveToNewLine',
    '`{          foo        }` must be placed on a new line',
    range,
  );
}

function expectMoveToNewLineError10(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'moveToNewLine',
    '`button` must be placed on a new line',
    range,
  );
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
        diagnostic(
          'moveToNewLine',
          '`foo` must be placed on a new line',
          [2, 14, 2, 17],
        ),
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
        diagnostic(
          'moveToNewLine',
          '` bar        ` must be placed on a new line',
          [3, 18, 4, 9],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`foo        ` must be placed on a new line',
          [3, 19, 4, 9],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`{ bar }` must be placed on a new line',
          [2, 34, 2, 41],
        ),
        diagnostic(
          'moveToNewLine',
          '`Text` must be placed on a new line',
          [2, 42, 2, 49],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`Bunk` must be placed on a new line',
          [2, 49, 2, 56],
        ),
        diagnostic(
          'moveToNewLine',
          '`Bruno` must be placed on a new line',
          [2, 57, 2, 65],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`span` must be placed on a new line',
          [3, 18, 3, 26],
        ),
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
        diagnostic(
          'moveToNewLine',
          '` foo        ` must be placed on a new line',
          [3, 20, 4, 9],
        ),
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
        diagnostic(
          'moveToNewLine',
          '` baz ` must be placed on a new line',
          [4, 18, 4, 23],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`foo` must be placed on a new line',
          [2, 14, 2, 17],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`foo        ` must be placed on a new line',
          [2, 14, 3, 9],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`baz` must be placed on a new line',
          [5, 10, 5, 13],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`foo        bar        ` must be placed on a new line',
          [2, 14, 4, 9],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`<></>` must be placed on a new line',
          [3, 18, 3, 23],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`a` must be placed on a new line',
          [3, 22, 3, 23],
        ),
        diagnostic(
          'moveToNewLine',
          '`{a}` must be placed on a new line',
          [4, 18, 4, 21],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`{"Hi people"}` must be placed on a new line',
          [3, 15, 3, 28],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`Welcome to your new Gatsby site.` must be placed on a new line',
          [3, 12, 3, 44],
        ),
        diagnostic(
          'moveToNewLine',
          '`Now go build something great.` must be placed on a new line',
          [4, 12, 4, 41],
        ),
        diagnostic(
          'moveToNewLine',
          '`Hi people` must be placed on a new line',
          [5, 13, 5, 22],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`Link` must be placed on a new line',
          [5, 17, 5, 56],
        ),
        diagnostic(
          'moveToNewLine',
          '`Go to page 2` must be placed on a new line',
          [5, 37, 5, 49],
        ),
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
        diagnostic(
          'moveToNewLine',
          '`Go to page 2` must be placed on a new line',
          [6, 21, 6, 33],
        ),
      ],
      languageOptions: { parserOptions: parserOptions },
    },
  ],
});
