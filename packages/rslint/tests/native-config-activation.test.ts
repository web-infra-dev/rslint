import { describe, expect, test } from '@rstest/core';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import {
  PluginHostLifecycle,
  stageNativeConfigActivation,
} from '../src/api/rslint.js';
import { ConfigModuleHost } from '../src/config/config-loader.js';

describe('native API config activation', () => {
  test('shuts down a newly created plugin host when its config changes during creation', async () => {
    const root = fs.mkdtempSync(
      path.join(os.tmpdir(), 'rslint-api-config-activation-'),
    );
    const configPath = path.join(root, 'rslint.config.mjs');
    fs.writeFileSync(configPath, '// original config bytes\n');
    const configHost = new ConfigModuleHost({
      loadCached: async () => [
        {
          plugins: { local: { rules: { example: {} } } },
        },
      ],
    });
    const request = {
      transactionId: 'api-prepare-race',
      effectiveConfigIds: ['root'],
    } as const;
    let createCalls = 0;
    let shutdownCalls = 0;

    try {
      await configHost.loadConfigs({
        transactionId: request.transactionId,
        loadMode: 'cached',
        candidates: [
          {
            id: 'root',
            configPath,
            configDirectory: root,
          },
        ],
      });

      await expect(
        stageNativeConfigActivation(
          configHost,
          request,
          async () => async () => {
            createCalls++;
            fs.writeFileSync(configPath, '// bytes imported by worker\n');
            return {
              async lint() {
                return { results: [] };
              },
              async shutdown() {
                shutdownCalls++;
              },
            };
          },
          () => undefined,
          () => false,
        ),
      ).rejects.toThrow('plugin host was being prepared');
      expect(createCalls).toBe(1);
      expect(shutdownCalls).toBe(1);
    } finally {
      configHost.deleteSession(request.transactionId);
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('registers a staged host before post-prepare verification so close can drain it', async () => {
    const root = fs.mkdtempSync(
      path.join(os.tmpdir(), 'rslint-api-config-staged-close-'),
    );
    const configPath = path.join(root, 'rslint.config.mjs');
    fs.writeFileSync(configPath, '// stable config bytes\n');
    let fingerprintReads = 0;
    let releasePostPrepare!: () => void;
    let markPostPrepareStarted!: () => void;
    const postPrepareStarted = new Promise<void>((resolve) => {
      markPostPrepareStarted = resolve;
    });
    const postPrepareGate = new Promise<void>((resolve) => {
      releasePostPrepare = resolve;
    });
    const configHost = new ConfigModuleHost({
      loadCached: async () => [
        { plugins: { local: { rules: { example: {} } } } },
      ],
      readSource: async (sourcePath) => {
        fingerprintReads++;
        if (fingerprintReads === 4) {
          markPostPrepareStarted();
          await postPrepareGate;
        }
        return fs.promises.readFile(sourcePath);
      },
    });
    const request = {
      transactionId: 'api-staged-close',
      effectiveConfigIds: ['root'],
    } as const;
    const lifecycle = new PluginHostLifecycle();
    let closing = false;
    let shutdownCalls = 0;

    try {
      await configHost.loadConfigs({
        transactionId: request.transactionId,
        loadMode: 'cached',
        candidates: [
          {
            id: 'root',
            configPath,
            configDirectory: root,
          },
        ],
      });

      const activation = stageNativeConfigActivation(
        configHost,
        request,
        async () => async () => ({
          async lint() {
            return { results: [] };
          },
          async shutdown() {
            shutdownCalls++;
          },
        }),
        () => undefined,
        () => closing,
        lifecycle,
      );
      await postPrepareStarted;
      closing = true;
      await lifecycle.shutdownAll();
      expect(shutdownCalls).toBe(1);

      releasePostPrepare();
      await expect(activation).rejects.toThrow('rslint service is closing');
      expect(shutdownCalls).toBe(1);
    } finally {
      releasePostPrepare();
      configHost.deleteSession(request.transactionId);
      fs.rmSync(root, { recursive: true, force: true });
    }
  });
});
