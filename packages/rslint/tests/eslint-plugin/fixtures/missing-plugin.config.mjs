// Intentionally references a non-existent plugin to exercise the
// worker's "config import failed" path. The worker pool's
// `loadPluginsFromConfigFile` should surface the import error verbatim
// in the init-error message, which the WorkerPool's init() promise
// rejects with. We assert the message includes the missing specifier.
import nope from 'eslint-plugin-this-does-not-exist';

export default [
  {
    plugins: {
      nope,
    },
  },
];
