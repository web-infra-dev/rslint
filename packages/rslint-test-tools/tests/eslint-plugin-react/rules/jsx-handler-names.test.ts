import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectBadHandlerNameError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'badHandlerName',
    "Handler function for onChange prop key must be a camelCase name beginning with 'handle' only",
    range,
  );
}

function expectBadPropKeyError2(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'badPropKey',
    "Prop key for handleChange must begin with 'on'",
    range,
  );
}

function expectBadPropKeyError3(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'badPropKey',
    "Prop key for handleChange must begin with 'when'",
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-handler-names', {
  valid: [
    // ---- Standard handle/on patterns ----
    { code: `var x = <TestComponent onChange={this.handleChange} />` },
    { code: `var x = <TestComponent onChange={this.handle123Change} />` },
    { code: `var x = <TestComponent onChange={this.props.onChange} />` },
    {
      code: `
        var x = <TestComponent
          onChange={
            this
              .handleChange
          } />;
      `,
    },
    {
      code: `
        var x = <TestComponent
          onChange={
            this
              .props
              .handleChange
          } />;
      `,
    },

    // ---- checkLocalVariables ----
    {
      code: `var x = <TestComponent onChange={handleChange} />`,
      options: [{ checkLocalVariables: true }],
    },
    {
      code: `var x = <TestComponent onChange={takeCareOfChange} />`,
      options: [{ checkLocalVariables: false }],
    },

    // ---- checkInlineFunction ----
    {
      code: `var x = <TestComponent onChange={event => window.alert(event.target.value)} />`,
      options: [{ checkInlineFunction: false }],
    },
    {
      code: `var x = <TestComponent onChange={() => handleChange()} />`,
      options: [{ checkInlineFunction: true, checkLocalVariables: true }],
    },
    {
      code: `var x = <TestComponent onChange={() => this.handleChange()} />`,
      options: [{ checkInlineFunction: true }],
    },
    { code: `var x = <TestComponent onChange={() => 42} />` },

    // ---- Well-named props with no handler shape ----
    { code: `var x = <TestComponent onChange={this.props.onFoo} />` },
    { code: `var x = <TestComponent isSelected={this.props.isSelected} />` },
    {
      code: `var x = <TestComponent shouldDisplay={this.state.shouldDisplay} />`,
    },
    { code: `var x = <TestComponent shouldDisplay={arr[0].prop} />` },
    { code: `var x = <TestComponent onChange={props.onChange} />` },

    // ---- ref is always allowed ----
    { code: `var x = <TestComponent ref={this.handleRef} />` },
    { code: `var x = <TestComponent ref={this.somethingRef} />` },

    // ---- Custom prefixes ----
    {
      code: `var x = <TestComponent test={this.props.content} />`,
      options: [{ eventHandlerPrefix: 'on', eventHandlerPropPrefix: 'on' }],
    },
    {
      code: `var x = <TestComponent only={this.only} />`,
    },

    // ---- Disabled prefixes (false) ----
    {
      code: `var x = <TestComponent onChange={this.someChange} />`,
      options: [{ eventHandlerPrefix: false, eventHandlerPropPrefix: 'on' }],
    },
    {
      code: `var x = <TestComponent somePrefixChange={this.someChange} />`,
      options: [
        { eventHandlerPrefix: false, eventHandlerPropPrefix: 'somePrefix' },
      ],
    },
    {
      code: `var x = <TestComponent someProp={this.handleChange} />`,
      options: [{ eventHandlerPropPrefix: false }],
    },
    {
      code: `var x = <TestComponent someProp={this.somePrefixChange} />`,
      options: [
        { eventHandlerPrefix: 'somePrefix', eventHandlerPropPrefix: false },
      ],
    },
    {
      code: `var x = <TestComponent someProp={props.onChange} />`,
      options: [{ eventHandlerPropPrefix: false }],
    },

    // ---- ignoreComponentNames ----
    {
      code: `var x = <ComponentFromOtherLibraryBar customPropNameBar={handleSomething} />;`,
      options: [
        {
          checkLocalVariables: true,
          ignoreComponentNames: ['ComponentFromOtherLibraryBar'],
        },
      ],
    },
    {
      code: `
        function App() {
          return (
            <div>
              <MyLibInput customPropNameBar={handleSomething} />
              <MyLibCheckbox customPropNameBar={handleSomething} />
              <MyLibButtom customPropNameBar={handleSomething} />
            </div>
          );
        }
      `,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['MyLib*'] },
      ],
    },
    {
      code: `var x = <A.TestComponent customPropNameBar={handleSomething} />`,
      options: [
        {
          checkLocalVariables: true,
          ignoreComponentNames: ['A.TestComponent'],
        },
      ],
    },
    {
      code: `
        function App() {
          return (
            <div>
              <A.MyLibInput customPropNameBar={handleSomething} />
              <A.MyLibCheckbox customPropNameBar={handleSomething} />
              <A.MyLibButtom customPropNameBar={handleSomething} />
            </div>
          );
        }
      `,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['A.MyLib*'] },
      ],
    },

    // ---- tsgo edge: parenthesized member access (flat-AST equivalence) ----
    { code: `var x = <TestComponent onChange={(this.handleChange)} />` },
    { code: `var x = <TestComponent onChange={((this.handleChange))} />` },

    // ---- tsgo edge: optional chain on receiver (matches ChainExpression) ----
    { code: `var x = <TestComponent onChange={this?.handleChange} />` },
    { code: `var x = <TestComponent onChange={this?.props?.onChange} />` },

    // ---- Universal edge: deep / chained member access ----
    { code: `var x = <TestComponent onChange={a.b.c.handleChange} />` },

    // ---- Universal edge: async inline arrow body ----
    {
      code: `var x = <TestComponent onChange={async () => this.handleChange()} />`,
      options: [{ checkInlineFunction: true }],
    },

    // ---- Universal edge: only-spread element doesn't crash ----
    { code: `var x = <TestComponent {...props} />` },

    // ---- Glob alignment: brace expansion ----
    {
      code: `var x = <Foo1 onChange={whateverHandler} />`,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['Foo{1,2}'] },
      ],
    },
    {
      code: `var x = <Foo2 onChange={whateverHandler} />`,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['Foo{1,2}'] },
      ],
    },

    // ---- Glob alignment: char class ----
    {
      code: `var x = <TestA onChange={whateverHandler} />`,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['Test[ABC]'] },
      ],
    },

    // ---- Glob alignment: negated char class ----
    {
      code: `var x = <TestB onChange={whateverHandler} />`,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['Test[!A]'] },
      ],
    },

    // ---- Glob alignment: leading `!` whole-pattern negation ----
    {
      code: `var x = <Bar onChange={whateverHandler} />`,
      options: [{ checkLocalVariables: true, ignoreComponentNames: ['!Foo'] }],
    },

    // ---- Glob alignment: extglob ----
    {
      code: `var x = <Testa onChange={whateverHandler} />`,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['Test+(a|b)'] },
      ],
    },

    // ---- Lock-in: invalid-regex prefix doesn't crash the linter ----
    // Each prefix below produces a regex Go's RE2 rejects (unbalanced `(`,
    // empty class `[]`, dangling escape `\`, inverted range, etc.). The
    // upgraded option-parser uses `regexp.Compile` + nil fallback to keep
    // the lint process alive — the failing half of the rule simply
    // becomes a no-op for the affected prefix.
    {
      code: `var x = <TestComponent onChange={this.handleChange} />`,
      options: [{ eventHandlerPrefix: '(' }],
    },
    {
      code: `var x = <TestComponent onChange={this.handleChange} />`,
      options: [{ eventHandlerPrefix: '[a-Z]' }],
    },
    {
      code: `var x = <TestComponent onChange={this.handleChange} />`,
      options: [{ eventHandlerPrefix: '\\' }],
    },
    {
      code: `var x = <TestComponent onChange={this.handleChange} />`,
      options: [{ eventHandlerPropPrefix: '*' }],
    },

    // ---- Lock-in: invalid glob in ignoreComponentNames doesn't crash ----
    {
      code: `var x = <TestComponent onChange={whateverHandler} />`,
      options: [
        {
          checkLocalVariables: true,
          ignoreComponentNames: ['[z-a]', 'TestComponent'],
        },
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<TestComponent onChange={this.handleChange} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<TestComponent onChange={this.handle123Change} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<TestComponent onChange={this.props.onChange} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        <TestComponent
          onChange={
            this
              .handleChange
          } />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        <TestComponent
          onChange={
            this
              .props
              .handleChange
          } />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<TestComponent onChange={handleChange} />`,
      options: [{ checkLocalVariables: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `<TestComponent onChange={takeCareOfChange} />`,
      options: [{ checkLocalVariables: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `<TestComponent onChange={event => window.alert(event.target.value)} />`,
      options: [{ checkInlineFunction: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<TestComponent onChange={() => handleChange()} />`,
      options: [{ checkInlineFunction: true, checkLocalVariables: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `<TestComponent onChange={() => this.handleChange()} />`,
      options: [{ checkInlineFunction: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<TestComponent onChange={() => 42} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `<TestComponent onChange={this.props.onFoo} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `<TestComponent isSelected={this.props.isSelected} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `<TestComponent shouldDisplay={this.state.shouldDisplay} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `<TestComponent shouldDisplay={arr[0].prop} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `<TestComponent onChange={props.onChange} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `<TestComponent ref={this.handleRef} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `<TestComponent ref={this.somethingRef} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `<TestComponent test={this.props.content} />`,
      options: [{ eventHandlerPrefix: 'on', eventHandlerPropPrefix: 'on' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax ('}' expected.).",
      code: `<TestComponent onChange={props::handleChange} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream bind operator syntax (Expression expected.).',
      code: `<TestComponent onChange={::props.onChange} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax ('}' expected.).",
      code: `<TestComponent onChange={props.foo::handleChange} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax ('}' expected.).",
      code: `<TestComponent onChange={() => props::handleChange()} />`,
      options: [{ checkInlineFunction: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream bind operator syntax (Expression expected.).',
      code: `<TestComponent onChange={() => ::props.onChange()} />`,
      options: [{ checkInlineFunction: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax ('}' expected.).",
      code: `<TestComponent onChange={() => props.foo::handleChange()} />`,
      options: [{ checkInlineFunction: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `<TestComponent only={this.only} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `<TestComponent onChange={this.someChange} />`,
      options: [{ eventHandlerPrefix: false, eventHandlerPropPrefix: 'on' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `<TestComponent somePrefixChange={this.someChange} />`,
      options: [
        { eventHandlerPrefix: false, eventHandlerPropPrefix: 'somePrefix' },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `<TestComponent someProp={this.handleChange} />`,
      options: [{ eventHandlerPropPrefix: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `<TestComponent someProp={this.somePrefixChange} />`,
      options: [
        { eventHandlerPrefix: 'somePrefix', eventHandlerPropPrefix: false },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `<TestComponent someProp={props.onChange} />`,
      options: [{ eventHandlerPropPrefix: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `<ComponentFromOtherLibraryBar customPropNameBar={handleSomething} />;`,
      options: [
        {
          checkLocalVariables: true,
          ignoreComponentNames: ['ComponentFromOtherLibraryBar'],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        function App() {
          return (
            <div>
              <MyLibInput customPropNameBar={handleSomething} />;
              <MyLibCheckbox customPropNameBar={handleSomething} />;
              <MyLibButtom customPropNameBar={handleSomething} />;
            </div>
          )
        }
      `,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['MyLib*'] },
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Bad handler name (default) ----
    {
      code: `var x = <TestComponent onChange={this.doSomethingOnChange} />`,
      errors: [expectBadHandlerNameError([1, 24, 1, 59])],
    },
    {
      code: `var x = <TestComponent onChange={this.handlerChange} />`,
      errors: [expectBadHandlerNameError([1, 24, 1, 53])],
    },
    {
      code: `var x = <TestComponent onChange={this.handle} />`,
      errors: [expectBadHandlerNameError([1, 24, 1, 46])],
    },
    {
      code: `var x = <TestComponent onChange={this.handle2} />`,
      errors: [expectBadHandlerNameError([1, 24, 1, 47])],
    },
    {
      code: `var x = <TestComponent onChange={this.handl3Change} />`,
      errors: [expectBadHandlerNameError([1, 24, 1, 52])],
    },
    {
      code: `var x = <TestComponent onChange={this.handle4change} />`,
      errors: [expectBadHandlerNameError([1, 24, 1, 53])],
    },

    // ---- checkLocalVariables ----
    {
      code: `var x = <TestComponent onChange={takeCareOfChange} />`,
      options: [{ checkLocalVariables: true }],
      errors: [expectBadHandlerNameError([1, 24, 1, 51])],
    },

    // ---- checkInlineFunction ----
    {
      code: `var x = <TestComponent onChange={() => this.takeCareOfChange()} />`,
      options: [{ checkInlineFunction: true }],
      errors: [expectBadHandlerNameError([1, 24, 1, 64])],
    },

    // ---- Bad prop key — non-`on` paired with `handle*` ----
    {
      code: `var x = <TestComponent only={this.handleChange} />`,
      errors: [expectBadPropKeyError2([1, 24, 1, 48])],
    },
    {
      code: `var x = <TestComponent2 only={this.handleChange} />`,
      errors: [expectBadPropKeyError2([1, 25, 1, 49])],
    },
    {
      code: `var x = <TestComponent handleChange={this.handleChange} />`,
      errors: [expectBadPropKeyError2([1, 24, 1, 56])],
    },
    {
      code: `var x = <TestComponent whenChange={handleChange} />`,
      options: [{ checkLocalVariables: true }],
      errors: [expectBadPropKeyError2([1, 24, 1, 49])],
    },
    {
      code: `var x = <TestComponent whenChange={() => handleChange()} />`,
      options: [{ checkInlineFunction: true, checkLocalVariables: true }],
      errors: [expectBadPropKeyError2([1, 24, 1, 57])],
    },

    // ---- Custom prefixes ----
    {
      code: `var x = <TestComponent onChange={handleChange} />`,
      options: [
        {
          checkLocalVariables: true,
          eventHandlerPrefix: 'handle',
          eventHandlerPropPrefix: 'when',
        },
      ],
      errors: [expectBadPropKeyError3([1, 24, 1, 47])],
    },
    {
      code: `var x = <TestComponent onChange={() => handleChange()} />`,
      options: [
        {
          checkInlineFunction: true,
          checkLocalVariables: true,
          eventHandlerPrefix: 'handle',
          eventHandlerPropPrefix: 'when',
        },
      ],
      errors: [expectBadPropKeyError3([1, 24, 1, 55])],
    },

    // ---- Handler named like prop fails the handle prefix ----
    {
      code: `var x = <TestComponent onChange={this.onChange} />`,
      errors: [expectBadHandlerNameError([1, 24, 1, 48])],
    },

    // ---- ignoreComponentNames non-matching pattern ----
    {
      code: `
        function App() {
          return (
            <div>
              <MyLibInput customPropNameBar={handleInput} />
              <MyLibCheckbox customPropNameBar={handleCheckbox} />
              <MyLibButtom customPropNameBar={handleButton} />
            </div>
          );
        }
      `,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['MyLibrary*'] },
      ],
      errors: [
        diagnostic(
          'badPropKey',
          "Prop key for handleInput must begin with 'on'",
          [5, 27, 5, 58],
        ),
        diagnostic(
          'badPropKey',
          "Prop key for handleCheckbox must begin with 'on'",
          [6, 30, 6, 64],
        ),
        diagnostic(
          'badPropKey',
          "Prop key for handleButton must begin with 'on'",
          [7, 28, 7, 60],
        ),
      ],
    },

    // ---- Namespaced component, ignoreComponentNames patterns don't match ----
    {
      code: `var x = <A.TestComponent onChange={onChange} />`,
      options: [
        {
          checkLocalVariables: true,
          ignoreComponentNames: ['B.TestComponent', 'TestComponent', 'Test*'],
        },
      ],
      errors: [expectBadHandlerNameError([1, 26, 1, 45])],
    },

    // ---- Lock-in: tsgo flat-AST equivalence with parens ----
    {
      code: `var x = <TestComponent onChange={(this.doSomethingOnChange)} />`,
      errors: [expectBadHandlerNameError([1, 24, 1, 61])],
    },

    // ---- Lock-in: optional chain still triggers regex under checkLocal ----
    {
      code: `var x = <TestComponent onChange={this?.takeCareOfChange} />`,
      options: [{ checkLocalVariables: true }],
      errors: [expectBadHandlerNameError([1, 24, 1, 57])],
    },

    // ---- Lock-in: multiple attributes on one element fire independently ----
    {
      code: `var x = <TestComponent disabled onChange={this.bad1} onClick={this.bad2} />`,
      errors: [
        expectBadHandlerNameError([1, 33, 1, 53]),
        diagnostic(
          'badHandlerName',
          "Handler function for onClick prop key must be a camelCase name beginning with 'handle' only",
          [1, 54, 1, 73],
        ),
      ],
    },

    // ---- Glob alignment (invalid side): patterns that DON'T match ----
    {
      code: `var x = <Foo3 onChange={whateverHandler} />`,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['Foo{1,2}'] },
      ],
      errors: [expectBadHandlerNameError([1, 15, 1, 41])],
    },
    {
      code: `var x = <TestD onChange={whateverHandler} />`,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['Test[ABC]'] },
      ],
      errors: [expectBadHandlerNameError([1, 16, 1, 42])],
    },
    {
      code: `var x = <Foo onChange={whateverHandler} />`,
      options: [{ checkLocalVariables: true, ignoreComponentNames: ['!Foo'] }],
      errors: [expectBadHandlerNameError([1, 14, 1, 40])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<TestComponent onChange={this.doSomethingOnChange} />`,
      errors: [expectBadHandlerNameError([1, 16, 1, 51])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<TestComponent onChange={this.handlerChange} />`,
      errors: [expectBadHandlerNameError([1, 16, 1, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `<TestComponent onChange={this.handle} />`,
      errors: [expectBadHandlerNameError([1, 16, 1, 38])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<TestComponent onChange={this.handle2} />`,
      errors: [expectBadHandlerNameError([1, 16, 1, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `<TestComponent onChange={this.handl3Change} />`,
      errors: [expectBadHandlerNameError([1, 16, 1, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `<TestComponent onChange={this.handle4change} />`,
      errors: [expectBadHandlerNameError([1, 16, 1, 45])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<TestComponent onChange={takeCareOfChange} />`,
      options: [{ checkLocalVariables: true }],
      errors: [expectBadHandlerNameError([1, 16, 1, 43])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `<TestComponent onChange={() => this.takeCareOfChange()} />`,
      options: [{ checkInlineFunction: true }],
      errors: [expectBadHandlerNameError([1, 16, 1, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `<TestComponent only={this.handleChange} />`,
      errors: [expectBadPropKeyError2([1, 16, 1, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `<TestComponent2 only={this.handleChange} />`,
      errors: [expectBadPropKeyError2([1, 17, 1, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `<TestComponent handleChange={this.handleChange} />`,
      errors: [expectBadPropKeyError2([1, 16, 1, 48])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `<TestComponent whenChange={handleChange} />`,
      options: [{ checkLocalVariables: true }],
      errors: [expectBadPropKeyError2([1, 16, 1, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `<TestComponent whenChange={() => handleChange()} />`,
      options: [{ checkInlineFunction: true, checkLocalVariables: true }],
      errors: [expectBadPropKeyError2([1, 16, 1, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `<TestComponent onChange={handleChange} />`,
      options: [
        {
          checkLocalVariables: true,
          eventHandlerPrefix: 'handle',
          eventHandlerPropPrefix: 'when',
        },
      ],
      errors: [expectBadPropKeyError3([1, 16, 1, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `<TestComponent onChange={() => handleChange()} />`,
      options: [
        {
          checkInlineFunction: true,
          checkLocalVariables: true,
          eventHandlerPrefix: 'handle',
          eventHandlerPropPrefix: 'when',
        },
      ],
      errors: [expectBadPropKeyError3([1, 16, 1, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `<TestComponent onChange={this.onChange} />`,
      errors: [expectBadHandlerNameError([1, 16, 1, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax ('}' expected.).",
      code: `<TestComponent onChange={props::onChange} />`,
      errors: [expectBadHandlerNameError([1, 16, 1, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      skip: "The native TypeScript parser cannot parse this upstream bind operator syntax ('}' expected.).",
      code: `<TestComponent onChange={props.foo::onChange} />`,
      errors: [expectBadHandlerNameError([1, 16, 1, 46])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        function App() {
          return (
            <div>
              <MyLibInput customPropNameBar={handleInput} />;
              <MyLibCheckbox customPropNameBar={handleCheckbox} />;
              <MyLibButtom customPropNameBar={handleButton} />;
            </div>
          )
        }
      `,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['MyLibrary*'] },
      ],
      errors: [
        diagnostic(
          'badPropKey',
          "Prop key for handleInput must begin with 'on'",
          [5, 27, 5, 58],
        ),
        diagnostic(
          'badPropKey',
          "Prop key for handleCheckbox must begin with 'on'",
          [6, 30, 6, 64],
        ),
        diagnostic(
          'badPropKey',
          "Prop key for handleButton must begin with 'on'",
          [7, 28, 7, 60],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
