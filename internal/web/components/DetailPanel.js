import { html } from 'htm/react';
import { nodeColor } from '../colors.js';

const PROP_LABELS = { owner: 'Schema', shortName: 'Short Name', file: 'File' };
const propLabel = (k) => PROP_LABELS[k] || (k.charAt(0).toUpperCase() + k.slice(1));

export default function DetailPanel({ node, nodeDetail, nodeDetailLoading, impactedSystems, onClose, onExploreFrom, onShowFullGraph }) {
  if (!node) return null;

  const extraProps = nodeDetail
    ? Object.entries(nodeDetail.properties).filter(([, v]) => v !== null && v !== '')
    : [];

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

      ${nodeDetailLoading && html`<div class="detail-props-loading">Loading…</div>`}

      ${!nodeDetailLoading && extraProps.length > 0 && html`
        <div class="detail-divider" />
        <table class="detail-props">
          <tbody>
            ${extraProps.map(([k, v]) => html`
              <tr key=${k}>
                <td class="prop-key">${propLabel(k)}</td>
                <td class="prop-val">${String(v)}</td>
              </tr>
            `)}
          </tbody>
        </table>
      `}

      ${impactedSystems !== null && html`
        <div class="detail-divider" />
        <div class="detail-impact">
          <div class="detail-impact-label">
            Impacted systems
            ${impactedSystems.length > 0 && html`<span class="detail-impact-count">${impactedSystems.length}</span>`}
          </div>
          ${impactedSystems.length === 0
            ? html`<div class="detail-impact-empty">No application reaches this node</div>`
            : impactedSystems.map((s) => html`
                <div key=${s.id} class="detail-impact-item">
                  <span class="detail-impact-dot" style=${{ background: nodeColor('APPLICATION') }} />
                  <span class="detail-impact-name">${s.name || s.id}</span>
                </div>
              `)
          }
        </div>
      `}
    </div>
  `;
}
