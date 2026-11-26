import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectMissingIterKeyError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingIterKey',
    message: 'Missing "key" prop for element in iterator',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectMissingArrayKeyError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingArrayKey',
    message: 'Missing "key" prop for element in array',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectKeyBeforeSpreadError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'keyBeforeSpread',
    message:
      '`key` prop must be placed before any `{...spread}, to avoid conflicting with React’s new JSX transform: https://reactjs.org/blog/2020/09/22/introducing-the-new-jsx-transform.html`',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNonUniqueKeysError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'nonUniqueKeys',
    message: '`key` prop must be unique',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-key', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `fn()`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `[1, 2, 3].map(function () {})`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `[1, 2, 3].map(function(x) { return <App key={x} /> });`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `[1, 2 ,3].map(x => x && <App x={x} key={x} />);`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `[1, 2 ,3].map(x => x ? <App x={x} key="1" /> : <OtherApp x={x} key="2" />);`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `Array.from([1, 2, 3], function(x) { return <App key={x} /> });`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `Array.from([1, 2, 3], (x => {return <App key={x} />}));`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `[1, 2, 3].map(function(x) { return; });`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        const spans = [
          <span key="notunique"/>,
          <span key="notunique"/>,
        ];
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        function Component(props) {
          return hasPayment ? (
            <div className="stuff">
              <BookingDetailSomething {...props} />
              {props.modal && props.calculatedPrice && (
                <SomeOtherThing items={props.something} discount={props.discount} />
              )}
            </div>
          ) : null;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        import React, { FC, useRef, useState } from 'react';

        import './ResourceVideo.sass';
        import VimeoVideoPlayInModal from '../vimeoVideoPlayInModal/VimeoVideoPlayInModal';

        type Props = {
          videoUrl: string;
          videoTitle: string;
        };
        const ResourceVideo: FC<Props> = ({
          videoUrl,
          videoTitle,
        }: Props): JSX.Element => {
          return (
            <div className="resource-video">
              <VimeoVideoPlayInModal videoUrl={videoUrl} />
              <h3>{videoTitle}</h3>
            </div>
          );
        };

        export default ResourceVideo;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        // testrule.jsx
        const trackLink = () => {};
        const getAnalyticsUiElement = () => {};

        const onTextButtonClick = (e, item) => trackLink([, getAnalyticsUiElement(item), item.name], e);
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        function Component({ allRatings }) {
          return (
            <RatingDetailsStyles>
              {Object.entries(allRatings)?.map(([key, value], index) => {
                const rate = value?.split(/(?=[%, /])/);

                if (!rate) return null;

                return (
                  <li key={\`\${entertainment.tmdbId}\${index}\`}>
                    <img src={\`/assets/rating/\${key}.png\`} />
                    <span className="rating-details--rate">{rate?.[0]}</span>
                    <span className="rating-details--rate-suffix">{rate?.[1]}</span>
                  </li>
                );
              })}
            </RatingDetailsStyles>
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        const baz = foo?.bar?.()?.[1] ?? 'qux';

        qux()?.map()

        const directiveRanges = comments?.map(tryParseTSDirective)
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        import { observable } from "mobx";

        export interface ClusterFrameInfo {
          frameId: number;
          processId: number;
        }

        export const clusterFrameMap = observable.map<string, ClusterFrameInfo>();
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `React.Children.toArray([1, 2 ,3].map(x => <App />));`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        import { Children } from "react";
        Children.toArray([1, 2 ,3].map(x => <App />));
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        import Act from 'react';
        import { Children as ReactChildren } from 'react';

        const { Children } = Act;
        const { toArray } = Children;

        Act.Children.toArray([1, 2 ,3].map(x => <App />));
        Act.Children.toArray(Array.from([1, 2 ,3], x => <App />));
        Children.toArray([1, 2 ,3].map(x => <App />));
        Children.toArray(Array.from([1, 2 ,3], x => <App />));
        // ReactChildren.toArray([1, 2 ,3].map(x => <App />));
        // ReactChildren.toArray(Array.from([1, 2 ,3], x => <App />));
        // toArray([1, 2 ,3].map(x => <App />));
        // toArray(Array.from([1, 2 ,3], x => <App />));
      `,
      settings: { react: { pragma: 'Act', fragment: 'Frag' } },
      languageOptions: upstreamModuleOptions,
    },
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
        expectMissingIterKeyError([3, 35, 3, 52]),
        expectMissingIterKeyError([3, 55, 3, 74]),
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
        expectMissingIterKeyError([3, 35, 3, 52]),
        expectMissingIterKeyError([3, 55, 3, 74]),
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
        expectMissingIterKeyError([3, 35, 3, 52]),
        expectMissingIterKeyError([3, 55, 3, 74]),
      ],
    },
    // Array literal — missing key on array member.
    {
      code: '[<App />];',
      errors: [expectMissingArrayKeyError([1, 2, 1, 9])],
    },
    // Spread-providing-key still counts as missing array key.
    {
      code: '[<App {...key} />];',
      errors: [expectMissingArrayKeyError([1, 2, 1, 18])],
    },
    // Mixed: first has key, second doesn't.
    {
      code: '[<App key={0} />, <App />];',
      errors: [expectMissingArrayKeyError([1, 19, 1, 26])],
    },
    // Plain map with FunctionExpression.
    {
      code: '[1, 2, 3].map(function (x) { return <App /> });',
      errors: [expectMissingIterKeyError([1, 37, 1, 44])],
    },
    // Plain map with arrow.
    {
      code: '[1, 2, 3].map(x => <App />);',
      errors: [expectMissingIterKeyError([1, 20, 1, 27])],
    },
    // Logical-right JSX inside map.
    {
      code: '[1, 2, 3].map(x => x && <App x={x} />);',
      errors: [expectMissingIterKeyError([1, 25, 1, 38])],
    },
    // Ternary with one keyed and one unkeyed.
    {
      code: '[1, 2, 3].map(x => x ? <App x={x} key="1" /> : <OtherApp x={x} />);',
      errors: [expectMissingIterKeyError([1, 48, 1, 66])],
    },
    // Array.from with arrow wrapped in parens.
    {
      code: 'Array.from([1, 2, 3], (x => <App />));',
      errors: [expectMissingIterKeyError([1, 29, 1, 36])],
    },
    // Optional chain call.
    {
      code: '[1, 2, 3]?.map(x => <App />);',
      errors: [expectMissingIterKeyError([1, 21, 1, 28])],
    },
    // Optional member access.
    {
      code: 'xs?.map(x => <App />);',
      errors: [expectMissingIterKeyError([1, 14, 1, 21])],
    },
    // Optional call.
    {
      code: '[1, 2, 3].map?.(x => <App />);',
      errors: [expectMissingIterKeyError([1, 22, 1, 29])],
    },
    // Fragment shorthand in iterator (option on).
    {
      code: '[1, 2, 3].map(x => <>{x}</>);',
      options: [{ checkFragmentShorthand: true }],
      errors: [
        {
          messageId: 'missingIterKeyUsePrag',
          message:
            'Missing "key" prop for element in iterator. Shorthand fragment syntax does not support providing keys. Use React.Fragment instead',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    // Fragment shorthand in array literal (option on).
    {
      code: '[<></>];',
      options: [{ checkFragmentShorthand: true }],
      errors: [
        {
          messageId: 'missingArrayKeyUsePrag',
          message:
            'Missing "key" prop for element in array. Shorthand fragment syntax does not support providing keys. Use React.Fragment instead',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    // checkKeyMustBeforeSpread — key after spread in array context.
    {
      code: '[<App {...obj} key="keyAfterSpread" />];',
      options: [{ checkKeyMustBeforeSpread: true }],
      errors: [expectKeyBeforeSpreadError([1, 1, 1, 40])],
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
        expectNonUniqueKeysError([3, 17, 3, 32]),
        expectNonUniqueKeysError([4, 17, 4, 32]),
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
        expectNonUniqueKeysError([4, 19, 4, 34]),
        expectNonUniqueKeysError([5, 19, 5, 34]),
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
        expectMissingIterKeyError([8, 38, 8, 55]),
        expectMissingIterKeyError([9, 43, 9, 50]),
        expectMissingIterKeyError([10, 29, 10, 36]),
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
      errors: [expectKeyBeforeSpreadError([7, 30, 7, 57])],
    },
    // Logical / disjunction in return argument.
    {
      code: '[1, 2, 3].map(x => { return x && <App />; });',
      errors: [expectMissingIterKeyError([1, 34, 1, 41])],
    },
    {
      code: '[1, 2, 3].map(x => { return x || y || <App />; });',
      errors: [expectMissingIterKeyError([1, 39, 1, 46])],
    },
    // Parenthesized JSX body.
    {
      code: '[1,2,3].map(x => (<App />));',
      errors: [expectMissingIterKeyError([1, 19, 1, 26])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `[<App key={0}/>, <App />];`,
      errors: [expectMissingArrayKeyError([1, 18, 1, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `[1, 2 ,3].map(function(x) { return <App /> });`,
      errors: [expectMissingIterKeyError([1, 36, 1, 43])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `[1, 2 ,3].map(x => <App />);`,
      errors: [expectMissingIterKeyError([1, 20, 1, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `[1, 2 ,3].map(x => x && <App x={x} />);`,
      errors: [expectMissingIterKeyError([1, 25, 1, 38])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `[1, 2 ,3].map(x => x ? <App x={x} key="1" /> : <OtherApp x={x} />);`,
      errors: [expectMissingIterKeyError([1, 48, 1, 66])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `[1, 2 ,3].map(x => x ? <App x={x} /> : <OtherApp x={x} key="2" />);`,
      errors: [expectMissingIterKeyError([1, 24, 1, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `[1, 2 ,3].map(x => { return <App /> });`,
      errors: [expectMissingIterKeyError([1, 29, 1, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `Array.from([1, 2 ,3], function(x) { return <App /> });`,
      errors: [expectMissingIterKeyError([1, 44, 1, 51])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `Array.from([1, 2 ,3], (x => { return <App /> }));`,
      errors: [expectMissingIterKeyError([1, 38, 1, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `Array.from([1, 2 ,3], (x => <App />));`,
      errors: [expectMissingIterKeyError([1, 29, 1, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `[1, 2, 3]?.map(x => <BabelEslintApp />)`,
      errors: [expectMissingIterKeyError([1, 21, 1, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `[1, 2, 3]?.map(x => <TypescriptEslintApp />)`,
      errors: [expectMissingIterKeyError([1, 21, 1, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `[1, 2, 3].map(x => <>{x}</>);`,
      options: [{ checkFragmentShorthand: true }],
      settings: { react: { pragma: 'Act', fragment: 'Frag' } },
      errors: [
        {
          messageId: 'missingIterKeyUsePrag',
          message:
            'Missing "key" prop for element in iterator. Shorthand fragment syntax does not support providing keys. Use Act.Frag instead',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 28,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `[<></>];`,
      options: [{ checkFragmentShorthand: true }],
      settings: { react: { pragma: 'Act', fragment: 'Frag' } },
      errors: [
        {
          messageId: 'missingArrayKeyUsePrag',
          message:
            'Missing "key" prop for element in array. Shorthand fragment syntax does not support providing keys. Use Act.Frag instead',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 7,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `[<App {...obj} key="keyAfterSpread" />];`,
      options: [{ checkKeyMustBeforeSpread: true }],
      settings: { react: { pragma: 'Act', fragment: 'Frag' } },
      errors: [expectKeyBeforeSpreadError([1, 1, 1, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `[<div {...obj} key="keyAfterSpread" />];`,
      options: [{ checkKeyMustBeforeSpread: true }],
      settings: { react: { pragma: 'Act', fragment: 'Frag' } },
      errors: [expectKeyBeforeSpreadError([1, 1, 1, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        const spans = [
          <span key="notunique"/>,
          <span key="notunique"/>,
        ];
      `,
      options: [{ warnOnDuplicates: true }],
      errors: [
        expectNonUniqueKeysError([3, 17, 3, 32]),
        expectNonUniqueKeysError([4, 17, 4, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        const div = (
          <div>
            <span key="notunique"/>
            <span key="notunique"/>
          </div>
        );
      `,
      options: [{ warnOnDuplicates: true }],
      errors: [
        expectNonUniqueKeysError([4, 19, 4, 34]),
        expectNonUniqueKeysError([5, 19, 5, 34]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        const Test = () => {
          const list = [1, 2, 3, 4, 5];

          return (
            <div>
              {list.map(item => {
                if (item < 2) {
                  return <div>{item}</div>;
                }

                return <div />;
              })}
            </div>
          );
        };
      `,
      errors: [
        expectMissingIterKeyError([9, 26, 9, 43]),
        expectMissingIterKeyError([12, 24, 12, 31]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        const TestO = () => {
          const list = [1, 2, 3, 4, 5];

          return (
            <div>
              {list.map(item => {
                if (item < 2) {
                  return <div>{item}</div>;
                } else if (item < 5) {
                  return <div></div>
                }  else {
                  return <div></div>
                }

                return <div />;
              })}
            </div>
          );
        };
      `,
      errors: [
        expectMissingIterKeyError([9, 26, 9, 43]),
        expectMissingIterKeyError([11, 26, 11, 37]),
        expectMissingIterKeyError([13, 26, 13, 37]),
        expectMissingIterKeyError([16, 24, 16, 31]),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
