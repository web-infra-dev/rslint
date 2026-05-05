import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-array-index-key', {} as never, {
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
  ],

  invalid: [
    // ---- Direct identifier ----
    {
      code: `foo.map((bar, i) => <Foo key={i} />);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `[{}, {}].map((bar, i) => <Foo key={i} />);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.map((bar, anything) => <Foo key={anything} />);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Template literal ----
    {
      code: 'foo.map((bar, i) => <Foo key={`foo-${i}`} />);',
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Concatenation chain ----
    {
      code: `foo.map((bar, i) => <Foo key={'foo-' + i} />);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.map((bar, i) => <Foo key={'foo-' + i + '-bar'} />);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- React.cloneElement ----
    {
      code: `foo.map((baz, i) => React.cloneElement(someChild, { ...someChild.props, key: i }));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        import { cloneElement } from 'react';

        foo.map((baz, i) => cloneElement(someChild, { ...someChild.props, key: i }));
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        foo.map((item, i) => {
          return React.cloneElement(someChild, {
            key: i
          })
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }],
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
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Every iterator method, JSX path ----
    {
      code: `foo.forEach((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.filter((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.some((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.every((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.find((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.findIndex((bar, i) => { baz.push(<Foo key={i} />); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.reduce((a, b, i) => a.concat(<Foo key={i} />), []);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.flatMap((a, i) => <Foo key={i} />);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.reduceRight((a, b, i) => a.concat(<Foo key={i} />), []);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- React.createElement props path ----
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: i }));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: "foo.map((bar, i) => React.createElement('Foo', { key: `foo-${i}` }));",
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' + i }));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' + i + '-bar' }));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.forEach((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.filter((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.some((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.every((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.find((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.findIndex((bar, i) => { baz.push(React.createElement('Foo', { key: i })); });`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- React.Children.* / Children.* (callback at args[1]) ----
    {
      code: `
        Children.map(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        import { cloneElement } from 'react';

        Children.map(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        React.Children.map(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        import { cloneElement } from 'react';

        React.Children.map(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        Children.forEach(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        import { cloneElement } from 'react';

        Children.forEach(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        React.Children.forEach(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        import { cloneElement } from 'react';

        React.Children.forEach(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Optional chaining iterator ----
    {
      code: `foo?.map((child, i) => <Foo key={i} />);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Nested iterators ----
    {
      code: `foo.map((a, i) => bar.map((c, j) => <X key={i} />));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.map((a, i) => bar.map((c, j) => <X key={j} />));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.map((a, i) => bar.map((c, i) => <X key={i} />));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.map((a, i) => bar.qux((c, j) => <X key={i} />));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Inner function inside iterator body still tracked ----
    {
      code: `
        foo.map((bar, i) => function inner() {
          return <Foo key={i} />;
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Identifier branch accepts ANY react-imported name ----
    // Upstream's Identifier branch does NOT verify name; any
    // ImportSpecifier from 'react' passes. Mirrored byte-for-byte.
    {
      code: `
        import { foo } from 'react';

        foo.map((bar, i) => foo(c, { key: i }));
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Optional-chain pragma — upstream listens on
    // `'CallExpression, OptionalCallExpression'` and accepts both
    // `MemberExpression` and `OptionalMemberExpression`. ----
    {
      code: `foo.map((bar, i) => React?.cloneElement(c, { key: i }));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.map((bar, i) => React?.createElement('Foo', { key: i }));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Multi-substitution / multi-leaf ----
    {
      code: 'foo.map((bar, i) => <Foo key={`${i}-${i}`} />);',
      errors: [{ messageId: 'noArrayIndex' }, { messageId: 'noArrayIndex' }],
    },
    {
      code: `foo.map((bar, i) => <Foo key={i + '-' + i} />);`,
      errors: [{ messageId: 'noArrayIndex' }, { messageId: 'noArrayIndex' }],
    },

    // ---- Multiple `key` keys in one props object ----
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: i, key: i }));`,
      errors: [{ messageId: 'noArrayIndex' }, { messageId: 'noArrayIndex' }],
    },

    // ---- Nested cloneElement ----
    {
      code: `foo.map((bar, i) => React.cloneElement(React.cloneElement(c, { key: i }), { key: i }));`,
      errors: [{ messageId: 'noArrayIndex' }, { messageId: 'noArrayIndex' }],
    },

    // ---- pragma swap settings ----
    {
      code: `foo.map((bar, i) => Act.cloneElement(c, { key: i }));`,
      settings: { react: { pragma: 'Act' } },
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Chained receiver `foo.bar.map(...)` ----
    {
      code: `foo.bar.map((c, i) => <X key={i} />);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Shorthand `{ key }` when the index parameter is named `key` ----
    {
      code: `foo.map((bar, key) => React.cloneElement(c, { key }));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- TS-wrapped iterator receiver still triggers ----
    // Upstream `getMapIndexParamName` only inspects `.property.name`,
    // so the receiver shape is opaque to it.
    {
      code: `(foo as any[]).map((bar, i) => <Foo key={i} />);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Mixed key kinds in same props: only Identifier `key` matches ----
    {
      code: `foo.map((bar, i) => React.createElement('Foo', { key: i, ["key"]: i }));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Computed property `[<Identifier "key">]: ...` ----
    // Upstream ESTree puts the inner expression directly as
    // `prop.key`. With `[key]: key` and the iterator's index parameter
    // named `key`, `prop.key.name === 'key'` matches and `prop.value`
    // is the index Identifier.
    {
      code: `foo.map((bar, key) => React.cloneElement(c, { [key]: key }));`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Children.toArray pipe ----
    {
      code: `React.Children.toArray(children).map((c, i) => <X key={i} />);`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Block body with multiple JSX returns ----
    {
      code: `
        foo.map((bar, i) => {
          if (cond) return <A key={i} />;
          return <B key={i} />;
        });
      `,
      errors: [{ messageId: 'noArrayIndex' }, { messageId: 'noArrayIndex' }],
    },

    // ---- JSX child expression containing iterator ----
    {
      code: `<div>{foo.map((bar, i) => <X key={i} />)}</div>;`,
      errors: [{ messageId: 'noArrayIndex' }],
    },

    // ---- Coercion via toString / String ----
    {
      code: `
        foo.map((bar, index) => (
          <Element key={index.toString()} bar={bar} />
        ));
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        foo.map((bar, index) => (
          <Element key={String(index)} bar={bar} />
        ));
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
    {
      code: `
        foo.map((bar, index) => (
          <Element key={index} bar={bar} />
        ));
      `,
      errors: [{ messageId: 'noArrayIndex' }],
    },
  ],
});
