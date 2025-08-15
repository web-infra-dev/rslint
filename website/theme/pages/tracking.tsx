import React from 'react';
import { TooltipProvider } from '../components/TooltipContext';
import HeatMapDevelopment from '../components/HeatMapDevelopment';
import HeatMapProduction from '../components/HeatMapProduction';
import { HeapMapExamples } from '../components/HeatMapExamples';

const TrackingPage: React.FC = () => {
  return (
    <div className="tracking-page">
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-4xl font-bold mb-8 text-center">
          RSLint Progress Tracking
        </h1>

        <div className="mb-12">
          <p className="text-lg text-gray-600 dark:text-gray-300 text-center mb-8">
            Track the progress of RSLint development through comprehensive test
            results and rule implementation status.
          </p>
        </div>

        <TooltipProvider>
          <div className="space-y-12">
            {/* Development Environment */}
            <section>
              <h2 className="text-3xl font-semibold mb-6">
                Development Environment
              </h2>
              <p className="text-gray-600 dark:text-gray-300 mb-4">
                Current test results from the development branch showing the
                latest features and fixes.
              </p>
              <HeatMapDevelopment />
            </section>

            {/* Production Environment */}
            <section>
              <h2 className="text-3xl font-semibold mb-6">
                Production Environment
              </h2>
              <p className="text-gray-600 dark:text-gray-300 mb-4">
                Stable test results from the production-ready codebase.
              </p>
              <HeatMapProduction />
            </section>

            {/* Rule Examples */}
            <section>
              <HeapMapExamples />
            </section>

            {/* Legend */}
            <section className="bg-gray-50 dark:bg-gray-800 p-6 rounded-lg">
              <h3 className="text-xl font-semibold mb-4">Legend</h3>
              <div className="flex flex-wrap gap-6">
                <div className="flex items-center gap-2">
                  <div className="w-4 h-4 bg-passing-square border border-gray-300"></div>
                  <span>Passing Tests</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-4 h-4 bg-failing-square border border-gray-300"></div>
                  <span>Failing Tests</span>
                </div>
              </div>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-4">
                Hover over individual squares to see detailed information about
                specific tests and rules. Click on squares to view the
                corresponding source code on GitHub.
              </p>
            </section>
          </div>
        </TooltipProvider>
      </div>
    </div>
  );
};

export default TrackingPage;
