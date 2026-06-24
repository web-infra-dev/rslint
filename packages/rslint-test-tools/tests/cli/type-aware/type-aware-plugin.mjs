// A self-contained third-party-style type-aware ESLint plugin.
//
// It deliberately does NOT import @typescript-eslint/utils: in this repo that
// name resolves to a minimal workspace stub (no ESLintUtils, and a bare `enum`
// that Node's strip-only TS loader rejects when the config imports it). Instead
// the rules read `context.sourceCode.parserServices` directly — which is exactly
// what ESLintUtils.getParserServices returns (a thin wrapper over the same
// object). So this still e2e's the bridge: the rules run in a worker and query
// types through the parserServices the worker reconstructed from Go's snapshot,
// the way every real type-aware plugin does.
function getServices(context) {
  const services =
    context.sourceCode?.parserServices ?? context.parserServices;
  if (!services || typeof services.getTypeAtLocation !== 'function') {
    // Fail loud: a missing parserServices means the bridge didn't attach, which
    // must surface as a rule error, not a silent pass that masks the regression.
    throw new Error(
      'type-aware fixture: parserServices.getTypeAtLocation is unavailable — the type snapshot bridge did not attach',
    );
  }
  return services;
}

export default {
  meta: { name: 'eslint-plugin-type-aware-fixture', version: '0.0.0' },
  rules: {
    // Flags a variable whose static type is a union that includes `undefined`.
    // Minimal but real: getTypeAtLocation → isUnion → member identity against
    // checker.getUndefinedType(), all through the bridge's parserServices.
    'no-undefined-union': {
      meta: {
        type: 'problem',
        docs: {
          description:
            'Disallow a variable whose type is a union that includes undefined.',
          // Gates the rule as type-aware: Go sees RequiresTypeInfo and builds +
          // ships the type snapshot for files this rule runs on.
          requiresTypeChecking: true,
        },
        schema: [],
        messages: {
          unionWithUndefined:
            "Variable '{{name}}' has a union type that includes undefined.",
        },
      },
      create(context) {
        const services = getServices(context);
        const checker = services.program.getTypeChecker();
        const undefinedType = checker.getUndefinedType();
        return {
          VariableDeclarator(node) {
            if (node.id.type !== 'Identifier') return;
            const type = services.getTypeAtLocation(node.id);
            if (typeof type.isUnion === 'function' && type.isUnion()) {
              // Intern identity: a union member with the undefined type-id
              // resolves to the same wrapper object as checker.getUndefinedType().
              const includesUndefined = type.types.some(
                (t) => t === undefinedType,
              );
              if (includesUndefined) {
                context.report({
                  node: node.id,
                  messageId: 'unionWithUndefined',
                  data: { name: node.id.name },
                });
              }
            }
          },
        };
      },
    },

    // Reports each variable's structural type shape, reading EVERY layout block
    // the snapshot can carry (union/intersection members via type.types, array
    // type args via checker.getTypeArguments, call signatures via
    // getCallSignatures). This guards the Go-encode ↔ Node-decode wire layout
    // across type kinds, not just the union path no-undefined-union exercises:
    // a layout drift on the array/intersection/callable blocks would surface
    // here as a wrong tag or a missing report.
    'report-type-shape': {
      meta: {
        type: 'suggestion',
        docs: {
          description: 'Reports the structural shape of a variable type.',
          requiresTypeChecking: true,
        },
        schema: [],
        messages: { shape: '{{name}}: {{tags}}' },
      },
      create(context) {
        const services = getServices(context);
        const checker = services.program.getTypeChecker();
        return {
          VariableDeclarator(node) {
            if (node.id.type !== 'Identifier') return;
            const type = services.getTypeAtLocation(node.id);
            const tags = [];
            if (typeof type.isUnion === 'function' && type.isUnion()) {
              tags.push(`union:${type.types.length}`);
            }
            if (
              typeof type.isIntersection === 'function' &&
              type.isIntersection()
            ) {
              tags.push(`intersection:${type.types.length}`);
            }
            if (checker.isArrayType(type)) {
              tags.push(`array:${checker.getTypeArguments(type).length}`);
            }
            const sigs =
              typeof type.getCallSignatures === 'function'
                ? type.getCallSignatures()
                : [];
            if (sigs.length > 0) {
              tags.push(`callable:${sigs.length}`);
            }
            if (tags.length > 0) {
              context.report({
                node: node.id,
                messageId: 'shape',
                data: { name: node.id.name, tags: tags.join('|') },
              });
            }
          },
        };
      },
    },

    // Fixable, purely-syntactic rule used by the --fix multi-pass e2e. It drops
    // an explicit `: any` annotation off a VariableDeclarator. The fix CHANGES
    // the variable's static type (any → the inferred type), so a later pass's
    // type-aware rule must re-snapshot from a fresh program to observe it.
    'drop-any-annotation': {
      meta: {
        type: 'suggestion',
        fixable: 'code',
        docs: { description: 'Drop an explicit any annotation.' },
        schema: [],
        messages: { drop: "Drop the explicit 'any' annotation." },
      },
      create(context) {
        return {
          VariableDeclarator(node) {
            const ann =
              node.id.type === 'Identifier' && node.id.typeAnnotation;
            if (ann && ann.typeAnnotation?.type === 'TSAnyKeyword') {
              context.report({
                node: ann,
                messageId: 'drop',
                fix: (fixer) => fixer.remove(ann),
              });
            }
          },
        };
      },
    },

    // Gate-integrity rule: a type-aware rule that does NOT declare
    // meta.docs.requiresTypeChecking, yet queries types. It must still get a
    // snapshot and work — proving gating is project-based (a program is present),
    // NOT keyed off the optional, non-standard requiresTypeChecking declaration.
    'no-undefined-union-undeclared': {
      meta: {
        type: 'problem',
        docs: {
          description:
            'Like no-undefined-union but WITHOUT a requiresTypeChecking declaration.',
        },
        schema: [],
        messages: {
          unionWithUndefined:
            "Variable '{{name}}' has a union type that includes undefined.",
        },
      },
      create(context) {
        const services = getServices(context);
        const checker = services.program.getTypeChecker();
        const undefinedType = checker.getUndefinedType();
        return {
          VariableDeclarator(node) {
            if (node.id.type !== 'Identifier') return;
            const type = services.getTypeAtLocation(node.id);
            if (
              typeof type.isUnion === 'function' &&
              type.isUnion() &&
              type.types.some((t) => t === undefinedType)
            ) {
              context.report({
                node: node.id,
                messageId: 'unionWithUndefined',
                data: { name: node.id.name },
              });
            }
          },
        };
      },
    },
  },
};
