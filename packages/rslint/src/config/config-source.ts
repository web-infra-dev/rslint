import { createHash } from 'node:crypto';

/** The selected config entry passed from activation to plugin workers. */
export interface PluginConfigDescriptor {
  /** Absolute filesystem path of the selected JS/TS config file. */
  configPath: string;
  /** Go-authoritative routing directory, preserved byte-for-byte as configKey. */
  configDirectory: string;
  /** Entry-module version selected by ConfigModuleHost for this activation. */
  sourceFingerprint?: string;
}

export function fingerprintConfigSource(contents: Uint8Array): string {
  return `${contents.byteLength}:${createHash('sha256').update(contents).digest('hex')}`;
}
