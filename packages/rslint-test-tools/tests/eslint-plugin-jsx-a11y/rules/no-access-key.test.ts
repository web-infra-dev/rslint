import { RuleTester } from '../rule-tester';

const errorMessage =
  'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.';
const expectedError = { message: errorMessage };

new RuleTester().run('no-access-key', null as never, {
  valid: [
    // ---- Upstream valid ----
    { code: '<div />;' },
    { code: '<div {...props} />' },
    { code: '<div accessKey={undefined} />' },

    // ---- rslint extras: falsy literal values ----
    { code: '<div accessKey="" />' },
    { code: '<div accessKey={""} />' },
    { code: '<div accessKey="false" />' },
    { code: '<div accessKey={"false"} />' },
    { code: '<div accessKey={0} />' },
    { code: '<div accessKey={null} />' },
    { code: '<div accessKey={false} />' },
    { code: '<div accessKey={void 0} />' },

    // ---- Logical / conditional short-circuits to falsy ----
    { code: '<div accessKey={false && "h"} />' },
    { code: '<div accessKey={"" || ""} />' },
    { code: '<div accessKey={true ? "" : "h"} />' },
    { code: '<div accessKey={false ? "h" : ""} />' },

    // ---- TS wrappers around falsy literal — staticEval unwraps ----
    { code: '<div accessKey={undefined as any} />' },
    { code: '<div accessKey={(undefined)} />' },
    { code: '<div accessKey={"" as string} />' },

    // ---- Spread literal of falsy values ----
    { code: '<div {...{accessKey: undefined}} />' },
    { code: '<div {...{accessKey: ""}} />' },
    { code: '<div {...{accesskey: false}} />' },

    // ---- Spread of non-literal — opaque ----
    { code: '<div {...this.props} />' },

    // ---- Element kind survey: falsy / no-attr forms ----
    { code: '<a />' },
    { code: '<input />' },
    { code: '<Component />' },
    { code: '<UX.Layout>x</UX.Layout>' },
    { code: '<div>content</div>' },

    // ---- Differential-locked valid (vs eslint-plugin-jsx-a11y v6.10.2) ----
    // Numeric / BigInt falsy.
    { code: '<div accessKey={-0} />' },
    { code: '<div accessKey={0n} />' },
    // String concat to empty.
    { code: '<div accessKey={"" + ""} />' },
    // Nullish coalescing → falsy right.
    { code: '<div accessKey={null ?? ""} />' },
    { code: '<div accessKey={undefined ?? null} />' },
    // `satisfies` is opaque (TYPES table miss → null → falsy). Mirrors
    // the staticEval skipTransparent's deliberate exclusion of
    // OEKSatisfies.
    { code: '<div accessKey={"h" satisfies string} />' },
    // Multiple spreads / duplicate keys — first-match wins, when it's
    // falsy the rule skips.
    { code: '<div {...{accessKey: ""}} {...{accessKey: "h"}} />' },
    { code: '<div {...{accessKey: ""}} accessKey="h" />' },
    { code: '<div accessKey="" {...{accessKey: "h"}} />' },
    { code: '<div {...{accessKey: "", accessKey: "h"}} />' },
    // TS-wrapped spread argument is opaque (upstream's strict
    // `argument.type === "ObjectExpression"` check).
    { code: '<div {...({accessKey: "h"} as any)} />' },
    { code: '<div {...({accessKey: "h"})!} />' },
    // Computed / non-Identifier property keys — getProp skips.
    { code: '<div {...{["accessKey"]: "h"}} />' },
    { code: '<div {...{["accesskey"]: "h"}} />' },
    { code: '<div {...{0: "h"}} />' },
    { code: '<div {...{"accesskey": "h"}} />' },
    // Namespaced attribute name does NOT match "accesskey".
    { code: '<div xml:accesskey="h" />' },
  ],
  invalid: [
    // ---- Upstream invalid ----
    { code: '<div accesskey="h" />', errors: [expectedError] },
    { code: '<div accessKey="h" />', errors: [expectedError] },
    { code: '<div accessKey="h" {...props} />', errors: [expectedError] },
    { code: '<div acCesSKeY="y" />', errors: [expectedError] },
    { code: '<div accessKey={"y"} />', errors: [expectedError] },
    { code: '<div accessKey={`${y}`} />', errors: [expectedError] },
    {
      code: '<div accessKey={`${undefined}y${undefined}`} />',
      errors: [expectedError],
    },
    { code: '<div accessKey={`This is ${bad}`} />', errors: [expectedError] },
    { code: '<div accessKey={accessKey} />', errors: [expectedError] },
    { code: '<div accessKey={`${undefined}`} />', errors: [expectedError] },
    {
      code: '<div accessKey={`${undefined}${undefined}`} />',
      errors: [expectedError],
    },

    // ---- Paired element: report on the OPENING element only ----
    { code: '<div accessKey="h">child</div>', errors: [expectedError] },

    // ---- Listener boundary: outer + inner each report ----
    {
      code: '<a accessKey="h"><span accessKey="i" /></a>',
      errors: [expectedError, expectedError],
    },

    // ---- Element kind survey ----
    { code: '<a accessKey="h" />', errors: [expectedError] },
    { code: '<input accessKey="h" />', errors: [expectedError] },
    { code: '<Component accessKey="h" />', errors: [expectedError] },
    { code: '<UX.Layout accessKey="h" />', errors: [expectedError] },

    // ---- Boolean attribute form ----
    { code: '<div accessKey />', errors: [expectedError] },

    // ---- "true" / non-zero literal coercions to truthy ----
    { code: '<div accessKey="true" />', errors: [expectedError] },
    { code: '<div accessKey={"true"} />', errors: [expectedError] },
    // `` `true` `` / `` `false` `` are TemplateLiteral, NOT Literal — no
    // boolean coercion; both extract to truthy strings.
    { code: '<div accessKey={`true`} />', errors: [expectedError] },
    { code: '<div accessKey={`false`} />', errors: [expectedError] },
    { code: '<div accessKey={1} />', errors: [expectedError] },
    { code: '<div accessKey={"0"} />', errors: [expectedError] },

    // ---- TS wrappers around truthy ----
    { code: '<div accessKey={"h" as string} />', errors: [expectedError] },
    { code: '<div accessKey={"h"!} />', errors: [expectedError] },
    { code: '<div accessKey={("h")} />', errors: [expectedError] },

    // ---- Conditional / logical resolving to truthy ----
    { code: '<div accessKey={true && "h"} />', errors: [expectedError] },
    { code: '<div accessKey={"" || "h"} />', errors: [expectedError] },
    { code: '<div accessKey={true ? "h" : ""} />', errors: [expectedError] },

    // ---- Call / member — upstream truthy synthesis ----
    { code: '<div accessKey={fn()} />', errors: [expectedError] },
    { code: '<div accessKey={obj.x} />', errors: [expectedError] },
    { code: '<div accessKey={obj?.x} />', errors: [expectedError] },

    // ---- Spread literal: truthy values ----
    { code: '<div {...{accessKey: "h"}} />', errors: [expectedError] },
    { code: '<div {...{ACCESSKEY: "h"}} />', errors: [expectedError] },
    { code: '<div {...{accessKey}} />', errors: [expectedError] },

    // ---- Differential-locked invalid (vs eslint-plugin-jsx-a11y v6.10.2) ----
    // Numeric / BigInt truthy edges.
    // `NaN` is an Identifier, not in JS_RESERVED — extracted as the
    // string "NaN" (truthy). Counter-intuitive but upstream-correct.
    { code: '<div accessKey={NaN} />', errors: [expectedError] },
    // `Infinity` IS in JS_RESERVED → +Infinity (truthy).
    { code: '<div accessKey={Infinity} />', errors: [expectedError] },
    { code: '<div accessKey={-1} />', errors: [expectedError] },
    { code: '<div accessKey={1n} />', errors: [expectedError] },
    // String concat: Identifier "NaN" treated as "NaN" string,
    // `+ 0` stringifies → "NaN0" truthy.
    { code: '<div accessKey={NaN + 0} />', errors: [expectedError] },
    // String concat truthy.
    { code: '<div accessKey={"h" + "i"} />', errors: [expectedError] },
    { code: '<div accessKey={"" + x} />', errors: [expectedError] },
    // Composite container literals — always truthy under jsx-ast-utils' TYPES.
    { code: '<div accessKey={[]} />', errors: [expectedError] },
    { code: '<div accessKey={[1, 2]} />', errors: [expectedError] },
    { code: '<div accessKey={{}} />', errors: [expectedError] },
    { code: '<div accessKey={{x: 1}} />', errors: [expectedError] },
    { code: '<div accessKey={/foo/} />', errors: [expectedError] },
    { code: '<div accessKey={<span />} />', errors: [expectedError] },
    { code: '<div accessKey={() => {}} />', errors: [expectedError] },
    { code: '<div accessKey={function() {}} />', errors: [expectedError] },
    // Nullish coalescing → truthy.
    { code: '<div accessKey={x ?? "h"} />', errors: [expectedError] },
    { code: '<div accessKey={null ?? "h"} />', errors: [expectedError] },
    { code: '<div accessKey={x || y || ""} />', errors: [expectedError] },
    // Tagged template — upstream truthy synthesis.
    { code: '<div accessKey={tag`x`} />', errors: [expectedError] },

    // Real-world component patterns.
    {
      code: 'function SubmitButton() { return <button accessKey="s" type="submit">Submit</button>; }',
      errors: [expectedError],
    },
    {
      code: 'function HomeLink() { return <a href="/home" accessKey="h" target="_blank">Home</a>; }',
      errors: [expectedError],
    },
    {
      code: 'function NameField() { return <input type="text" accessKey="n" placeholder="name" />; }',
      errors: [expectedError],
    },
    {
      code: 'function ToolbarBtn() { return <div role="button" tabIndex={0} accessKey="t" onClick={fn} />; }',
      errors: [expectedError],
    },
    {
      code: 'const x = <>{cond && <div accessKey="h" />}</>',
      errors: [expectedError],
    },
    {
      code: 'const x = items.map(item => <li accessKey={item.key} key={item.id} />)',
      errors: [expectedError],
    },
    {
      code: 'const x = cond ? <div accessKey="h" /> : <div />',
      errors: [expectedError],
    },
    {
      code: 'class MyForm { render() { return <form><input accessKey="u" name="username" /><input accessKey="p" name="password" type="password" /></form>; } }',
      errors: [expectedError, expectedError],
    },
    {
      code: 'function Stateful() { return <input value={v} onChange={onChange} accessKey="i" />; }',
      errors: [expectedError],
    },

    // Spread + direct attribute mix (multiple orderings).
    {
      code: '<div {...spread1} {...spread2} accessKey="h" />',
      errors: [expectedError],
    },
    {
      code: '<div accessKey="h" {...spread1} {...spread2} />',
      errors: [expectedError],
    },
    {
      code: '<div {...spread1} accessKey="h" {...spread2} />',
      errors: [expectedError],
    },
    // First-match wins (truthy first).
    {
      code: '<div {...{accessKey: "h"}} {...{accessKey: ""}} />',
      errors: [expectedError],
    },
    {
      code: '<div {...{accessKey: "h", accessKey: ""}} />',
      errors: [expectedError],
    },
    // Sibling property doesn't interfere.
    {
      code: '<div {...{className: "x", accessKey: "h"}} />',
      errors: [expectedError],
    },
    // Nested SpreadAssignment — inner spread opaque, sibling property
    // still found.
    {
      code: '<div {...{...other, accessKey: "h"}} />',
      errors: [expectedError],
    },
    {
      code: '<div {...{accessKey: "h", ...other}} />',
      errors: [expectedError],
    },
  ],
});
