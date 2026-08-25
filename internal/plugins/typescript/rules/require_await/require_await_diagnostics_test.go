package require_await

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireAwaitDiagnosticPayloads(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: `class DampingSwipe {
  protected async isSwipeHorizontalDisAllow(left: number) {
    return left < 0;
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'isSwipeHorizontalDisAllow' has no 'await' expression.",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 44,
				}},
			},
			{
				Code: `const ClipboardHelper = {
  async loadClipboardSecuritySDK() {
    return;
  },
};`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'loadClipboardSecuritySDK' has no 'await' expression.",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 33,
				}},
			},
			{
				Code: `const mockStore = {
  async getAllKeys() {
    return [];
  },
  async getItem() {
    return {};
  },
  async getItems() {
    return [];
  },
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Async method 'getAllKeys' has no 'await' expression.",
						Line:      2,
						Column:    3,
						EndLine:   2,
						EndColumn: 19,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method 'getItem' has no 'await' expression.",
						Line:      5,
						Column:    3,
						EndLine:   5,
						EndColumn: 16,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method 'getItems' has no 'await' expression.",
						Line:      8,
						Column:    3,
						EndLine:   8,
						EndColumn: 17,
					},
				},
			},
			{
				Code: `const manager = {
  async warmup() {
    return {};
  },
  async preload() {
    return;
  },
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Async method 'warmup' has no 'await' expression.",
						Line:      2,
						Column:    3,
						EndLine:   2,
						EndColumn: 15,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method 'preload' has no 'await' expression.",
						Line:      5,
						Column:    3,
						EndLine:   5,
						EndColumn: 16,
					},
				},
			},
			{
				Code: `function render() {
  const { data: meta, loading: isSearchingMetaData } = useRequest(async () => {
    return 1;
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function has no 'await' expression.",
					Line:      2,
					Column:    76,
					EndLine:   2,
					EndColumn: 78,
				}},
			},
			{
				Code: `const element = (
  <LazyReadDocChatInput
          onBeforeSendMessage={async () =>
            true
          }
  />
);`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function has no 'await' expression.",
					Line:      3,
					Column:    41,
					EndLine:   3,
					EndColumn: 43,
				}},
			},
			{
				Code: `const rawSubmit = useSubmit({
    onBeforeSendMessage: async () => onBeforeSendMessage?.() ?? true,
});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'onBeforeSendMessage' has no 'await' expression.",
					Line:      2,
					Column:    5,
					EndLine:   2,
					EndColumn: 32,
				}},
			},
			{
				Code: `const getGlobalInfoService = async () => {
  return getGlobalContainer();
};`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function 'getGlobalInfoService' has no 'await' expression.",
					Line:      1,
					Column:    39,
					EndLine:   1,
					EndColumn: 41,
				}},
			},
			{
				Code: `class ActionManager {
  runResetTaskFn = async () => {
    return;
  };
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'runResetTaskFn' has no 'await' expression.",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 26,
				}},
			},
			{
				Code: `const commentSDK = {
  commentSDKManager: async () => {
    return {};
  },
};`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'commentSDKManager' has no 'await' expression.",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 28,
				}},
			},
			{
				Code: `const wrapped = {
  field: ((async () => 1)),
};`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'field' has no 'await' expression.",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 18,
				}},
			},
			{
				Code: `class AutoAccessor {
  accessor field = async () => 1;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function has no 'await' expression.",
					Line:      2,
					Column:    29,
					EndLine:   2,
					EndColumn: 31,
				}},
			},
			{
				Code: `class PrivateMembers {
  static async #method() {
    return 1;
  }
  static #field = async () => 1;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Static private async method #method has no 'await' expression.",
					},
					{
						MessageId: "missingAwait",
						Message:   "Static private async method #field has no 'await' expression.",
					},
				},
			},
			{
				Code: `const computed = {
  ['field']: async () => 1,
};`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'field' has no 'await' expression.",
				}},
			},
			{
				Code: `class EmptyNames {
  async ''() {
    return 1;
  }
  '' = async () => 2;
}
const object = {
  '': async () => 3,
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
					},
				},
			},
			{
				Code: `declare function consume(value: unknown): void;
const dynamic = 1;
const object = {
  [dynamic]: async function ownName() { consume(1); },
  [1 + 1]: async () => consume(2),
  [true ? 4 : dynamic]: async () => consume(3),
  [String(5)]: async () => consume(4),
  [({ key: 6 }).key]: async () => consume(5),
};
class Fields {
  [dynamic] = async function ownField() { consume(6); };
  [1 + 2] = async () => consume(7);
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
						Line:      4,
						Column:    3,
						EndLine:   4,
						EndColumn: 36,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method '2' has no 'await' expression.",
						Line:      5,
						Column:    3,
						EndLine:   5,
						EndColumn: 18,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method '4' has no 'await' expression.",
						Line:      6,
						Column:    3,
						EndLine:   6,
						EndColumn: 31,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
						Line:      7,
						Column:    3,
						EndLine:   7,
						EndColumn: 22,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method '6' has no 'await' expression.",
						Line:      8,
						Column:    3,
						EndLine:   8,
						EndColumn: 29,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
						Line:      11,
						Column:    3,
						EndLine:   11,
						EndColumn: 38,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method '3' has no 'await' expression.",
						Line:      12,
						Column:    3,
						EndLine:   12,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `let assigned: () => Promise<void>;
assigned = async () => consume(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function 'assigned' has no 'await' expression.",
				}},
			},
			{
				Code: `function defaults(callback = async () => consume(1)) {
  return callback;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function 'callback' has no 'await' expression.",
				}},
			},
			{
				Code: `export default async () => consume(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function 'default' has no 'await' expression.",
				}},
			},
			{
				Code: `declare function consume(value: unknown): void;

function bindingElement({ callback = async () => consume(1) } = {}) {
  return callback;
}

let shorthand: () => Promise<void>;
({ shorthand = async () => consume(2) } = {});

let compound: (() => Promise<void>) | undefined;
compound ??= async () => consume(3);

export default async function () {
  consume(4);
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Async arrow function 'callback' has no 'await' expression.",
						Line:      3,
						Column:    47,
						EndLine:   3,
						EndColumn: 49,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async arrow function 'shorthand' has no 'await' expression.",
						Line:      8,
						Column:    25,
						EndLine:   8,
						EndColumn: 27,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async arrow function 'compound' has no 'await' expression.",
						Line:      11,
						Column:    23,
						EndLine:   11,
						EndColumn: 25,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async function 'default' has no 'await' expression.",
						Line:      13,
						Column:    16,
						EndLine:   13,
						EndColumn: 31,
					},
				},
			},
		},
	)
}
