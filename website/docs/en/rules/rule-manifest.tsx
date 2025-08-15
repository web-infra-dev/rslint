import React, { useEffect, useState } from 'react';
import { usePageData } from '@rspress/core/runtime';

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

export default function RuleManifestTable() {
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [groupFilter, setGroupFilter] = useState<string>('all');
  const pageData = usePageData();
  useEffect(() => {
    if (pageData.page.ruleManifest) {
      let data: any = pageData.page.ruleManifest;
      setRules(data.rules ?? []);
      setLoading(false);
      return;
    }
    fetch(RULE_MANIFEST_URL)
      .then(res => {
        if (!res.ok) throw new Error('Failed to fetch rule-manifest.json');
        return res.json();
      })
      .then(data => {
        setRules(data.rules ?? []);
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
    <>
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
              style={{ fontSize: '24px', fontWeight: 'bold', color: '#6b7280' }}
            >
              {rules.length}
            </div>
            <div style={{ color: '#6b7280' }}>Total Rules</div>
          </div>
        </div>

        {/* Filters */}
        <div
          style={{
            display: 'flex',
            gap: '16px',
            marginBottom: '24px',
            flexWrap: 'wrap',
            alignItems: 'center',
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
                borderRadius: '4px',
                border: '1px solid #d1d5db',
              }}
            >
              {statuses.map(status => (
                <option key={status} value={status}>
                  {status === 'all'
                    ? 'All Statuses'
                    : status === 'full'
                      ? '✅ Full'
                      : status === 'partial-impl'
                        ? '🔄 Partial Implementation'
                        : status === 'partial-test'
                          ? '🧪 Partial Test'
                          : status}
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
                borderRadius: '4px',
                border: '1px solid #d1d5db',
              }}
            >
              {groups.map(group => (
                <option key={group} value={group}>
                  {group === 'all' ? 'All Groups' : group}
                </option>
              ))}
            </select>
          </div>
          <div style={{ color: '#6b7280' }}>
            Showing {filteredRules.length} of {rules.length} rules
          </div>
        </div>

        {/* Rules Table */}
        <div style={{ overflowX: 'auto' }}>
          <table
            style={{
              borderCollapse: 'collapse',
              width: '100%',
              backgroundColor: 'white',
              borderRadius: '8px',
              overflow: 'hidden',
              boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1)',
            }}
          >
            <thead>
              <tr style={{ backgroundColor: '#f9fafb' }}>
                <th
                  style={{
                    padding: '12px 16px',
                    textAlign: 'left',
                    borderBottom: '1px solid #e5e7eb',
                    fontWeight: '600',
                    fontSize: '14px',
                    width: '35%',
                  }}
                >
                  Rule Name
                </th>
                <th
                  style={{
                    padding: '12px 16px',
                    textAlign: 'left',
                    borderBottom: '1px solid #e5e7eb',
                    fontWeight: '600',
                    fontSize: '14px',
                    width: '25%',
                  }}
                >
                  Group
                </th>
                <th
                  style={{
                    padding: '12px 16px',
                    textAlign: 'left',
                    borderBottom: '1px solid #e5e7eb',
                    fontWeight: '600',
                    fontSize: '14px',
                    width: '15%',
                  }}
                >
                  Status
                </th>
                <th
                  style={{
                    padding: '12px 16px',
                    textAlign: 'left',
                    borderBottom: '1px solid #e5e7eb',
                    fontWeight: '600',
                    fontSize: '14px',
                    width: '25%',
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
                    backgroundColor: index % 2 === 0 ? 'white' : '#f9fafb',
                    borderBottom: '1px solid #f3f4f6',
                  }}
                >
                  <td
                    style={{
                      padding: '12px 16px',
                      fontWeight: '500',
                      fontFamily: 'monospace',
                      fontSize: '14px',
                      wordBreak: 'break-word',
                    }}
                  >
                    <a
                      href={`https://typescript-eslint.io/rules/${rule.name}/`}
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{
                        color: '#2563eb',
                        textDecoration: 'none',
                        borderBottom: '2px solid #2563eb',
                        paddingBottom: '2px',
                        fontWeight: '600',
                        transition: 'all 0.2s ease-in-out',
                      }}
                      onMouseEnter={e => {
                        e.currentTarget.style.color = '#1d4ed8';
                        e.currentTarget.style.borderBottomColor = '#1d4ed8';
                        e.currentTarget.style.backgroundColor = '#eff6ff';
                        e.currentTarget.style.padding = '2px 4px';
                        e.currentTarget.style.borderRadius = '4px';
                      }}
                      onMouseLeave={e => {
                        e.currentTarget.style.color = '#2563eb';
                        e.currentTarget.style.borderBottomColor = '#2563eb';
                        e.currentTarget.style.backgroundColor = 'transparent';
                        e.currentTarget.style.padding = '2px 0';
                        e.currentTarget.style.borderRadius = '0';
                      }}
                      title={`View ${rule.name} rule documentation in typescript-eslint`}
                    >
                      {rule.name}
                    </a>
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <GroupBadge group={rule.group} />
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <StatusBadge status={rule.status} />
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    {rule.failing_case.length > 0 ? (
                      <div>
                        <div style={{ fontWeight: '500', marginBottom: '4px' }}>
                          {rule.failing_case.length} failing case
                          {rule.failing_case.length !== 1 ? 's' : ''}
                        </div>
                        {rule.failing_case.map((failingCase, idx) => (
                          <div
                            key={idx}
                            style={{
                              fontSize: '12px',
                              color: '#6b7280',
                              marginBottom: '2px',
                            }}
                          >
                            •{' '}
                            <a
                              href={`https://github.com/web-infra-dev/rslint/blob/main/${failingCase.url}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              style={{
                                color: '#3b82f6',
                                textDecoration: 'none',
                                borderBottom: '1px dotted #3b82f6',
                              }}
                              onMouseEnter={e => {
                                e.currentTarget.style.textDecoration =
                                  'underline';
                              }}
                              onMouseLeave={e => {
                                e.currentTarget.style.textDecoration = 'none';
                              }}
                            >
                              {failingCase.name}
                            </a>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <span style={{ color: '#10b981', fontWeight: '500' }}>
                        ✅ All tests passing
                      </span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Legend */}
        <div
          style={{
            marginTop: '24px',
            padding: '16px',
            backgroundColor: '#f9fafb',
            borderRadius: '8px',
            fontSize: '14px',
          }}
        >
          <h3 style={{ marginTop: 0, marginBottom: '12px' }}>Status Legend</h3>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
              gap: '8px',
            }}
          >
            <div>
              <StatusBadge status="full" /> - Rule is fully implemented and all
              tests pass
            </div>
            <div>
              <StatusBadge status="partial-impl" /> - Rule is partially
              implemented with some failing cases
            </div>
            <div>
              <StatusBadge status="partial-test" /> - Rule has partial test
              coverage
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
