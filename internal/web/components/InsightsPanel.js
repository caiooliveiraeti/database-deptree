import { html } from 'htm/react';
import { nodeColor } from '../colors.js';

const SECTIONS = [
  {
    key: 'understanding',
    title: 'System understanding',
    items: [
      { field: 'orphan_entities',    label: 'Entities with no repository' },
      { field: 'native_queries',     label: 'Native SQL queries (bypass ORM)' },
    ],
  },
  {
    key: 'impact',
    title: 'Change impact',
    items: [
      { field: 'dual_access',   label: 'Dual access — JPA entity + direct procedure' },
      { field: 'mapped_views',  label: 'Views mapped by Java (MAPS_TO)' },
    ],
  },
  {
    key: 'refactor',
    title: 'Refactoring',
    items: [
      { field: 'single_entity_tables', label: 'Tables used by only one entity' },
      { field: 'single_caller_procs',  label: 'Procedures with only one caller' },
      { field: 'unused_procs',         label: 'Procedures with no callers' },
    ],
  },
  {
    key: 'modernization',
    title: 'Modernization / migration',
    items: [
      { field: 'hotspots',         label: 'Hotspots — most incoming dependencies' },
      { field: 'shared_tables',    label: 'Tables shared across systems' },
      { field: 'heavy_packages',   label: 'Packages with most procedures' },
      { field: 'deep_call_chains', label: 'Longest procedure call chains' },
      { field: 'uncovered_procs',  label: 'Oracle procedures with no Java coverage' },
    ],
  },
];

function InsightItem({ node, onSelect }) {
  const color = nodeColor(node.label);
  const shortName = node.name.includes('.') ? node.name.split('.').pop() : node.name;
  const truncated = shortName.length > 30 ? shortName.slice(0, 28) + '…' : shortName;

  return html`
    <div class="insight-item" onClick=${() => onSelect(node)} title=${node.name}>
      <span class="insight-dot" style=${{ background: color }} />
      <span class="insight-name">${truncated}</span>
      ${node.count > 0 && html`<span class="insight-count">${node.count}</span>`}
      ${node.tags?.length > 0 && html`
        <span class="insight-tags" title=${node.tags.join(', ')}>
          ${node.tags.slice(0, 2).join(', ')}${node.tags.length > 2 ? ` +${node.tags.length - 2}` : ''}
        </span>
      `}
      <span class="insight-arrow">→</span>
    </div>
  `;
}

function InsightGroup({ title, nodes, onSelect }) {
  if (!nodes?.length) return html`
    <div class="insight-group-empty">${title} <span class="insight-empty-badge">0</span></div>
  `;

  return html`
    <div class="insight-group">
      <div class="insight-group-header">
        ${title}
        <span class="insight-count-badge">${nodes.length}</span>
      </div>
      ${nodes.map((n) => html`<${InsightItem} key=${n.id} node=${n} onSelect=${onSelect} />`)}
    </div>
  `;
}

export default function InsightsPanel({ data, loading, onClose, onNodeSelect, onRefresh }) {
  return html`
    <div class="insights-panel">
      <div class="insights-header">
        <span class="insights-title">Insights</span>
        <div style=${{ display: 'flex', gap: 6 }}>
          <button class="detail-close" title="Atualizar" onClick=${onRefresh}>↺</button>
          <button class="detail-close" onClick=${onClose}>×</button>
        </div>
      </div>

      ${loading && html`<div class="insights-loading">Loading analyses…</div>`}

      ${!loading && data && html`
        <div class="insights-body">
          ${SECTIONS.map((section) => html`
            <div key=${section.key} class="insights-section">
              <div class="insights-section-title">${section.title}</div>
              ${section.items.map((item) => html`
                <${InsightGroup}
                  key=${item.field}
                  title=${item.label}
                  nodes=${data[item.field]}
                  onSelect=${onNodeSelect}
                />
              `)}
            </div>
          `)}
        </div>
      `}
    </div>
  `;
}
