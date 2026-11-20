import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNoArrayIndexError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noArrayIndex',
    message: 'Do not use Array index in keys',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-array-index-key', {
  valid: [
    // ---- Outside any tracked iterator ----
    { code: `<Foo key="foo" />;` },
    { code: `<Foo key={i} />;` },
    { code: `<Foo key />;` },
    { code: '<Foo key={`foo-${i}`} />;' },
    { code: `<Foo key={'foo-' + i} />;` },

    // ---- Untracked iterators (.bar) ----
    { code: `foo.bar((baz, i) => <Foo key={i} />);` },
    { code: 'foo.bar((bar, i) => <Foo key={`foo-${i}`} />);' },
    { code: `foo.bar((bar, i) => <Foo key={'foo-' + i} />);` },

    // ---- Tracked iterator, key does not reference the index ----
    { code: `foo.map((baz) => <Foo key={baz.id} />);` },
    { code: `foo.map((baz, i) => <Foo key={baz.id} />);` },
    { code: `foo.map((baz, i) => <Foo key={'foo' + baz.id} />);` },
    {
      code: `foo.map((baz, i) => React.cloneElement(someChild, { ...someChild.props }));`,
    },
    {
      code: `foo.map((baz, i) => cloneElement(someChild, { ...someChild.props }));`,
    },
    {
      code: `
        foo.map((item, i) => {
          return React.cloneElement(someChild, {
            key: item.id
          })
        });
      `,
    },
    {
      code: `
        foo.map((item, i) => {
          return cloneElement(someChild, {
            key: item.id
          })
        });
      `,
    },
    { code: `foo.map((baz, i) => <Foo key />);` },
    { code: `foo.reduce((a, b) => a.concat(<Foo key={b.id} />), []);` },

    // ---- Index-flavored expressions that the rule does not match ----
    { code: `foo.map((bar, i) => <Foo key={i.baz.toString()} />);` },
    { code: `foo.map((bar, i) => <Foo key={i.toString} />);` },
    { code: `foo.map((bar, i) => <Foo key={String()} />);` },
    { code: `foo.map((bar, i) => <Foo key={String(baz)} />);` },

    // ---- Insufficient params for the index slot ----
    { code: `foo.flatMap((a) => <Foo key={a} />);` },
    { code: `foo.reduce((a, b, i) => a.concat(<Foo key={b.id} />), []);` },
    { code: `foo.reduceRight((a, b) => a.concat(<Foo key={b.id} />), []);` },
    {
      code: `foo.reduceRight((a, b, i) => a.concat(<Foo key={b.id} />), []);`,
    },

    // ---- React.Children.* / Children.* with non-index key ----
    {
      code: `
        React.Children.map(this.props.children, (child, index, arr) => {
          return React.cloneElement(child, { key: child.id });
        });
      `,
    },
    {
      code: `
        React.Children.map(this.props.children, (child, index, arr) => {
          return cloneElement(child, { key: child.id });
        });
      `,
    },
    {
      code: `
        Children.forEach(this.props.children, (child, index, arr) => {
          return React.cloneElement(child, { key: child.id });
        });
      `,
    },
    {
      code: `
        Children.forEach(this.props.children, (child, index, arr) => {
          return cloneElement(child, { key: child.id });
        });
      `,
    },

    // ---- Optional chaining with non-index key ----
    { code: `foo?.map(child => <Foo key={child.i} />);` },

    // ---- Real-user scenarios ----
    // createElement / cloneElement OUTSIDE any iterator — stack empty.
    { code: `React.createElement('Foo', { key: i });` },
    { code: `React.cloneElement(child, { key: i });` },
    // thisArg as third argument — irrelevant.
    { code: `foo.map((bar) => <Foo key={bar.id} />, this);` },
    // Empty array.
    { code: `[].map((bar, i) => <Foo key={bar.id} />);` },
    // Computed-key with StringLiteral inside (`["key"]: i`) — opaque
    // upstream (`prop.key.name` undefined on Literal).
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { ["key"]: i }));`,
    },
    // String-literal key `"key"` — same opaque shape.
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { "key": i }));`,
    },
    // Computed-key with non-`key` Identifier inside (`[other]: i`) —
    // `prop.key.name === 'other' !== 'key'`. NOT reported.
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { [other]: i }));`,
    },
    // Shorthand `{ key }` when the index parameter is NOT named `key` —
    // the local `key` binding shadows the index slot.
    {
      code: `foo.map((bar, i) => { const key = bar.id; return React.cloneElement(c, { key }); });`,
    },
    // Iterator returning a non-JSX value.
    { code: `foo.map((bar, i) => bar + i);` },
    // Empty arrow body.
    { code: `foo.map((bar, i) => {});` },
    // Inner shadows outer index name — pop balance preserved.
    { code: `foo.map((a, i) => bar.map((c, i) => <X key={c.id} />));` },
    // Numeric / boolean literal keys.
    { code: `foo.map((bar, i) => <Foo key={42} />);` },
    { code: `foo.map((bar, i) => <Foo key={true} />);` },
    // Reduce accumulator (pos 0) is not the index slot (pos 2).
    { code: `foo.reduce((a, b) => a.concat(<Foo key={a} />), []);` },
    // Default-valued index parameter.
    { code: `foo.map((bar, i = 0) => <Foo key={i} />);` },
    // Pragma swap — `Act.cloneElement` is the factory now;
    // `key: bar.id` is fine.
    {
      code: `foo.map((bar, i) => Act.cloneElement(c, { key: bar.id }));`,
      settings: { react: { pragma: 'Act' } },
    },
    // Bare `cloneElement` without an `import { cloneElement }
    // from 'react'` — the bare-callee branch returns false.
    {
      code: `
        const cloneElement = (...args) => args;
        foo.map((baz, i) => cloneElement(someChild, { key: i }));
      `,
    },
    // Identifier branch is gated on ImportSpecifier — destructuring
    // / aliasing / require do NOT count.
    {
      code: `
        const { cloneElement } = React;
        foo.map((baz, i) => cloneElement(someChild, { key: i }));
      `,
    },
    {
      code: `
        const cloneElement = React.cloneElement;
        foo.map((baz, i) => cloneElement(someChild, { key: i }));
      `,
    },
    {
      code: `
        const { cloneElement } = require('react');
        foo.map((baz, i) => cloneElement(someChild, { key: i }));
      `,
    },
    // Module specifier hardcoded `'react'` — pragma swap doesn't redirect.
    {
      code: `
        import { cloneElement } from 'act';
        foo.map((baz, i) => cloneElement(someChild, { key: i }));
      `,
      settings: { react: { pragma: 'Act' } },
    },
    // Tagged template — KindTaggedTemplateExpression, not template.
    { code: 'foo.map((bar, i) => <Foo key={tag`foo-${i}`} />);' },
    // No-substitution template — opaque.
    { code: 'foo.map((bar, i) => <Foo key={`foo-i`} />);' },
    // Conditional inside binary chain — opaque.
    { code: `foo.map((bar, i) => <Foo key={'a' + (i ? 'x' : 'y')} />);` },
    // Identity-wrapped index — not String / .toString().
    { code: `foo.map((bar, i) => <Foo key={identity(i)} />);` },
    // Object-literal as key value.
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: { nested: i } }));`,
    },
    // Bracket access `i['toString']` — not the dotted shape.
    { code: `foo.map((bar, i) => <Foo key={i['toString']()} />);` },
    // Destructured index parameter — undefined name.
    { code: `foo.map((bar, [i]) => <Foo key={i} />);` },
    // Rest-only index slot.
    { code: `foo.map((bar, ...rest) => <Foo key={rest[0]} />);` },
    // String() with member-access argument — not a direct index ref.
    { code: `foo.map((bar, i) => <Foo key={String(i.foo)} />);` },
    // TS expression wrappers on pragma / key value — opaque, matching
    // ESLint's JS-only AST exactly.
    {
      code: `foo.map((bar, i) => (React as any).cloneElement(c, { key: i }));`,
    },
    {
      code: `foo.map((bar, i) => React!.createElement('Foo', { key: i }));`,
    },
    { code: `foo.map((bar, i) => <Foo key={(i as any)} />);` },
    // Optional-chain CallExpression as the key value — opaque to
    // upstream's `astUtil.isCallExpression` (rejects OptionalCallExpression)
    // and to `node.callee.type === 'MemberExpression'` (rejects
    // OptionalMemberExpression).
    { code: `foo.map((bar, i) => <Foo key={i?.toString()} />);` },
    { code: `foo.map((bar, i) => <Foo key={String?.(i)} />);` },
    { code: `foo.map((bar, i) => <Foo key={i?.toString?.()} />);` },

    // ---- Logical-operator fallback (ESTree LogicalExpression) ----
    // `||` / `&&` / `??` are LogicalExpression upstream, not a
    // BinaryExpression, so `key={fallback || index}` is never flagged.
    { code: `foo.map((bar, i) => <Foo key={bar.id || i} />);` },
    { code: `foo.map((bar, i) => <Foo key={bar.id ?? i} />);` },
    { code: `foo.map((bar, i) => <Foo key={i && bar.id} />);` },
    { code: `foo.map((bar, i) => <Foo key={bar.a || bar.b || i} />);` },
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: bar.id || i }));`,
    },
    { code: `foo.map((bar, i) => <Foo key={'x' + (bar.id || i)} />);` },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `foo.bar((baz, i) => <Foo key={i} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `foo.bar((bar, i) => <Foo key={\`foo-\${i}\`} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `foo.bar((bar, i) => <Foo key={'foo-' + i} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `foo.map((baz) => <Foo key={baz.id} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `foo.map((baz, i) => <Foo key={baz.id} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `foo.map((baz, i) => <Foo key={'foo' + baz.id} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `foo.map((baz, i) => React.cloneElement(someChild, { ...someChild.props }))`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `foo.map((baz, i) => cloneElement(someChild, { ...someChild.props }))`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        foo.map((item, i) => {
          return React.cloneElement(someChild, {
            key: item.id
          })
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        foo.map((item, i) => {
          return cloneElement(someChild, {
            key: item.id
          })
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `foo.map((baz, i) => <Foo key />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `foo.reduce((a, b) => a.concat(<Foo key={b.id} />), [])`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `foo.map((bar, i) => <Foo key={i.baz.toString()} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `foo.map((bar, i) => <Foo key={i.toString} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `foo.map((bar, i) => <Foo key={String()} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `foo.map((bar, i) => <Foo key={String(baz)} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `foo.flatMap((a) => <Foo key={a} />)`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `foo.reduce((a, b, i) => a.concat(<Foo key={b.id} />), [])`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `foo.reduceRight((a, b) => a.concat(<Foo key={b.id} />), [])`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `foo.reduceRight((a, b, i) => a.concat(<Foo key={b.id} />), [])`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        React.Children.map(this.props.children, (child, index, arr) => {
          return React.cloneElement(child, { key: child.id });
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        React.Children.map(this.props.children, (child, index, arr) => {
          return cloneElement(child, { key: child.id });
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        Children.forEach(this.props.children, (child, index, arr) => {
          return React.cloneElement(child, { key: child.id });
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        Children.forEach(this.props.children, (child, index, arr) => {
          return cloneElement(child, { key: child.id });
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `foo?.map(child => <Foo key={child.i} />)`,
      languageOptions: upstreamModuleOptions,
    },
  ],

  invalid: [
    // ---- Direct identifier ----
    {
      code: `foo.map((bar, i) => <Foo key={i} />);`,
      errors: [expectNoArrayIndexError([1, 31, 1, 32])],
    },
    {
      code: `[{}, {}].map((bar, i) => <Foo key={i} />);`,
      errors: [expectNoArrayIndexError([1, 36, 1, 37])],
    },
    {
      code: `foo.map((bar, anything) => <Foo key={anything} />);`,
      errors: [expectNoArrayIndexError([1, 38, 1, 46])],
    },

    // ---- Template literal ----
    {
      code: 'foo.map((bar, i) => <Foo key={`foo-${i}`} />);',
      errors: [expectNoArrayIndexError([1, 31, 1, 41])],
    },

    // ---- Concatenation chain ----
    {
      code: `foo.map((bar, i) => <Foo key={'foo-' + i} />);`,
      errors: [expectNoArrayIndexError([1, 31, 1, 41])],
    },
    {
      code: `foo.map((bar, i) => <Foo key={'foo-' + i + '-bar'} />);`,
      errors: [expectNoArrayIndexError([1, 31, 1, 50])],
    },

    // Real BinaryExpression whose right operand is a logical subtree —
    // the `+` reports the index on the left; the `||` subtree is skipped.
    {
      code: `foo.map((bar, i) => <Foo key={i + (bar.id || 0)} />);`,
      errors: [expectNoArrayIndexError([1, 31, 1, 48])],
    },

    // ---- React.cloneElement ----
    {
      code: `foo.map((baz, i) => React.cloneElement(someChild, { ...someChild.props, key: i }));`,
      errors: [expectNoArrayIndexError([1, 78, 1, 79])],
    },
    {
      code: `
        import { cloneElement } from 'react';

        foo.map((baz, i) => cloneElement(someChild, { ...someChild.props, key: i }));
      `,
      errors: [expectNoArrayIndexError([4, 80, 4, 81])],
    },
    {
      code: `
        foo.map((item, i) => {
          return React.cloneElement(someChild, {
            key: i
          })
        });
      `,
      errors: [expectNoArrayIndexError([4, 18, 4, 19])],
    },
    {
      code: `
        import { cloneElement } from 'react';

        foo.map((item, i) => {
          return cloneElement(someChild, {
            key: i
          })
        });
      `,
      errors: [expectNoArrayIndexError([6, 18, 6, 19])],
    },

    // ---- Every iterator method, JSX path ----
    {
      code: `foo.forEach((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [expectNoArrayIndexError([1, 46, 1, 47])],
    },
    {
      code: `foo.filter((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [expectNoArrayIndexError([1, 45, 1, 46])],
    },
    {
      code: `foo.some((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [expectNoArrayIndexError([1, 43, 1, 44])],
    },
    {
      code: `foo.every((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [expectNoArrayIndexError([1, 44, 1, 45])],
    },
    {
      code: `foo.find((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [expectNoArrayIndexError([1, 43, 1, 44])],
    },
    {
      code: `foo.findIndex((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [expectNoArrayIndexError([1, 48, 1, 49])],
    },
    {
      code: `foo.reduce((a, b, i) => a.concat(<Foo key={i} />), []);`,
      errors: [expectNoArrayIndexError([1, 44, 1, 45])],
    },
    {
      code: `foo.flatMap((a, i) => <Foo key={i} />);`,
      errors: [expectNoArrayIndexError([1, 33, 1, 34])],
    },
    {
      code: `foo.reduceRight((a, b, i) => a.concat(<Foo key={i} />), []);`,
      errors: [expectNoArrayIndexError([1, 49, 1, 50])],
    },

    // ---- React.createElement props path ----
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: i }));`,
      errors: [expectNoArrayIndexError([1, 55, 1, 56])],
    },
    {
      code: "foo.map((bar, i) => React.createElement('Foo', { key: `foo-${i}` }));",
      errors: [expectNoArrayIndexError([1, 55, 1, 65])],
    },
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' + i }));`,
      errors: [expectNoArrayIndexError([1, 55, 1, 65])],
    },
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' + i + '-bar' }));`,
      errors: [expectNoArrayIndexError([1, 55, 1, 74])],
    },
    {
      code: `foo.forEach((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [expectNoArrayIndexError([1, 70, 1, 71])],
    },
    {
      code: `foo.filter((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [expectNoArrayIndexError([1, 69, 1, 70])],
    },
    {
      code: `foo.some((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [expectNoArrayIndexError([1, 67, 1, 68])],
    },
    {
      code: `foo.every((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [expectNoArrayIndexError([1, 68, 1, 69])],
    },
    {
      code: `foo.find((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [expectNoArrayIndexError([1, 67, 1, 68])],
    },
    {
      code: `foo.findIndex((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [expectNoArrayIndexError([1, 72, 1, 73])],
    },

    // ---- React.Children.* / Children.* (callback at args[1]) ----
    {
      code: `
        Children.map(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        });
      `,
      errors: [expectNoArrayIndexError([3, 51, 3, 56])],
    },
    {
      code: `
        import { cloneElement } from 'react';

        Children.map(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        });
      `,
      errors: [expectNoArrayIndexError([5, 45, 5, 50])],
    },
    {
      code: `
        React.Children.map(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        });
      `,
      errors: [expectNoArrayIndexError([3, 51, 3, 56])],
    },
    {
      code: `
        import { cloneElement } from 'react';

        React.Children.map(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        });
      `,
      errors: [expectNoArrayIndexError([5, 45, 5, 50])],
    },
    {
      code: `
        Children.forEach(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        });
      `,
      errors: [expectNoArrayIndexError([3, 51, 3, 56])],
    },
    {
      code: `
        import { cloneElement } from 'react';

        Children.forEach(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        });
      `,
      errors: [expectNoArrayIndexError([5, 45, 5, 50])],
    },
    {
      code: `
        React.Children.forEach(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        });
      `,
      errors: [expectNoArrayIndexError([3, 51, 3, 56])],
    },
    {
      code: `
        import { cloneElement } from 'react';

        React.Children.forEach(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        });
      `,
      errors: [expectNoArrayIndexError([5, 45, 5, 50])],
    },

    // ---- Optional chaining iterator ----
    {
      code: `foo?.map((child, i) => <Foo key={i} />);`,
      errors: [expectNoArrayIndexError([1, 34, 1, 35])],
    },

    // ---- Nested iterators ----
    {
      code: `foo.map((a, i) => bar.map((c, j) => <X key={i} />));`,
      errors: [expectNoArrayIndexError([1, 45, 1, 46])],
    },
    {
      code: `foo.map((a, i) => bar.map((c, j) => <X key={j} />));`,
      errors: [expectNoArrayIndexError([1, 45, 1, 46])],
    },
    {
      code: `foo.map((a, i) => bar.map((c, i) => <X key={i} />));`,
      errors: [expectNoArrayIndexError([1, 45, 1, 46])],
    },
    {
      code: `foo.map((a, i) => bar.qux((c, j) => <X key={i} />));`,
      errors: [expectNoArrayIndexError([1, 45, 1, 46])],
    },

    // ---- Inner function inside iterator body still tracked ----
    {
      code: `
        foo.map((bar, i) => function inner() {
          return <Foo key={i} />;
        });
      `,
      errors: [expectNoArrayIndexError([3, 28, 3, 29])],
    },

    // ---- Identifier branch accepts ANY react-imported name ----
    // Upstream's Identifier branch does NOT verify name; any
    // ImportSpecifier from 'react' passes. Mirrored byte-for-byte.
    {
      code: `
        import { foo } from 'react';

        foo.map((bar, i) => foo(c, { key: i }));
      `,
      errors: [expectNoArrayIndexError([4, 43, 4, 44])],
    },

    // ---- Optional-chain pragma — upstream listens on
    // `'CallExpression, OptionalCallExpression'` and accepts both
    // `MemberExpression` and `OptionalMemberExpression`. ----
    {
      code: `foo.map((bar, i) => React?.cloneElement(c, { key: i }));`,
      errors: [expectNoArrayIndexError([1, 51, 1, 52])],
    },
    {
      code: `foo.map((bar, i) => React?.createElement('Foo', { key: i }));`,
      errors: [expectNoArrayIndexError([1, 56, 1, 57])],
    },

    // ---- Multi-substitution / multi-leaf ----
    {
      code: 'foo.map((bar, i) => <Foo key={`${i}-${i}`} />);',
      errors: [
        expectNoArrayIndexError([1, 31, 1, 42]),
        expectNoArrayIndexError([1, 31, 1, 42]),
      ],
    },
    {
      code: `foo.map((bar, i) => <Foo key={i + '-' + i} />);`,
      errors: [
        expectNoArrayIndexError([1, 31, 1, 42]),
        expectNoArrayIndexError([1, 31, 1, 42]),
      ],
    },

    // ---- Multiple `key` keys in one props object ----
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: i, key: i }));`,
      errors: [
        expectNoArrayIndexError([1, 55, 1, 56]),
        expectNoArrayIndexError([1, 63, 1, 64]),
      ],
    },

    // ---- Nested cloneElement ----
    {
      code: `foo.map((bar, i) => React.cloneElement(React.cloneElement(c, { key: i }), { key: i }));`,
      errors: [
        expectNoArrayIndexError([1, 69, 1, 70]),
        expectNoArrayIndexError([1, 82, 1, 83]),
      ],
    },

    // ---- pragma swap settings ----
    {
      code: `foo.map((bar, i) => Act.cloneElement(c, { key: i }));`,
      settings: { react: { pragma: 'Act' } },
      errors: [expectNoArrayIndexError([1, 48, 1, 49])],
    },

    // ---- Chained receiver `foo.bar.map(...)` ----
    {
      code: `foo.bar.map((c, i) => <X key={i} />);`,
      errors: [expectNoArrayIndexError([1, 31, 1, 32])],
    },

    // ---- Shorthand `{ key }` when the index parameter is named `key` ----
    {
      code: `foo.map((bar, key) => React.cloneElement(c, { key }));`,
      errors: [expectNoArrayIndexError([1, 47, 1, 50])],
    },

    // ---- TS-wrapped iterator receiver still triggers ----
    // Upstream `getMapIndexParamName` only inspects `.property.name`,
    // so the receiver shape is opaque to it.
    {
      code: `(foo as any[]).map((bar, i) => <Foo key={i} />);`,
      errors: [expectNoArrayIndexError([1, 42, 1, 43])],
    },

    // ---- Mixed key kinds in same props: only Identifier `key` matches ----
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: i, ["key"]: i }));`,
      errors: [expectNoArrayIndexError([1, 55, 1, 56])],
    },

    // ---- Computed property `[<Identifier "key">]: ...` ----
    // Upstream ESTree puts the inner expression directly as
    // `prop.key`. With `[key]: key` and the iterator's index parameter
    // named `key`, `prop.key.name === 'key'` matches and `prop.value`
    // is the index Identifier.
    {
      code: `foo.map((bar, key) => React.cloneElement(c, { [key]: key }));`,
      errors: [expectNoArrayIndexError([1, 54, 1, 57])],
    },

    // ---- Children.toArray pipe ----
    {
      code: `React.Children.toArray(children).map((c, i) => <X key={i} />);`,
      errors: [expectNoArrayIndexError([1, 56, 1, 57])],
    },

    // ---- Block body with multiple JSX returns ----
    {
      code: `
        foo.map((bar, i) => {
          if (cond) return <A key={i} />;
          return <B key={i} />;
        });
      `,
      errors: [
        expectNoArrayIndexError([3, 36, 3, 37]),
        expectNoArrayIndexError([4, 26, 4, 27]),
      ],
    },

    // ---- JSX child expression containing iterator ----
    {
      code: `<div>{foo.map((bar, i) => <X key={i} />)}</div>;`,
      errors: [expectNoArrayIndexError([1, 35, 1, 36])],
    },

    // ---- Coercion via toString / String ----
    {
      code: `
        foo.map((bar, index) => (
          <Element key={index.toString()} bar={bar} />
        ));
      `,
      errors: [expectNoArrayIndexError([3, 25, 3, 41])],
    },
    {
      code: `
        foo.map((bar, index) => (
          <Element key={String(index)} bar={bar} />
        ));
      `,
      errors: [expectNoArrayIndexError([3, 32, 3, 37])],
    },
    {
      code: `
        foo.map((bar, index) => (
          <Element key={index} bar={bar} />
        ));
      `,
      errors: [expectNoArrayIndexError([3, 25, 3, 30])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `foo.map((bar, i) => <Foo key={i} />)`,
      errors: [expectNoArrayIndexError([1, 31, 1, 32])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `[{}, {}].map((bar, i) => <Foo key={i} />)`,
      errors: [expectNoArrayIndexError([1, 36, 1, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `foo.map((bar, anything) => <Foo key={anything} />)`,
      errors: [expectNoArrayIndexError([1, 38, 1, 46])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `foo.map((bar, i) => <Foo key={\`foo-\${i}\`} />)`,
      errors: [expectNoArrayIndexError([1, 31, 1, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `foo.map((bar, i) => <Foo key={'foo-' + i} />)`,
      errors: [expectNoArrayIndexError([1, 31, 1, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `foo.map((bar, i) => <Foo key={'foo-' + i + '-bar'} />)`,
      errors: [expectNoArrayIndexError([1, 31, 1, 50])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `foo.map((baz, i) => React.cloneElement(someChild, { ...someChild.props, key: i }))`,
      errors: [expectNoArrayIndexError([1, 78, 1, 79])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        import { cloneElement } from 'react';

        foo.map((baz, i) => cloneElement(someChild, { ...someChild.props, key: i }))
      `,
      errors: [expectNoArrayIndexError([4, 80, 4, 81])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        foo.map((item, i) => {
          return React.cloneElement(someChild, {
            key: i
          })
        })
      `,
      errors: [expectNoArrayIndexError([4, 18, 4, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        import { cloneElement } from 'react';

        foo.map((item, i) => {
          return cloneElement(someChild, {
            key: i
          })
        })
      `,
      errors: [expectNoArrayIndexError([6, 18, 6, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `foo.forEach((bar, i) => { baz.push(<Foo key={i} />); })`,
      errors: [expectNoArrayIndexError([1, 46, 1, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `foo.filter((bar, i) => { baz.push(<Foo key={i} />); })`,
      errors: [expectNoArrayIndexError([1, 45, 1, 46])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `foo.some((bar, i) => { baz.push(<Foo key={i} />); })`,
      errors: [expectNoArrayIndexError([1, 43, 1, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `foo.every((bar, i) => { baz.push(<Foo key={i} />); })`,
      errors: [expectNoArrayIndexError([1, 44, 1, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `foo.find((bar, i) => { baz.push(<Foo key={i} />); })`,
      errors: [expectNoArrayIndexError([1, 43, 1, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `foo.findIndex((bar, i) => { baz.push(<Foo key={i} />); })`,
      errors: [expectNoArrayIndexError([1, 48, 1, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `foo.reduce((a, b, i) => a.concat(<Foo key={i} />), [])`,
      errors: [expectNoArrayIndexError([1, 44, 1, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `foo.flatMap((a, i) => <Foo key={i} />)`,
      errors: [expectNoArrayIndexError([1, 33, 1, 34])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `foo.reduceRight((a, b, i) => a.concat(<Foo key={i} />), [])`,
      errors: [expectNoArrayIndexError([1, 49, 1, 50])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `foo.map((bar, i) => React.createElement('Foo', { key: i }))`,
      errors: [expectNoArrayIndexError([1, 55, 1, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `foo.map((bar, i) => React.createElement('Foo', { key: \`foo-\${i}\` }))`,
      errors: [expectNoArrayIndexError([1, 55, 1, 65])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' + i }))`,
      errors: [expectNoArrayIndexError([1, 55, 1, 65])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' + i + '-bar' }))`,
      errors: [expectNoArrayIndexError([1, 55, 1, 74])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `foo.forEach((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
      errors: [expectNoArrayIndexError([1, 70, 1, 71])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `foo.filter((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
      errors: [expectNoArrayIndexError([1, 69, 1, 70])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `foo.some((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
      errors: [expectNoArrayIndexError([1, 67, 1, 68])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `foo.every((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
      errors: [expectNoArrayIndexError([1, 68, 1, 69])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `foo.find((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
      errors: [expectNoArrayIndexError([1, 67, 1, 68])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `foo.findIndex((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`,
      errors: [expectNoArrayIndexError([1, 72, 1, 73])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        Children.map(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        })
      `,
      errors: [expectNoArrayIndexError([3, 51, 3, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        import { cloneElement } from 'react';

        Children.map(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        })
      `,
      errors: [expectNoArrayIndexError([5, 45, 5, 50])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        React.Children.map(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        })
      `,
      errors: [expectNoArrayIndexError([3, 51, 3, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        import { cloneElement } from 'react';

        React.Children.map(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        })
      `,
      errors: [expectNoArrayIndexError([5, 45, 5, 50])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
        Children.forEach(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        })
      `,
      errors: [expectNoArrayIndexError([3, 51, 3, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        import { cloneElement } from 'react';

        Children.forEach(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        })
      `,
      errors: [expectNoArrayIndexError([5, 45, 5, 50])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        React.Children.forEach(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        })
      `,
      errors: [expectNoArrayIndexError([3, 51, 3, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        import { cloneElement } from 'react';

        React.Children.forEach(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        })
      `,
      errors: [expectNoArrayIndexError([5, 45, 5, 50])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `foo?.map((child, i) => <Foo key={i} />)`,
      errors: [expectNoArrayIndexError([1, 34, 1, 35])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        foo.map((bar, index) => (
          <Element key={index.toString()} bar={bar} />
        ))
      `,
      errors: [expectNoArrayIndexError([3, 25, 3, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        foo.map((bar, index) => (
          <Element key={String(index)} bar={bar} />
        ))
      `,
      errors: [expectNoArrayIndexError([3, 32, 3, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        foo.map((bar, index) => (
          <Element key={index} bar={bar} />
        ))
      `,
      errors: [expectNoArrayIndexError([3, 25, 3, 30])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
