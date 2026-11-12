import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('tabindex-no-positive', {
  valid: [
    // ============================================================
    // Upstream-suite valid cases
    // ============================================================
    { code: '<div />;' },
    { code: '<div {...props} />' },
    { code: '<div id="main" />' },
    { code: '<div tabIndex={undefined} />' },
    { code: '<div tabIndex={`${undefined}`} />' },
    { code: '<div tabIndex={`${undefined}${undefined}`} />' },
    { code: '<div tabIndex={0} />' },
    { code: '<div tabIndex={-1} />' },
    { code: '<div tabIndex={null} />' },
    { code: '<div tabIndex={bar()} />' },
    { code: '<div tabIndex={bar} />' },
    { code: '<div tabIndex={"foobar"} />' },
    { code: '<div tabIndex="0" />' },
    { code: '<div tabIndex="-1" />' },
    { code: '<div tabIndex="-5" />' },
    { code: '<div tabIndex="-5.5" />' },
    { code: '<div tabIndex={-5.5} />' },
    { code: '<div tabIndex={-5} />' },

    // ============================================================
    // Boolean / coerce-to-false extraction
    // ============================================================
    { code: '<div tabIndex={false} />' },
    { code: '<div tabIndex="false" />' },
    { code: '<div tabIndex="False" />' },
    { code: '<div tabIndex="FALSE" />' },
    { code: '<div tabIndex={"false"} />' },
    // NoSubstitutionTemplate `false` keeps as string, not coerced to bool.
    { code: '<div tabIndex={`false`} />' },
    { code: '<div tabIndex={`true`} />' }, // same path; Number("true") = NaN
    { code: '<div tabIndex={`0`} />' },
    { code: '<div tabIndex={`-1`} />' },
    { code: '<div tabIndex={`abc`} />' },

    // ============================================================
    // String numeric edge cases
    // ============================================================
    { code: '<div tabIndex="" />' },
    { code: '<div tabIndex={""} />' },
    { code: '<div tabIndex=" " />' },
    { code: '<div tabIndex="   " />' },
    { code: '<div tabIndex="-0x10" />' }, // signed-prefix non-decimal rejected
    { code: '<div tabIndex="0x" />' },
    { code: '<div tabIndex="0xZZ" />' },
    { code: '<div tabIndex="+0" />' },
    { code: '<div tabIndex="-0" />' },
    { code: '<div tabIndex=" -1 " />' },
    { code: '<div tabIndex="abc" />' },

    // ============================================================
    // Negative / zero / non-positive numerics
    // ============================================================
    { code: '<div tabIndex={-100} />' },
    { code: '<div tabIndex={-0.5} />' },
    { code: '<div tabIndex={-0} />' },
    { code: '<div tabIndex={NaN} />' },
    // LITERAL_TYPES.Identifier strips JS_RESERVED, so Infinity → null → 0 → skip
    { code: '<div tabIndex={Infinity} />' },
    { code: '<div tabIndex={-Infinity} />' },

    // ============================================================
    // BigInt
    // ============================================================
    { code: '<div tabIndex={0n} />' },
    { code: '<div tabIndex={-1n} />' },

    // ============================================================
    // Identifier / Call / Member / Conditional / Logical / Binary / New
    // — all LITERAL_TYPES noop → null → 0 → skip
    // ============================================================
    { code: '<div tabIndex={obj.x} />' },
    { code: '<div tabIndex={obj?.x} />' },
    { code: '<div tabIndex={fn?.()} />' },
    { code: '<div tabIndex={Number(5)} />' },
    { code: '<div tabIndex={-(-(-1))} />' },
    { code: '<div tabIndex={-(-(-2))} />' },
    { code: '<div tabIndex={true ? 1n : 0n} />' },
    { code: '<div tabIndex={cond ? 1 : 2} />' },
    { code: '<div tabIndex={true ? 1 : 2} />' },
    { code: '<div tabIndex={1 || 2} />' },
    { code: '<div tabIndex={null ?? 1} />' },
    { code: '<div tabIndex={1 + 1} />' },
    { code: '<div tabIndex={new X()} />' },
    { code: '<div tabIndex={{x:1}} />' },
    { code: '<div tabIndex={/foo/} />' },
    { code: '<div tabIndex={<X/>} />' },
    { code: '<div tabIndex={() => 5} />' },
    { code: '<div tabIndex={function(){}} />' },
    { code: '<div tabIndex={(0, 5)} />' },
    { code: '<div tabIndex={Math.PI} />' },

    // ============================================================
    // Template with substitution → placeholder rendering
    // ============================================================
    { code: '<div tabIndex={`${5}`} />' },
    { code: '<div tabIndex={`${0}`} />' },
    { code: '<div tabIndex={`${false}`} />' },
    { code: '<div tabIndex={`${"5"}`} />' },
    { code: '<div tabIndex={`${`5`}`} />' },
    { code: '<div tabIndex={`${`${5}`}`} />' },
    { code: '<div tabIndex={`${cond ? 1 : 2}`} />' },
    { code: '<div tabIndex={tag`${5}`} />' },

    // ============================================================
    // ArrayLiteralExpression — Array.join semantics
    // ============================================================
    { code: '<div tabIndex={[]} />' },
    { code: '<div tabIndex={[null]} />' },
    { code: '<div tabIndex={[undefined]} />' },
    { code: '<div tabIndex={[bar]} />' }, // Identifier name string → NaN
    { code: '<div tabIndex={[5,6]} />' }, // join "," → "5,6" → NaN
    { code: '<div tabIndex={[true]} />' },
    { code: '<div tabIndex={[false]} />' },
    { code: '<div tabIndex={["abc"]} />' },
    { code: '<div tabIndex={[1, ""]} />' },
    { code: '<div tabIndex={[,5]} />' },
    { code: '<div tabIndex={[,,]} />' },
    { code: '<div tabIndex={[[0]]} />' },
    { code: '<div tabIndex={[[5,6]]} />' },
    { code: '<div tabIndex={[+null]} />' },
    { code: '<div tabIndex={[+undefined]} />' },
    { code: '<div tabIndex={[-true]} />' },
    { code: '<div tabIndex={[!0]} />' },
    { code: '<div tabIndex={[~0]} />' },
    { code: '<div tabIndex={[typeof x]} />' },
    { code: '<div tabIndex={[void 0]} />' },
    { code: '<div tabIndex={[delete a.b]} />' },
    { code: '<div tabIndex={[NaN]} />' },
    { code: '<div tabIndex={[Math.PI]} />' },
    { code: '<div tabIndex={[0]} />' },
    { code: '<div tabIndex={[0,0,0]} />' },
    { code: '<div tabIndex={[5,0]} />' },
    { code: '<div tabIndex={[{}]} />' },
    { code: '<div tabIndex={[/foo/]} />' },
    { code: 'function F(){ let x = 0; return <div tabIndex={[x++]} />; }' },
    { code: '<div tabIndex={[<X/>]} />' },
    { code: '<div tabIndex={[function(){}]} />' },
    { code: '<div tabIndex={[() => 5]} />' },
    { code: '<div tabIndex={[`abc`]} />' },
    { code: '<div tabIndex={[`${5}`]} />' },
    // Template with raw escape sequences (raw text, not cooked).
    { code: '<div tabIndex={`\\t1\\t`} />' },
    { code: '<div tabIndex={`\\n5`} />' },
    { code: '<div tabIndex={`\\u0035`} />' },
    { code: '<div tabIndex={0o0} />' },
    { code: '<div tabIndex={0b0} />' },
    { code: '<div tabIndex={0x0} />' },

    // Extreme small negative.
    { code: '<div tabIndex={-0.0000001} />' },
    // MemberExpression numerics (Number.EPSILON, Number.MAX_VALUE).
    { code: '<div tabIndex={Number.EPSILON} />' },
    { code: '<div tabIndex={Number.MAX_VALUE} />' },
    // Non-Latin / fullwidth numeral.
    { code: '<div tabIndex="５" />' },
    // HTML entities that don't decode to a positive number — `&nbsp;` is
    // U+00A0, `&amp;` is "&", `&#48;` is "0", unknown entities stay as-is.
    { code: '<div tabIndex="&nbsp;" />' },
    { code: '<div tabIndex="&amp;" />' },
    { code: '<div tabIndex="&lt;1" />' },
    { code: '<div tabIndex="&#48;" />' },
    { code: '<div tabIndex="&unknown;" />' },
    { code: '<div tabIndex="&#0;" />' },
    // JsxExpression-wrapped string is NOT an attribute string — entities
    // stay as the literal 5-char sequence.
    { code: '<div tabIndex={"&#49;"} />' },
    { code: '<div tabIndex={`&#49;`} />' },
    // JsxText with "tabIndex=5" content — not an attribute.
    { code: '<div>tabIndex=5</div>' },
    // Element named tabIndex without prop.
    { code: '<tabIndex />' },
    // Spread of non-literal identifier — opaque, no extraction.
    {
      code: 'const props = cond ? {tabIndex:1} : {}; const Q = () => <div {...props} />;',
    },
    // Spread inside array — element is SpreadElement, not extractable.
    { code: '<div tabIndex={[...arr]} />' },
    { code: '<div tabIndex={[...arr, 5]} />' },
    // Map+index real-world patterns.
    {
      code: 'function F(){return [1,2].map((x, i) => <li tabIndex={i} />);}',
    },
    {
      code: 'function F(){return [1,2].map((x, i) => <li tabIndex={i + 1} />);}',
    },
    // Tagged template with substitution.
    { code: '<div tabIndex={String.raw`${5}`} />' },

    // ============================================================
    // Unary on coerced operand — non-positive results
    // ============================================================
    { code: '<div tabIndex={+null} />' },
    { code: '<div tabIndex={+undefined} />' },
    { code: '<div tabIndex={-true} />' },
    { code: '<div tabIndex={+"abc"} />' },
    { code: '<div tabIndex={!1} />' },
    { code: '<div tabIndex={~0} />' },
    { code: '<div tabIndex={~-1} />' },

    // ============================================================
    // typeof / void / postfix update
    // ============================================================
    { code: '<div tabIndex={typeof x} />' },
    { code: '<div tabIndex={void 0} />' },
    { code: 'function F(){ let x = 0; return <div tabIndex={x++} />; }' },

    // ============================================================
    // TS wrappers — opaque under LITERAL_TYPES
    // ============================================================
    { code: '<div tabIndex={(-1)} />' },
    { code: '<div tabIndex={((-1))} />' },
    { code: '<div tabIndex={(-1) as number} />' },
    { code: '<div tabIndex={5 as any} />' },
    { code: '<div tabIndex={5 satisfies number} />' },
    { code: '<div tabIndex={(5)!} />' },

    // ============================================================
    // Spread — listener only fires on JsxAttribute, not JsxSpreadAttribute
    // ============================================================
    { code: '<div {...{tabIndex: 5}} />' },
    { code: '<div {...props} />' },
    { code: '<div {...{tabIndex: 5, role: "x"}} />' },

    // ============================================================
    // Attribute name variants — only exact case-insensitive match
    // ============================================================
    { code: '<div xml:tabIndex="5" />' },
    { code: '<div data-tabIndex="5" />' },

    // ============================================================
    // Comments don't change extraction
    // ============================================================
    { code: '<div /* before */ tabIndex={-1} /* after */ />' },
    { code: '<div tabIndex={/* note */ -1} />' },
  ],
  invalid: [
    // ============================================================
    // Upstream-suite invalid cases
    // ============================================================
    {
      code: '<div tabIndex="1" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div tabIndex={1} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div tabIndex={"1"} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={`1`} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={1.589} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // ============================================================
    // String "true" → boolean coercion
    // ============================================================
    {
      code: '<div tabIndex="true" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex="True" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex="TRUE" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex={"true"} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // ============================================================
    // Boolean attribute form / direct true
    // ============================================================
    {
      code: '<div tabIndex />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<div tabIndex={true} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },

    // ============================================================
    // Non-integer numerics (explicit divergence from no-noninteractive-tabindex)
    // ============================================================
    {
      code: '<div tabIndex={0.5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={1.5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex="0.5" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex="1.5" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ============================================================
    // Numeric variations
    // ============================================================
    {
      code: '<div tabIndex={5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div tabIndex={100} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={0x10} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex={0o10} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex={0b10} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex={1e2} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={1_000} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex={1e1000} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    }, // +Infinity
    {
      code: '<div tabIndex={(5)} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={((5))} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // ============================================================
    // String numerics — hex / oct / bin / decimal
    // ============================================================
    {
      code: '<div tabIndex="5" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div tabIndex="100" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex="0x10" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex="0X10" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex="0o7" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex="0b10" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex="+1" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex=" 1 " />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ============================================================
    // NoSubstitutionTemplateLiteral / TaggedTemplate
    // ============================================================
    {
      code: '<div tabIndex={`5`} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={tag`1`} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // ============================================================
    // BigInt — Number(bigint) coercion
    // ============================================================
    {
      code: '<div tabIndex={2n} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex={1n} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex={9007199254740992n} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<div tabIndex={1.0e2} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex={[1 + 1]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div tabIndex={[cond ? 1 : 2]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div tabIndex={[`5`]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // ============================================================
    // Extreme small positive numerics, separators, small BigInt
    // ============================================================
    {
      code: '<div tabIndex={1e-10} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex={0.0000001} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div tabIndex={5_000_000} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div tabIndex={5n} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },

    // ============================================================
    // Real-world component / framework placement
    // ============================================================
    {
      code: 'export default <div tabIndex={1} />;',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'const f = () => <div tabIndex={1} />;',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'const f = function() { return <div tabIndex={1} />; };',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: 'const o = { render() { return <div tabIndex={1} />; } };',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: 'async function* g() { yield <div tabIndex={1} />; await x; }',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: 'const Comp = ({render}) => render(<div tabIndex={1} />);',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<div>{cond && <span tabIndex={1} />}</div>',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div>{cond ? <span tabIndex={1} /> : null}</div>',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },

    // ============================================================
    // JSX tag shapes
    // ============================================================
    {
      code: '<svg:foo tabIndex={1} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<A.B tabIndex={1} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<A.B.C tabIndex={1} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<tabIndex tabIndex={1} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // ============================================================
    // Coexisting attributes
    // ============================================================
    {
      code: '<li key="k" tabIndex={1}>x</li>',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<input required tabIndex={1} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div {...{key:"x"}} tabIndex={1} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },

    // ============================================================
    // Whitespace / newlines around attribute
    // ============================================================
    {
      code: '<div  tabIndex={1}  />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div\n\ttabIndex={1}\n/>',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },

    // ============================================================
    // Array of JSX with own positive tabIndex — only inner reports
    // ============================================================
    {
      code: '<div tabIndex={[<X tabIndex={1} />]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },

    // ============================================================
    // Self-closing vs paired form
    // ============================================================
    {
      code: '<div tabIndex={1}></div>',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // ============================================================
    // Cross-function / fragment counts
    // ============================================================
    {
      code: 'function A(){return <div tabIndex={1} />;} function B(){return <span tabIndex={5} />;}',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 38,
        },
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 70,
          endLine: 1,
          endColumn: 82,
        },
      ],
    },
    {
      code: '<><div tabIndex={1} /><span tabIndex={2} /><p tabIndex={0} /></>',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },

    // ============================================================
    // JsxText doesn't count; the attribute does
    // ============================================================
    {
      code: '<div tabIndex={5}>tabIndex=5</div>',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // ============================================================
    // Tab/newline in string — JS Number trims all whitespace
    // ============================================================
    {
      code: '<div tabIndex="\t1\t" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex="\n1\n" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 3,
          endColumn: 2,
        },
      ],
    },

    // ============================================================
    // HTML entities — decoded by jsxtransforms.DecodeEntities
    // ============================================================
    {
      code: '<div tabIndex="&#49;" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex="&#x31;" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div tabIndex="&nbsp;5" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div tabIndex="&#53;" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex="&#x35;" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // ============================================================
    // ArrayLiteralExpression → Array.join → Number
    // ============================================================
    {
      code: '<div tabIndex={[5]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={["5"]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex={[Infinity]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    }, // TYPES.Identifier resolves Infinity
    {
      code: '<div tabIndex={[1n]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex={[[5]]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    }, // nested
    {
      code: '<div tabIndex={[[[5]]]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    }, // triple-nested
    // Array with Unary on non-numeric operand — TYPES.UnaryExpression
    // ToNumber-on-operand makes `[+true]` resolve to "1" → 1 → report.
    {
      code: '<div tabIndex={[+true]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div tabIndex={[+"5"]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div tabIndex={[~-2]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    // Large BigInt within Float64 precision.
    {
      code: '<div tabIndex={[9007199254740992n]} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ============================================================
    // Unary on coerced operand
    // ============================================================
    {
      code: '<div tabIndex={+true} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex={+"5"} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex={-"-5"} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex={+"0x10"} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div tabIndex={+5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex={-(-1)} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex={-(-5)} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex={+"  5  "} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<div tabIndex={!0} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex={~-2} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ============================================================
    // delete operator
    // ============================================================
    {
      code: '<div tabIndex={delete a.b} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },

    // ============================================================
    // Custom components / namespaced / member tag names — rule fires
    // regardless of element type (unlike no-noninteractive-tabindex).
    // ============================================================
    {
      code: '<MyButton tabIndex={5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<UX.Layout tabIndex={5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<svg:circle tabIndex={3} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<this.Foo tabIndex={5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<Foo.Bar.Baz tabIndex={5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },

    // ============================================================
    // Case-insensitive name match — propName.toUpperCase() === 'TABINDEX'
    // ============================================================
    {
      code: '<div TABINDEX="5" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div tabindex="5" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div TabIndex="5" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div Tabindex="5" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div TaBiNdEx="5" />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // ============================================================
    // Multiple tabIndex props — each visited independently
    // ============================================================
    {
      code: '<div tabIndex={0} tabIndex={1} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div tabIndex={1} tabIndex={0} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div tabIndex={1} tabIndex={5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ============================================================
    // Comments don't suppress
    // ============================================================
    {
      code: '<div /* a */ tabIndex={5} /* b */ />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div tabIndex={/* truthy */ 5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ============================================================
    // Listener boundary — nested invalid elements each report
    // ============================================================
    {
      code: '<div tabIndex={5}><span tabIndex={1} /></div>',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<><div tabIndex={5} /><span tabIndex={1} /></>',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },

    // ============================================================
    // Real-world component patterns
    // ============================================================
    {
      code: 'function Outer() { return <div tabIndex={5}>focusable</div>; }',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: 'const items = arr.map(item => <li key={item.id} tabIndex={1}>{item.name}</li>);',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: 'const Pane = React.forwardRef((props, ref) => <div ref={ref} tabIndex={3} {...props} />);',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 62,
          endLine: 1,
          endColumn: 74,
        },
      ],
    },
    {
      code: 'class Form extends React.Component { render() { return <div tabIndex={5}>ready</div>; } }',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 61,
          endLine: 1,
          endColumn: 73,
        },
      ],
    },

    // ============================================================
    // Generic JSX
    // ============================================================
    {
      code: '<Map<string, number> tabIndex={5} />',
      errors: [
        {
          messageId: 'tabIndexNoPositive',
          message: 'Avoid positive integer values for tabIndex.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
  ],
});
