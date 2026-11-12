import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Migrated from the existing suite. Each active case pins the full diagnostic and edits.
new RuleTester(config).run('no-unnecessary-boolean-literal-compare', {
  valid: [
    {
      code: "\n      declare const varAny: any;\n      varAny === true;\n    ",
    },
    {
      code: "\n      declare const varAny: any;\n      varAny == false;\n    ",
    },
    {
      code: "\n      declare const varString: string;\n      varString === false;\n    ",
    },
    {
      code: "\n      declare const varString: string;\n      varString === true;\n    ",
    },
    {
      code: "\n      declare const varObject: {};\n      varObject === true;\n    ",
    },
    {
      code: "\n      declare const varObject: {};\n      varObject == false;\n    ",
    },
    {
      code: "\n      declare const varNullOrUndefined: null | undefined;\n      varNullOrUndefined === false;\n    ",
    },
    {
      code: "\n      declare const varBooleanOrString: boolean | string;\n      varBooleanOrString === false;\n    ",
    },
    {
      code: "\n      declare const varBooleanOrString: boolean | string;\n      varBooleanOrString == true;\n    ",
    },
    {
      code: "\n      declare const varTrueOrStringOrUndefined: true | string | undefined;\n      varTrueOrStringOrUndefined == true;\n    ",
    },
    {
      code: "\n      const test: <T>(someCondition: T) => void = someCondition => {\n        if (someCondition === true) {\n        }\n      };\n    ",
    },
    {
      code: "\n      const test: <T>(someCondition: boolean | string) => void = someCondition => {\n        if (someCondition === true) {\n        }\n      };\n    ",
    },
    {
      code: "\n      declare const varBooleanOrUndefined: boolean | undefined;\n      varBooleanOrUndefined === true;\n    ",
    },
    {
      code: "\n        declare const varBooleanOrUndefined: boolean | undefined;\n        varBooleanOrUndefined === true;\n      ",
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
    },
    {
      code: "\n        declare const varBooleanOrUndefined: boolean | undefined;\n        varBooleanOrUndefined === false;\n      ",
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
        },
      ],
    },
    {
      code: "\n        const test: <T extends boolean | undefined>(\n          someCondition: T,\n        ) => void = someCondition => {\n          if (someCondition === true) {\n          }\n        };\n      ",
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
    },
    {
      code: "\n        const test: <T extends boolean | undefined>(\n          someCondition: T,\n        ) => void = someCondition => {\n          if (someCondition === false) {\n          }\n        };\n      ",
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
        },
      ],
    },
    {
      code: "'false' === true;",
    },
    {
      code: "'true' === false;",
    },
    {
      code: "\nconst unconstrained: <T>(someCondition: T) => void = someCondition => {\n  if (someCondition === true) {\n  }\n};\n    ",
    },
    {
      code: "\nconst extendsUnknown: <T extends unknown>(\n  someCondition: T,\n) => void = someCondition => {\n  if (someCondition === true) {\n  }\n};\n    ",
    },
    {
      code: "\nfunction test(a?: boolean): boolean {\n  // eslint-disable-next-line\n  return a !== false;\n}\n      ",
      options: [
        {
          allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.unstrict.json",
        },
      },
    },
  ],
  invalid: [
    {
      code: "true === true;",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
      output: "true;",
    },
    {
      code: "false !== true;",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: "!false;",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (varBoolean !== false) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 33,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varTrue: true;\n        if (varTrue !== true) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 29,
        },
      ],
      output: "\n        declare const varTrue: true;\n        if (!varTrue) {\n        }\n      ",
    },
    {
      code: "\n        declare const varTrueOrUndefined: true | undefined;\n        if (varTrueOrUndefined === true) {\n        }\n      ",
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
        },
      ],
      errors: [
        {
          messageId: "comparingNullableToTrueDirect",
          message: "This expression unnecessarily compares a nullable boolean value to true instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 40,
        },
      ],
      output: "\n        declare const varTrueOrUndefined: true | undefined;\n        if (varTrueOrUndefined) {\n        }\n      ",
    },
    {
      code: "\n        declare const varFalseOrNull: false | null;\n        if (varFalseOrNull !== true) {\n        }\n      ",
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
        },
      ],
      errors: [
        {
          messageId: "comparingNullableToTrueNegated",
          message: "This expression unnecessarily compares a nullable boolean value to true instead of negating it.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 36,
        },
      ],
      output: "\n        declare const varFalseOrNull: false | null;\n        if (!varFalseOrNull) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBooleanOrNull: boolean | null;\n        declare const otherBoolean: boolean;\n        if (varBooleanOrNull === false && otherBoolean) {\n        }\n      ",
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: "comparingNullableToFalse",
          message: "This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.",
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 39,
        },
      ],
      output: "\n        declare const varBooleanOrNull: boolean | null;\n        declare const otherBoolean: boolean;\n        if (!(varBooleanOrNull ?? true) && otherBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBooleanOrNull: boolean | null;\n        declare const otherBoolean: boolean;\n        if (!(varBooleanOrNull === false) || otherBoolean) {\n        }\n      ",
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: "comparingNullableToFalse",
          message: "This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.",
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 41,
        },
      ],
      output: "\n        declare const varBooleanOrNull: boolean | null;\n        declare const otherBoolean: boolean;\n        if ((varBooleanOrNull ?? true) || otherBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varTrueOrFalseOrUndefined: true | false | undefined;\n        declare const otherBoolean: boolean;\n        if (varTrueOrFalseOrUndefined !== false && !otherBoolean) {\n        }\n      ",
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: "comparingNullableToFalse",
          message: "This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.",
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 48,
        },
      ],
      output: "\n        declare const varTrueOrFalseOrUndefined: true | false | undefined;\n        declare const otherBoolean: boolean;\n        if ((varTrueOrFalseOrUndefined ?? true) && !otherBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (false !== varBoolean) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 33,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (true !== varBoolean) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const x;\n        if ((x instanceof Error) === false) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 43,
        },
      ],
      output: "\n        declare const x;\n        if (!(x instanceof Error)) {\n        }\n      ",
    },
    {
      code: "\n        declare const x;\n        if (false === (x instanceof Error)) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 43,
        },
      ],
      output: "\n        declare const x;\n        if (!(x instanceof Error)) {\n        }\n      ",
    },
    {
      code: "\n        declare const x;\n        if (x instanceof Error === false) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 41,
        },
      ],
      output: "\n        declare const x;\n        if (!(x instanceof Error)) {\n        }\n      ",
    },
    {
      code: "\n        declare const x;\n        if (typeof x === 'string' === false) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 44,
        },
      ],
      output: "\n        declare const x;\n        if (!(typeof x === 'string')) {\n        }\n      ",
    },
    {
      code: "\n        declare const x;\n        if (x instanceof Error === (false)) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 43,
        },
      ],
      output: "\n        declare const x;\n        if (!(x instanceof Error)) {\n        }\n      ",
    },
    {
      code: "\n        declare const x;\n        if ((false) === x instanceof Error) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 43,
        },
      ],
      output: "\n        declare const x;\n        if (!(x instanceof Error)) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (!(varBoolean !== false)) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (!(varBoolean === false)) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (!(varBoolean instanceof Event == false)) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 51,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean instanceof Event) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (varBoolean instanceof Event == false) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 49,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!(varBoolean instanceof Event)) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (!((varBoolean ?? false) !== false)) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 46,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!(varBoolean ?? false)) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (!((varBoolean ?? false) === false)) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 46,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean ?? false) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (!((varBoolean ?? true) !== false)) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 45,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!(varBoolean ?? true)) {\n        }\n      ",
    },
    {
      code: "\n        const test: <T extends boolean>(someCondition: T) => void = someCondition => {\n          if (someCondition === true) {\n          }\n        };\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 37,
        },
      ],
      output: "\n        const test: <T extends boolean>(someCondition: T) => void = someCondition => {\n          if (someCondition) {\n          }\n        };\n      ",
    },
    {
      code: "\n        const test: <T extends boolean>(someCondition: T) => void = someCondition => {\n          if (!(someCondition !== false)) {\n          }\n        };\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 40,
        },
      ],
      output: "\n        const test: <T extends boolean>(someCondition: T) => void = someCondition => {\n          if (!someCondition) {\n          }\n        };\n      ",
    },
    {
      code: "\n        const test: <T extends boolean>(someCondition: T) => void = someCondition => {\n          if (!((someCondition ?? true) !== false)) {\n          }\n        };\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 50,
        },
      ],
      output: "\n        const test: <T extends boolean>(someCondition: T) => void = someCondition => {\n          if (!(someCondition ?? true)) {\n          }\n        };\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (varBoolean === true) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (varBoolean === false) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 33,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (varBoolean !== true) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (varBoolean == true) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 31,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (varBoolean == false) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (varBoolean != true) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 31,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (varBoolean != false) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (true == varBoolean) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 31,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (true === varBoolean) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (false === varBoolean) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 33,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (false == varBoolean) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (true != varBoolean) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 31,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (!varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBoolean: boolean;\n        if (false != varBoolean) {\n        }\n      ",
      errors: [
        {
          messageId: "negated",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of negating it.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 32,
        },
      ],
      output: "\n        declare const varBoolean: boolean;\n        if (varBoolean) {\n        }\n      ",
    },
    {
      code: "\n        declare const varFalse: false;\n        if (varFalse === false) {\n        }\n      ",
      errors: [
        {
          messageId: "direct",
          message: "This expression unnecessarily compares a boolean value to a boolean instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 31,
        },
      ],
      output: "\n        declare const varFalse: false;\n        if (!varFalse) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBooleanOrNull: boolean | null;\n        if (varBooleanOrNull === true) {\n        }\n      ",
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: "comparingNullableToTrueDirect",
          message: "This expression unnecessarily compares a nullable boolean value to true instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 38,
        },
      ],
      output: "\n        declare const varBooleanOrNull: boolean | null;\n        if (varBooleanOrNull) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBooleanOrNull: boolean | null;\n        if (varBooleanOrNull === false) {\n        }\n      ",
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: "comparingNullableToFalse",
          message: "This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 39,
        },
      ],
      output: "\n        declare const varBooleanOrNull: boolean | null;\n        if (!(varBooleanOrNull ?? true)) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBooleanOrUndefined: boolean | undefined;\n        if (true === varBooleanOrUndefined) {\n        }\n      ",
      options: [
        {
          allowComparingNullableBooleansToTrue: false,
        },
      ],
      errors: [
        {
          messageId: "comparingNullableToTrueDirect",
          message: "This expression unnecessarily compares a nullable boolean value to true instead of using it directly.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 43,
        },
      ],
      output: "\n        declare const varBooleanOrUndefined: boolean | undefined;\n        if (varBooleanOrUndefined) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBooleanOrNull: boolean | null;\n        if (false === varBooleanOrNull) {\n        }\n      ",
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: "comparingNullableToFalse",
          message: "This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 39,
        },
      ],
      output: "\n        declare const varBooleanOrNull: boolean | null;\n        if (!(varBooleanOrNull ?? true)) {\n        }\n      ",
    },
    {
      code: "\n        declare const varBooleanOrNull: boolean | null;\n        if (!(varBooleanOrNull !== false)) {\n        }\n      ",
      options: [
        {
          allowComparingNullableBooleansToFalse: false,
        },
      ],
      errors: [
        {
          messageId: "comparingNullableToFalse",
          message: "This expression unnecessarily compares a nullable boolean value to false instead of using the ?? operator to provide a default.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 41,
        },
      ],
      output: "\n        declare const varBooleanOrNull: boolean | null;\n        if (!(varBooleanOrNull ?? true)) {\n        }\n      ",
    },
    {
      code: "\nfunction foo(): boolean {}\n      ",
      options: [
        {
          allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: false,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.unstrict.json",
        },
      },
      errors: [
        {
          messageId: "noStrictNullCheck",
          message: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
    },
  ],
});
