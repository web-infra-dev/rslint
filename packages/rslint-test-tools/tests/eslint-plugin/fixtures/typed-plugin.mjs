export default {
  rules: {
    value: {
      meta: { schema: [] },
      create(context) {
        const services = context.sourceCode.parserServices;
        if (!services.program) {
          return {
            Program(node) {
              const forward = services.esTreeNodeToTSNodeMap;
              const reverse = services.tsNodeToESTreeNodeMap;
              const source = forward.get(node);
              if (
                // A standalone parser-services tree uses TypeScript's Latest target.
                source.languageVersion !== 99 ||
                source.hasNoDefaultLib !== false ||
                source.impliedNodeFormat !== undefined ||
                services.program !== null ||
                services.getTypeAtLocation !== undefined ||
                services.emitDecoratorMetadata !== false ||
                !forward.has(node) ||
                !reverse.has(forward.get(node)) ||
                reverse.get(forward.get(node)) !== node
              ) {
                throw new Error('incorrect source-only parser services');
              }
              context.report({ node, message: 'source-only' });
            },
          };
        }
        const checker = services.program.getTypeChecker();
        return {
          CallExpression(node) {
            if (
              node.callee.type !== 'Identifier' ||
              node.callee.name !== 'choose'
            )
              return;
            const native = services.esTreeNodeToTSNodeMap.get(node);
            if (checker.typeToString(checker.getTypeAtLocation(native)) !== '1')
              throw new Error('unexpected initial overload');
            for (const count of [2, 1, 2]) {
              const candidates = [];
              const signature = checker.getResolvedSignature(
                native,
                candidates,
                count,
              );
              if (
                candidates.length !== 2 ||
                checker.typeToString(signature.getReturnType()) !==
                  String(count) ||
                checker.typeToString(checker.getTypeAtLocation(native)) !==
                  String(count)
              )
                throw new Error('stale signature response through host cache');
            }
            context.report({ node, message: 'ordered overload queries' });
          },
          ChainExpression(node) {
            const prefix = node.expression.callee?.object;
            if (prefix?.type !== 'MemberExpression') return;
            const native = services.esTreeNodeToTSNodeMap.get(prefix);
            const reverse = services.tsNodeToESTreeNodeMap;
            const alias = reverse.get(native);
            if (
              !reverse.has(native) ||
              alias?.type !== 'ChainExpression' ||
              alias.expression !== prefix ||
              reverse.get(native) !== alias ||
              services.esTreeNodeToTSNodeMap.get(alias) !== native
            )
              throw new Error('detached optional chain identity');
            context.report({
              node,
              message: checker.typeToString(services.getTypeAtLocation(node)),
            });
          },
          SequenceExpression(node) {
            const [first, second] = node.expressions;
            node.expressions.reverse();
            try {
              if (!services.esTreeNodeToTSNodeMap.has(node))
                throw new Error('sequence origin missing');
              const native = services.esTreeNodeToTSNodeMap.get(node);
              const prefix = services.tsNodeToESTreeNodeMap.get(native.left);
              if (
                !services.tsNodeToESTreeNodeMap.has(native.left) ||
                prefix.expressions[0] !== first ||
                prefix.expressions[1] !== second
              )
                throw new Error('lazy alias used mutated sequence membership');
            } finally {
              node.expressions.reverse();
            }
            context.report({ node, message: 'original sequence prefix' });
          },
          Identifier(node) {
            if (
              node.name !== 'value' ||
              node.parent.type !== 'ExpressionStatement'
            )
              return;
            const native = services.esTreeNodeToTSNodeMap.get(node);
            if (services.tsNodeToESTreeNodeMap.get(native) !== node)
              throw new Error('node map identity');
            const type = checker.getTypeAtLocation(native);
            if (checker.getTypeAtLocation(native) !== type)
              throw new Error('type identity');
            context.report({ node, message: checker.typeToString(type) });
          },
        };
      },
    },
  },
};
