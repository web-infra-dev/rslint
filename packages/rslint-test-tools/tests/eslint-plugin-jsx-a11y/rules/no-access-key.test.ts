import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('no-access-key', {
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
    {
      code: '<div accesskey="h" />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div accessKey="h" />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div accessKey="h" {...props} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div acCesSKeY="y" />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div accessKey={"y"} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div accessKey={`${y}`} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div accessKey={`${undefined}y${undefined}`} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<div accessKey={`This is ${bad}`} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<div accessKey={accessKey} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div accessKey={`${undefined}`} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '<div accessKey={`${undefined}${undefined}`} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },

    // ---- Paired element: report on the OPENING element only ----
    {
      code: '<div accessKey="h">child</div>',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ---- Listener boundary: outer + inner each report ----
    {
      code: '<a accessKey="h"><span accessKey="i" /></a>',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },

    // ---- Element kind survey ----
    {
      code: '<a accessKey="h" />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<input accessKey="h" />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<Component accessKey="h" />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<UX.Layout accessKey="h" />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },

    // ---- Boolean attribute form ----
    {
      code: '<div accessKey />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // ---- "true" / non-zero literal coercions to truthy ----
    {
      code: '<div accessKey="true" />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<div accessKey={"true"} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    // `` `true` `` / `` `false` `` are TemplateLiteral, NOT Literal — no
    // boolean coercion; both extract to truthy strings.
    {
      code: '<div accessKey={`true`} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div accessKey={`false`} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<div accessKey={1} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div accessKey={"0"} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },

    // ---- TS wrappers around truthy ----
    {
      code: '<div accessKey={"h" as string} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<div accessKey={"h"!} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<div accessKey={("h")} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },

    // ---- Conditional / logical resolving to truthy ----
    {
      code: '<div accessKey={true && "h"} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div accessKey={"" || "h"} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div accessKey={true ? "h" : ""} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ---- Call / member — upstream truthy synthesis ----
    {
      code: '<div accessKey={fn()} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<div accessKey={obj.x} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div accessKey={obj?.x} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },

    // ---- Spread literal: truthy values ----
    {
      code: '<div {...{accessKey: "h"}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div {...{ACCESSKEY: "h"}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div {...{accessKey}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },

    // ---- Differential-locked invalid (vs eslint-plugin-jsx-a11y v6.10.2) ----
    // Numeric / BigInt truthy edges.
    // `NaN` is an Identifier, not in JS_RESERVED — extracted as the
    // string "NaN" (truthy). Counter-intuitive but upstream-correct.
    {
      code: '<div accessKey={NaN} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    // `Infinity` IS in JS_RESERVED → +Infinity (truthy).
    {
      code: '<div accessKey={Infinity} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div accessKey={-1} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div accessKey={1n} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    // String concat: Identifier "NaN" treated as "NaN" string,
    // `+ 0` stringifies → "NaN0" truthy.
    {
      code: '<div accessKey={NaN + 0} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    // String concat truthy.
    {
      code: '<div accessKey={"h" + "i"} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div accessKey={"" + x} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    // Composite container literals — always truthy under jsx-ast-utils' TYPES.
    {
      code: '<div accessKey={[]} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div accessKey={[1, 2]} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div accessKey={{}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div accessKey={{x: 1}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div accessKey={/foo/} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div accessKey={<span />} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div accessKey={() => {}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div accessKey={function() {}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    // Nullish coalescing → truthy.
    {
      code: '<div accessKey={x ?? "h"} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div accessKey={null ?? "h"} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div accessKey={x || y || ""} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    // Tagged template — upstream truthy synthesis.
    {
      code: '<div accessKey={tag`x`} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },

    // Real-world component patterns.
    {
      code: 'function SubmitButton() { return <button accessKey="s" type="submit">Submit</button>; }',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 70,
        },
      ],
    },
    {
      code: 'function HomeLink() { return <a href="/home" accessKey="h" target="_blank">Home</a>; }',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 76,
        },
      ],
    },
    {
      code: 'function NameField() { return <input type="text" accessKey="n" placeholder="name" />; }',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 85,
        },
      ],
    },
    {
      code: 'function ToolbarBtn() { return <div role="button" tabIndex={0} accessKey="t" onClick={fn} />; }',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 93,
        },
      ],
    },
    {
      code: 'const x = <>{cond && <div accessKey="h" />}</>',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'const x = items.map(item => <li accessKey={item.key} key={item.id} />)',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 70,
        },
      ],
    },
    {
      code: 'const x = cond ? <div accessKey="h" /> : <div />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'class MyForm { render() { return <form><input accessKey="u" name="username" /><input accessKey="p" name="password" type="password" /></form>; } }',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 79,
        },
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 79,
          endLine: 1,
          endColumn: 134,
        },
      ],
    },
    {
      code: 'function Stateful() { return <input value={v} onChange={onChange} accessKey="i" />; }',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 83,
        },
      ],
    },

    // Spread + direct attribute mix (multiple orderings).
    {
      code: '<div {...spread1} {...spread2} accessKey="h" />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<div accessKey="h" {...spread1} {...spread2} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<div {...spread1} accessKey="h" {...spread2} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    // First-match wins (truthy first).
    {
      code: '<div {...{accessKey: "h"}} {...{accessKey: ""}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: '<div {...{accessKey: "h", accessKey: ""}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    // Sibling property doesn't interfere.
    {
      code: '<div {...{className: "x", accessKey: "h"}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    // Nested SpreadAssignment — inner spread opaque, sibling property
    // still found.
    {
      code: '<div {...{...other, accessKey: "h"}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '<div {...{accessKey: "h", ...other}} />',
      errors: [
        {
          messageId: 'noAccessKey',
          message:
            'No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
  ],
});
