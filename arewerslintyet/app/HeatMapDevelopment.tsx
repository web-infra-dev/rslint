import { getDevelopmentLintResults } from './data';
import { HeapMap } from './HeatMap';

export default async function HeatMapDevelopment() {
  const data = await getDevelopmentLintResults();

  if (!data) {
    return null;
  }

  return (
    <section className="HeatMap">
      <HeapMap lintResults={data} />
    </section>
  );
}
