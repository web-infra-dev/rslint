import React, { useState } from 'react';
import { ChevronRightIcon } from 'lucide-react';

interface ExpandableSectionProps {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
}

export const ExpandableSection: React.FC<ExpandableSectionProps> = ({
  title,
  children,
  defaultOpen = false,
}) => {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div className="my-1">
      <button
        className="flex w-full items-center gap-1 rounded px-1 py-0.5 text-left hover:bg-gray-100"
        onClick={() => setOpen(!open)}
      >
        <ChevronRightIcon
          className={`h-4 w-4 text-gray-500 transition-transform ${open ? 'rotate-90' : ''}`}
        />
        <span className="font-medium text-blue-700">{title}</span>
      </button>
      {open && (
        <div className="ml-4 border-l border-gray-200 pl-2">{children}</div>
      )}
    </div>
  );
};
