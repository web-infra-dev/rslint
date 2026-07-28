import { existsSync, writeFileSync } from 'node:fs';

export default {
  rules: {
    hang: {
      meta: { type: 'problem', schema: [] },
      create(context) {
        return {
          Program(node) {
            const started = process.env.RSLINT_API_CLOSE_STARTED;
            const release = process.env.RSLINT_API_CLOSE_RELEASE;
            if (!started || !release) {
              throw new Error('close gate paths are required');
            }
            writeFileSync(started, 'started');
            const waitArray = new Int32Array(new SharedArrayBuffer(4));
            // Synchronous polling deliberately keeps the worker unable to
            // consume its shutdown message. The parent releases this gate only
            // after close() has started. A missing shutdown drain lets this
            // unique diagnostic escape and fails the outer assertion.
            while (!existsSync(release)) {
              // Keep the worker synchronously blocked while yielding the CPU,
              // so the fixture does not amplify host starvation on Windows.
              Atomics.wait(waitArray, 0, 0, 10);
            }
            context.report({
              node,
              message: 'CLOSE_GATE_LEAKED',
            });
          },
        };
      },
    },
  },
};
