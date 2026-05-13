import { RuleTester } from '../rule-tester';

// failMessage mirrors axe-core's `lib/checks/forms/autocomplete-valid.json`
// `messages.fail` — the upstream ESLint rule extracts the same string via
// `violations[0].nodes[0].all[0].message`.
const failMessage = 'the autocomplete attribute is incorrectly formatted';

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Input: 'input',
    },
  },
};

new RuleTester().run('autocomplete-valid', null as never, {
  valid: [
    // Upstream `valid` block — INAPPLICABLE / PASSED AUTOCOMPLETE.
    { code: '<input type="text" />;' },
    { code: '<input type="text" autocomplete="name" />;' },
    { code: '<input type="text" autocomplete="" />;' },
    { code: '<input type="text" autocomplete="off" />;' },
    { code: '<input type="text" autocomplete="on" />;' },
    { code: '<input type="text" autocomplete="billing family-name" />;' },
    {
      code: '<input type="text" autocomplete="section-blue shipping street-address" />;',
    },
    {
      code: '<input type="text" autocomplete="section-somewhere shipping work email" />;',
    },
    { code: '<input type="text" autocomplete />;' },
    { code: '<input type="text" autocomplete={dynValue} />;' },
    { code: '<input type="text" autocomplete={dynValue || "name"} />;' },
    { code: '<input type="text" autocomplete={dynValue || "foo"} />;' },
    { code: '<Foo autocomplete="bar"></Foo>;' },
    {
      code: '<input type={isEmail ? "email" : "text"} autocomplete="none" />;',
    },
    {
      code: '<Input type="text" autocomplete="name" />',
      settings: componentsSettings,
    },
    { code: '<Input type="text" autocomplete="baz" />' },
    // Upstream `valid` block — PASSED "autocomplete-appropriate" (these are
    // syntactically valid; only the separate autocomplete-appropriate axe-core
    // rule would flag them, and the ESLint plugin doesn't run that rule).
    { code: '<input type="date" autocomplete="email" />;' },
    { code: '<input type="number" autocomplete="url" />;' },
    { code: '<input type="month" autocomplete="tel" />;' },
    {
      code: '<Foo type="month" autocomplete="tel"></Foo>;',
      options: [{ inputComponents: ['Foo'] }],
    },

    // Extra rslint lockdowns:
    // axe-core's extended stateTerms — every entry must accept.
    { code: '<input autocomplete="none" />;' },
    { code: '<input autocomplete="false" />;' },
    { code: '<input autocomplete="true" />;' },
    { code: '<input autocomplete="disabled" />;' },
    { code: '<input autocomplete="enabled" />;' },
    { code: '<input autocomplete="undefined" />;' },
    { code: '<input autocomplete="null" />;' },
    { code: '<input autocomplete="xoff" />;' },
    { code: '<input autocomplete="xon" />;' },
    // ignoredValues — axe-core returns "incomplete" (undefined), not a
    // violation, so the ESLint plugin reports nothing.
    { code: '<input autocomplete="text" />;' },
    { code: '<input autocomplete="pronouns" />;' },
    { code: '<input autocomplete="gender" />;' },
    { code: '<input autocomplete="message" />;' },
    { code: '<input autocomplete="content" />;' },
    // Case-insensitivity.
    { code: '<input autocomplete="NAME" />;' },
    { code: '<input autocomplete="OFF" />;' },
    { code: '<input autocomplete="Section-Blue Shipping Street-Address" />;' },
    // Whitespace handling: trim + \s+ split.
    { code: '<input autocomplete="   name   " />;' },
    { code: '<input autocomplete="  " />;' },
    { code: '<input autocomplete="billing  family-name" />;' },
    // webauthn after a valid token.
    { code: '<input autocomplete="name webauthn" />;' },
    { code: '<input autocomplete="shipping street-address webauthn" />;' },
    { code: '<input autocomplete="home email webauthn" />;' },
    // Qualified term combinations.
    { code: '<input autocomplete="home tel" />;' },
    { code: '<input autocomplete="work email" />;' },
    { code: '<input autocomplete="mobile tel-extension" />;' },
    { code: '<input autocomplete="fax impp" />;' },
    { code: '<input autocomplete="pager tel-country-code" />;' },
    // All optional prefixes used at once.
    { code: '<input autocomplete="section-payment billing home tel" />;' },
    // Section- length boundary on the valid side: 9-char prefix stripped.
    { code: '<input autocomplete="section-x name" />;' },
    // JSX shape variants.
    { code: '<input autocomplete="name"></input>' },
    { code: '<input autocomplete="name"/>' },
    // Spread props mixed with literal autocomplete.
    { code: '<input {...rest} autocomplete="name" />;' },
    // JsxExpression wrapping a literal.
    { code: '<input autocomplete={"name"} />;' },
    { code: '<input autocomplete={`name`} />;' },
    // jsx-ast-utils' TemplateLiteral.js extracts `${undefined}` as the
    // bare string "undefined", which axe-core's extended stateTerms
    // accepts as a valid state term. Aligned with upstream verbatim.
    { code: '<input autocomplete={`${undefined}`} />;' },
    // TS wrappers.
    { code: '<input autocomplete={"name" as const} />;' },
    { code: '<input autocomplete={("name")} />;' },
    // Components map points at a non-input HTML tag.
    {
      code: '<Block autocomplete="foo" />',
      settings: { 'jsx-a11y': { components: { Block: 'div' } } },
    },
    // polymorphicPropName resolves a polymorphic component to a non-input.
    {
      code: '<Box as="div" autocomplete="foo" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    // Non-input elements with autocomplete are silently passed through.
    { code: '<select autocomplete="foo" />;' },
    { code: '<textarea autocomplete="foo" />;' },
    { code: '<button autocomplete="foo" />;' },
    // Member-expression / namespaced tag names — neither matches "input".
    { code: '<Foo.input autocomplete="foo" />;' },
    { code: '<svg:input autocomplete="foo" />;' },
    // Empty inputComponents array.
    { code: '<Foo autocomplete="foo" />;', options: [{ inputComponents: [] }] },
    // Literal `null` autocomplete coerces to the string "null" (in stateTerms).
    { code: '<input autocomplete={null} />;' },
    // CallExpression / MemberExpression / Conditional inside JsxExpression.
    { code: '<input autocomplete={getValue()} />;' },
    { code: '<input autocomplete={config.autocomplete} />;' },
    { code: '<input autocomplete={cond ? "name" : "foo"} />;' },
    // Multi-line JSX with valid literal autocomplete.
    { code: '<input\n  type="text"\n  autocomplete="name"\n/>' },

    // axe-core matches() type-filter: excluded input types skip the check
    // even with an invalid autocomplete value.
    { code: '<input type="hidden" autocomplete="foo" />;' },
    { code: '<input type="submit" autocomplete="foo" />;' },
    { code: '<input type="reset" autocomplete="invalid garbage" />;' },
    { code: '<input type="button" autocomplete="bogus" />;' },
    // Case-insensitive on type comparison.
    { code: '<input type="HIDDEN" autocomplete="foo" />;' },
    { code: '<input type="Submit" autocomplete="foo" />;' },
    // Literal type via JsxExpression / template.
    { code: '<input type={"hidden"} autocomplete="foo" />;' },
    { code: '<input type={`hidden`} autocomplete="foo" />;' },
    // TS-wrapped autocomplete value: jsx-ast-utils' LITERAL_TYPES maps
    // TSAsExpression / TSSatisfiesExpression to noop → null, so the rule
    // returns early without checking the value.
    { code: '<input autocomplete={"foo" satisfies string} />' },
    { code: '<input {...{autocomplete: "foo" as const}} />' },
    // Custom inputComponent with excluded type.
    {
      code: '<MyInput type="hidden" autocomplete="foo" />;',
      options: [{ inputComponents: ['MyInput'] }],
    },
    // Components-map mapping with excluded type.
    {
      code: '<Input type="submit" autocomplete="baz" />',
      settings: componentsSettings,
    },

    // Spread-of-literal lockdowns.
    { code: '<input {...{autocomplete: "name"}} />;' },
    { code: '<input {...{autocomplete}} />;' },

    // Logical short-circuit lockdowns.
    { code: '<input autocomplete={cond && "name"} />;' },
    { code: '<input autocomplete={x ?? "name"} />;' },

    // Deeply-nested JSX patterns.
    {
      code: 'function L({xs}) { return xs.map(x => <input autocomplete="name" key={x} />); }',
    },
    {
      code: 'function F() { return <div>{cond && <input autocomplete="name" />}</div>; }',
    },
    {
      code: 'function f<T>(x: T) { return <input autocomplete="name" />; }',
    },
    { code: '<form><input autocomplete="name" /></form>' },
    { code: '<><><input autocomplete="name" /></></>' },
    {
      code: 'function F() { useEffect(() => { renderer(<input autocomplete="name" />); }); }',
    },

    // Tab / newline whitespace handling.
    { code: '<input autocomplete="home\ttel" />' },
    { code: '<input autocomplete="name\nwebauthn" />' },

    // Attribute-name case-insensitive match — both HTML-correct and
    // React-camelCase forms must trigger.
    { code: '<input autoComplete="name" />;' },
    { code: '<input AUTOCOMPLETE="name" />;' },
    { code: '<input AutoComplete="name" />;' },
    { code: '<input autoComplete={dynamicValue} />;' },

    // Real-world login form patterns: every standalone term we ship.
    { code: '<input type="text" autocomplete="username" />;' },
    { code: '<input type="password" autocomplete="current-password" />;' },
    { code: '<input type="password" autocomplete="new-password" />;' },
    { code: '<input type="text" autocomplete="one-time-code" />;' },
    { code: '<input type="text" autocomplete="street-address" />;' },
    { code: '<input type="text" autocomplete="address-line1" />;' },
    { code: '<input type="text" autocomplete="address-line2" />;' },
    { code: '<input type="text" autocomplete="address-line3" />;' },
    { code: '<input type="text" autocomplete="address-level1" />;' },
    { code: '<input type="text" autocomplete="address-level2" />;' },
    { code: '<input type="text" autocomplete="address-level3" />;' },
    { code: '<input type="text" autocomplete="address-level4" />;' },
    { code: '<input type="text" autocomplete="country" />;' },
    { code: '<input type="text" autocomplete="country-name" />;' },
    { code: '<input type="text" autocomplete="postal-code" />;' },
    { code: '<input type="text" autocomplete="honorific-prefix" />;' },
    { code: '<input type="text" autocomplete="given-name" />;' },
    { code: '<input type="text" autocomplete="additional-name" />;' },
    { code: '<input type="text" autocomplete="family-name" />;' },
    { code: '<input type="text" autocomplete="honorific-suffix" />;' },
    { code: '<input type="text" autocomplete="nickname" />;' },
    { code: '<input type="text" autocomplete="organization-title" />;' },
    { code: '<input type="text" autocomplete="organization" />;' },
    { code: '<input type="date" autocomplete="bday" />;' },
    { code: '<input type="number" autocomplete="bday-day" />;' },
    { code: '<input type="number" autocomplete="bday-month" />;' },
    { code: '<input type="number" autocomplete="bday-year" />;' },
    { code: '<input type="text" autocomplete="sex" />;' },
    { code: '<input type="text" autocomplete="cc-name" />;' },
    { code: '<input type="text" autocomplete="cc-given-name" />;' },
    { code: '<input type="text" autocomplete="cc-additional-name" />;' },
    { code: '<input type="text" autocomplete="cc-family-name" />;' },
    { code: '<input type="text" autocomplete="cc-number" />;' },
    { code: '<input type="text" autocomplete="cc-exp" />;' },
    { code: '<input type="text" autocomplete="cc-exp-month" />;' },
    { code: '<input type="text" autocomplete="cc-exp-year" />;' },
    { code: '<input type="text" autocomplete="cc-csc" />;' },
    { code: '<input type="text" autocomplete="cc-type" />;' },
    { code: '<input type="text" autocomplete="transaction-currency" />;' },
    { code: '<input type="number" autocomplete="transaction-amount" />;' },
    { code: '<input type="text" autocomplete="language" />;' },
    { code: '<input type="url" autocomplete="url" />;' },
    { code: '<input type="url" autocomplete="photo" />;' },

    // Real-world component patterns.
    {
      code: 'function LoginForm() { return (<form><input type="text" autocomplete="username" /><input type="password" autocomplete="current-password" /><button type="submit">Login</button></form>); }',
    },
    {
      code: 'function DynamicForm({ fields }) { return fields.map(f => <input key={f.id} autocomplete={f.autocomplete} />); }',
    },
    {
      code: 'const Input = forwardRef((props, ref) => <input {...props} ref={ref} autocomplete="name" />);',
    },
    {
      code: '<form><fieldset><legend>Address</legend><input autocomplete="street-address" /></fieldset></form>',
    },
    {
      code: 'const inputs = [<input autocomplete="name" key="1" />, <input autocomplete="email" key="2" />];',
    },
    {
      code: 'const map = { name: <input autocomplete="name" />, email: <input autocomplete="email" /> };',
    },
    {
      code: 'function F(child = <input autocomplete="name" />) { return child; }',
    },

    // Spread interaction patterns (first wins).
    { code: '<input autocomplete="name" {...rest} />;' },
    { code: '<input {...a} {...b} autocomplete="name" />;' },
    { code: '<input {...{key: "x", autocomplete: "name"}} />;' },
    { code: '<input autocomplete="name" autocomplete="foo" />;' },

    // Optional chaining and complex expression shapes.
    { code: '<input autocomplete={config?.autocomplete} />;' },
    { code: '<input autocomplete={fn?.()} />;' },
    { code: '<input autocomplete={a?.b?.c?.d} />;' },
    { code: '<input autocomplete={tag`name`} />;' },
    { code: '<input autocomplete={new String("name")} />;' },
    { code: '<input autocomplete={this.value} />;' },
    { code: '<input autocomplete={class {}} />;' },
    { code: '<input autocomplete={["name"]} />;' },
    { code: '<input autocomplete={{}} />;' },
    { code: '<input autocomplete={"name" satisfies string} />;' },
    { code: '<input autocomplete={value!} />;' },
    { code: '<input autocomplete={(("name" as string) satisfies any)} />;' },
    { code: '<input autocomplete={/* leading */ "name" /* trailing */} />;' },
  ],
  invalid: [
    // Upstream `invalid` block.
    {
      code: '<input type="text" autocomplete="foo" />;',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input type="text" autocomplete="name invalid" />;',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input type="text" autocomplete="invalid name" />;',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input type="text" autocomplete="home url" />;',
      errors: [{ message: failMessage }],
    },
    {
      code: '<Bar autocomplete="baz"></Bar>;',
      options: [{ inputComponents: ['Bar'] }],
      errors: [{ message: failMessage }],
    },
    {
      code: '<Input type="text" autocomplete="baz" />',
      settings: componentsSettings,
      errors: [{ message: failMessage }],
    },

    // Extra grammar-boundary lockdowns.
    // webauthn alone is invalid.
    {
      code: '<input autocomplete="webauthn" />;',
      errors: [{ message: failMessage }],
    },
    // section- alone (after stripping prefix) leaves no purpose token.
    {
      code: '<input autocomplete="section-foo" />;',
      errors: [{ message: failMessage }],
    },
    // Bare "section-" (8 chars, NOT stripped — length must be > 8).
    {
      code: '<input autocomplete="section-" />;',
      errors: [{ message: failMessage }],
    },
    // Two consecutive locations.
    {
      code: '<input autocomplete="billing shipping name" />;',
      errors: [{ message: failMessage }],
    },
    // Two consecutive qualifiers.
    {
      code: '<input autocomplete="home work tel" />;',
      errors: [{ message: failMessage }],
    },
    // qualifier + standaloneTerm — only qualifiedTerms allowed after a qualifier.
    {
      code: '<input autocomplete="home name" />;',
      errors: [{ message: failMessage }],
    },
    // qualifier alone — no purpose token.
    {
      code: '<input autocomplete="home" />;',
      errors: [{ message: failMessage }],
    },
    // location alone — no purpose token.
    {
      code: '<input autocomplete="billing" />;',
      errors: [{ message: failMessage }],
    },
    // section + location + qualifier without final field-name token.
    {
      code: '<input autocomplete="section-foo shipping work" />;',
      errors: [{ message: failMessage }],
    },
    // <Input autocomplete="baz"> WITHOUT a type attribute, with components map
    // — locks that the type attribute does not affect autocomplete-valid.
    {
      code: '<Input autocomplete="baz" />',
      settings: componentsSettings,
      errors: [{ message: failMessage }],
    },
    // polymorphic resolves to <input>.
    {
      code: '<Box as="input" autocomplete="foo" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [{ message: failMessage }],
    },
    // Multiple invalid <input> elements at the same level.
    {
      code: '<><input autocomplete="foo" /><input autocomplete="bar" /></>',
      errors: [{ message: failMessage }, { message: failMessage }],
    },
    // Mixed valid + invalid — only the invalid one reports.
    {
      code: '<><input autocomplete="name" /><input autocomplete="foo" /></>',
      errors: [{ message: failMessage }],
    },
    // Uppercase invalid value.
    {
      code: '<input autocomplete="FOO" />',
      errors: [{ message: failMessage }],
    },
    // Whitespace around an invalid value.
    {
      code: '<input autocomplete="  foo  " />',
      errors: [{ message: failMessage }],
    },
    // Paired (non-self-closing) form.
    {
      code: '<input autocomplete="foo"></input>',
      errors: [{ message: failMessage }],
    },
    // Multi-line invalid attribute.
    {
      code: '<input\n  autocomplete="foo"\n/>',
      errors: [{ message: failMessage }],
    },
    // Spread-of-literal with invalid autocomplete — must report.
    {
      code: '<input {...{autocomplete: "foo"}} />',
      errors: [{ message: failMessage }],
    },
    // Non-excluded type still reports.
    {
      code: '<input type="email" autocomplete="foo" />',
      errors: [{ message: failMessage }],
    },
    // Dynamic type → undefined to axe-core → check still runs.
    {
      code: '<input type={cond ? "hidden" : "text"} autocomplete="foo" />',
      errors: [{ message: failMessage }],
    },
    // Boolean type form — getLiteralPropValue returns true (not a string),
    // so the type filter doesn't skip.
    {
      code: '<input type autocomplete="foo" />',
      errors: [{ message: failMessage }],
    },
    // TemplateExpression with substitutions — synthesized placeholder string fails grammar.
    {
      code: '<input autocomplete={`name${suffix}`} />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input autocomplete={`${dynVar}`} />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input autocomplete={`name ${suffix}`} />',
      errors: [{ message: failMessage }],
    },

    // Attribute name case variants.
    {
      code: '<input autoComplete="foo" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input AUTOCOMPLETE="foo" />',
      errors: [{ message: failMessage }],
    },

    // Duplicate attribute first wins.
    {
      code: '<input autocomplete="foo" autocomplete="name" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input {...{autocomplete: "foo"}} autocomplete="name" />',
      errors: [{ message: failMessage }],
    },

    // disabled / readonly / aria-* / tabindex / role NOT excluded —
    // upstream's runVirtualRule call doesn't forward them to axe-core.
    {
      code: '<input disabled autocomplete="foo" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input readOnly autocomplete="foo" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input aria-disabled="true" autocomplete="foo" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input aria-readonly="true" autocomplete="foo" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input tabIndex={-1} autocomplete="foo" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input role="presentation" autocomplete="foo" />',
      errors: [{ message: failMessage }],
    },

    // Type via spread that resolves to a non-excluded type → check runs.
    {
      code: '<input {...{type: "text"}} autocomplete="foo" />',
      errors: [{ message: failMessage }],
    },

    // Multi-element with progressively complex bad values.
    {
      code: '<><input autocomplete="foo" /><input autocomplete="name invalid" /><input autocomplete="home url" /></>',
      errors: [
        { message: failMessage },
        { message: failMessage },
        { message: failMessage },
      ],
    },

    // Grammar boundaries.
    {
      code: '<input autocomplete="webauthn webauthn" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input autocomplete="name section-foo" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input autocomplete="section-a section-b name" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input autocomplete="name extra webauthn" />',
      errors: [{ message: failMessage }],
    },
    {
      code: '<input autocomplete={/* leading */ "foo" /* trailing */} />',
      errors: [{ message: failMessage }],
    },
    // `type={"hidden" as const}` — jsx-ast-utils' LITERAL_TYPES maps
    // TSAsExpression to noop → null, so the type is unknown and the
    // autocomplete value gets validated → reported.
    {
      code: '<input type={"hidden" as const} autocomplete="foo" />;',
      errors: [{ message: failMessage }],
    },
  ],
});
