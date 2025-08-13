import { getProductionLintRuns } from './data';
import Graph from './Graph';

export default async function GraphDataProduction() {
  const { graphData } = await getProductionLintRuns();
  if (graphData.length === 0) {
    return null;
  }

  return <Graph graphData={graphData} />;
}
