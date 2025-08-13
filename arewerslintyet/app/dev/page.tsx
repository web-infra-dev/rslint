import { HeapMapExamples } from 'app/HeatMapExamples';
import { Suspense } from 'react';
import Footer from '../Footer';
import GraphDataDevelopment from '../GraphDataDevelopment';
import HeatMapDevelopment from '../HeatMapDevelopment';
import IsItReadyDevelopment from '../IsItReadyDevelopment';
import ProgressBarDevelopment from '../ProgressBarDevelopment';
import { TooltipProvider } from '../TooltipContext';

export default function DevelopmentPage() {
  return (
    <TooltipProvider>
      {/* Development */}
      <Suspense fallback={null}>
        <ProgressBarDevelopment />
      </Suspense>
      <div className="mx-4">
        <Suspense fallback={null}>
          <IsItReadyDevelopment />
        </Suspense>
        <Suspense fallback={null}>
          <GraphDataDevelopment />
        </Suspense>
        <h2 className="text-4xl my-2">Lint Tests</h2>
        <Suspense fallback={null}>
          <HeatMapDevelopment />
        </Suspense>

        <Suspense fallback={null}>
          <HeapMapExamples />
        </Suspense>
      </div>

      <Suspense fallback={null}>
        <Footer />
      </Suspense>
    </TooltipProvider>
  );
}
