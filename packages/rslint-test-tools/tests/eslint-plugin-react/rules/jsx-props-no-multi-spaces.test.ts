import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectOnlyOneSpaceError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'onlyOneSpace',
    'Expected only one space between "foo" and "bar"',
    range,
  );
}

function expectNoLineGapError2(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'noLineGap',
    'Expected no line gap between "title" and "onClick"',
    range,
  );
}

function expectNoLineGapError3(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'noLineGap',
    'Expected no line gap between "onClick" and "type"',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-props-no-multi-spaces', {
  valid: [
    {
      code: `<Foo bar="baz" />`,
    },
    {
      code: `<Foo bar="baz" qux="quux" />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        <App />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        <App foo />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        <App foo bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        <App foo="with  spaces   " bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        <App
          foo bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        <App
          foo
          bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        <App
          foo {...test}
          bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `<App<T> foo bar />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<Foo.Bar baz="quux" />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `<Foobar.Foo.Bar.Baz.Qux.Quux.Quuz.Corge.Grault.Garply.Waldo.Fred.Plugh xyzzy="thud" />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        <button
          title="Some button"
          type="button"
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        <button
          title="Some button"
          onClick={(value) => {
            console.log(value);
          }}
          type="button"
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
          <button
            title="Some button"
            // this is a comment
            onClick={(value) => {
              console.log(value);
            }}
            type="button"
          />
        `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
          <button
            title="Some button"
            // this is a comment
            // this is a second comment
            onClick={(value) => {
              console.log(value);
            }}
            type="button"
          />
        `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
          <App
            foo="Some button" // comment
            // comment
            bar=""
          />
        `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
          <button
            title="Some button"
            /* this is a multiline comment
                ...
                ... */
            onClick={(value) => {
              console.log(value);
            }}
            type="button"
          />
        `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<Foo  bar="baz" />`,
      errors: [
        diagnostic(
          'onlyOneSpace',
          'Expected only one space between "Foo" and "bar"',
          [1, 7, 1, 16],
        ),
      ],
      output: `<Foo bar="baz" />`,
    },
    {
      code: `<Foo bar="baz"  qux="quux" />`,
      errors: [
        diagnostic(
          'onlyOneSpace',
          'Expected only one space between "bar" and "qux"',
          [1, 17, 1, 27],
        ),
      ],
      output: `<Foo bar="baz" qux="quux" />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        <App  foo />
      `,
      errors: [
        diagnostic(
          'onlyOneSpace',
          'Expected only one space between "App" and "foo"',
          [2, 15, 2, 18],
        ),
      ],
      output: `
        <App foo />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        <App foo="with  spaces   "   bar />
      `,
      errors: [expectOnlyOneSpaceError([2, 38, 2, 41])],
      output: `
        <App foo="with  spaces   " bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        <App foo  bar />
      `,
      errors: [expectOnlyOneSpaceError([2, 19, 2, 22])],
      output: `
        <App foo bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        <App  foo   bar />
      `,
      errors: [
        diagnostic(
          'onlyOneSpace',
          'Expected only one space between "App" and "foo"',
          [2, 15, 2, 18],
        ),
        expectOnlyOneSpaceError([2, 21, 2, 24]),
      ],
      output: `
        <App foo bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        <App foo  {...test}  bar />
      `,
      errors: [
        diagnostic(
          'onlyOneSpace',
          'Expected only one space between "foo" and "test"',
          [2, 19, 2, 28],
        ),
        diagnostic(
          'onlyOneSpace',
          'Expected only one space between "test" and "bar"',
          [2, 30, 2, 33],
        ),
      ],
      output: `
        <App foo {...test} bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `<Foo.Bar  baz="quux" />`,
      errors: [
        diagnostic(
          'onlyOneSpace',
          'Expected only one space between "Foo.Bar" and "baz"',
          [1, 11, 1, 21],
        ),
      ],
      output: `<Foo.Bar baz="quux" />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        <Foobar.Foo.Bar.Baz.Qux.Quux.Quuz.Corge.Grault.Garply.Waldo.Fred.Plugh  xyzzy="thud" />
      `,
      errors: [
        diagnostic(
          'onlyOneSpace',
          'Expected only one space between "Foobar.Foo.Bar.Baz.Qux.Quux.Quuz.Corge.Grault.Garply.Waldo.Fred.Plugh" and "xyzzy"',
          [2, 81, 2, 93],
        ),
      ],
      output: `
        <Foobar.Foo.Bar.Baz.Qux.Quux.Quuz.Corge.Grault.Garply.Waldo.Fred.Plugh xyzzy="thud" />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        <button
          title='Some button'

          type="button"
        />
      `,
      errors: [
        diagnostic(
          'noLineGap',
          'Expected no line gap between "title" and "type"',
          [5, 11, 5, 24],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        <button
          title="Some button"

          onClick={(value) => {
            console.log(value);
          }}

          type="button"
        />
      `,
      errors: [
        expectNoLineGapError2([5, 11, 7, 13]),
        expectNoLineGapError3([9, 11, 9, 24]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
          <button
            title="Some button"
            // this is a comment
            onClick={(value) => {
              console.log(value);
            }}

            type="button"
          />
        `,
      errors: [expectNoLineGapError3([9, 13, 9, 26])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
          <button
            title="Some button"
            // this is a comment
            // second comment

            onClick={(value) => {
              console.log(value);
            }}

            type="button"
          />
        `,
      errors: [
        expectNoLineGapError2([7, 13, 9, 15]),
        expectNoLineGapError3([11, 13, 11, 26]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
          <button
            title="Some button"
            /*this is a
              multiline
              comment
            */

            onClick={(value) => {
              console.log(value);
            }}

            type="button"
          />
        `,
      errors: [
        expectNoLineGapError2([9, 13, 11, 15]),
        expectNoLineGapError3([13, 13, 13, 26]),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
