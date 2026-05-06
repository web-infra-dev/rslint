import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-handler-names', {} as never, {
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
  ],
  invalid: [
    // ---- Bad handler name (default) ----
    {
      code: `var x = <TestComponent onChange={this.doSomethingOnChange} />`,
      errors: [
        {
          messageId: 'badHandlerName',
          message:
            "Handler function for onChange prop key must be a camelCase name beginning with 'handle' only",
        },
      ],
    },
    {
      code: `var x = <TestComponent onChange={this.handlerChange} />`,
      errors: [{ messageId: 'badHandlerName' }],
    },
    {
      code: `var x = <TestComponent onChange={this.handle} />`,
      errors: [{ messageId: 'badHandlerName' }],
    },
    {
      code: `var x = <TestComponent onChange={this.handle2} />`,
      errors: [{ messageId: 'badHandlerName' }],
    },
    {
      code: `var x = <TestComponent onChange={this.handl3Change} />`,
      errors: [{ messageId: 'badHandlerName' }],
    },
    {
      code: `var x = <TestComponent onChange={this.handle4change} />`,
      errors: [{ messageId: 'badHandlerName' }],
    },

    // ---- checkLocalVariables ----
    {
      code: `var x = <TestComponent onChange={takeCareOfChange} />`,
      options: [{ checkLocalVariables: true }],
      errors: [{ messageId: 'badHandlerName' }],
    },

    // ---- checkInlineFunction ----
    {
      code: `var x = <TestComponent onChange={() => this.takeCareOfChange()} />`,
      options: [{ checkInlineFunction: true }],
      errors: [{ messageId: 'badHandlerName' }],
    },

    // ---- Bad prop key — non-`on` paired with `handle*` ----
    {
      code: `var x = <TestComponent only={this.handleChange} />`,
      errors: [
        {
          messageId: 'badPropKey',
          message: "Prop key for handleChange must begin with 'on'",
        },
      ],
    },
    {
      code: `var x = <TestComponent2 only={this.handleChange} />`,
      errors: [{ messageId: 'badPropKey' }],
    },
    {
      code: `var x = <TestComponent handleChange={this.handleChange} />`,
      errors: [{ messageId: 'badPropKey' }],
    },
    {
      code: `var x = <TestComponent whenChange={handleChange} />`,
      options: [{ checkLocalVariables: true }],
      errors: [{ messageId: 'badPropKey' }],
    },
    {
      code: `var x = <TestComponent whenChange={() => handleChange()} />`,
      options: [{ checkInlineFunction: true, checkLocalVariables: true }],
      errors: [{ messageId: 'badPropKey' }],
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
      errors: [
        {
          messageId: 'badPropKey',
          message: "Prop key for handleChange must begin with 'when'",
        },
      ],
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
      errors: [{ messageId: 'badPropKey' }],
    },

    // ---- Handler named like prop fails the handle prefix ----
    {
      code: `var x = <TestComponent onChange={this.onChange} />`,
      errors: [{ messageId: 'badHandlerName' }],
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
        { messageId: 'badPropKey' },
        { messageId: 'badPropKey' },
        { messageId: 'badPropKey' },
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
      errors: [{ messageId: 'badHandlerName' }],
    },

    // ---- Lock-in: tsgo flat-AST equivalence with parens ----
    {
      code: `var x = <TestComponent onChange={(this.doSomethingOnChange)} />`,
      errors: [{ messageId: 'badHandlerName' }],
    },

    // ---- Lock-in: optional chain still triggers regex under checkLocal ----
    {
      code: `var x = <TestComponent onChange={this?.takeCareOfChange} />`,
      options: [{ checkLocalVariables: true }],
      errors: [{ messageId: 'badHandlerName' }],
    },

    // ---- Lock-in: multiple attributes on one element fire independently ----
    {
      code: `var x = <TestComponent disabled onChange={this.bad1} onClick={this.bad2} />`,
      errors: [
        { messageId: 'badHandlerName' },
        { messageId: 'badHandlerName' },
      ],
    },

    // ---- Glob alignment (invalid side): patterns that DON'T match ----
    {
      code: `var x = <Foo3 onChange={whateverHandler} />`,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['Foo{1,2}'] },
      ],
      errors: [{ messageId: 'badHandlerName' }],
    },
    {
      code: `var x = <TestD onChange={whateverHandler} />`,
      options: [
        { checkLocalVariables: true, ignoreComponentNames: ['Test[ABC]'] },
      ],
      errors: [{ messageId: 'badHandlerName' }],
    },
    {
      code: `var x = <Foo onChange={whateverHandler} />`,
      options: [{ checkLocalVariables: true, ignoreComponentNames: ['!Foo'] }],
      errors: [{ messageId: 'badHandlerName' }],
    },
  ],
});
