import { Linter, getLinter } from './bundler';
import { getDevelopmentLintRuns } from './data';
import IsItReady from './IsItReady';

export default async function IsItReadyDevelopment() {
  const { mostRecent } = await getDevelopmentLintRuns();

  if (!mostRecent) {
    return null;
  }

  return (
    <IsItReady
      title="Development"
      description="lint tests"
      percent={mostRecent.percent}
      decision={mostRecent.percent >= 90 ? 'yes' : 'no'}
    />
  );
}
