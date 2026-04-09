import React, { ReactNode } from 'react';

// Badge component
export const Badge: React.FC<{ children: ReactNode }> = ({ children }) => (
  <span className="inline-flex items-center rounded-xs bg-gray-50 dark:bg-gray-800 px-2 py-1 text-xs font-medium text-gray-600 dark:text-gray-300 inset-ring inset-ring-gray-500/10 dark:inset-ring-gray-400/20">
    {children}
  </span>
);

/** Emoji per preset level. Extend here when adding strict / stylistic. */
const PRESET_EMOJI: Record<string, string> = {
  recommended: '\u2705',
};

export const PresetBadge: React.FC<{ preset: string }> = ({ preset }) => (
  <span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 dark:bg-emerald-950/40 px-2.5 py-0.5 text-xs font-medium text-emerald-800 dark:text-emerald-300">
    <span className="text-[0.7rem] leading-none">
      {PRESET_EMOJI[preset] || '\u2705'}
    </span>
    {preset}
  </span>
);

// Heading component
export const Heading: React.FC<{ children: ReactNode }> = ({ children }) => (
  <p className="scroll-m-20 text-xl tracking-tight first:mt-0">{children}</p>
);

// Text component
export const Text: React.FC<{ className?: string; children?: ReactNode }> = ({
  className,
  children,
}) => (
  <div className={`leading-7 [&:not(:first-child)]:mt-1 ${className || ''}`}>
    {children}
  </div>
);
