import { html } from 'htm/react';
import { nodeColor } from '../colors.js';

export default function DetailPanel({ node, onClose, onExploreFrom, onShowFullGraph }) {
  if (!node) return null;

  return html`
    <div class="detail-panel">
      <div class="detail-header">
        <div class="detail-badge" style=${{ background: nodeColor(node.label) + '22', color: nodeColor(node.label), border: '1px solid ' + nodeColor(node.label) }}>
          ${node.label}
        </div>
        <button class="detail-close" onClick=${onClose}>×</button>
      </div>

      <div class="detail-name" title=${node.id}>${node.name || node.id}</div>
      ${node.id !== node.name && html`<div class="detail-id">${node.id}</div>`}
      ${node.system && html`<div class="detail-system">⬡ ${node.system}</div>`}

      <div class="detail-stats">
        <div class="stat-box">
          <span class="stat-num">${node.indegree ?? '—'}</span>
          <span class="stat-lbl">used by</span>
        </div>
        <div class="stat-box">
          <span class="stat-num">${node.outdegree ?? '—'}</span>
          <span class="stat-lbl">depends on</span>
        </div>
      </div>

      <div class="detail-actions">
        <button class="btn-primary" onClick=${() => onExploreFrom(node)}>
          Explore from here ↗
        </button>
        <button class="btn-secondary" onClick=${onShowFullGraph}>
          Full graph
        </button>
      </div>
    </div>
  `;
}
