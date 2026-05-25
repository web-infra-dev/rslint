import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('button-has-type', {} as never, {
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
  ],
  invalid: [
    {
      code: `<button/>`,
      errors: [{ messageId: 'missingType' }],
    },
    {
      code: `<button type="foo"/>`,
      errors: [
        {
          messageId: 'invalidValue',
          message: `"foo" is an invalid value for button type attribute`,
        },
      ],
    },
    {
      code: `<button type={foo}/>`,
      errors: [{ messageId: 'complexType' }],
    },
    {
      code: `<button type={"foo"}/>`,
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: '<button type={`foo`}/>',
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: '<button type={`button${foo}`}/>',
      errors: [{ messageId: 'complexType' }],
    },
    {
      code: `<button type="reset"/>`,
      options: [{ reset: false }],
      errors: [
        {
          messageId: 'forbiddenValue',
          message: `"reset" is an invalid value for button type attribute`,
        },
      ],
    },
    {
      code: `<button type={condition ? "button" : foo}/>`,
      errors: [{ messageId: 'complexType' }],
    },
    {
      code: `<button type={condition ? "button" : "foo"}/>`,
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: `<button type={condition ? "button" : "reset"}/>`,
      options: [{ reset: false }],
      errors: [{ messageId: 'forbiddenValue' }],
    },
    {
      code: `<button type={condition ? foo : "button"}/>`,
      errors: [{ messageId: 'complexType' }],
    },
    {
      code: `<button type={condition ? "foo" : "button"}/>`,
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: `<button type/>`,
      errors: [
        {
          messageId: 'invalidValue',
          message: `"true" is an invalid value for button type attribute`,
        },
      ],
    },
    {
      code: `React.createElement("button")`,
      errors: [{ messageId: 'missingType' }],
    },
    {
      code: `React.createElement("button", {type: foo})`,
      errors: [{ messageId: 'complexType' }],
    },
    {
      code: `React.createElement("button", {type: "foo"})`,
      errors: [{ messageId: 'invalidValue' }],
    },
    {
      code: `React.createElement("button", {type: "reset"})`,
      options: [{ reset: false }],
      errors: [{ messageId: 'forbiddenValue' }],
    },
    {
      code: `React.createElement("button", {type: condition ? "button" : foo})`,
      errors: [{ messageId: 'complexType' }],
    },
    {
      code: `React.createElement("button", {...extraProps})`,
      errors: [{ messageId: 'missingType' }],
    },
    // Numeric value — normalized like ESLint's String(0x1) === "1"
    {
      code: `<button type={0x1}/>`,
      errors: [
        {
          messageId: 'invalidValue',
          message: `"1" is an invalid value for button type attribute`,
        },
      ],
    },
    // Boolean value expression
    {
      code: `<button type={true}/>`,
      errors: [
        {
          messageId: 'invalidValue',
          message: `"true" is an invalid value for button type attribute`,
        },
      ],
    },
    // Spread without type
    {
      code: `<button {...props}/>`,
      errors: [{ messageId: 'missingType' }],
    },
    // LogicalOr — complex expression
    {
      code: `<button type={foo || "button"}/>`,
      errors: [{ messageId: 'complexType' }],
    },
    // Shorthand object key { type } — value is Identifier → complex
    {
      code: `React.createElement("button", {type})`,
      errors: [{ messageId: 'complexType' }],
    },
    // Ternary with both sides invalid — expect two reports
    {
      code: `<button type={a ? "foo" : "bar"}/>`,
      errors: [{ messageId: 'invalidValue' }, { messageId: 'invalidValue' }],
    },
    // Complex expression types
    {
      code: `<button type={getType()}/>`,
      errors: [{ messageId: 'complexType' }],
    },
    {
      code: `<button type={obj.type}/>`,
      errors: [{ messageId: 'complexType' }],
    },
    {
      code: `<button type={arr[0]}/>`,
      errors: [{ messageId: 'complexType' }],
    },
    // TS `as` assertion — not unwrapped → complex
    {
      code: `<button type={foo as "button"}/>`,
      errors: [{ messageId: 'complexType' }],
    },
    // createElement with non-object second argument → missingType
    {
      code: `React.createElement("button", null)`,
      errors: [{ messageId: 'missingType' }],
    },
    {
      code: `React.createElement("button", "foo")`,
      errors: [{ messageId: 'missingType' }],
    },
    // BigInt literal — decimal-normalized
    {
      code: `<button type={1n}/>`,
      errors: [
        {
          messageId: 'invalidValue',
          message: `"1" is an invalid value for button type attribute`,
        },
      ],
    },
    // Paren-wrapped callee / args — ESTree-flattening parity
    {
      code: `(React).createElement("button")`,
      errors: [{ messageId: 'missingType' }],
    },
    {
      code: `(React.createElement)("button")`,
      errors: [{ messageId: 'missingType' }],
    },
    {
      code: `React.createElement(("button"))`,
      errors: [{ messageId: 'missingType' }],
    },
    {
      code: `React.createElement("button", ({type: "foo"}))`,
      errors: [
        {
          messageId: 'invalidValue',
          message: `"foo" is an invalid value for button type attribute`,
        },
      ],
    },
  ],
});
