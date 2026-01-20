import fs from 'node:fs';
import path from 'node:path';
import { execFileSync, spawnSync } from 'node:child_process';

import { colorizeDiagnostics, colorizeError } from './color.js';
import { parseEnvBool } from './env.js';
import {
  MESSAGE_TYPE_REQUEST,
  MESSAGE_TYPE_RESPONSE,
  MESSAGE_TYPE_ERROR,
  encodeMessage,
  decodeMessage,
} from './msgpack.js';
import { findProject, isInNodeModules } from './project.js';
import { resolveTsgoExecutable } from './tsgo.js';
import { createRequire } from 'node:module';

// eslint-disable-next-line @typescript-eslint/no-require-imports
const esmRequire = createRequire(import.meta.url);

export interface RegisterOptions {
  project?: string;
  module?: string;
  target?: string;
  jsx?: string;
  typecheck?: boolean;
  inlineSourceMap?: boolean;
  sourceMap?: boolean;
  extensions?: string[];
  transpileNodeModules?: boolean;
  cwd?: string;
  ignore?: (filename: string) => boolean;
}

export interface RegisterResult {
  extensions: string[];
  project: string | null;
  module: string;
  target?: string;
  jsx?: string;
  typecheck: boolean;
  inlineSourceMap: boolean;
  sourceMap: boolean;
  undo(): void;
}

interface CacheEntry {
  mtimeMs: number;
  code: string;
}

interface TranspilePayload {
  fileName: string;
  configFileName?: string;
  module: string;
  target?: string;
  jsx?: string;
  inlineSourceMap: boolean;
  sourceMap: boolean;
  typecheck: boolean;
}

interface TranspileResponse {
  outputText?: string;
  diagnosticsText?: string;
  hasError?: boolean;
}

type ExtensionHandler = (module: NodeModule, filename: string) => void;

const DEFAULT_EXTENSIONS = ['.ts', '.tsx', '.cts'];

export function register(options: RegisterOptions = {}): RegisterResult {
  const env = process.env;
  const cwd = options.cwd ? path.resolve(options.cwd) : process.cwd();
  const project = options.project ?? env.RSRUN_PROJECT ?? findProject(cwd);
  const projectPath = project ? path.resolve(cwd, project) : null;
  const moduleKind = options.module ?? env.RSRUN_MODULE ?? 'commonjs';
  const target = options.target ?? env.RSRUN_TARGET;
  const jsx = options.jsx ?? env.RSRUN_JSX;
  const typecheck =
    options.typecheck ?? parseEnvBool(env.RSRUN_TYPECHECK, false);
  const inlineSourceMap =
    options.inlineSourceMap ?? parseEnvBool(env.RSRUN_INLINE_SOURCE_MAP, true);
  const sourceMap =
    options.sourceMap ?? parseEnvBool(env.RSRUN_SOURCE_MAP, false);
  const extensions = options.extensions ?? DEFAULT_EXTENSIONS;
  const transpileNodeModules = options.transpileNodeModules ?? false;
  const ignore =
    options.ignore ??
    ((filename: string) => !transpileNodeModules && isInNodeModules(filename));
  const useTranspileServer = typecheck;

  const tsgoPath = resolveTsgoExecutable();
  const cache = new Map<string, CacheEntry>();
  const errorBaseDir = projectPath ? path.dirname(projectPath) : cwd;

  const baseArgs = ['--transpile', '--module', moduleKind];
  if (projectPath) {
    baseArgs.push('--config', projectPath);
  }
  if (target) {
    baseArgs.push('--target', target);
  }
  if (jsx) {
    baseArgs.push('--jsx', jsx);
  }
  if (typecheck) {
    baseArgs.push('--typecheck');
  }
  if (inlineSourceMap) {
    baseArgs.push('--inlineSourceMap');
  }
  if (sourceMap) {
    baseArgs.push('--sourceMap');
  }

  const previousHandlers = new Map<string, ExtensionHandler | undefined>();
  const defaultJsHandler = esmRequire.extensions['.js'];

  function formatErrorPath(filePath: string): string {
    if (!filePath || !errorBaseDir) {
      return filePath;
    }
    let rel = path.relative(errorBaseDir, filePath);
    if (!rel || rel === '') {
      rel = path.basename(filePath);
    }
    if (path.isAbsolute(rel)) {
      return filePath;
    }
    return rel.split(path.sep).join('/');
  }

  function failTranspile(absPath: string, detail?: string): never {
    if (detail) {
      const message = detail.endsWith('\n') ? detail : `${detail}\n`;
      process.stderr.write(colorizeError(message));
    }
    const message = `rsrun: failed to transpile ${formatErrorPath(absPath)}\n`;
    process.stderr.write(colorizeError(message));
    process.exit(1);
  }

  function compileWithServer(absPath: string): string {
    const payload: TranspilePayload = {
      fileName: absPath,
      configFileName: projectPath || undefined,
      module: moduleKind,
      target: target || undefined,
      jsx: jsx || undefined,
      inlineSourceMap,
      sourceMap,
      typecheck,
    };
    const input = encodeMessage(
      MESSAGE_TYPE_REQUEST,
      'transpile',
      Buffer.from(JSON.stringify(payload), 'utf8'),
    );
    const serverCwd = projectPath ? path.dirname(projectPath) : cwd;
    const result = spawnSync(
      tsgoPath,
      ['--transpileServer', '--cwd', serverCwd],
      {
        input,
        cwd: serverCwd,
        maxBuffer: 1024 * 1024 * 20,
        encoding: 'buffer',
      },
    );
    if (result.error) {
      failTranspile(absPath, result.error.message);
    }
    if (result.stderr && result.stderr.length > 0) {
      process.stderr.write(result.stderr);
    }
    if (result.status != null && result.status !== 0) {
      failTranspile(absPath, `tsgo exited with status ${result.status}`);
    }
    const stdout = result.stdout || Buffer.alloc(0);
    let decodedMessage: ReturnType<typeof decodeMessage>;
    try {
      decodedMessage = decodeMessage(stdout, 0);
    } catch (error) {
      const detail =
        error instanceof Error ? error.message : 'invalid api response';
      failTranspile(absPath, detail);
    }
    if (decodedMessage.messageType === MESSAGE_TYPE_ERROR) {
      failTranspile(absPath, decodedMessage.payload.toString('utf8'));
    }
    if (decodedMessage.messageType !== MESSAGE_TYPE_RESPONSE) {
      failTranspile(absPath, 'unexpected api response');
    }
    let parsed: TranspileResponse;
    try {
      parsed = JSON.parse(decodedMessage.payload.toString('utf8'));
    } catch {
      failTranspile(absPath, 'failed to parse api response');
    }
    if (parsed && parsed.diagnosticsText) {
      const text = String(parsed.diagnosticsText);
      const output = text.endsWith('\n') ? text : `${text}\n`;
      process.stderr.write(colorizeDiagnostics(output));
    }
    if (parsed && parsed.hasError) {
      failTranspile(absPath);
    }
    if (!parsed || !parsed.outputText) {
      failTranspile(absPath, 'no output produced');
    }
    return String(parsed.outputText);
  }

  function compile(filename: string): string {
    const absPath = path.resolve(filename);
    const stat = fs.statSync(absPath);
    const cached = cache.get(absPath);
    if (cached && cached.mtimeMs === stat.mtimeMs) {
      return cached.code;
    }
    let output: string;
    if (useTranspileServer) {
      output = compileWithServer(absPath);
    } else {
      const args = baseArgs.concat(['--file', absPath]);
      try {
        output = execFileSync(tsgoPath, args, {
          encoding: 'utf8',
          cwd: projectPath ? path.dirname(projectPath) : cwd,
          maxBuffer: 1024 * 1024 * 20,
        });
      } catch (error) {
        const stderr =
          error && typeof error === 'object' && 'stderr' in error
            ? String(error.stderr)
            : '';
        if (stderr) {
          process.stderr.write(stderr.endsWith('\n') ? stderr : `${stderr}\n`);
        }
        const detail = stderr
          ? ''
          : error instanceof Error
            ? error.message
            : 'unknown error';
        if (detail) {
          process.stderr.write(detail.endsWith('\n') ? detail : `${detail}\n`);
        }
        failTranspile(absPath);
      }
    }
    cache.set(absPath, { mtimeMs: stat.mtimeMs, code: output });
    return output;
  }

  function makeLoader(ext: string): ExtensionHandler {
    const previous = previousHandlers.get(ext) || defaultJsHandler;
    return function loader(module: NodeModule, filename: string): void {
      if (filename.endsWith('.d.ts')) {
        throw new Error(`rsrun: cannot execute declaration file: ${filename}`);
      }
      if (ignore(filename)) {
        if (previous) {
          previous(module, filename);
        }
        return;
      }
      const code = compile(filename);
      (
        module as NodeModule & {
          _compile: (code: string, filename: string) => void;
        }
      )._compile(code, filename);
    };
  }

  for (const ext of extensions) {
    previousHandlers.set(ext, esmRequire.extensions[ext]);
    esmRequire.extensions[ext] = makeLoader(ext);
  }

  return {
    extensions: [...extensions],
    project: projectPath,
    module: moduleKind,
    target,
    jsx,
    typecheck,
    inlineSourceMap,
    sourceMap,
    undo(): void {
      for (const ext of extensions) {
        const previous = previousHandlers.get(ext);
        if (previous) {
          esmRequire.extensions[ext] = previous;
        } else {
          delete esmRequire.extensions[ext];
        }
      }
    },
  };
}
