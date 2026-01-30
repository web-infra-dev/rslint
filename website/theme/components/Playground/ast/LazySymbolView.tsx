import React, { useState, useCallback } from 'react';
import { useAstInfoContextOptional } from './AstInfoContext';
import type { SymbolInfo } from './types';
import { SymbolInfoView } from './SymbolInfoView';
import { ControlledCollapsibleItem } from './CollapsibleItem';

interface LazySymbolViewProps {
  pos?: number;
  label: string;
  preview: string;
  shallow: SymbolInfo;
}

export const LazySymbolView: React.FC<LazySymbolViewProps> = ({
  pos,
  label,
  preview,
  shallow,
}) => {
  const context = useAstInfoContextOptional();
  const [expanded, setExpanded] = useState(false);
  const [loading, setLoading] = useState(false);
  const [fullData, setFullData] = useState<SymbolInfo | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Check if shallow data is already full data (has id field)
  const isAlreadyFull = shallow.id !== undefined;
  // Check if this symbol has a valid position to fetch from
  // pos is always valid now (external files have real positions + fileName)
  const hasValidPos = pos !== undefined;

  const handleToggle = useCallback(async () => {
    if (expanded) {
      setExpanded(false);
      return;
    }

    // If already have full data, or no valid position, just expand
    if (isAlreadyFull || !hasValidPos || !context) {
      setExpanded(true);
      return;
    }

    setLoading(true);
    setError(null);
    try {
      // Pass fileName (for external files) to fetch from the correct source file
      const result = await context.fetchAstInfo(
        pos!,
        undefined,
        undefined,
        shallow.fileName,
      );
      if (result?.symbol) {
        setFullData(result.symbol);
      }
      setExpanded(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to fetch');
      setExpanded(true);
    } finally {
      setLoading(false);
    }
  }, [expanded, context, pos, shallow.fileName, isAlreadyFull, hasValidPos]);

  // Title: Symbol(name)
  const title = `Symbol(${shallow.name})`;

  return (
    <ControlledCollapsibleItem
      label={label}
      title={title}
      preview={preview}
      expanded={expanded}
      loading={loading}
      onToggle={handleToggle}
    >
      {loading ? (
        <div className="text-gray-400 text-xs py-1">Loading...</div>
      ) : error ? (
        <div className="text-red-500 text-xs py-1">{error}</div>
      ) : (
        <SymbolInfoView info={fullData ?? shallow} />
      )}
    </ControlledCollapsibleItem>
  );
};
