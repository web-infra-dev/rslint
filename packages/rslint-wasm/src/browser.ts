import { BrowserRslintService } from '@rslint/core/browser';
import { RSLintService } from '@rslint/core/service';
export async function initialize(options: { wasmURL: string }) {
  let blob = new Blob([WEB_WORKER_SOURCE_CODE], { type: 'text/javascript' });
  const service = new RSLintService(
    new BrowserRslintService({
      workerUrl: URL.createObjectURL(blob),
      wasmUrl: options.wasmURL,
    }),
  );

  return service;
}
