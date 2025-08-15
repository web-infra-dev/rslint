import { Linter, getLinter } from './bundler';
import { getProductionLintRuns } from './data';
import IsItReady from './IsItReady';

export default async function IsItReadyProduction() {
  const { mostRecent } = await getProductionLintRuns();

  if (!mostRecent) {
    return null;
  }

  return (
    <IsItReady
      title="Production"
      description="lint tests"
      percent={mostRecent.percent}
      decision={mostRecent.percent >= 90 ? 'yes' : 'no'}
    />
  );
}
