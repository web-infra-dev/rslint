export type PlaygroundConfig = Record<string, unknown>[];

const CORE_IMPORT = /from ['"]@rslint\/core['"]/g;

/**
 * A config is arbitrary JavaScript, and a playground link carries it, so the
 * page must never be the thing that runs it. It executes inside a sandboxed
 * iframe instead: without `allow-same-origin` the frame gets an opaque origin,
 * which puts this document's DOM, storage, caches and credentialed same-origin
 * requests out of its reach. It talks to us only through `postMessage`.
 */
const SANDBOX_ATTRIBUTE = 'allow-scripts';

/**
 * The opaque origin keeps a config away from this page's data; this keeps it
 * from phoning home. Without it a config could still reach an attacker's server
 * — `fetch`, an image beacon, `sendBeacon`, a dynamic import — and a link that
 * looks like it points at the docs would quietly hand over the reader's address
 * the moment they opened it.
 *
 * Package CDNs stay reachable because a config is meant to import from one, and
 * allowing them costs nothing here: an attacker cannot read jsDelivr's logs, so
 * a request smuggled into a CDN URL tells them nothing about who made it.
 *
 * `'unsafe-inline'` is not a concession. Every script in this frame is the
 * config the reader is asking to run; the policy exists to bound where that
 * script can reach, not to keep it out.
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
].join('; ');

/**
 * Generous, because the first evaluation of a session waits on `esm.sh` to
 * serve the core package — measured at 2.4s on a fast connection, and a reader
 * on a slow one should not be told their config failed. Nothing is lost by
 * waiting: a spinning frame cannot block this document, and a config that never
 * finishes never finishes at five seconds either.
 */
const EVALUATION_TIMEOUT_MS = 20000;

/**
 * Receives the config source, imports it as a module and hands back the default
 * export. The source travels as a message rather than inlined into this markup
 * so that a `</script` sequence in the config cannot break out of it.
 *
 * The result is serialized here because that is what the lint service does with
 * it anyway: anything that cannot survive JSON would not have reached the
 * linter under the old in-page evaluation either.
 */
const SANDBOX_MARKUP = `<!doctype html><meta charset="utf-8"><meta http-equiv="Content-Security-Policy" content="${SANDBOX_POLICY}"><script type="module">
const post = (message) => parent.postMessage(message, '*');
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
});
post({ ready: true });
<\/script>`;

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

    let timer: number | undefined;
    function settle(action: () => void) {
      if (timer !== undefined) window.clearTimeout(timer);
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
    timer = window.setTimeout(() => {
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
