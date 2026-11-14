import { ConfigCommentParser } from '@eslint/plugin-kit';
import type { Comment } from '../../languages/js/token-builder.js';
import type { Diagnostic } from './diagnostic-builder.js';

/** Original comment location, not a next-line directive's target location. */
export interface DisableDirectiveUsage {
  /** Zero-based ECMAScript line and UTF-16 column, against BOM-stripped text. */
  line: number;
  column: number;
  /** The declared rule entry; null identifies the wildcard entry. */
  ruleId: string | null;
}

/** Per-file summary shared across the worker and Go transport boundaries. */
export interface DisableDirectiveUsageSummary {
  /** Declared rule IDs; null identifies the wildcard entry. */
  rules: (string | null)[];
  /** Original comment's zero-based line, UTF-16 column, and rule-table index. */
  entries: [line: number, column: number, ruleIndex: number][];
}

export function encodeDisableDirectiveUsage(
  usages: readonly DisableDirectiveUsage[],
): DisableDirectiveUsageSummary {
  const rules: (string | null)[] = [];
  const indices = new Map<string | null, number>();
  const entries = usages.map(
    ({ line, column, ruleId }): [number, number, number] => {
      let index = indices.get(ruleId);
      if (index === undefined) {
        index = rules.length;
        indices.set(ruleId, index);
        rules.push(ruleId);
      }
      return [line, column, index];
    },
  );
  return { rules, entries };
}

interface Directive {
  type: 'disable' | 'enable';
  pos: number;
  ruleId: string | null;
  usage?: DisableDirectiveUsage;
}

interface Options {
  comments: readonly Comment[];
  diagnostics: readonly Diagnostic[];
  lineStartOffsets: readonly number[];
  textLength: number;
  collectUsage: boolean;
}

const parser = new ConfigCommentParser();

/** Filter before byte conversion and serialization; return each used entry once. */
export function applyDisableDirectives(options: Options): {
  diagnostics: readonly Diagnostic[];
  usedDisableDirectives?: DisableDirectiveUsage[];
} {
  const used = options.collectUsage
    ? new Set<DisableDirectiveUsage>()
    : undefined;
  if (options.diagnostics.length === 0) {
    return {
      diagnostics: options.diagnostics,
      ...(used ? { usedDisableDirectives: [] } : {}),
    };
  }
  const blocks: Directive[] = [];
  const lines: Directive[] = [];
  for (const comment of options.comments) {
    if (comment.type !== 'Line' && comment.type !== 'Block') continue;
    const parsed = parser.parseDirective(comment.value);
    if (!parsed) continue;
    let block = false;
    let nextLine = false;
    let type: 'disable' | 'enable' = 'disable';
    switch (parsed.label) {
      case 'eslint-disable':
      case 'rslint-disable':
        block = true;
        break;
      case 'eslint-enable':
      case 'rslint-enable':
        block = true;
        type = 'enable';
        break;
      case 'eslint-disable-next-line':
      case 'rslint-disable-next-line':
        nextLine = true;
        break;
      case 'eslint-disable-line':
      case 'rslint-disable-line':
        if (comment.loc.start.line !== comment.loc.end.line) continue;
        break;
      default:
        continue;
    }
    const names = parseRuleNames(parsed.value);
    for (const ruleId of names.length ? names : [null]) {
      const usage =
        used && type === 'disable'
          ? {
              line: comment.loc.start.line - 1,
              column: comment.loc.start.column,
              ruleId,
            }
          : undefined;
      if (block) {
        const directive: Directive = { type, pos: comment.range[0], ruleId };
        if (usage) directive.usage = usage;
        blocks.push(directive);
      } else {
        const target = nextLine
          ? comment.loc.end.line + 1
          : comment.loc.start.line;
        // A nonexistent following line must not clear a disable at EOF.
        const directive: Directive = {
          type: 'disable',
          pos: options.lineStartOffsets[target - 1] ?? Infinity,
          ruleId,
        };
        if (usage) directive.usage = usage;
        lines.push(directive);
        lines.push({
          type: 'enable',
          pos: options.lineStartOffsets[target] ?? Infinity,
          ruleId,
        });
      }
    }
  }
  blocks.sort((a, b) => a.pos - b.pos);
  // Expire the previous line before applying suppression for the new line.
  // Stable sorting still preserves source order among disables at that line.
  lines.sort(
    (a, b) =>
      a.pos - b.pos ||
      Number(a.type === 'disable') - Number(b.type === 'disable'),
  );
  if (
    !used &&
    ((blocks.length === 0 && lines.length === 0) ||
      options.diagnostics.every(
        (diagnostic) =>
          Number.isInteger(diagnostic.startPos) &&
          diagnostic.startPos >= 0 &&
          diagnostic.startPos <= options.textLength,
      ))
  ) {
    return {
      diagnostics: filterSuppressed(
        filterSuppressed(options.diagnostics, blocks),
        lines,
      ),
    };
  }
  // Block suppression gets usage attribution before line suppression. Stable
  // sorting preserves source ordering when line directives target the same line.
  const afterBlocks = filterWithUsage(
    options.diagnostics,
    blocks,
    options.textLength,
    used,
  );
  const diagnostics = filterWithUsage(
    afterBlocks,
    lines,
    options.textLength,
    used,
  );
  return {
    diagnostics,
    ...(used ? { usedDisableDirectives: [...used] } : {}),
  };
}

function parseRuleNames(value: string): string[] {
  if (!value.includes(',')) {
    const name = trimRuleName(value);
    return name ? Object.keys({ [name]: true }) : [];
  }
  const names: Record<string, true> = Object.create(null);
  for (const part of value.split(',')) {
    const name = trimRuleName(part);
    if (name) names[name] = true;
  }
  return Object.keys(names);
}

function trimRuleName(value: string): string {
  const name = value.trim();
  if (
    name.length >= 2 &&
    (name[0] === '"' || name[0] === "'") &&
    name[name.length - 1] === name[0]
  ) {
    return name.slice(1, -1);
  }
  return name;
}

function diagnosticPosition(
  diagnostic: Diagnostic,
  textLength: number,
): number {
  return Number.isFinite(diagnostic.startPos)
    ? Math.min(textLength, Math.max(0, Math.trunc(diagnostic.startPos)))
    : 0;
}

// Ordinary reports with reporting off keep the original scalar-state loop. It never reads an owner
// or checks collection state while walking directives for each diagnostic.
function filterSuppressed(
  diagnostics: readonly Diagnostic[],
  directives: readonly Directive[],
): readonly Diagnostic[] {
  if (directives.length === 0) return diagnostics;
  const visible: Diagnostic[] = [];
  for (const diagnostic of diagnostics) {
    let active = 0;
    for (let i = 0; i < directives.length; i++) {
      const directive = directives[i];
      if (directive.pos > diagnostic.startPos) break;
      if (
        directive.ruleId === null ||
        directive.ruleId === diagnostic.ruleName
      ) {
        if (directive.type === 'disable') active++;
        else active = 0;
      }
    }
    if (active === 0) visible.push(diagnostic);
  }
  return visible;
}

function filterWithUsage(
  diagnostics: readonly Diagnostic[],
  directives: readonly Directive[],
  textLength: number,
  used: Set<DisableDirectiveUsage> | undefined,
): readonly Diagnostic[] {
  if (directives.length === 0) return diagnostics;
  const visible: Diagnostic[] = [];
  for (const diagnostic of diagnostics) {
    // Match the byte-boundary normalization applied to visible diagnostics.
    const pos = diagnosticPosition(diagnostic, textLength);
    let active: Directive | undefined;
    for (const directive of directives) {
      if (directive.pos > pos) break;
      if (
        directive.ruleId === null ||
        directive.ruleId === diagnostic.ruleName
      ) {
        active = directive.type === 'disable' ? directive : undefined;
      }
    }
    if (active) {
      if (used && active.usage) used.add(active.usage);
    } else {
      visible.push(diagnostic);
    }
  }
  return visible;
}
