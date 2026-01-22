import React from 'react';
import { ObjectTree, EnumValueMap } from './ObjectTree';
import type {
  TypeDetails,
  SymbolDetails,
  SignatureDetails,
  FlowNodeDetails,
  NodeLocation,
} from '@rslint/wasm';
import { SyntaxKind } from '@rslint/api';

// Node detail info passed from parent
export interface NodeDetailInfo {
  kind?: number;
  kindName?: string;
  flags?: number;
  flagNames?: string[];
  modifierFlags?: number;
  modifierFlagNames?: string[];
  pos: number;
  end: number;
  text?: string;
  // Type info (when available) - full TypeDetails from Go
  type?: TypeDetails;
  contextualType?: TypeDetails;
  // Symbol info (when available) - full SymbolDetails from Go
  symbol?: SymbolDetails;
  // Signature info (when available) - full SignatureDetails from Go
  signature?: SignatureDetails;
  // FlowNode info (when available) - full FlowNodeDetails from Go
  flowNode?: FlowNodeDetails;
  // Related objects for reference lookup (Type ID -> TypeDetails, Symbol ID -> SymbolDetails)
  relatedTypes?: Record<number, TypeDetails>;
  relatedSymbols?: Record<number, SymbolDetails>;
}

interface NodeDetailPanelProps {
  detail?: NodeDetailInfo;
}

export const NodeDetailPanel: React.FC<NodeDetailPanelProps> = ({ detail }) => {
  if (!detail) {
    return (
      <div className="h-full flex flex-col border border-gray-200 rounded-md bg-white overflow-hidden">
        <div className="flex-1 flex items-center justify-center p-4 text-gray-500 text-sm">
          Select a node to view details
        </div>
      </div>
    );
  }

  // Helper to resolve Type ID to TypeString
  const resolveTypeId = (typeId: number): string => {
    const type = detail.relatedTypes?.[typeId];
    return type ? type.TypeString : `Type#${typeId}`;
  };

  // Helper to resolve Symbol ID to SymbolString
  const resolveSymbolId = (symbolId: number): string => {
    const symbol = detail.relatedSymbols?.[symbolId];
    return symbol ? symbol.SymbolString : `Symbol#${symbolId}`;
  };

  // Helper to convert Types array (of IDs) to resolved TypeStrings
  const resolveTypesArray = (typeIds?: number[]): Array<{ id: number; typeString: string }> | undefined => {
    if (!typeIds) return undefined;
    return typeIds.map(id => ({ id, typeString: resolveTypeId(id) }));
  };

  // Helper to convert Properties array (of Symbol IDs) to resolved info
  const resolvePropertiesArray = (symbolIds?: number[]): Array<{ id: number; symbolString: string }> | undefined => {
    if (!symbolIds) return undefined;
    return symbolIds.map(id => ({ id, symbolString: resolveSymbolId(id) }));
  };

  // Helper to format NodeLocation to a readable string
  const formatNodeLocation = (loc: NodeLocation): string => {
    const kindName = SyntaxKind[loc.kind] || `Kind(${loc.kind})`;
    return `${kindName} @ ${loc.filePath}:${loc.pos}`;
  };

  // Helper to format Declarations array (of NodeLocation) to readable strings
  const formatDeclarations = (locs?: NodeLocation[]): string[] | undefined => {
    if (!locs || locs.length === 0) return undefined;
    return locs.map(formatNodeLocation);
  };

  // Build node object for tree display
  const nodeData: Record<string, unknown> = {
    kind: detail.kind,
    pos: detail.pos,
    end: detail.end,
  };
  if (detail.text) {
    nodeData.text = detail.text;
  }
  // Always show flags
  if (detail.flags !== undefined) {
    nodeData.flags = detail.flags;
  }
  if (detail.modifierFlags !== undefined) {
    nodeData.modifierFlags = detail.modifierFlags;
  }

  // Build enum values map for tooltips
  const nodeEnumValues: EnumValueMap = {
    kind: detail.kindName || `Unknown(${detail.kind})`,
    flags: detail.flagNames?.length ? detail.flagNames.join('\n') : 'None',
    modifierFlags: detail.modifierFlagNames?.length ? detail.modifierFlagNames.join('\n') : 'None',
  };

  return (
    <div className="h-full flex flex-col border border-gray-200 rounded-md bg-white overflow-hidden">
      <div className="flex-1 overflow-auto p-3">
        {/* Node */}
        <div className="pb-3">
          <div className="text-xs font-semibold text-gray-800 mb-2">Node</div>
          <ObjectTree
            data={nodeData}
            name={detail.kindName}
            defaultExpanded
            enumValues={nodeEnumValues}
          />
        </div>

        {/* Divider */}
        <div className="border-t border-gray-200 my-3" />

        {/* Type */}
        <div className="pb-3">
          <div className="text-xs font-semibold text-gray-800 mb-2">Type</div>
          {detail.type ? (
            <>
              {/* TypeToString() result */}
              <div className="mb-2">
                <span className="text-xs text-gray-600">TypeToString(): </span>
                <span className="text-xs font-mono text-blue-700">{detail.type.TypeString}</span>
              </div>
              {/* Full Type object */}
              <ObjectTree
                data={{
                  Id: detail.type.Id,
                  Flags: detail.type.Flags,
                  ObjectFlags: detail.type.ObjectFlags,
                  ...(detail.type.Symbol !== undefined && { Symbol: `${resolveSymbolId(detail.type.Symbol)} (id: ${detail.type.Symbol})` }),
                  ...(detail.type.IntrinsicName && { IntrinsicName: detail.type.IntrinsicName }),
                  ...(detail.type.Value !== undefined && { Value: detail.type.Value }),
                  ...(detail.type.Types && { Types: resolveTypesArray(detail.type.Types) }),
                  ...(detail.type.Properties && { Properties: resolvePropertiesArray(detail.type.Properties) }),
                  ...(detail.type.CallSignatures && { CallSignatures: detail.type.CallSignatures }),
                  ...(detail.type.Target !== undefined && { Target: `${resolveTypeId(detail.type.Target)} (id: ${detail.type.Target})` }),
                }}
                name="Type"
                defaultExpanded
                enumValues={{
                  Flags: detail.type.FlagNames?.join('\n') || 'None',
                  ObjectFlags: detail.type.ObjectFlagNames?.join('\n') || 'None',
                }}
              />
            </>
          ) : (
            <div className="text-xs text-gray-400 italic font-mono">Not available</div>
          )}
        </div>

        {/* ContextualType */}
        {detail.contextualType && (
          <>
            <div className="border-t border-gray-200 my-3" />
            <div className="pb-3">
              <div className="text-xs font-semibold text-gray-800 mb-2">ContextualType</div>
              {/* TypeToString() result */}
              <div className="mb-2">
                <span className="text-xs text-gray-600">TypeToString(): </span>
                <span className="text-xs font-mono text-blue-700">{detail.contextualType.TypeString}</span>
              </div>
              {/* Full ContextualType object */}
              <ObjectTree
                data={{
                  Id: detail.contextualType.Id,
                  Flags: detail.contextualType.Flags,
                  ObjectFlags: detail.contextualType.ObjectFlags,
                  ...(detail.contextualType.Symbol !== undefined && { Symbol: `${resolveSymbolId(detail.contextualType.Symbol)} (id: ${detail.contextualType.Symbol})` }),
                  ...(detail.contextualType.IntrinsicName && { IntrinsicName: detail.contextualType.IntrinsicName }),
                  ...(detail.contextualType.Types && { Types: resolveTypesArray(detail.contextualType.Types) }),
                  ...(detail.contextualType.Target !== undefined && { Target: `${resolveTypeId(detail.contextualType.Target)} (id: ${detail.contextualType.Target})` }),
                }}
                name="ContextualType"
                defaultExpanded
                enumValues={{
                  Flags: detail.contextualType.FlagNames?.join('\n') || 'None',
                  ObjectFlags: detail.contextualType.ObjectFlagNames?.join('\n') || 'None',
                }}
              />
            </div>
          </>
        )}

        {/* Divider */}
        <div className="border-t border-gray-200 my-3" />

        {/* Symbol */}
        <div className="pb-3">
          <div className="text-xs font-semibold text-gray-800 mb-2">Symbol</div>
          {detail.symbol ? (
            <>
              {/* symbolToString() result */}
              <div className="mb-2">
                <span className="text-xs text-gray-600">SymbolToString(): </span>
                <span className="text-xs font-mono text-blue-700">{detail.symbol.SymbolString}</span>
              </div>
              {/* Full Symbol object */}
              <ObjectTree
                data={{
                  Id: detail.symbol.Id,
                  Name: detail.symbol.Name,
                  Flags: detail.symbol.Flags,
                  CheckFlags: detail.symbol.CheckFlags,
                  ...(detail.symbol.Declarations && { Declarations: formatDeclarations(detail.symbol.Declarations) }),
                  ...(detail.symbol.ValueDeclaration && { ValueDeclaration: formatNodeLocation(detail.symbol.ValueDeclaration) }),
                  ...(detail.symbol.Members && Object.keys(detail.symbol.Members).length > 0 && {
                    Members: Object.fromEntries(
                      Object.entries(detail.symbol.Members).map(([name, id]) => [name, `${resolveSymbolId(id)} (id: ${id})`])
                    )
                  }),
                  ...(detail.symbol.Exports && Object.keys(detail.symbol.Exports).length > 0 && {
                    Exports: Object.fromEntries(
                      Object.entries(detail.symbol.Exports).map(([name, id]) => [name, `${resolveSymbolId(id)} (id: ${id})`])
                    )
                  }),
                  ...(detail.symbol.Parent !== undefined && { Parent: `${resolveSymbolId(detail.symbol.Parent)} (id: ${detail.symbol.Parent})` }),
                }}
                name="Symbol"
                defaultExpanded
                enumValues={{
                  Flags: detail.symbol.FlagNames?.join('\n') || 'None',
                  CheckFlags: detail.symbol.CheckFlagNames?.join('\n') || 'None',
                }}
              />
            </>
          ) : (
            <div className="text-xs text-gray-400 italic font-mono">Not available</div>
          )}
        </div>

        {/* Divider */}
        <div className="border-t border-gray-200 my-3" />

        {/* Signature */}
        <div className="pb-3">
          <div className="text-xs font-semibold text-gray-800 mb-2">Signature</div>
          {detail.signature ? (
            <>
              {/* signatureToString() result */}
              <div className="mb-2">
                <span className="text-xs text-gray-600">signatureToString(): </span>
                <span className="text-xs font-mono text-blue-700">{detail.signature.SignatureString}</span>
              </div>
              {/* Full Signature object */}
              <ObjectTree
                data={{
                  HasRestParameter: detail.signature.HasRestParameter,
                  ...(detail.signature.TypeParameters && { TypeParameters: resolveTypesArray(detail.signature.TypeParameters) }),
                  ...(detail.signature.Parameters && { Parameters: detail.signature.Parameters }),
                  ...(detail.signature.ThisParameter && { ThisParameter: detail.signature.ThisParameter }),
                  ...(detail.signature.ReturnType !== undefined && { ReturnType: `${resolveTypeId(detail.signature.ReturnType)} (id: ${detail.signature.ReturnType})` }),
                  ...(detail.signature.Declaration && { Declaration: formatNodeLocation(detail.signature.Declaration) }),
                }}
                name="Signature"
                defaultExpanded
              />
            </>
          ) : (
            <div className="text-xs text-gray-400 italic font-mono">Not available</div>
          )}
        </div>

        {/* Divider */}
        <div className="border-t border-gray-200 my-3" />

        {/* FlowNode */}
        <div>
          <div className="text-xs font-semibold text-gray-800 mb-2">FlowNode</div>
          {detail.flowNode ? (
            <ObjectTree
              data={{
                Flags: detail.flowNode.Flags,
                ...(detail.flowNode.Node && { Node: formatNodeLocation(detail.flowNode.Node) }),
                ...(detail.flowNode.Antecedent && { Antecedent: detail.flowNode.Antecedent }),
                ...(detail.flowNode.Antecedents && { Antecedents: detail.flowNode.Antecedents }),
              }}
              name="FlowNode"
              defaultExpanded
              enumValues={{
                Flags: detail.flowNode.FlagNames?.join('\n') || 'None',
              }}
            />
          ) : (
            <div className="text-xs text-gray-400 italic font-mono">Not available</div>
          )}
        </div>
      </div>
    </div>
  );
};
