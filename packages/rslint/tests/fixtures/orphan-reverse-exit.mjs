// A reverse config import can outlive the outer lint request because native
// dynamic import() cannot be cancelled. Once the fake Go peer settles that
// outer request, the orphaned handler must not keep this process alive.
import { NodeRslintService } from '@rslint/core/internal';
import { fileURLToPath } from 'node:url';

const service = new NodeRslintService({
  rslintPath: fileURLToPath(new URL('./fake-api-binary.cjs', import.meta.url)),
});
service.setInboundHandler(() => new Promise(() => {}));

await service.sendMessage('orphanReverse', {});
// Intentionally no terminate(): the settled outer request must release the
// child/stdin/stdout refs even though the reverse handler never settles.
