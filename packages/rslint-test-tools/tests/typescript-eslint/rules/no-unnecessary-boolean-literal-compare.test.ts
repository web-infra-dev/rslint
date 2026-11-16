import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester(config).run('no-unnecessary-boolean-literal-compare', {
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
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
    },
    {
      code: `
        declare const varBooleanOrUndefined: boolean | undefined;
        varBooleanOrUndefined === false;
      `,
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
        },
      ],
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
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
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
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
        },
      ],
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
      code: `
function test(a?: boolean): boolean {
  // eslint-disable-next-line
  return a !== false;
}
      `,
      options: [
        {
          allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: './tsconfig.unstrict.json',
        },
      },
    },
  ],
  invalid: [
    {
      code: 'true === true;',
      errors: [
        {
          messageId: 'direct',
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
      output: 'true;',
    },
    {
      code: 'false !== true;',
      errors: [
        {
          messageId: 'negated',
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 33,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 29,
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
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
        },
      ],
      errors: [
        {
          messageId: 'comparingNullableToTrueDirect',
          message:
            'This expression unnecessarily compares a nullable boolean value to true instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 40,
        },
      ],
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
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
        },
      ],
      errors: [
        {
          messageId: 'comparingNullableToTrueNegated',
          message:
            'This expression unnecessarily compares a nullable boolean value to true instead of negating it.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 36,
        },
      ],
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
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: 'comparingNullableToFalse',
          message:
            'This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 39,
        },
      ],
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
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: 'comparingNullableToFalse',
          message:
            'This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 41,
        },
      ],
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
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: 'comparingNullableToFalse',
          message:
            'This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 48,
        },
      ],
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 33,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
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
        declare const x;
        if ((x instanceof Error) === false) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 43,
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
        if (false === (x instanceof Error)) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 43,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 41,
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
        if (typeof x === 'string' === false) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 44,
        },
      ],
      output: `
        declare const x;
        if (!(typeof x === 'string')) {
        }
      `,
    },
    {
      code: `
        declare const x;
        if (x instanceof Error === (false)) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 43,
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
        if ((false) === x instanceof Error) {
        }
      `,
      errors: [
        {
          messageId: 'direct',
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 43,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 51,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 49,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 46,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 46,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 45,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 37,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 40,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 50,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 33,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 31,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 31,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 31,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 33,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 31,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of negating it.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
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
          message:
            'This expression unnecessarily compares a boolean value to a boolean instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 31,
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
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: 'comparingNullableToTrueDirect',
          message:
            'This expression unnecessarily compares a nullable boolean value to true instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 38,
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
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: 'comparingNullableToFalse',
          message:
            'This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 39,
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
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
        },
      ],
      errors: [
        {
          messageId: 'comparingNullableToTrueDirect',
          message:
            'This expression unnecessarily compares a nullable boolean value to true instead of using it directly.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 43,
        },
      ],
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
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: 'comparingNullableToFalse',
          message:
            'This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 39,
        },
      ],
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
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: 'comparingNullableToFalse',
          message:
            'This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 41,
        },
      ],
      output: `
        declare const varBooleanOrNull: boolean | null;
        if (!(varBooleanOrNull ?? true)) {
        }
      `,
    },
    {
      code: `
function foo(): boolean {}
      `,
      options: [
        {
          allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: false,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: './tsconfig.unstrict.json',
        },
      },
      errors: [
        {
          messageId: 'noStrictNullCheck',
          message:
            'This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
    },
  ],
});
