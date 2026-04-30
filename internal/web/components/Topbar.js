import { html } from 'htm/react';
import { useState, useEffect, useRef, useCallback } from 'react';
import { searchNodes } from '../api.js';
import { nodeColor } from '../colors.js';

function debounce(fn, delay) {
  let t;
  return (...args) => { clearTimeout(t); t = setTimeout(() => fn(...args), delay); };
}

export default function Topbar({ onNodeSelect, onShowFullGraph, stats }) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const inputRef = useRef(null);

  const search = useCallback(debounce(async (term) => {
    if (term.length < 2) { setResults([]); setOpen(false); return; }
    setLoading(true);
    try {
      const data = await searchNodes(term);
      setResults(data);
      setOpen(data.length > 0);
    } finally {
      setLoading(false);
    }
  }, 300), []);

  const handleInput = (e) => {
    setQuery(e.target.value);
    search(e.target.value);
  };

  const handleSelect = (node) => {
    setQuery(node.name || node.id);
    setOpen(false);
    onNodeSelect(node);
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Escape') { setOpen(false); inputRef.current?.blur(); }
  };

  useEffect(() => {
    const close = (e) => { if (!e.target.closest('.topbar-search')) setOpen(false); };
    document.addEventListener('mousedown', close);
    return () => document.removeEventListener('mousedown', close);
  }, []);

  return html`
    <header class="topbar">
      <span class="topbar-logo">deptree</span>

      <div class="topbar-search">
        <div class="search-input-wrap">
          <span class="search-icon">⌕</span>
          <input
            ref=${inputRef}
            class="search-input"
            placeholder="Search nodes… (e.g. owners, calculate_tax)"
            value=${query}
            onInput=${handleInput}
            onKeyDown=${handleKeyDown}
            onFocus=${() => results.length && setOpen(true)}
          />
          ${loading && html`<span class="search-spinner">…</span>`}
        </div>

        ${open && html`
          <ul class="search-dropdown">
            ${results.map((n) => html`
              <li key=${n.id} class="search-item" onClick=${() => handleSelect(n)}>
                <span class="search-dot" style=${{ background: nodeColor(n.label) }} />
                <span class="search-name">${n.name || n.id}</span>
                <span class="search-label">${n.label}</span>
                ${n.system && html`<span class="search-system">${n.system}</span>`}
              </li>
            `)}
          </ul>
        `}
      </div>

      <button class="ctrl-btn" onClick=${onShowFullGraph}>Full graph</button>
      ${stats && html`<span class="topbar-stats">${stats.nodes}n · ${stats.edges}e</span>`}
    </header>
  `;
}
