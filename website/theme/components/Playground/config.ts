export type PlaygroundConfig = Record<string, unknown>[];

const CORE_IMPORT = /from ['"]@rslint\/core['"]/g;

/**
 * A config is arbitrary JavaScript, and a playground link carries it, so the
 * page must never be the thing that runs it. It executes in a dedicated worker
 * inside a sandboxed iframe: without `allow-same-origin` the frame gets an
 * opaque origin, which puts this document's DOM, storage, caches and
 * credentialed same-origin requests out of its reach. It talks to us only
 * through `postMessage`.
 */
const SANDBOX_ATTRIBUTE = 'allow-scripts';

/**
 * The Blob worker inherits this policy, which restricts its network access to
 * the package CDNs. A worker has no DOM and cannot navigate its own location:
 * running the config in the iframe itself would let it bypass this policy by
 * navigating the frame to an arbitrary URL. Only the trusted bridge below runs
 * in the iframe; config source is passed straight to the worker.
 */
const MODULE_HOSTS = [
  'https://esm.sh',
  'https://cdn.jsdelivr.net',
  'https://unpkg.com',
].join(' ');
const SANDBOX_POLICY = [
  "default-src 'none'",
  `script-src 'unsafe-inline' blob: ${MODULE_HOSTS}`,
  `connect-src ${MODULE_HOSTS}`,
  'worker-src blob:',
].join('; ');

/**
 * Generous, because the first evaluation of a session waits on `esm.sh` to
 * serve the core package — measured at 2.4s on a fast connection, and a reader
 * on a slow one should not be told their config failed. Nothing is lost by
 * waiting: a spinning worker cannot block this document, and a config that never
 * finishes never finishes at five seconds either.
 */
const EVALUATION_TIMEOUT_MS = 20000;

/**
 * Receives the config source in a worker, imports it as a module and hands back
 * the default export. Source is sent as a message, never inlined into markup.
 *
 * The result is serialized here because that is what the lint service does with
 * it anyway: anything that cannot survive JSON would not have reached the
 * linter under the old in-page evaluation either.
 */
const SANDBOX_WORKER = `
const post = (message) => self.postMessage(message);
addEventListener('error', (event) => post({ error: String(event.message) }));
addEventListener('unhandledrejection', (event) => post({ error: String(event.reason) }));
addEventListener('message', async (event) => {
  const url = URL.createObjectURL(new Blob([event.data], { type: 'text/javascript' }));
  try {
    const value = (await import(url)).default;
    post({ config: JSON.stringify(value === undefined ? null : value) });
  } catch (error) {
    post({ error: String(error) });
  } finally {
    URL.revokeObjectURL(url);
  }
});`;

// The closing tag is escaped so it cannot close an inlined application bundle.
const SANDBOX_MARKUP = `<!doctype html><meta charset="utf-8"><meta http-equiv="Content-Security-Policy" content="${SANDBOX_POLICY}"><script type="module">
const post = (message) => parent.postMessage(message, '*');
addEventListener('error', (event) => post({ error: String(event.message) }));
const url = URL.createObjectURL(new Blob([${JSON.stringify(SANDBOX_WORKER)}], { type: 'text/javascript' }));
// A classic bootstrap can load its Blob URL from an opaque origin. It still
// imports the user config as an ES module.
const worker = new Worker(url);
worker.addEventListener('error', (event) => post({ error: event.message || 'Unable to start the config worker.' }));
worker.addEventListener('message', ({ data }) => {
  if (typeof data?.config !== 'string' && typeof data?.error !== 'string') return;
  worker.terminate();
  URL.revokeObjectURL(url);
  post(typeof data.config === 'string' ? { config: data.config } : { error: data.error });
});
addEventListener('message', (event) => {
  if (event.source === parent && typeof event.data === 'string') {
    worker.postMessage(event.data);
  }
});
post({ ready: true });
<\u002Fscript>`;

interface SandboxMessage {
  ready?: boolean;
  config?: string;
  error?: string;
}

function prepareModule(source: string, wasmVersion: string): string {
  const coreUrl = `https://esm.sh/@rslint/core@${encodeURIComponent(wasmVersion)}?target=es2022`;
  return source.replace(CORE_IMPORT, `from ${JSON.stringify(coreUrl)}`);
}

function normalizeConfig(value: unknown): PlaygroundConfig {
  if (!Array.isArray(value)) {
    throw new Error('rslint.config.js must default-export an array.');
  }
  const entries = value.flat();
  if (
    entries.some(
      (entry) =>
        entry === null || typeof entry !== 'object' || Array.isArray(entry),
    )
  ) {
    throw new Error('rslint.config.js must contain only config objects.');
  }
  return entries as PlaygroundConfig;
}

function runInSandbox(source: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const iframe = document.createElement('iframe');
    iframe.setAttribute('sandbox', SANDBOX_ATTRIBUTE);
    iframe.setAttribute('aria-hidden', 'true');
    iframe.style.display = 'none';

    function settle(action: () => void) {
      window.clearTimeout(timer);
      window.removeEventListener('message', onMessage);
      iframe.remove();
      action();
    }

    function onMessage(event: MessageEvent) {
      // The frame has an opaque origin, so identity comes from the window, not
      // from `event.origin` — which is the string "null" for every such frame.
      if (event.source !== iframe.contentWindow) return;
      const message = event.data as SandboxMessage;
      if (message?.ready) {
        iframe.contentWindow?.postMessage(source, '*');
        return;
      }
      if (typeof message?.config === 'string') {
        const serialized = message.config;
        settle(() => resolve(serialized));
        return;
      }
      if (message?.error !== undefined) {
        const reason = message.error;
        settle(() => reject(new Error(reason)));
      }
    }

    window.addEventListener('message', onMessage);
    // Assigned before the frame is attached, so nothing can settle before it
    // exists even though `settle` closes over it.
    const timer = window.setTimeout(() => {
      settle(() =>
        reject(
          new Error(
            `rslint.config.js did not finish within ${EVALUATION_TIMEOUT_MS}ms.`,
          ),
        ),
      );
    }, EVALUATION_TIMEOUT_MS);

    iframe.srcdoc = SANDBOX_MARKUP;
    document.body.appendChild(iframe);
  });
}

/** Execute a config against the selected @rslint/wasm release's core package. */
export async function evaluateConfig(
  source: string,
  wasmVersion: string,
): Promise<PlaygroundConfig> {
  const serialized = await runInSandbox(prepareModule(source, wasmVersion));
  return normalizeConfig(JSON.parse(serialized));
}
