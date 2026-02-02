import React, { useState, useCallback } from 'react';
import { useAstInfoContextOptional } from './AstInfoContext';
import type { NodeInfo } from './types';
import { NodeInfoView } from './NodeInfoView';
import { ControlledCollapsibleItem } from './CollapsibleItem';

interface LazyNodeViewProps {
  pos?: number;
  label: string;
  preview: string;
  kindName?: string;
  shallow: NodeInfo;
}

export const LazyNodeView: React.FC<LazyNodeViewProps> = ({
  pos,
  label,
  preview,
  kindName,
  shallow,
}) => {
  const context = useAstInfoContextOptional();
  const [expanded, setExpanded] = useState(false);
  const [loading, setLoading] = useState(false);
  const [fullData, setFullData] = useState<NodeInfo | null>(null);
  const [error, setError] = useState<string | null>(null);

  const nodePos = pos ?? shallow.pos;
  // Check if shallow data is already full data (has id field)
  const isAlreadyFull = shallow.id !== undefined;
  // Check if this node has a valid position to fetch from
  // pos is always valid now (external files have real positions + fileName)
  const hasValidPos = nodePos !== undefined;

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
      // Pass pos, end, kind, and fileName (for external files) to fetch the exact node
      const result = await context.fetchAstInfo(
        nodePos!,
        shallow.end,
        shallow.kind,
        shallow.fileName,
      );
      if (result?.node) {
        setFullData(result.node);
      }
      setExpanded(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to fetch');
      setExpanded(true);
    } finally {
      setLoading(false);
    }
  }, [
    expanded,
    context,
    nodePos,
    shallow.end,
    shallow.kind,
    shallow.fileName,
    isAlreadyFull,
    hasValidPos,
  ]);

  const handleMouseEnter = useCallback(() => {
    // Only highlight for nodes in current file (no fileName means current file)
    if (
      context &&
      !shallow.fileName &&
      shallow.pos !== undefined &&
      shallow.end !== undefined
    ) {
      context.highlightRange(shallow.pos, shallow.end);
    }
  }, [context, shallow.fileName, shallow.pos, shallow.end]);

  const handleMouseLeave = useCallback(() => {
    if (context) {
      context.clearHighlight();
    }
  }, [context]);

  // Title: KindName (e.g., "Identifier", "VariableDeclaration")
  const title = kindName || shallow.kindName?.replace('ast.', '') || 'Node';

  return (
    <ControlledCollapsibleItem
      label={label}
      title={title}
      preview={preview}
      expanded={expanded}
      loading={loading}
      onToggle={handleToggle}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {loading ? (
        <div className="text-gray-400 text-xs py-1">Loading...</div>
      ) : error ? (
        <div className="text-red-500 text-xs py-1">{error}</div>
      ) : (
        <NodeInfoView info={fullData ?? shallow} />
      )}
    </ControlledCollapsibleItem>
  );
};
