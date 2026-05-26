/**
 * Fixture plugin used by worker-pool-respawn tests.
 *
 * The `hang` rule's listener enters an infinite loop on the first
 * `Program` node — there's no yield point, so the per-task cancel
 * SAB flag never gets read. The only way to free the worker is for
 * the pool's `taskTimeoutMs` to fire `worker.terminate()`, which
 * triggers the exit handler's respawn path.
 *
 * The `noop` rule is a non-hanging companion used to drive a follow-
 * up task on the newly respawned worker. Its presence verifies the
 * pool isn't permanently broken after the timeout.
 */
export default {
  meta: { name: 'hang-plugin', version: '0.0.0' },
  rules: {
    hang: {
      meta: { type: 'problem' },
      create() {
        return {
          Program() {
            // Spin forever. Worker can only exit via terminate().
            // eslint-disable-next-line no-constant-condition
            while (true) {
              // empty
            }
          },
        };
      },
    },
    noop: {
      meta: { type: 'problem' },
      create(ctx) {
        return {
          Identifier(node) {
            if (node.name === 'TRIGGER') {
              ctx.report({ node, message: 'noop fired' });
            }
          },
        };
      },
    },
  },
};
