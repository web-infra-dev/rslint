import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectMissingTypeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingType',
    message: 'Missing an explicit type attribute for button',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectInvalidValueError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'invalidValue',
    message: '"foo" is an invalid value for button type attribute',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectComplexTypeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'complexType',
    message:
      'The button type attribute must be specified by a static string or a trivial ternary expression',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectForbiddenValueError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'forbiddenValue',
    message: '"reset" is an invalid value for button type attribute',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('button-has-type', {
  valid: [
    { code: `<span/>` },
    { code: `<span type="foo"/>` },
    { code: `<button type="button"/>` },
    { code: `<button type="submit"/>` },
    { code: `<button type="reset"/>` },
    { code: `<button type={"button"}/>` },
    { code: `<button type={'button'}/>` },
    { code: '<button type={`button`}/>' },
    { code: `<button type={condition ? "button" : "submit"}/>` },
    { code: `<button type={condition ? 'button' : 'submit'}/>` },
    { code: '<button type={condition ? `button` : `submit`}/>' },
    { code: `<button type="button"/>`, options: [{ reset: false }] },
    { code: `React.createElement("span")` },
    { code: `React.createElement("span", {type: "foo"})` },
    { code: `React.createElement("button", {type: "button"})` },
    { code: `React.createElement("button", {type: 'button'})` },
    { code: 'React.createElement("button", {type: `button`})' },
    { code: `React.createElement("button", {type: "submit"})` },
    { code: `React.createElement("button", {type: "reset"})` },
    {
      code: `React.createElement("button", {type: condition ? "button" : "submit"})`,
    },
    {
      code: `React.createElement("button", {type: "button"})`,
      options: [{ reset: false }],
    },
    { code: `document.createElement("button")` },
    // Paren-wrapped expression
    { code: `<button type={("button")}/>` },
    // Nested ternary
    { code: `<button type={a ? "button" : b ? "submit" : "reset"}/>` },
    // Button with children
    { code: `<button type="button">Click</button>` },
    // Spread mixed with valid type (order doesn't matter for our detection)
    { code: `<button type="button" {...props}/>` },
    { code: `<button {...props} type="button"/>` },
    // Not a button
    { code: `<Button/>` },
    { code: `<div/>` },
    // createElement first arg is Component, not "button"
    { code: `React.createElement(Component, {type: "foo"})` },
    // Spread before explicit type prop
    { code: `React.createElement("button", {...extraProps, type: "button"})` },
    // Paren-wrapped callee / args — ESTree-flattening parity
    { code: `(React).createElement("button", {type: "button"})` },
    { code: `(React.createElement)("button", {type: "button"})` },
    { code: `React.createElement(("button"), {type: "button"})` },
    { code: `React.createElement("button", ({type: "button"}))` },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `React.createElement("button", {type: 'submit'})`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `React.createElement("button", {type: \`submit\`})`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `React.createElement("button", {type: 'reset'})`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `React.createElement("button", {type: \`reset\`})`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `React.createElement("button", {type: condition ? 'button' : 'submit'})`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `React.createElement("button", {type: condition ? \`button\` : \`submit\`})`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `Foo.createElement("span")`,
      settings: { react: { pragma: 'Foo' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        function MyComponent(): ReactElement {
          const buttonProps: (Required<Attributes> & ButtonHTMLAttributes<HTMLButtonElement>)[] = [
            {
              children: 'test',
              key: 'test',
              onClick: (): void => {
                return;
              },
            },
          ];

          return <>
            {
              buttonProps.map(
                ({ key, ...props }: Required<Attributes> & ButtonHTMLAttributes<HTMLButtonElement>): ReactElement =>
                  <button key={key} type="button" {...props} />
              )
            }
          </>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<button/>`,
      errors: [expectMissingTypeError([1, 1, 1, 10])],
    },
    {
      code: `<button type="foo"/>`,
      errors: [expectInvalidValueError([1, 1, 1, 21])],
    },
    {
      code: `<button type={foo}/>`,
      errors: [expectComplexTypeError([1, 15, 1, 18])],
    },
    {
      code: `<button type={"foo"}/>`,
      errors: [expectInvalidValueError([1, 1, 1, 23])],
    },
    {
      code: '<button type={`foo`}/>',
      errors: [expectInvalidValueError([1, 1, 1, 23])],
    },
    {
      code: '<button type={`button${foo}`}/>',
      errors: [expectComplexTypeError([1, 15, 1, 29])],
    },
    {
      code: `<button type="reset"/>`,
      options: [{ reset: false }],
      errors: [expectForbiddenValueError([1, 1, 1, 23])],
    },
    {
      code: `<button type={condition ? "button" : foo}/>`,
      errors: [expectComplexTypeError([1, 38, 1, 41])],
    },
    {
      code: `<button type={condition ? "button" : "foo"}/>`,
      errors: [expectInvalidValueError([1, 1, 1, 46])],
    },
    {
      code: `<button type={condition ? "button" : "reset"}/>`,
      options: [{ reset: false }],
      errors: [expectForbiddenValueError([1, 1, 1, 48])],
    },
    {
      code: `<button type={condition ? foo : "button"}/>`,
      errors: [expectComplexTypeError([1, 27, 1, 30])],
    },
    {
      code: `<button type={condition ? "foo" : "button"}/>`,
      errors: [expectInvalidValueError([1, 1, 1, 46])],
    },
    {
      code: `<button type/>`,
      errors: [
        {
          messageId: 'invalidValue',
          message: '"true" is an invalid value for button type attribute',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: `React.createElement("button")`,
      errors: [expectMissingTypeError([1, 1, 1, 30])],
    },
    {
      code: `React.createElement("button", {type: foo})`,
      errors: [expectComplexTypeError([1, 38, 1, 41])],
    },
    {
      code: `React.createElement("button", {type: "foo"})`,
      errors: [expectInvalidValueError([1, 1, 1, 45])],
    },
    {
      code: `React.createElement("button", {type: "reset"})`,
      options: [{ reset: false }],
      errors: [expectForbiddenValueError([1, 1, 1, 47])],
    },
    {
      code: `React.createElement("button", {type: condition ? "button" : foo})`,
      errors: [expectComplexTypeError([1, 61, 1, 64])],
    },
    {
      code: `React.createElement("button", {...extraProps})`,
      errors: [expectMissingTypeError([1, 1, 1, 47])],
    },
    // Numeric value — normalized like ESLint's String(0x1) === "1"
    {
      code: `<button type={0x1}/>`,
      errors: [
        {
          messageId: 'invalidValue',
          message: '"1" is an invalid value for button type attribute',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    // Boolean value expression
    {
      code: `<button type={true}/>`,
      errors: [
        {
          messageId: 'invalidValue',
          message: '"true" is an invalid value for button type attribute',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    // Spread without type
    {
      code: `<button {...props}/>`,
      errors: [expectMissingTypeError([1, 1, 1, 21])],
    },
    // LogicalOr — complex expression
    {
      code: `<button type={foo || "button"}/>`,
      errors: [expectComplexTypeError([1, 15, 1, 30])],
    },
    // Shorthand object key { type } — value is Identifier → complex
    {
      code: `React.createElement("button", {type})`,
      errors: [expectComplexTypeError([1, 32, 1, 36])],
    },
    // Ternary with both sides invalid — expect two reports
    {
      code: `<button type={a ? "foo" : "bar"}/>`,
      errors: [
        expectInvalidValueError([1, 1, 1, 35]),
        {
          messageId: 'invalidValue',
          message: '"bar" is an invalid value for button type attribute',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    // Complex expression types
    {
      code: `<button type={getType()}/>`,
      errors: [expectComplexTypeError([1, 15, 1, 24])],
    },
    {
      code: `<button type={obj.type}/>`,
      errors: [expectComplexTypeError([1, 15, 1, 23])],
    },
    {
      code: `<button type={arr[0]}/>`,
      errors: [expectComplexTypeError([1, 15, 1, 21])],
    },
    // TS `as` assertion — not unwrapped → complex
    {
      code: `<button type={foo as "button"}/>`,
      errors: [expectComplexTypeError([1, 15, 1, 30])],
    },
    // createElement with non-object second argument → missingType
    {
      code: `React.createElement("button", null)`,
      errors: [expectMissingTypeError([1, 1, 1, 36])],
    },
    {
      code: `React.createElement("button", "foo")`,
      errors: [expectMissingTypeError([1, 1, 1, 37])],
    },
    // BigInt literal — decimal-normalized
    {
      code: `<button type={1n}/>`,
      errors: [
        {
          messageId: 'invalidValue',
          message: '"1" is an invalid value for button type attribute',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    // Paren-wrapped callee / args — ESTree-flattening parity
    {
      code: `(React).createElement("button")`,
      errors: [expectMissingTypeError([1, 1, 1, 32])],
    },
    {
      code: `(React.createElement)("button")`,
      errors: [expectMissingTypeError([1, 1, 1, 32])],
    },
    {
      code: `React.createElement(("button"))`,
      errors: [expectMissingTypeError([1, 1, 1, 32])],
    },
    {
      code: `React.createElement("button", ({type: "foo"}))`,
      errors: [expectInvalidValueError([1, 1, 1, 47])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `<button type={'foo'}/>`,
      errors: [expectInvalidValueError([1, 1, 1, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `<button type={condition ? "reset" : "button"}/>`,
      options: [{ reset: false }],
      errors: [expectForbiddenValueError([1, 1, 1, 48])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `React.createElement("button", {type: condition ? "button" : "foo"})`,
      errors: [expectInvalidValueError([1, 1, 1, 68])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `React.createElement("button", {type: condition ? "button" : "reset"})`,
      options: [{ reset: false }],
      errors: [expectForbiddenValueError([1, 1, 1, 70])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `React.createElement("button", {type: condition ? foo : "button"})`,
      errors: [expectComplexTypeError([1, 50, 1, 53])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `React.createElement("button", {type: condition ? "foo" : "button"})`,
      errors: [expectInvalidValueError([1, 1, 1, 68])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `React.createElement("button", {type: condition ? "reset" : "button"})`,
      options: [{ reset: false }],
      errors: [expectForbiddenValueError([1, 1, 1, 70])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `Foo.createElement("button")`,
      settings: { react: { pragma: 'Foo' } },
      errors: [expectMissingTypeError([1, 1, 1, 28])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `function Button({ type, ...extraProps }) { const button = type; return <button type={button} {...extraProps} />; }`,
      errors: [expectComplexTypeError([1, 86, 1, 92])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
