import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';
import * as path from 'node:path';


import { getFixturesRootDir } from '../RuleTester';

const rootDir = getFixturesRootDir();
const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootDir,
    },
  },
});

ruleTester.run('no-unnecessary-boolean-literal-compare', {
  valid: [
    `
      declare const varAny: any;
      varAny === true;
    `,
    `
      declare const varAny: any;
      varAny == false;
    `,
    `
      declare const varString: string;
      varString === false;
    `,
    `
      declare const varString: string;
      varString === true;
    `,
    `
      declare const varObject: {};
      varObject === true;
    `,
    `
      declare const varObject: {};
      varObject == false;
    `,
    `
      declare const varNullOrUndefined: null | undefined;
      varNullOrUndefined === false;
    `,
    `
      declare const varBooleanOrString: boolean | string;
      varBooleanOrString === false;
    `,
    `
      declare const varBooleanOrString: boolean | string;
      varBooleanOrString == true;
    `,
    `
      declare const varTrueOrStringOrUndefined: true | string | undefined;
      varTrueOrStringOrUndefined == true;
    `,
    `
      const test: <T>(someCondition: T) => void = someCondition => {
        if (someCondition === true) {
        }
      };
    `,
    `
      const test: <T>(someCondition: boolean | string) => void = someCondition => {
        if (someCondition === true) {
        }
      };
    `,
    `
      declare const varBooleanOrUndefined: boolean | undefined;
      varBooleanOrUndefined === true;
    `,
    {
      code: `
        declare const varBooleanOrUndefined: boolean | undefined;
        varBooleanOrUndefined === true;
      `,
      options: [{ allowComparingNullableBooleansToFalse: false }],
    },
    {
      code: `
        declare const varBooleanOrUndefined: boolean | undefined;
        varBooleanOrUndefined === false;
      `,
      options: [{ allowComparingNullableBooleansToTrue: false }],
    },
    {
      code: `
        const test: <T extends boolean | undefined>(
          someCondition: T,
        ) => void = someCondition => {
          if (someCondition === true) {
          }
        };
      `,
      options: [{ allowComparingNullableBooleansToFalse: false }],
    },
    {
      code: `
        const test: <T extends boolean | undefined>(
          someCondition: T,
        ) => void = someCondition => {
          if (someCondition === false) {
          }
        };
      `,
      options: [{ allowComparingNullableBooleansToTrue: false }],
    },
    "'false' === true;",
    "'true' === false;",
    `
const unconstrained: <T>(someCondition: T) => void = someCondition => {
  if (someCondition === true) {
  }
};
    `,
    `
const extendsUnknown: <T extends unknown>(
  someCondition: T,
) => void = someCondition => {
  if (someCondition === true) {
  }
};
    `,
    {
      // Skip: requires per-test-case tsconfigRootDir which the test framework doesn't support
      skip: true,
      code: `
function test(a?: boolean): boolean {
  // eslint-disable-next-line
  return a !== false;
}
      `,
      languageOptions: {
        parserOptions: {
          tsconfigRootDir: path.join(rootDir, 'unstrict'),
        },
      },
      options: [
        {
          allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: true,
        },
      ],
    },
  ],

  invalid: [
    {
      code: 'true === true;',
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: 'true;',
    },
    {
      code: 'false !== true;',
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: '!false;',
    },
    {
      code: `
        declare const varBoolean: boolean;
        if (varBoolean !== false) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
    },
    {
      code: `
        declare const varTrue: true;
        if (varTrue !== true) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varTrue: true;
        if (!varTrue) {
        }
      `,
    },
    {
      code: `
        declare const varTrueOrUndefined: true | undefined;
        if (varTrueOrUndefined === true) {
        }
      `,
      errors: [
        {
          messageId: 'comparingNullableToTrueDirect',
        },
      ],
      options: [{ allowComparingNullableBooleansToTrue: false }],
      output: `
        declare const varTrueOrUndefined: true | undefined;
        if (varTrueOrUndefined) {
        }
      `,
    },
    {
      code: `
        declare const varFalseOrNull: false | null;
        if (varFalseOrNull !== true) {
        }
      `,
      errors: [
        {
          messageId: 'comparingNullableToTrueNegated',
        },
      ],
      options: [{ allowComparingNullableBooleansToTrue: false }],
      output: `
        declare const varFalseOrNull: false | null;
        if (!varFalseOrNull) {
        }
      `,
    },
    {
      code: `
        declare const varBooleanOrNull: boolean | null;
        declare const otherBoolean: boolean;
        if (varBooleanOrNull === false && otherBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'comparingNullableToFalse',
        },
      ],
      options: [{ allowComparingNullableBooleansToFalse: false }],
      output: `
        declare const varBooleanOrNull: boolean | null;
        declare const otherBoolean: boolean;
        if (!(varBooleanOrNull ?? true) && otherBoolean) {
        }
      `,
    },
    {
      code: `
        declare const varBooleanOrNull: boolean | null;
        declare const otherBoolean: boolean;
        if (!(varBooleanOrNull === false) || otherBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'comparingNullableToFalse',
        },
      ],
      options: [{ allowComparingNullableBooleansToFalse: false }],
      output: `
        declare const varBooleanOrNull: boolean | null;
        declare const otherBoolean: boolean;
        if ((varBooleanOrNull ?? true) || otherBoolean) {
        }
      `,
    },
    {
      code: `
        declare const varTrueOrFalseOrUndefined: true | false | undefined;
        declare const otherBoolean: boolean;
        if (varTrueOrFalseOrUndefined !== false && !otherBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'comparingNullableToFalse',
        },
      ],
      options: [{ allowComparingNullableBooleansToFalse: false }],
      output: `
        declare const varTrueOrFalseOrUndefined: true | false | undefined;
        declare const otherBoolean: boolean;
        if ((varTrueOrFalseOrUndefined ?? true) && !otherBoolean) {
        }
      `,
    },
    {
      code: `
        declare const varBoolean: boolean;
        if (false !== varBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
    },
    {
      code: `
        declare const varBoolean: boolean;
        if (true !== varBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!varBoolean) {
        }
      `,
    },
    {
      code: noFormat`
        declare const x;
        if ((x instanceof Error) === false) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const x;
        if (!(x instanceof Error)) {
        }
      `,
    },
    {
      code: noFormat`
        declare const x;
        if (false === (x instanceof Error)) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const x;
        if (!(x instanceof Error)) {
        }
      `,
    },
    {
      code: `
        declare const x;
        if (x instanceof Error === false) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const x;
        if (!(x instanceof Error)) {
        }
      `,
    },
    {
      code: noFormat`
        declare const x;
        if (typeof x === 'string' === false) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const x;
        if (!(typeof x === 'string')) {
        }
      `,
    },
    {
      code: noFormat`
        declare const x;
        if (x instanceof Error === (false)) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const x;
        if (!(x instanceof Error)) {
        }
      `,
    },
    {
      code: noFormat`
        declare const x;
        if ((false) === x instanceof Error) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const x;
        if (!(x instanceof Error)) {
        }
      `,
    },
    {
      code: `
        declare const varBoolean: boolean;
        if (!(varBoolean !== false)) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!varBoolean) {
        }
      `,
    },
    {
      code: `
        declare const varBoolean: boolean;
        if (!(varBoolean === false)) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
    },
    {
      code: `
        declare const varBoolean: boolean;
        if (!(varBoolean instanceof Event == false)) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean instanceof Event) {
        }
      `,
    },
    {
      code: `
        declare const varBoolean: boolean;
        if (varBoolean instanceof Event == false) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!(varBoolean instanceof Event)) {
        }
      `,
    },
    {
      code: `
        declare const varBoolean: boolean;
        if (!((varBoolean ?? false) !== false)) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!(varBoolean ?? false)) {
        }
      `,
    },
    {
      code: `
        declare const varBoolean: boolean;
        if (!((varBoolean ?? false) === false)) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean ?? false) {
        }
      `,
    },
    {
      code: `
        declare const varBoolean: boolean;
        if (!((varBoolean ?? true) !== false)) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!(varBoolean ?? true)) {
        }
      `,
    },
    {
      code: `
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (someCondition === true) {
          }
        };
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (someCondition) {
          }
        };
      `,
    },
    {
      code: `
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (!(someCondition !== false)) {
          }
        };
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (!someCondition) {
          }
        };
      `,
    },
    {
      code: `
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (!((someCondition ?? true) !== false)) {
          }
        };
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        const test: <T extends boolean>(someCondition: T) => void = someCondition => {
          if (!(someCondition ?? true)) {
          }
        };
      `,
    },
    // basic === true
    {
      code: `
        declare const varBoolean: boolean;
        if (varBoolean === true) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
    },
    // basic === false
    {
      code: `
        declare const varBoolean: boolean;
        if (varBoolean === false) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!varBoolean) {
        }
      `,
    },
    // basic !== true
    {
      code: `
        declare const varBoolean: boolean;
        if (varBoolean !== true) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!varBoolean) {
        }
      `,
    },
    // loose == true
    {
      code: `
        declare const varBoolean: boolean;
        if (varBoolean == true) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
    },
    // loose == false
    {
      code: `
        declare const varBoolean: boolean;
        if (varBoolean == false) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!varBoolean) {
        }
      `,
    },
    // loose != true
    {
      code: `
        declare const varBoolean: boolean;
        if (varBoolean != true) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!varBoolean) {
        }
      `,
    },
    // loose != false
    {
      code: `
        declare const varBoolean: boolean;
        if (varBoolean != false) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
    },
    // loose == with literal on left
    {
      code: `
        declare const varBoolean: boolean;
        if (true == varBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
    },
    // strict === with true literal on left
    {
      code: `
        declare const varBoolean: boolean;
        if (true === varBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
    },
    // strict === with false literal on left
    {
      code: `
        declare const varBoolean: boolean;
        if (false === varBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!varBoolean) {
        }
      `,
    },
    // loose == with false literal on left
    {
      code: `
        declare const varBoolean: boolean;
        if (false == varBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!varBoolean) {
        }
      `,
    },
    // loose != with true literal on left
    {
      code: `
        declare const varBoolean: boolean;
        if (true != varBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (!varBoolean) {
        }
      `,
    },
    // loose != with false literal on left (existing)
    {
      code: `
        declare const varBoolean: boolean;
        if (false != varBoolean) {
        }
      `,
      errors: [
        {
          messageId: 'negated',
        },
      ],
      output: `
        declare const varBoolean: boolean;
        if (varBoolean) {
        }
      `,
    },
    // literal false type
    {
      code: `
        declare const varFalse: false;
        if (varFalse === false) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
        },
      ],
      output: `
        declare const varFalse: false;
        if (!varFalse) {
        }
      `,
    },
    // both nullable options false — comparing to true
    {
      code: `
        declare const varBooleanOrNull: boolean | null;
        if (varBooleanOrNull === true) {
        }
      `,
      errors: [
        {
          messageId: 'comparingNullableToTrueDirect',
        },
      ],
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      output: `
        declare const varBooleanOrNull: boolean | null;
        if (varBooleanOrNull) {
        }
      `,
    },
    // both nullable options false — comparing to false
    {
      code: `
        declare const varBooleanOrNull: boolean | null;
        if (varBooleanOrNull === false) {
        }
      `,
      errors: [
        {
          messageId: 'comparingNullableToFalse',
        },
      ],
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      output: `
        declare const varBooleanOrNull: boolean | null;
        if (!(varBooleanOrNull ?? true)) {
        }
      `,
    },
    // nullable with literal on left side
    {
      code: `
        declare const varBooleanOrUndefined: boolean | undefined;
        if (true === varBooleanOrUndefined) {
        }
      `,
      errors: [
        {
          messageId: 'comparingNullableToTrueDirect',
        },
      ],
      options: [{ allowComparingNullableBooleansToTrue: false }],
      output: `
        declare const varBooleanOrUndefined: boolean | undefined;
        if (varBooleanOrUndefined) {
        }
      `,
    },
    // nullable with false literal on left side
    {
      code: `
        declare const varBooleanOrNull: boolean | null;
        if (false === varBooleanOrNull) {
        }
      `,
      errors: [
        {
          messageId: 'comparingNullableToFalse',
        },
      ],
      options: [{ allowComparingNullableBooleansToFalse: false }],
      output: `
        declare const varBooleanOrNull: boolean | null;
        if (!(varBooleanOrNull ?? true)) {
        }
      `,
    },
    // nullable + unary negation + !== false: both ! and ?? true blocks fire
    {
      code: `
        declare const varBooleanOrNull: boolean | null;
        if (!(varBooleanOrNull !== false)) {
        }
      `,
      errors: [
        {
          messageId: 'comparingNullableToFalse',
        },
      ],
      options: [{ allowComparingNullableBooleansToFalse: false }],
      output: `
        declare const varBooleanOrNull: boolean | null;
        if (!(varBooleanOrNull ?? true)) {
        }
      `,
    },
    {
      // Skip: requires per-test-case tsconfigRootDir which the test framework doesn't support
      skip: true,
      code: `
function foo(): boolean {}
      `,
      errors: [
        {
          messageId: 'noStrictNullCheck',
        },
      ],
      languageOptions: {
        parserOptions: {
          tsconfigRootDir: path.join(rootDir, 'unstrict'),
        },
      },
      options: [
        {
          allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: false,
        },
      ],
    },
  ],
});
