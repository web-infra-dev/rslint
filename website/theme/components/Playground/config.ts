export type PlaygroundConfig = Record<string, unknown>[];

const CORE_IMPORT = /from ['"]@rslint\/core['"]/g;

function prepareModule(source: string, wasmVersion: string): string {
  const coreUrl = `https://esm.sh/@rslint/core@${encodeURIComponent(wasmVersion)}?target=es2022`;
  return source.replace(CORE_IMPORT, `from ${JSON.stringify(coreUrl)}`);
}

function normalizeConfig(value: unknown): PlaygroundConfig {
  if (!Array.isArray(value)) {
    throw new Error("rslint.config.js must default-export an array.");
  }
  const entries = value.flat();
  if (
    entries.some(
      (entry) =>
        entry === null || typeof entry !== "object" || Array.isArray(entry),
    )
  ) {
    throw new Error("rslint.config.js must contain only config objects.");
  }
  return entries as PlaygroundConfig;
}

/** Execute a config against the selected @rslint/wasm release's core package. */
export async function evaluateConfig(
  source: string,
  wasmVersion: string,
): Promise<PlaygroundConfig> {
  const moduleUrl = URL.createObjectURL(
    new Blob([prepareModule(source, wasmVersion)], {
      type: "text/javascript",
    }),
  );
  try {
    const module = await import(/* rspackIgnore: true */ moduleUrl);
    return normalizeConfig(module.default);
  } finally {
    URL.revokeObjectURL(moduleUrl);
  }
}
