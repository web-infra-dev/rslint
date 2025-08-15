import React, { useEffect, useState } from 'react';
import { TooltipProvider } from '../../../theme/components/TooltipContext';
import HeatMapItem from '../../../theme/components/HeatMapItem';
import { HeapMap } from '../../../theme/components/OriginalHeatMap';

type FailingCase = {
  name: string;
  url: string;
};

type Rule = {
  name: string;
  group: string;
  status: string;
  failing_case: FailingCase[];
};

const RULE_MANIFEST_URL =
  'https://raw.githubusercontent.com/web-infra-dev/rslint/main/packages/rslint-test-tools/rule-manifest.json';

// Status badge component with color coding
const StatusBadge = ({ status }: { status: string }) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'full':
        return { bg: '#10b981', color: 'white' }; // green
      case 'partial-impl':
        return { bg: '#f59e0b', color: 'white' }; // amber
      case 'partial-test':
        return { bg: '#ef4444', color: 'white' }; // red
      default:
        return { bg: '#6b7280', color: 'white' }; // gray
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'full':
        return '✅';
      case 'partial-impl':
        return '🔄';
      case 'partial-test':
        return '🧪';
      default:
        return '❓';
    }
  };

  const colors = getStatusColor(status);

  return (
    <span
      style={{
        backgroundColor: colors.bg,
        color: colors.color,
        padding: '4px 8px',
        borderRadius: '12px',
        fontSize: '14px',
        fontWeight: 'bold',
        display: 'inline-block',
        minWidth: '40px',
        textAlign: 'center',
      }}
      title={
        status === 'full'
          ? 'Full Implementation'
          : status === 'partial-impl'
            ? 'Partial Implementation'
            : status === 'partial-test'
              ? 'Partial Test'
              : 'Unknown Status'
      }
    >
      {getStatusText(status)}
    </span>
  );
};

// Group badge component
const GroupBadge = ({ group }: { group: string }) => {
  const getGroupColor = (group: string) => {
    switch (group) {
      case 'typescript-eslint':
        return { bg: '#3b82f6', color: 'white' }; // blue
      default:
        return { bg: '#6b7280', color: 'white' }; // gray
    }
  };

  const colors = getGroupColor(group);

  return (
    <span
      style={{
        backgroundColor: colors.bg,
        color: colors.color,
        padding: '2px 6px',
        borderRadius: '8px',
        fontSize: '11px',
        fontWeight: '500',
      }}
    >
      {group}
    </span>
  );
};

// Mock data that matches original arewerslintyet format
const getMockLintResults = () => {
  return {
    passing: `internal/rules/array_type/array_type.go
  ✓ should detect array type violations
  ✓ should suggest proper array syntax
  ✓ should handle nested arrays
  ✓ should work with generic arrays

internal/rules/no_unused_vars/no_unused_vars.go
  ✓ should detect unused variables
  ✓ should ignore used variables
  ✓ should handle destructuring
  ✓ should work with function parameters
  ✓ should detect unused imports

internal/rules/prefer_const/prefer_const.go
  ✓ should suggest const over let
  ✓ should handle destructuring
  ✓ should work with block scope
  ✓ should ignore reassigned variables

internal/rules/no_var/no_var.go
  ✓ should detect var declarations
  ✓ should suggest let or const
  ✓ should work in function scope

internal/rules/eqeqeq/eqeqeq.go
  ✓ should enforce strict equality
  ✓ should detect loose equality
  ✓ should handle null comparisons
  ✓ should work with type coercion

internal/rules/curly/curly.go
  ✓ should enforce braces
  ✓ should detect missing braces
  ✓ should work with if statements
  ✓ should handle else clauses

internal/rules/dot_notation/dot_notation.go
  ✓ should prefer dot notation
  ✓ should detect bracket notation
  ✓ should handle computed properties

internal/rules/no_empty/no_empty.go
  ✓ should detect empty blocks
  ✓ should allow comments in blocks
  ✓ should work with try-catch

internal/rules/no_eval/no_eval.go
  ✓ should detect eval usage
  ✓ should detect indirect eval
  ✓ should work with Function constructor

internal/rules/no_implied_eval/no_implied_eval.go
  ✓ should detect setTimeout with string
  ✓ should detect setInterval with string
  ✓ should allow function callbacks`,
    failing: `internal/rules/no_console/no_console.go
  ✗ should detect console statements
  ✗ should handle console methods
  ✗ should work with console.log
  ✗ should detect console.error

internal/rules/semicolon/semicolon.go
  ✗ should enforce semicolons
  ✗ should detect missing semicolons
  ✗ should work with ASI

internal/rules/no_debugger/no_debugger.go
  ✗ should detect debugger statements
  ✗ should work in strict mode

internal/rules/no_alert/no_alert.go
  ✗ should detect alert usage
  ✗ should detect confirm usage
  ✗ should detect prompt usage

internal/rules/no_with/no_with.go
  ✗ should detect with statements
  ✗ should work in strict mode

internal/rules/radix/radix.go
  ✗ should enforce radix parameter
  ✗ should detect parseInt without radix

internal/rules/no_new_object/no_new_object.go
  ✗ should detect new Object()
  ✗ should suggest object literal

internal/rules/no_new_array/no_new_array.go
  ✗ should detect new Array()
  ✗ should suggest array literal

internal/rules/no_new_func/no_new_func.go
  ✗ should detect new Function()
  ✗ should suggest function declaration`
  };
};

// Heat Map component using original logic
const RulesHeatMap = () => {
  const mockData = getMockLintResults();

  return (
    <section className="mb-8">
      <h2 className="text-2xl font-semibold mb-4">RSLint Tests</h2>
      <p className="text-gray-600 dark:text-gray-300 mb-4">
        Visual overview of RSLint test results. Each square represents a test case:
      </p>
      <div className="flex items-center gap-4 mb-4 text-sm">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 bg-passing-square border border-gray-300"></div>
          <span>Passing Tests</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 bg-failing-square border border-gray-300"></div>
          <span>Failing Tests</span>
        </div>
      </div>
      <section className="HeatMap" style={{ display: 'flex', flexWrap: 'wrap', gap: '1px', maxWidth: '800px' }}>
        <HeapMap lintResults={mockData} />
      </section>
    </section>
  );
};

export default function RulesWithHeatMap() {
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [groupFilter, setGroupFilter] = useState<string>('all');

  useEffect(() => {
    fetch(RULE_MANIFEST_URL)
      .then(res => {
        if (!res.ok) throw new Error('Failed to fetch rule-manifest.json');
        return res.json();
      })
      .then(data => {
        setRules(Array.isArray(data) ? data : data.rules || []);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  }, []);

  if (loading)
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        Loading rule manifest...
      </div>
    );
  if (error)
    return <div style={{ padding: '24px', color: 'red' }}>Error: {error}</div>;
  if (!rules.length)
    return <div style={{ padding: '24px' }}>No rules found.</div>;

  // Filter rules based on selected filters
  const filteredRules = rules.filter(rule => {
    const statusMatch = statusFilter === 'all' || rule.status === statusFilter;
    const groupMatch = groupFilter === 'all' || rule.group === groupFilter;
    return statusMatch && groupMatch;
  });

  // Get unique statuses and groups for filters
  const statuses = ['all', ...Array.from(new Set(rules.map(r => r.status)))];
  const groups = ['all', ...Array.from(new Set(rules.map(r => r.group)))];

  // Count rules by status
  const statusCounts = rules.reduce(
    (acc, rule) => {
      acc[rule.status] = (acc[rule.status] || 0) + 1;
      return acc;
    },
    {} as Record<string, number>,
  );

  return (
    <TooltipProvider>
      <div style={{ padding: '24px' }}>
        {/* Summary Statistics */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
            gap: '16px',
            marginBottom: '24px',
          }}
        >
          <div
            style={{
              backgroundColor: '#f3f4f6',
              padding: '16px',
              borderRadius: '8px',
              textAlign: 'center',
            }}
          >
            <div
              style={{ fontSize: '24px', fontWeight: 'bold', color: '#10b981' }}
            >
              {statusCounts['full'] || 0}
            </div>
            <div style={{ color: '#6b7280' }}>Fully Implemented</div>
          </div>
          <div
            style={{
              backgroundColor: '#f3f4f6',
              padding: '16px',
              borderRadius: '8px',
              textAlign: 'center',
            }}
          >
            <div
              style={{ fontSize: '24px', fontWeight: 'bold', color: '#f59e0b' }}
            >
              {statusCounts['partial-impl'] || 0}
            </div>
            <div style={{ color: '#6b7280' }}>Partial Implementation</div>
          </div>
          <div
            style={{
              backgroundColor: '#f3f4f6',
              padding: '16px',
              borderRadius: '8px',
              textAlign: 'center',
            }}
          >
            <div
              style={{ fontSize: '24px', fontWeight: 'bold', color: '#ef4444' }}
            >
              {statusCounts['partial-test'] || 0}
            </div>
            <div style={{ color: '#6b7280' }}>Partial Test</div>
          </div>
          <div
            style={{
              backgroundColor: '#f3f4f6',
              padding: '16px',
              borderRadius: '8px',
              textAlign: 'center',
            }}
          >
            <div
              style={{ fontSize: '24px', fontWeight: 'bold', color: '#374151' }}
            >
              {rules.length}
            </div>
            <div style={{ color: '#6b7280' }}>Total Rules</div>
          </div>
        </div>

        {/* Heat Map Visualization */}
        <RulesHeatMap />

        {/* Filters */}
        <div
          style={{
            display: 'flex',
            gap: '16px',
            marginBottom: '24px',
            flexWrap: 'wrap',
          }}
        >
          <div>
            <label style={{ marginRight: '8px', fontWeight: '500' }}>
              Status:
            </label>
            <select
              value={statusFilter}
              onChange={e => setStatusFilter(e.target.value)}
              style={{
                padding: '4px 8px',
                border: '1px solid #d1d5db',
                borderRadius: '4px',
              }}
            >
              {statuses.map(status => (
                <option key={status} value={status}>
                  {status === 'all' ? 'All Statuses' : status}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label style={{ marginRight: '8px', fontWeight: '500' }}>
              Group:
            </label>
            <select
              value={groupFilter}
              onChange={e => setGroupFilter(e.target.value)}
              style={{
                padding: '4px 8px',
                border: '1px solid #d1d5db',
                borderRadius: '4px',
              }}
            >
              {groups.map(group => (
                <option key={group} value={group}>
                  {group === 'all' ? 'All Groups' : group}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Rules Table */}
        <div style={{ overflowX: 'auto' }}>
          <table
            style={{
              width: '100%',
              borderCollapse: 'collapse',
              backgroundColor: 'white',
              borderRadius: '8px',
              overflow: 'hidden',
              boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)',
            }}
          >
            <thead>
              <tr style={{ backgroundColor: '#f9fafb' }}>
                <th
                  style={{
                    padding: '12px',
                    textAlign: 'left',
                    fontWeight: '600',
                    borderBottom: '1px solid #e5e7eb',
                  }}
                >
                  Rule Name
                </th>
                <th
                  style={{
                    padding: '12px',
                    textAlign: 'left',
                    fontWeight: '600',
                    borderBottom: '1px solid #e5e7eb',
                  }}
                >
                  Group
                </th>
                <th
                  style={{
                    padding: '12px',
                    textAlign: 'left',
                    fontWeight: '600',
                    borderBottom: '1px solid #e5e7eb',
                  }}
                >
                  Status
                </th>
                <th
                  style={{
                    padding: '12px',
                    textAlign: 'left',
                    fontWeight: '600',
                    borderBottom: '1px solid #e5e7eb',
                  }}
                >
                  Failing Cases
                </th>
              </tr>
            </thead>
            <tbody>
              {filteredRules.map((rule, index) => (
                <tr
                  key={rule.name}
                  style={{
                    borderBottom: '1px solid #f3f4f6',
                    backgroundColor: index % 2 === 0 ? 'white' : '#f9fafb',
                  }}
                >
                  <td style={{ padding: '12px', fontFamily: 'monospace' }}>
                    {rule.name}
                  </td>
                  <td style={{ padding: '12px' }}>
                    <GroupBadge group={rule.group} />
                  </td>
                  <td style={{ padding: '12px' }}>
                    <StatusBadge status={rule.status} />
                  </td>
                  <td style={{ padding: '12px' }}>
                    {rule.failing_case && rule.failing_case.length > 0 ? (
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px' }}>
                        {rule.failing_case.slice(0, 3).map((failCase, idx) => (
                          <a
                            key={idx}
                            href={failCase.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            style={{
                              fontSize: '12px',
                              padding: '2px 6px',
                              backgroundColor: '#fef3c7',
                              color: '#92400e',
                              borderRadius: '4px',
                              textDecoration: 'none',
                            }}
                          >
                            {failCase.name}
                          </a>
                        ))}
                        {rule.failing_case.length > 3 && (
                          <span style={{ fontSize: '12px', color: '#6b7280' }}>
                            +{rule.failing_case.length - 3} more
                          </span>
                        )}
                      </div>
                    ) : (
                      <span style={{ color: '#6b7280', fontSize: '14px' }}>None</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div style={{ marginTop: '24px', fontSize: '14px', color: '#6b7280' }}>
          Showing {filteredRules.length} of {rules.length} rules
        </div>
      </div>
    </TooltipProvider>
  );
}