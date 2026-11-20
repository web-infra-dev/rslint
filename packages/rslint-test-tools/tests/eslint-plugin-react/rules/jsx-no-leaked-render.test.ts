import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNoPotentialLeakedRenderError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noPotentialLeakedRender',
    message:
      'Potential leaked value that might cause unintentionally rendered values or rendering crashes',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-no-leaked-render', {
  valid: [
    // ---- Default options (validStrategies = [ternary, coerce]) ----
    { code: `const C = () => <div>{customTitle || defaultTitle}</div>` },
    { code: `const C = ({ elements }) => <div>{elements}</div>` },
    {
      code: `const C = ({ elements }) => <div>There are {elements.length} elements</div>`,
    },
    {
      code: `const C = ({ count }) => <div>{!count && 'No results found'}</div>`,
    },
    {
      code: `const C = ({ elements }) => <div>{!!elements.length && <List/>}</div>`,
    },
    {
      code: `const C = ({ elements }) => <div>{Boolean(elements.length) && <List/>}</div>`,
    },
    {
      code: `const C = ({ elements }) => <div>{elements.length > 0 && <List/>}</div>`,
    },
    {
      code: `const C = ({ elements }) => <div>{elements.length ? <List/> : null}</div>`,
    },

    // ---- coerce-only ----
    {
      code: `const C = ({ count }) => <div>{!!count && <List/>}</div>`,
      options: [{ validStrategies: ['coerce'] }],
    },
    {
      code: `const C = ({ direction }) => <div>{!!direction && direction === "down" && "▼"}</div>`,
      options: [{ validStrategies: ['coerce'] }],
    },

    // ---- ternary-only ----
    {
      code: `const C = ({ count }) => <div>{count ? <List/> : null}</div>`,
      options: [{ validStrategies: ['ternary'] }],
    },

    // ---- ignoreAttributes mutes attribute-only reports ----
    {
      code: `const C = ({ enabled, checked }) => <CheckBox checked={enabled && checked} />`,
      options: [{ ignoreAttributes: true }],
    },

    // ---- Default react version (latest) → React 18+ behavior:
    //      '' empty string is safe on the left of `&&` ----
    { code: `const C = () => <>{'' && <Foo/>}</>` },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        const Component = () => {
          return <div>{customTitle || defaultTitle}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{elements}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>There are {elements.length} elements</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        const Component = ({ elements, count }) => {
          return <div>{!count && 'No results found'}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{!!elements.length && <List elements={elements}/>}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{Boolean(elements.length) && <List elements={elements}/>}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{elements.length > 0 && <List elements={elements}/>}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{elements.length ? <List elements={elements}/> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        const Component = ({ elements, count }) => {
          return <div>{count ? <List elements={elements}/> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        const Component = ({ elements, count }) => {
          return <div>{count ? <List elements={elements}/> : null}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        const Component = ({ elements, count }) => {
          return <div>{!!count && <List elements={elements}/>}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        const Component = ({ elements, count }) => {
          return <div>{count ? <List elements={elements}/> : null}</div>
        }
      `,
      options: [{ validStrategies: ['coerce', 'ternary'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        const Component = ({ elements, count }) => {
          return <div>{!!count && <List elements={elements}/>}</div>
        }
      `,
      options: [{ validStrategies: ['coerce', 'ternary'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        const Component = ({ elements, count }) => {
          return (
            <div>
              <div> {direction ? (direction === "down" ? "▼" : "▲") : ""} </div>
              <div>{ containerName.length > 0 ? "Loading several stuff" : "Loading" }</div>
            </div>
          )
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        const Component = ({ elements, count }) => {
          return <div>{direction ? (direction === "down" ? "▼" : "▲") : ""}</div>
        }
      `,
      options: [{ validStrategies: ['coerce', 'ternary'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        const Component = ({ direction }) => {
          return (
            <div>
              <div>{!!direction && direction === "down" && "▼"}</div>
              <div>{direction === "down" && !!direction && "▼"}</div>
              <div>{direction === "down" || !!direction && "▼"}</div>
              <div>{(!display || display === DISPLAY.WELCOME) && <span>foo</span>}</div>
            </div>
          )
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        const Component = ({ elements, count }) => {
          return <div>{count ? <List elements={elements}/> : <EmptyList />}</div>
        }
      `,
      options: [{ validStrategies: ['coerce', 'ternary'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        const Component = ({ elements, count }) => {
          return <div>{count ? <List elements={elements}/> : <EmptyList />}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        const isOpen = true;
        const Component = () => {
          return <Popover open={isOpen && items.length > 0} />
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        const isOpen = false;
        const Component = () => {
          return <Popover open={isOpen && items.length > 0} />
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Default: ternary fix ----
    {
      code: `const C = ({ count, title }) => <div>{count && title}</div>`,
      errors: [expectNoPotentialLeakedRenderError([1, 39, 1, 53])],
      output: `const C = ({ count, title }) => <div>{count ? title : null}</div>`,
    },
    {
      code: `const C = ({ count }) => <div>{count && <span>{count}</span>}</div>`,
      errors: [expectNoPotentialLeakedRenderError([1, 32, 1, 61])],
      output: `const C = ({ count }) => <div>{count ? <span>{count}</span> : null}</div>`,
    },
    {
      code: `const C = ({ elements }) => <div>{elements.length && <List/>}</div>`,
      errors: [expectNoPotentialLeakedRenderError([1, 35, 1, 61])],
      output: `const C = ({ elements }) => <div>{elements.length ? <List/> : null}</div>`,
    },
    {
      code: `const C = ({ a, b }) => <div>{(a || b) && <Results/>}</div>`,
      errors: [expectNoPotentialLeakedRenderError([1, 31, 1, 53])],
      output: `const C = ({ a, b }) => <div>{(a || b) ? <Results/> : null}</div>`,
    },

    // ---- coerce-only ----
    {
      code: `const C = ({ count, title }) => <div>{count && title}</div>`,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([1, 39, 1, 53])],
      output: `const C = ({ count, title }) => <div>{!!count && title}</div>`,
    },
    {
      code: `const C = ({ a, b }) => <div>{(a || b) && <Results/>}</div>`,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([1, 31, 1, 53])],
      output: `const C = ({ a, b }) => <div>{!!(a || b) && <Results/>}</div>`,
    },
    // ternary not allowed under coerce-only.
    {
      code: `const C = ({ count, title }) => <div>{count ? title : null}</div>`,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([1, 39, 1, 59])],
      output: `const C = ({ count, title }) => <div>{!!count && title}</div>`,
    },
    // Inverse ternary `cond ? false : alt`.
    {
      code: `const C = () => <Something checked={isIndeterminate ? false : isChecked} />`,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([1, 37, 1, 72])],
      output: `const C = () => <Something checked={!isIndeterminate && isChecked} />`,
    },

    // ---- ternary-only — boolean coercion on the left is not a free pass ----
    {
      code: `const C = ({ count, title }) => <div>{!!count && title}</div>`,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([1, 39, 1, 55])],
      output: `const C = ({ count, title }) => <div>{count ? title : null}</div>`,
    },
    {
      code: `const C = ({ count, title }) => <div>{count > 0 && title}</div>`,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([1, 39, 1, 57])],
      output: `const C = ({ count, title }) => <div>{count > 0 ? title : null}</div>`,
    },

    // ---- ignoreAttributes does NOT silence children-level reports ----
    {
      code: `const C = ({ enabled }) => <Foo bar={<Something>{enabled && <MuchWow/>}</Something>} />`,
      options: [{ ignoreAttributes: true }],
      errors: [expectNoPotentialLeakedRenderError([1, 50, 1, 71])],
      output: `const C = ({ enabled }) => <Foo bar={<Something>{enabled ? <MuchWow/> : null}</Something>} />`,
    },

    // ---- Const-bound non-boolean Identifier still leaks ----
    {
      code: `const isOpen = 0; const C = () => <Popover open={isOpen && items.length > 0} />`,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([1, 50, 1, 76])],
      output: `const isOpen = 0; const C = () => <Popover open={!!isOpen && items.length > 0} />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        const Example = () => {
          return (
            <>
              {0 && <Something/>}
              {'' && <Something/>}
              {NaN && <Something/>}
            </>
          )
        }
      `,
      settings: { react: { version: '17.999.999' } },
      errors: [
        expectNoPotentialLeakedRenderError([5, 16, 5, 33]),
        expectNoPotentialLeakedRenderError([6, 16, 6, 34]),
        expectNoPotentialLeakedRenderError([7, 16, 7, 35]),
      ],
      output: `
        const Example = () => {
          return (
            <>
              {0 ? <Something/> : null}
              {'' ? <Something/> : null}
              {NaN ? <Something/> : null}
            </>
          )
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        const Example = () => {
          return (
            <>
              {0 && <Something/>}
              {'' && <Something/>}
              {NaN && <Something/>}
            </>
          )
        }
      `,
      settings: { react: { version: '18.0.0' } },
      errors: [
        expectNoPotentialLeakedRenderError([5, 16, 5, 33]),
        expectNoPotentialLeakedRenderError([7, 16, 7, 35]),
      ],
      output: `
        const Example = () => {
          return (
            <>
              {0 ? <Something/> : null}
              {'' && <Something/>}
              {NaN ? <Something/> : null}
            </>
          )
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        const Component = ({ count, title }) => {
          return <div>{count && title}</div>
        }
      `,
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 38])],
      output: `
        const Component = ({ count, title }) => {
          return <div>{count ? title : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        const Component = ({ count }) => {
          return <div>{count && <span>There are {count} results</span>}</div>
        }
      `,
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 71])],
      output: `
        const Component = ({ count }) => {
          return <div>{count ? <span>There are {count} results</span> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{elements.length && <List elements={elements}/>}</div>
        }
      `,
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 70])],
      output: `
        const Component = ({ elements }) => {
          return <div>{elements.length ? <List elements={elements}/> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        const Component = ({ nestedCollection }) => {
          return <div>{nestedCollection.elements.length && <List elements={nestedCollection.elements}/>}</div>
        }
      `,
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 104])],
      output: `
        const Component = ({ nestedCollection }) => {
          return <div>{nestedCollection.elements.length ? <List elements={nestedCollection.elements}/> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{elements[0] && <List elements={elements}/>}</div>
        }
      `,
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 66])],
      output: `
        const Component = ({ elements }) => {
          return <div>{elements[0] ? <List elements={elements}/> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 84])],
      output: `
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) ? <Results>{numberA+numberB}</Results> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
      options: [{ validStrategies: ['coerce', 'ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 84])],
      output: `
        const Component = ({ numberA, numberB }) => {
          return <div>{!!(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        const Component = ({ count, title }) => {
          return <div>{count && title}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 38])],
      output: `
        const Component = ({ count, title }) => {
          return <div>{count ? title : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        const Component = ({ count }) => {
          return <div>{count && <span>There are {count} results</span>}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 71])],
      output: `
        const Component = ({ count }) => {
          return <div>{count ? <span>There are {count} results</span> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{elements.length && <List elements={elements}/>}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 70])],
      output: `
        const Component = ({ elements }) => {
          return <div>{elements.length ? <List elements={elements}/> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        const Component = ({ nestedCollection }) => {
          return <div>{nestedCollection.elements.length && <List elements={nestedCollection.elements}/>}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 104])],
      output: `
        const Component = ({ nestedCollection }) => {
          return <div>{nestedCollection.elements.length ? <List elements={nestedCollection.elements}/> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{elements[0] && <List elements={elements}/>}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 66])],
      output: `
        const Component = ({ elements }) => {
          return <div>{elements[0] ? <List elements={elements}/> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 84])],
      output: `
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) ? <Results>{numberA+numberB}</Results> : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        const Component = ({ someCondition, title }) => {
          return <div>{!someCondition && title}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 47])],
      output: `
        const Component = ({ someCondition, title }) => {
          return <div>{!someCondition ? title : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        const Component = ({ count, title }) => {
          return <div>{!!count && title}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 40])],
      output: `
        const Component = ({ count, title }) => {
          return <div>{count ? title : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        const Component = ({ count, title }) => {
          return <div>{count > 0 && title}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 42])],
      output: `
        const Component = ({ count, title }) => {
          return <div>{count > 0 ? title : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        const Component = ({ count, title }) => {
          return <div>{0 != count && title}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 43])],
      output: `
        const Component = ({ count, title }) => {
          return <div>{0 != count ? title : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        const Component = ({ count, total, title }) => {
          return <div>{count < total && title}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 46])],
      output: `
        const Component = ({ count, total, title }) => {
          return <div>{count < total ? title : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        const Component = ({ count, title, somethingElse }) => {
          return <div>{!!(count && somethingElse) && title}</div>
        }
      `,
      options: [{ validStrategies: ['ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 59])],
      output: `
        const Component = ({ count, title, somethingElse }) => {
          return <div>{count && somethingElse ? title : null}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        const Component = ({ count, title }) => {
          return <div>{count && title}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 38])],
      output: `
        const Component = ({ count, title }) => {
          return <div>{!!count && title}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        const Component = ({ count }) => {
          return <div>{count && <span>There are {count} results</span>}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 71])],
      output: `
        const Component = ({ count }) => {
          return <div>{!!count && <span>There are {count} results</span>}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{elements.length && <List elements={elements}/>}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 70])],
      output: `
        const Component = ({ elements }) => {
          return <div>{!!elements.length && <List elements={elements}/>}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        const Component = ({ nestedCollection }) => {
          return <div>{nestedCollection.elements.length && <List elements={nestedCollection.elements}/>}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 104])],
      output: `
        const Component = ({ nestedCollection }) => {
          return <div>{!!nestedCollection.elements.length && <List elements={nestedCollection.elements}/>}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        const Component = ({ elements }) => {
          return <div>{elements[0] && <List elements={elements}/>}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 66])],
      output: `
        const Component = ({ elements }) => {
          return <div>{!!elements[0] && <List elements={elements}/>}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 84])],
      output: `
        const Component = ({ numberA, numberB }) => {
          return <div>{!!(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        const Component = ({ connection, hasError, hasErrorUpdate}) => {
          return <div>{connection && (hasError || hasErrorUpdate)}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 66])],
      output: `
        const Component = ({ connection, hasError, hasErrorUpdate}) => {
          return <div>{!!connection && (hasError || hasErrorUpdate)}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        const Component = ({ count, title }) => {
          return <div>{count ? title : null}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 44])],
      output: `
        const Component = ({ count, title }) => {
          return <div>{!!count && title}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        const Component = ({ count, title }) => {
          return <div>{!count ? title : null}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 45])],
      output: `
        const Component = ({ count, title }) => {
          return <div>{!count && title}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        const Component = ({ count, somethingElse, title }) => {
          return <div>{count && somethingElse ? title : null}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 61])],
      output: `
        const Component = ({ count, somethingElse, title }) => {
          return <div>{!!count && !!somethingElse && title}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        const Component = ({ items, somethingElse, title }) => {
          return <div>{items.length > 0 && somethingElse && title}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 66])],
      output: `
        const Component = ({ items, somethingElse, title }) => {
          return <div>{items.length > 0 && !!somethingElse && title}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        const MyComponent = () => {
          const items = []
          const breakpoint = { phones: true }

          return <div>{items.length > 0 && breakpoint.phones && <span />}</div>
        }
      `,
      options: [{ validStrategies: ['coerce', 'ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([6, 24, 6, 73])],
      output: `
        const MyComponent = () => {
          const items = []
          const breakpoint = { phones: true }

          return <div>{items.length > 0 && !!breakpoint.phones && <span />}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
        const MyComponent = () => {
          return <div>{maybeObject && (isFoo ? <Aaa /> : <Bbb />)}</div>
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 24, 3, 66])],
      output: `
        const MyComponent = () => {
          return <div>{!!maybeObject && (isFoo ? <Aaa /> : <Bbb />)}</div>
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        const MyComponent = () => {
          return <Something checked={isIndeterminate ? false : isChecked} />
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 38, 3, 73])],
      output: `
        const MyComponent = () => {
          return <Something checked={!isIndeterminate && isChecked} />
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        const MyComponent = () => {
          return <Something checked={cond && isIndeterminate ? false : isChecked} />
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([3, 38, 3, 81])],
      output: `
        const MyComponent = () => {
          return <Something checked={!!cond && !!isIndeterminate ? false : isChecked} />
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        const MyComponent = () => {
          return (
            <>
              {someCondition && (
                <div>
                  <p>hello</p>
                </div>
              )}
            </>
          )
        }
      `,
      options: [{ validStrategies: ['coerce', 'ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([5, 16, 9, 16])],
      output: `
        const MyComponent = () => {
          return (
            <>
              {!!someCondition && (
                <div>
                  <p>hello</p>
                </div>
              )}
            </>
          )
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `
        const MyComponent = () => {
          return (
            <>
              {someCondition && (
                <SomeComponent
                  prop1={val1}
                  prop2={val2}
                />
              )}
            </>
          )
        }
      `,
      options: [{ validStrategies: ['coerce', 'ternary'] }],
      errors: [expectNoPotentialLeakedRenderError([5, 16, 10, 16])],
      output: `
        const MyComponent = () => {
          return (
            <>
              {!!someCondition && (
                <SomeComponent
                  prop1={val1}
                  prop2={val2}
                />
              )}
            </>
          )
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        const isOpen = 0;
        const Component = () => {
          return <Popover open={isOpen && items.length > 0} />
        }
      `,
      options: [{ validStrategies: ['coerce'] }],
      errors: [expectNoPotentialLeakedRenderError([4, 33, 4, 59])],
      output: `
        const isOpen = 0;
        const Component = () => {
          return <Popover open={!!isOpen && items.length > 0} />
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
