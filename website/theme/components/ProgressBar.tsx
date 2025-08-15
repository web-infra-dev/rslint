import React from 'react';
import { Linter } from './bundler';
import Switcher from './Switcher';

export function ProgressBar({ linter, mostRecent, dev }) {
  const testsLeft = mostRecent.total - mostRecent.passing;
  return (
    <div className="flex items-center justify-between px-4 py-2 bg-secondary border-b" data-testid="progress-bar">
      <section className="text-lg">
        <span className="block sm:inline-block">
          <span className="font-semibold mr-4">🦀 RSLint</span>
          <span className="font-semibold">
            {mostRecent.passing} of {mostRecent.total}{' '}
            {dev ? 'development' : 'production'} lint tests passing&nbsp;
          </span>
        </span>
        <span className="block sm:inline-block">
          ({testsLeft > 0 ? <>{testsLeft} left for 100%</> : '100%'})
        </span>
      </section>
      <div className="flex gap-x-4">
        <Switcher />
      </div>
    </div>
  );
}
