// The following tests are adapted from the tests in eslint.
// Original Code: https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/tests/lib/rules/no-unused-vars.js
// License      : https://github.com/eslint/eslint/blob/eb76282e0a2db8aa10a3d5659f5f9237d9729121/LICENSE

import type { TestCaseError } from '@typescript-eslint/rule-tester';
import type { TSESTree } from '@typescript-eslint/utils';

import { RuleTester } from '@typescript-eslint/rule-tester';
import { AST_NODE_TYPES } from '@typescript-eslint/utils';

import type { MessageIds } from '../../../src/rules/no-unused-vars';
import { getFixturesRootDir } from '../../RuleTester';



export const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.lenient.json',
      tsconfigRootDir: getFixturesRootDir(),
    },
  },
});

ruleTester.defineRule('use-every-a', {
  create: context => {
    /**
     * Mark a variable as used
     */
    function useA(node: TSESTree.Node): void {
      context.sourceCode.markVariableAsUsed('a', node);
    }
    return {
      ReturnStatement: useA,
      VariableDeclaration: useA,
    };
  },
  defaultOptions: [],
  meta: {
    messages: {},
    schema: [],
    type: 'problem',
  },
});

/**
 * Returns an expected error for defined-but-not-used variables.
 * @param varName The name of the variable
 * @param [additional] The additional text for the message data
 * @param [type] The node type (defaults to "Identifier")
 * @returns An expected error object
 */
export function definedError(
  varName: string,
  additional = '',
): TestCaseError<MessageIds> {
  return {
    data: {
      action: 'defined',
      additional,
      varName,
    },
    messageId: 'unusedVar',
  };
}

/**
 * Returns an expected error for assigned-but-not-used variables.
 * @param varName The name of the variable
 * @param [additional] The additional text for the message data
 * @param [type] The node type (defaults to "Identifier")
 * @returns An expected error object
 */
export function assignedError(
  varName: string,
  additional = '',
): TestCaseError<MessageIds> {
  return {
    data: {
      action: 'assigned a value',
      additional,
      varName,
    },
    messageId: 'unusedVar',
  };
}

/**
 * Returns an expected error for used-but-ignored variables.
 * @param varName The name of the variable
 * @param [additional] The additional text for the message data
 * @param [type] The node type (defaults to "Identifier")
 * @returns An expected error object
 */
export function usedIgnoredError(
  varName: string,
  additional = '',
  type = AST_NODE_TYPES.Identifier,
): TestCaseError<MessageIds> {
  return {
    data: {
      additional,
      varName,
    },
    messageId: 'usedIgnoredVar',
    type,
  };
}
