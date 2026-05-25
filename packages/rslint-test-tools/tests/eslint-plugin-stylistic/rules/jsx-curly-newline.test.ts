import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const EXPECTED_AFTER = "Expected newline after '{'.";
const EXPECTED_BEFORE = "Expected newline before '}'.";
const UNEXPECTED_AFTER = "Unexpected newline after '{'.";
const UNEXPECTED_BEFORE = "Unexpected newline before '}'.";

const CONSISTENT = ['consistent'];
const NEVER = ['never'];
const MULTILINE_REQUIRE = [{ singleline: 'consistent', multiline: 'require' }];

ruleTester.run('jsx-curly-newline', null as never, {
  valid: [
    // consistent (default)
    { code: `<div>{foo}</div>`, options: CONSISTENT },
    {
      code: `<div>\n          {\n            foo\n          }\n        </div>`,
      options: CONSISTENT,
    },
    {
      code: `<div>\n          { foo &&\n            foo.bar }\n        </div>`,
      options: CONSISTENT,
    },
    {
      code: `<div>\n          {\n            foo &&\n            foo.bar\n          }\n        </div>`,
      options: CONSISTENT,
    },
    { code: `<div foo={\n          bar\n        } />`, options: CONSISTENT },
    // { singleline: consistent, multiline: require }
    { code: `<div>{foo}</div>`, options: MULTILINE_REQUIRE },
    { code: `<div foo={bar} />`, options: MULTILINE_REQUIRE },
    {
      code: `<div>\n          {\n            foo &&\n            foo.bar\n          }\n        </div>`,
      options: MULTILINE_REQUIRE,
    },
    {
      code: `<div>\n          {\n            foo\n          }\n        </div>`,
      options: MULTILINE_REQUIRE,
    },
    // never
    { code: `<div>{foo}</div>`, options: NEVER },
    { code: `<div foo={bar} />`, options: NEVER },
    {
      code: `<div>\n          { foo &&\n            foo.bar }\n        </div>`,
      options: NEVER,
    },
  ],

  invalid: [
    // consistent: newline before } but not after {
    {
      code: `<div>\n          { foo \n}\n        </div>`,
      options: CONSISTENT,
      errors: [{ messageId: 'unexpectedBefore', message: UNEXPECTED_BEFORE }],
    },
    {
      code: `<div>\n          { foo &&\n            foo.bar \n}\n        </div>`,
      options: CONSISTENT,
      errors: [{ messageId: 'unexpectedBefore', message: UNEXPECTED_BEFORE }],
    },
    {
      code: `<div>\n          { foo &&\n            bar\n          }\n        </div>`,
      options: CONSISTENT,
      errors: [{ messageId: 'unexpectedBefore', message: UNEXPECTED_BEFORE }],
    },
    // { multiline: require }: newline before } unexpected on single-line expr
    {
      code: `<div>{foo\n}</div>`,
      options: MULTILINE_REQUIRE,
      errors: [{ messageId: 'unexpectedBefore', message: UNEXPECTED_BEFORE }],
    },
    {
      code: `<div>{\nfoo}</div>`,
      options: MULTILINE_REQUIRE,
      errors: [{ messageId: 'expectedBefore', message: EXPECTED_BEFORE }],
    },
    {
      code: `<div>\n          { foo &&\n            bar }\n        </div>`,
      options: MULTILINE_REQUIRE,
      errors: [
        { messageId: 'expectedAfter', message: EXPECTED_AFTER },
        { messageId: 'expectedBefore', message: EXPECTED_BEFORE },
      ],
    },
    {
      code: `<div style={foo &&\n          foo.bar\n        } />`,
      options: MULTILINE_REQUIRE,
      errors: [{ messageId: 'expectedAfter', message: EXPECTED_AFTER }],
    },
    // never: newlines on both sides unexpected
    {
      code: `<div>\n          {\nfoo\n}\n        </div>`,
      options: NEVER,
      errors: [
        { messageId: 'unexpectedAfter', message: UNEXPECTED_AFTER },
        { messageId: 'unexpectedBefore', message: UNEXPECTED_BEFORE },
      ],
    },
    {
      code: `<div>\n          {\n            foo &&\n            foo.bar\n          }\n        </div>`,
      options: NEVER,
      errors: [
        { messageId: 'unexpectedAfter', message: UNEXPECTED_AFTER },
        { messageId: 'unexpectedBefore', message: UNEXPECTED_BEFORE },
      ],
    },
    {
      code: `<div>\n          { foo &&\n            foo.bar\n          }\n        </div>`,
      options: NEVER,
      errors: [{ messageId: 'unexpectedBefore', message: UNEXPECTED_BEFORE }],
    },
    // never: comment in the gap suppresses the fix
    {
      code: `<div>\n          { /* not fixed due to comment */\n            foo }\n        </div>`,
      options: NEVER,
      errors: [{ messageId: 'unexpectedAfter', message: UNEXPECTED_AFTER }],
    },
    {
      code: `<div>\n          { foo\n            /* not fixed due to comment */}\n        </div>`,
      options: NEVER,
      errors: [{ messageId: 'unexpectedBefore', message: UNEXPECTED_BEFORE }],
    },
  ],
});
