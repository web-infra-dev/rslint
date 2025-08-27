import { BrowserRslintService } from '@rslint/core/browser';
import { RSLintService, Diagnostic } from '@rslint/core/service';
declare const WEB_WORKER_SOURCE_CODE: string;
async function initialize(options: { wasmURL: string }) {
  let blob = new Blob([WEB_WORKER_SOURCE_CODE], { type: 'text/javascript' });
  const service = new RSLintService(
    new BrowserRslintService({
      workerUrl: URL.createObjectURL(blob),
      wasmUrl: options.wasmURL,
    }),
  );

  return service;
}

export { initialize, type RSLintService, type Diagnostic };
