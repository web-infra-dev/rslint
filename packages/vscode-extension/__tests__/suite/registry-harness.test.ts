import * as assert from 'assert';
import { randomUUID } from 'node:crypto';
import * as vscode from 'vscode';
import {
  CodeActionRegistryProbe,
  saveDocumentOnce,
  waitForConsecutiveSuccessfulProbeWindows,
} from '../utils/codeActionRegistry';
import { runBeforeDeadline } from '../utils/deadline';

suite('VS Code test harness fail-closed guards', function () {
  this.timeout(15_000);

  test('an interrupted window resets consecutive readiness', async () => {
    const outcomes = [true, false, true, true];
    const result = await waitForConsecutiveSuccessfulProbeWindows(
      async (attempt) => outcomes[attempt - 1] ?? false,
      {
        consecutiveSuccessfulWindows: 2,
        timeoutMs: 1_000,
        retryDelayMs: 0,
        description: 'the injected readiness sequence',
      },
    );

    assert.deepStrictEqual(result, {
      attempts: 4,
      interruptedWindows: 1,
    });
  });

  test('continuous interruption times out instead of passing partially', async () => {
    let attempts = 0;
    await assert.rejects(
      waitForConsecutiveSuccessfulProbeWindows(
        async () => {
          attempts += 1;
          return false;
        },
        {
          consecutiveSuccessfulWindows: 2,
          timeoutMs: 50,
          retryDelayMs: 5,
          description: 'the continuously interrupted injected probe',
        },
      ),
      /Timed out.*interruptedWindows=/,
    );
    assert.ok(attempts > 1, 'The probe should retry readiness windows');
  });

  test('a non-settling probe window is bounded by the hard deadline', async () => {
    let attempts = 0;
    let saveCalls = 0;
    await assert.rejects(
      saveDocumentOnce(
        {
          save: async () => {
            saveCalls += 1;
            return true;
          },
        },
        'non-settling readiness failure',
        () =>
          waitForConsecutiveSuccessfulProbeWindows(
            () => {
              attempts += 1;
              return new Promise<boolean>(() => {});
            },
            {
              consecutiveSuccessfulWindows: 2,
              timeoutMs: 50,
              retryDelayMs: 0,
              description: 'the non-settling injected probe',
            },
          ),
      ),
      /Timed out.*attempts=1/,
    );
    assert.strictEqual(attempts, 1);
    assert.strictEqual(saveCalls, 0, 'A timed-out probe must block the save');
  });

  test('an unexpected probe error fails immediately and blocks save', async () => {
    let attempts = 0;
    let saveCalls = 0;
    const readiness = () =>
      waitForConsecutiveSuccessfulProbeWindows(
        async () => {
          attempts += 1;
          throw new Error('unexpected probe failure');
        },
        {
          consecutiveSuccessfulWindows: 2,
          timeoutMs: 1_000,
          retryDelayMs: 0,
          description: 'the unexpectedly failing probe',
        },
      );

    await assert.rejects(
      saveDocumentOnce(
        {
          save: async () => {
            saveCalls += 1;
            return true;
          },
        },
        'unexpected readiness failure',
        readiness,
      ),
      /unexpected probe failure/,
    );
    assert.strictEqual(attempts, 1, 'Unexpected errors must not be retried');
    assert.strictEqual(saveCalls, 0, 'An unexpected error must block the save');
  });

  test('a never-settling startup operation is bounded by its shared deadline', async () => {
    let attempts = 0;
    await assert.rejects(
      runBeforeDeadline(
        () => {
          attempts += 1;
          return new Promise<void>(() => {});
        },
        Date.now() + 50,
        'the injected startup operation',
      ),
      /Timed out waiting for the injected startup operation.*shared startup deadline/,
    );
    assert.strictEqual(
      attempts,
      1,
      'The startup operation must not be retried',
    );
  });

  test('an expired startup deadline prevents the operation from starting', async () => {
    let attempts = 0;
    await assert.rejects(
      runBeforeDeadline(
        () => {
          attempts += 1;
        },
        Date.now() - 1,
        'the expired injected startup operation',
      ),
      /shared startup deadline has expired/,
    );
    assert.strictEqual(attempts, 0, 'Expired startup work must not begin');
  });

  test('readiness failure blocks save and a false save is never retried', async () => {
    let saveCalls = 0;
    const document = {
      save: async () => {
        saveCalls += 1;
        return false;
      },
    };

    await assert.rejects(
      saveDocumentOnce(document, 'injected save failure', async () => {
        throw new Error('injected readiness failure');
      }),
      /injected readiness failure/,
    );
    assert.strictEqual(
      saveCalls,
      0,
      'A failed readiness probe must block save',
    );

    await assert.rejects(
      saveDocumentOnce(document, 'injected save failure', async () => {}),
      /returned false; the real save was not retried/,
    );
    assert.strictEqual(
      saveCalls,
      1,
      'The real save must be invoked exactly once',
    );
  });

  test('an unrelated delayed provider mutation interrupts vulnerable VS Code', async () => {
    const probe = new CodeActionRegistryProbe();
    let injected = false;
    let unrelatedProvider: vscode.Disposable | undefined;
    let resolveMutation: (() => void) | undefined;
    const mutationCompleted = new Promise<void>((resolve) => {
      resolveMutation = resolve;
    });

    try {
      const result = await probe.wait({
        quietWindowMs: 100,
        consecutiveSuccessfulWindows: 2,
        timeoutMs: 5_000,
        retryDelayMs: 5,
        onAttemptStarted: (attempt) => {
          if (attempt !== 1 || injected) return;
          injected = true;
          setTimeout(() => {
            unrelatedProvider = vscode.languages.registerCodeActionsProvider(
              { scheme: `rslint-unrelated-${randomUUID()}` },
              { provideCodeActions: () => [] },
            );
            setTimeout(() => {
              unrelatedProvider?.dispose();
              unrelatedProvider = undefined;
              resolveMutation?.();
            }, 10);
          }, 20);
        },
      });
      await mutationCompleted;

      assert.ok(injected, 'The provider mutation must be injected in-flight');
      if (result.interruptedWindows > 0) {
        assert.ok(
          result.attempts >= 3,
          'A cancelled window must reset the two-window readiness sequence',
        );
      } else {
        // Forward-compatible path: once VS Code fixes the registry comparison,
        // unrelated providers no longer cancel the filtered request.
        assert.strictEqual(result.attempts, 2);
      }
    } finally {
      unrelatedProvider?.dispose();
      probe.dispose();
    }
  });

  test('a relevant delayed provider mutation always resets readiness', async () => {
    const probe = new CodeActionRegistryProbe();
    let relevantProvider: vscode.Disposable | undefined;
    let mutationCompleted: Promise<void> | undefined;

    try {
      const result = await probe.wait({
        quietWindowMs: 100,
        consecutiveSuccessfulWindows: 2,
        timeoutMs: 5_000,
        retryDelayMs: 5,
        onAttemptStarted: (attempt, probeUri) => {
          if (attempt !== 1 || mutationCompleted) return;
          mutationCompleted = new Promise<void>((resolve) => {
            setTimeout(() => {
              // No kind metadata means this provider is part of the filtered
              // request on both vulnerable and fixed VS Code versions.
              relevantProvider = vscode.languages.registerCodeActionsProvider(
                { scheme: probeUri.scheme },
                { provideCodeActions: () => [] },
              );
              setTimeout(() => {
                relevantProvider?.dispose();
                relevantProvider = undefined;
                resolve();
              }, 10);
            }, 20);
          });
        },
      });
      await mutationCompleted;

      assert.ok(
        result.interruptedWindows > 0,
        'A relevant registry mutation must interrupt an in-flight probe',
      );
      assert.ok(
        result.attempts >= 3,
        'The interrupted window must reset the two-window readiness sequence',
      );
    } finally {
      relevantProvider?.dispose();
      probe.dispose();
    }
  });
});
