import React, { useState, useCallback } from 'react';

interface CollapsibleItemProps {
  /** Optional prefix label (e.g., array index "0", property name "Parent") */
  label?: string;
  /** Main title (e.g., "KindIdentifier", "Symbol(foo)", "Type[Object]") */
  title: string;
  /** Preview content shown when collapsed */
  preview?: string;
  /** Children to render when expanded */
  children: React.ReactNode;
  /** Whether expanded by default */
  defaultExpanded?: boolean;
  /** Whether the title should be bold (for top-level items) */
  bold?: boolean;
  /** Whether currently loading */
  loading?: boolean;
  /** Callback when mouse enters the header */
  onMouseEnter?: () => void;
  /** Callback when mouse leaves the header */
  onMouseLeave?: () => void;
  /** Custom click handler (for lazy loading) */
  onToggle?: () => void;
}

export const CollapsibleItem: React.FC<CollapsibleItemProps> = ({
  label,
  title,
  preview,
  children,
  defaultExpanded = false,
  bold = false,
  loading = false,
  onMouseEnter,
  onMouseLeave,
  onToggle,
}) => {
  const [expanded, setExpanded] = useState(defaultExpanded);

  const handleToggle = useCallback(() => {
    if (onToggle) {
      onToggle();
    }
    setExpanded(prev => !prev);
  }, [onToggle]);

  return (
    <div className="font-mono text-xs leading-relaxed">
      <div
        className="flex items-center py-0.5"
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}
      >
        <span
          className="text-[10px] text-gray-500 w-4 flex-shrink-0 text-center cursor-pointer hover:bg-gray-200 rounded"
          onClick={handleToggle}
        >
          {loading ? '⟳' : expanded ? '▼' : '▶'}
        </span>
        {label && (
          <>
            <span
              className="text-purple-600 cursor-pointer hover:underline"
              onClick={handleToggle}
            >
              {label}
            </span>
            <span className="text-gray-500">:</span>
          </>
        )}
        <span
          className={`text-blue-600 cursor-pointer hover:underline ${label ? 'ml-1' : ''} ${bold ? 'font-semibold' : ''}`}
          onClick={handleToggle}
        >
          {title}
        </span>
        {!expanded && preview && (
          <span className="text-gray-400 ml-1 truncate max-w-[300px]">
            {'{'}
            {preview}
            {'}'}
          </span>
        )}
        {expanded && <span className="text-gray-500 ml-1">{'{'}</span>}
      </div>
      {expanded && (
        <>
          <div className="ml-4 py-1 border-l border-gray-200 pl-2">
            {children}
          </div>
          <div className="text-gray-500 ml-4">{'}'}</div>
        </>
      )}
    </div>
  );
};

// Controlled version for lazy loading scenarios
interface ControlledCollapsibleItemProps {
  label?: string;
  title: string;
  preview?: string;
  children: React.ReactNode;
  expanded: boolean;
  loading?: boolean;
  onToggle: () => void;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
  bold?: boolean;
}

export const ControlledCollapsibleItem: React.FC<
  ControlledCollapsibleItemProps
> = ({
  label,
  title,
  preview,
  children,
  expanded,
  loading = false,
  onToggle,
  onMouseEnter,
  onMouseLeave,
  bold = false,
}) => {
  return (
    <div className="font-mono text-xs leading-relaxed">
      <div
        className="flex items-center py-0.5"
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}
      >
        <span
          className="text-[10px] text-gray-500 w-4 flex-shrink-0 text-center cursor-pointer hover:bg-gray-200 rounded"
          onClick={onToggle}
        >
          {loading ? '⟳' : expanded ? '▼' : '▶'}
        </span>
        {label && (
          <>
            <span
              className="text-purple-600 cursor-pointer hover:underline"
              onClick={onToggle}
            >
              {label}
            </span>
            <span className="text-gray-500">:</span>
          </>
        )}
        <span
          className={`text-blue-600 cursor-pointer hover:underline ${label ? 'ml-1' : ''} ${bold ? 'font-semibold' : ''}`}
          onClick={onToggle}
        >
          {title}
        </span>
        {!expanded && preview && (
          <span className="text-gray-400 ml-1 truncate max-w-[300px]">
            {'{'}
            {preview}
            {'}'}
          </span>
        )}
        {expanded && <span className="text-gray-500 ml-1">{'{'}</span>}
      </div>
      {expanded && (
        <>
          <div className="ml-4 py-1 border-l border-gray-200 pl-2">
            {children}
          </div>
          <div className="text-gray-500 ml-4">{'}'}</div>
        </>
      )}
    </div>
  );
};
