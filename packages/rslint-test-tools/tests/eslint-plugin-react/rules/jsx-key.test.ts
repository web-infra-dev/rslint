import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-key', {} as never, {
  valid: [
    // Keyed map callback — no report.
    {
      code: `
        [1, 2, 3].map((item) => {
          return item === 'bar' ? <div key={item}>{item}</div> : <span key={item}>{item}</span>;
        });
      `,
    },
    { code: 'fn();' },
    { code: '[1, 2, 3].map(function () {});' },
    { code: '<App />;' },
    { code: '[<App key={0} />, <App key={1} />];' },
    { code: '[1, 2, 3].map(function (x) { return <App key={x} /> });' },
    { code: '[1, 2, 3].map(x => <App key={x} />);' },
    { code: '[1, 2, 3].map(x => x && <App x={x} key={x} />);' },
    {
      code: '[1, 2, 3].map(x => x ? <App x={x} key="1" /> : <OtherApp x={x} key="2" />);',
    },
    { code: '[1, 2, 3].map(x => { return <App key={x} /> });' },
    { code: 'Array.from([1, 2, 3], function (x) { return <App key={x} /> });' },
    { code: 'Array.from([1, 2, 3], (x => <App key={x} />));' },
    { code: 'Array.from([1, 2, 3], (x => { return <App key={x} /> }));' },
    { code: 'Array.from([1, 2, 3], someFn);' },
    { code: 'Array.from([1, 2, 3]);' },
    // .foo() is not .map() — ignored.
    { code: '[1, 2, 3].foo(x => <App />);' },
    { code: 'var App = () => <div />;' },
    { code: '[1, 2, 3].map(function (x) { return; });' },
    { code: 'foo(() => <div />);' },
    { code: 'foo(() => <></>);' },
    { code: '<></>;' },
    { code: '<App {...{}} />;' },
    // Key before spread on a JSX element — not an array / iterator context,
    // so checkKeyMustBeforeSpread doesn't kick in.
    {
      code: '<App key="keyBeforeSpread" {...{}} />;',
      options: [{ checkKeyMustBeforeSpread: true }],
    },
    {
      code: '<div key="keyBeforeSpread" {...{}} />;',
      options: [{ checkKeyMustBeforeSpread: true }],
    },
    // warnOnDuplicates default (false) — duplicate keys allowed.
    {
      code: `
        const spans = [
          <span key="notunique" />,
          <span key="notunique" />,
        ];
      `,
    },
    // React.Children.toArray — rule goes dormant inside.
    { code: 'React.Children.toArray([1, 2, 3].map(x => <App />));' },
    // Destructured Children.toArray — also dormant.
    {
      code: `
        import { Children } from "react";
        Children.toArray([1, 2, 3].map(x => <App />));
      `,
    },
    // Logical returns with keyed JSX.
    { code: '[1, 2, 3].map(x => { return x && <App key={x} />; });' },
    { code: '[1, 2, 3].map(x => { return x && y && <App key={x} />; });' },
    // Logical returns where right-hand side is not JSX.
    { code: '[1, 2, 3].map(x => { return x && foo(); });' },
    // Nested JSX without keys — direct JSX children of JSX don't need keys
    // according to this rule.
    { code: '<ul><li /><li /></ul>;' },
    // Optional chaining on the map call itself.
    { code: 'xs?.map(x => <App key={x} />);' },
    { code: 'xs.map?.(x => <App key={x} />);' },
  ],
  invalid: [
    // Ternary inside map callback — two missing iter keys.
    {
      code: `
        [1, 2, 3].map((item) => {
          return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
        });
      `,
      errors: [
        { message: 'Missing "key" prop for element in iterator' },
        { message: 'Missing "key" prop for element in iterator' },
      ],
    },
    // FunctionExpression body variant.
    {
      code: `
        [1, 2, 3].map(function (item) {
          return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
        });
      `,
      errors: [
        { message: 'Missing "key" prop for element in iterator' },
        { message: 'Missing "key" prop for element in iterator' },
      ],
    },
    // Array.from callback.
    {
      code: `
        Array.from([1, 2, 3], (item) => {
          return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
        });
      `,
      errors: [
        { message: 'Missing "key" prop for element in iterator' },
        { message: 'Missing "key" prop for element in iterator' },
      ],
    },
    // Array literal — missing key on array member.
    {
      code: '[<App />];',
      errors: [{ message: 'Missing "key" prop for element in array' }],
    },
    // Spread-providing-key still counts as missing array key.
    {
      code: '[<App {...key} />];',
      errors: [{ message: 'Missing "key" prop for element in array' }],
    },
    // Mixed: first has key, second doesn't.
    {
      code: '[<App key={0} />, <App />];',
      errors: [{ message: 'Missing "key" prop for element in array' }],
    },
    // Plain map with FunctionExpression.
    {
      code: '[1, 2, 3].map(function (x) { return <App /> });',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
    // Plain map with arrow.
    {
      code: '[1, 2, 3].map(x => <App />);',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
    // Logical-right JSX inside map.
    {
      code: '[1, 2, 3].map(x => x && <App x={x} />);',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
    // Ternary with one keyed and one unkeyed.
    {
      code: '[1, 2, 3].map(x => x ? <App x={x} key="1" /> : <OtherApp x={x} />);',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
    // Array.from with arrow wrapped in parens.
    {
      code: 'Array.from([1, 2, 3], (x => <App />));',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
    // Optional chain call.
    {
      code: '[1, 2, 3]?.map(x => <App />);',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
    // Optional member access.
    {
      code: 'xs?.map(x => <App />);',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
    // Optional call.
    {
      code: '[1, 2, 3].map?.(x => <App />);',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
    // Fragment shorthand in iterator (option on).
    {
      code: '[1, 2, 3].map(x => <>{x}</>);',
      options: [{ checkFragmentShorthand: true }],
      errors: [
        {
          message:
            'Missing "key" prop for element in iterator. Shorthand fragment syntax does not support providing keys. Use React.Fragment instead',
        },
      ],
    },
    // Fragment shorthand in array literal (option on).
    {
      code: '[<></>];',
      options: [{ checkFragmentShorthand: true }],
      errors: [
        {
          message:
            'Missing "key" prop for element in array. Shorthand fragment syntax does not support providing keys. Use React.Fragment instead',
        },
      ],
    },
    // checkKeyMustBeforeSpread — key after spread in array context.
    {
      code: '[<App {...obj} key="keyAfterSpread" />];',
      options: [{ checkKeyMustBeforeSpread: true }],
      errors: [
        {
          message:
            '`key` prop must be placed before any `{...spread}, to avoid conflicting with React\u2019s new JSX transform: https://reactjs.org/blog/2020/09/22/introducing-the-new-jsx-transform.html`',
        },
      ],
    },
    // warnOnDuplicates — two identical keys.
    {
      code: `
        const spans = [
          <span key="notunique" />,
          <span key="notunique" />,
        ];
      `,
      options: [{ warnOnDuplicates: true }],
      errors: [
        { message: '`key` prop must be unique' },
        { message: '`key` prop must be unique' },
      ],
    },
    // warnOnDuplicates inside JSX parent.
    {
      code: `
        const div = (
          <div>
            <span key="notunique" />
            <span key="notunique" />
          </div>
        );
      `,
      options: [{ warnOnDuplicates: true }],
      errors: [
        { message: '`key` prop must be unique' },
        { message: '`key` prop must be unique' },
      ],
    },
    // if/else if/else with missing keys.
    {
      code: `
        const TestCase = () => {
          const list = [1, 2, 3, 4, 5];

          return (
            <div>
              {list.map(item => {
                if (item < 2) return <div>{item}</div>;
                else if (item < 5) return <div />;
                else return <div />;
              })}
            </div>
          );
        };
      `,
      errors: [
        { message: 'Missing "key" prop for element in iterator' },
        { message: 'Missing "key" prop for element in iterator' },
        { message: 'Missing "key" prop for element in iterator' },
      ],
    },
    // checkKeyMustBeforeSpread inside a map callback — reports on the
    // element itself, not the containing array.
    {
      code: `
        const TestCase = () => {
          const list = [1, 2, 3, 4, 5];

          return (
            <div>
              {list.map(x => <div {...spread} key={x} />)}
            </div>
          );
        };
      `,
      options: [{ checkKeyMustBeforeSpread: true }],
      errors: [
        {
          message:
            '`key` prop must be placed before any `{...spread}, to avoid conflicting with React\u2019s new JSX transform: https://reactjs.org/blog/2020/09/22/introducing-the-new-jsx-transform.html`',
        },
      ],
    },
    // Logical / disjunction in return argument.
    {
      code: '[1, 2, 3].map(x => { return x && <App />; });',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
    {
      code: '[1, 2, 3].map(x => { return x || y || <App />; });',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
    // Parenthesized JSX body.
    {
      code: '[1,2,3].map(x => (<App />));',
      errors: [{ message: 'Missing "key" prop for element in iterator' }],
    },
  ],
});
