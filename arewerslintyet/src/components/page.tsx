import { HeapMapExamples } from './HeatMapExamples';
import { Suspense } from 'react';
import Footer from './Footer';
import GraphDataProduction from './GraphDataProduction';
import HeatMapProduction from './HeatMapProduction';
import IsitReadyProduction from './IsitReadyProduction';
import ProgressBarProduction from './ProgressBarProduction';
import { TooltipProvider } from './TooltipContext';

export default function Homepage() {
  return (
    <TooltipProvider>
      {/* Production */}
      <Suspense fallback={null}>
        <ProgressBarProduction />
      </Suspense>
      <div className="mx-4">
        <Suspense fallback={null}>
          <IsitReadyProduction />
        </Suspense>
        <Suspense fallback={null}>
          <GraphDataProduction />
        </Suspense>
        <h2 className="text-4xl my-2">RSLint Tests</h2>
        <Suspense fallback={null}>
          <HeatMapProduction />
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
