import { html } from 'htm/react';
import { nodeColor, edgeColor, systemColor } from '../colors.js';

const DIRECTIONS = [
  { value: 'outgoing', label: '→',  title: 'What this node depends on' },
  { value: 'incoming', label: '←',  title: 'Who uses this node' },
  { value: 'both',     label: '↔', title: 'Both directions' },
];

export default function Sidebar({
  depth, onDepthChange,
  direction, onDirectionChange,
  meta, filters, onFiltersChange,
  focalNode,
}) {
  const toggle = (key, value) => {
    const next = new Set(filters[key]);
    next.has(value) ? next.delete(value) : next.add(value);
    onFiltersChange({ ...filters, [key]: next });
  };

  const allOn = (key) => onFiltersChange({ ...filters, [key]: new Set(meta[key] || []) });
  const allOff = (key) => onFiltersChange({ ...filters, [key]: new Set() });

  return html`
    <aside class="sidebar">
      <div class="sidebar-section">
        <div class="sidebar-label">Direction</div>
        <div class="dir-group">
          ${DIRECTIONS.map((d) => html`
            <button
              key=${d.value}
              class=${'dir-btn' + (direction === d.value ? ' active' : '')}
              title=${d.title}
              onClick=${() => onDirectionChange(d.value)}
            >${d.label}</button>
          `)}
        </div>
        ${!focalNode && html`<p class="sidebar-hint">Search for a node above to start exploring</p>`}
      </div>

      <div class="sidebar-section">
        <div class="sidebar-label">Depth</div>
        <div class="depth-group">
          ${[1,2,3,4,5].map((d) => html`
            <button
              key=${d}
              class=${'depth-btn' + (depth === d ? ' active' : '')}
              onClick=${() => onDepthChange(d)}
            >${d}</button>
          `)}
        </div>
      </div>

      ${(meta.labels || []).length > 0 && html`
        <div class="sidebar-divider" />
        <div class="sidebar-section">
          <div class="sidebar-label-row">
            <span class="sidebar-label">Node Types</span>
            <span class="sidebar-toggle" onClick=${() => allOn('labels')}>all</span>
            <span class="sidebar-toggle" onClick=${() => allOff('labels')}>none</span>
          </div>
          ${(meta.labels || []).map((l) => html`
            <label key=${l} class="filter-item">
              <input type="checkbox"
                checked=${filters.labels.has(l)}
                onChange=${() => toggle('labels', l)}
              />
              <span class="filter-dot" style=${{ background: nodeColor(l) }} />
              <span class="filter-name">${l}</span>
            </label>
          `)}
        </div>
      `}

      ${(meta.rels || []).length > 0 && html`
        <div class="sidebar-divider" />
        <div class="sidebar-section">
          <div class="sidebar-label-row">
            <span class="sidebar-label">Relationships</span>
            <span class="sidebar-toggle" onClick=${() => allOn('rels')}>all</span>
            <span class="sidebar-toggle" onClick=${() => allOff('rels')}>none</span>
          </div>
          ${(meta.rels || []).map((r) => html`
            <label key=${r} class="filter-item">
              <input type="checkbox"
                checked=${filters.rels.has(r)}
                onChange=${() => toggle('rels', r)}
              />
              <span class="filter-dot" style=${{ background: edgeColor(r) }} />
              <span class="filter-name">${r}</span>
            </label>
          `)}
        </div>
      `}

      ${(meta.systems || []).length > 0 && html`
        <div class="sidebar-divider" />
        <div class="sidebar-section">
          <div class="sidebar-label-row">
            <span class="sidebar-label">Systems</span>
            <span class="sidebar-toggle" onClick=${() => allOn('systems')}>all</span>
            <span class="sidebar-toggle" onClick=${() => allOff('systems')}>none</span>
          </div>
          ${(meta.systems || []).map((s, i) => html`
            <label key=${s} class="filter-item">
              <input type="checkbox"
                checked=${filters.systems.has(s)}
                onChange=${() => toggle('systems', s)}
              />
              <span class="filter-dot" style=${{ background: systemColor(i) }} />
              <span class="filter-name">${s}</span>
            </label>
          `)}
        </div>
      `}
    </aside>
  `;
}
