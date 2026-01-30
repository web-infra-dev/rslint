import React from 'react';
import type { GetAstInfoResponse } from './types';
import { NodeInfoView, getNodeTitle, getNodePreview } from './NodeInfoView';
import { TypeInfoView, getTypeTitle, getTypePreview } from './TypeInfoView';
import {
  SymbolInfoView,
  getSymbolTitle,
  getSymbolPreview,
} from './SymbolInfoView';
import {
  SignatureInfoView,
  getSignatureTitle,
  getSignaturePreview,
} from './SignatureInfoView';
import {
  FlowInfoView,
  FlowGraphSection,
  getFlowTitle,
  getFlowPreview,
} from './FlowInfoView';
import { AstInfoProvider } from './AstInfoContext';
import { CollapsibleItem } from './CollapsibleItem';

interface AstInfoPanelProps {
  info?: GetAstInfoResponse;
  loading?: boolean;
  onRequestAstInfo?: (
    position: number,
    end?: number,
    kind?: number,
    fileName?: string,
  ) => Promise<GetAstInfoResponse | null>;
  /** Fetch AST info for lazy loading - does not update global state */
  onFetchAstInfoForLazy?: (
    position: number,
    end?: number,
    kind?: number,
    fileName?: string,
  ) => Promise<GetAstInfoResponse | null>;
  /** Highlight a range in the editor (on hover) */
  onHighlightRange?: (pos: number, end: number) => void;
  /** Clear the highlight decoration */
  onClearHighlight?: () => void;
}

interface SectionProps {
  title: string;
  children: React.ReactNode;
}

const Section: React.FC<SectionProps> = ({ title, children }) => (
  <div className="border-b border-gray-200 last:border-b-0">
    <div className="bg-gray-50 px-3 py-2">
      <h3 className="text-xs font-semibold tracking-wide text-gray-600">
        {title}
      </h3>
    </div>
    <div className="p-3">{children}</div>
  </div>
);

export const AstInfoPanel: React.FC<AstInfoPanelProps> = ({
  info,
  loading,
  onRequestAstInfo,
  onFetchAstInfoForLazy,
  onHighlightRange,
  onClearHighlight,
}) => {
  if (loading) {
    return (
      <div className="flex h-full items-center justify-center text-center text-sm text-gray-400">
        Loading...
      </div>
    );
  }

  if (!info) {
    return (
      <div className="flex h-full items-center justify-center text-center text-sm text-gray-400">
        Click on a node in the AST tree to view detailed information
      </div>
    );
  }

  return (
    <AstInfoProvider
      onRequestAstInfo={onFetchAstInfoForLazy ?? onRequestAstInfo}
      onHighlightRange={onHighlightRange}
      onClearHighlight={onClearHighlight}
    >
      <div className="h-full overflow-auto">
        <Section title="Node">
          {info.node ? (
            <CollapsibleItem
              title={getNodeTitle(info.node)}
              preview={getNodePreview(info.node)}
              defaultExpanded
              bold
            >
              <NodeInfoView
                info={info.node}
                onNodeClick={(pos, end) => onRequestAstInfo?.(pos, end)}
              />
            </CollapsibleItem>
          ) : (
            <div className="py-2 text-center text-xs text-gray-400">
              No node information
            </div>
          )}
        </Section>

        <Section title="Type">
          {info.type ? (
            <CollapsibleItem
              title={getTypeTitle(info.type)}
              preview={getTypePreview(info.type)}
              defaultExpanded
              bold
            >
              <TypeInfoView info={info.type} />
            </CollapsibleItem>
          ) : (
            <div className="py-2 text-center text-xs text-gray-400">
              No type information
            </div>
          )}
        </Section>

        <Section title="Symbol">
          {info.symbol ? (
            <CollapsibleItem
              title={getSymbolTitle(info.symbol)}
              preview={getSymbolPreview(info.symbol)}
              defaultExpanded
              bold
            >
              <SymbolInfoView info={info.symbol} />
            </CollapsibleItem>
          ) : (
            <div className="py-2 text-center text-xs text-gray-400">
              No symbol information
            </div>
          )}
        </Section>

        <Section title="Signature">
          {info.signature ? (
            <CollapsibleItem
              title={getSignatureTitle(info.signature)}
              preview={getSignaturePreview(info.signature)}
              defaultExpanded
              bold
            >
              <SignatureInfoView info={info.signature} />
            </CollapsibleItem>
          ) : (
            <div className="py-2 text-center text-xs text-gray-400">
              No signature information
            </div>
          )}
        </Section>

        <Section title="FlowNode">
          {info.flow ? (
            <>
              <CollapsibleItem
                title={getFlowTitle(info.flow)}
                preview={getFlowPreview(info.flow)}
                defaultExpanded
                bold
              >
                <FlowInfoView info={info.flow} />
              </CollapsibleItem>
              <FlowGraphSection info={info.flow} />
            </>
          ) : (
            <div className="py-2 text-center text-xs text-gray-400">
              No flow information
            </div>
          )}
        </Section>
      </div>
    </AstInfoProvider>
  );
};
