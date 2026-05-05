import { html } from 'htm/react';
import { useState, useEffect, useMemo } from 'react';
import { fetchMeta, fetchFullGraph, fetchTraversal, fetchNodeDetail, fetchInsights } from './api.js';
import Topbar from './components/Topbar.js';
import Sidebar from './components/Sidebar.js';
import Graph from './components/Graph.js';
import DetailPanel from './components/DetailPanel.js';
import InsightsPanel from './components/InsightsPanel.js';

export default function App() {
  const [meta, setMeta] = useState({ labels: [], rels: [], systems: [] });
  const [rawGraph, setRawGraph] = useState({ nodes: [], edges: [] });
  const [selectedNode, setSelectedNode] = useState(null);
  const [focalNode, setFocalNode] = useState(null);  // node used as traversal start
  const [depth, setDepth] = useState(2);
  const [direction, setDirection] = useState('outgoing');
  const [loading, setLoading] = useState(false);
  const [nodeDetail, setNodeDetail] = useState(null);
  const [nodeDetailLoading, setNodeDetailLoading] = useState(false);
  const [error, setError] = useState(null);
  const [showInsights, setShowInsights] = useState(false);
  const [insightsData, setInsightsData] = useState(null);
  const [insightsLoading, setInsightsLoading] = useState(false);
  const [filters, setFilters] = useState({ labels: new Set(), rels: new Set(), systems: new Set() });

  // Load metadata once on mount
  useEffect(() => {
    fetchMeta()
      .then((m) => {
        setMeta(m);
        setFilters({
          labels:  new Set(m.labels  || []),
          rels:    new Set(m.rels    || []),
          systems: new Set(m.systems || []),
        });
      })
      .catch((e) => setError(e.message));
  }, []);

  // Fetch graph whenever focal node, depth, or direction change
  useEffect(() => {
    if (!focalNode) return;
    setLoading(true);
    setError(null);
    fetchTraversal(focalNode.id, depth, direction)
      .then(setRawGraph)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [focalNode, depth, direction]);

  const showFullGraph = () => {
    setFocalNode(null);
    setSelectedNode(null);
    setNodeDetail(null);
    setNodeDetailLoading(false);
    setLoading(true);
    fetchFullGraph()
      .then(setRawGraph)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  };

  // Client-side filtering applied on top of rawGraph
  const graphData = useMemo(() => {
    const { labels, rels, systems } = filters;
    const visibleIds = new Set();

    rawGraph.nodes.forEach((n) => {
      const labelOk  = labels.size  === 0 || labels.has(n.label);
      const systemOk = systems.size === 0 || !n.system || systems.has(n.system);
      if (labelOk && systemOk) visibleIds.add(n.id);
    });

    return {
      nodes: rawGraph.nodes.filter((n) => visibleIds.has(n.id)),
      edges: rawGraph.edges.filter((e) =>
        visibleIds.has(e.source) &&
        visibleIds.has(e.target) &&
        (rels.size === 0 || rels.has(e.rel))
      ),
    };
  }, [rawGraph, filters]);

  const handleSearchSelect = (node) => {
    setFocalNode(node);
    setSelectedNode(null);
  };

  const handleNodeSelect = (node) => {
    setSelectedNode(node);
    if (node) {
      setNodeDetail(null);
      setNodeDetailLoading(true);
      fetchNodeDetail(node.id)
        .then(setNodeDetail)
        .catch(() => setNodeDetail(null))
        .finally(() => setNodeDetailLoading(false));
    } else {
      setNodeDetail(null);
      setNodeDetailLoading(false);
    }
  };

  const handleExploreFrom = (node) => {
    setFocalNode(node);
    setSelectedNode(null);
    setNodeDetail(null);
    setNodeDetailLoading(false);
  };

  const loadInsights = () => {
    setInsightsLoading(true);
    fetchInsights()
      .then(setInsightsData)
      .catch(() => setInsightsData(null))
      .finally(() => setInsightsLoading(false));
  };

  const handleToggleInsights = () => {
    if (showInsights) {
      setShowInsights(false);
    } else {
      setShowInsights(true);
      setSelectedNode(null);
      setNodeDetail(null);
      if (!insightsData) loadInsights();
    }
  };

  const stats = { nodes: graphData.nodes.length, edges: graphData.edges.length };

  return html`
    <div class="app">
      <${Topbar}
        onNodeSelect=${handleSearchSelect}
        onShowFullGraph=${showFullGraph}
        onToggleInsights=${handleToggleInsights}
        insightsActive=${showInsights}
        stats=${focalNode || rawGraph.nodes.length > 0 ? stats : null}
      />

      <div class="workspace">
        <${Sidebar}
          depth=${depth}
          onDepthChange=${(d) => { setDepth(d); }}
          direction=${direction}
          onDirectionChange=${(d) => { setDirection(d); }}
          meta=${meta}
          filters=${filters}
          onFiltersChange=${setFilters}
          focalNode=${focalNode}
        />

        <div class="canvas-area">
          ${loading && html`<div class="overlay-msg">Loading…</div>`}
          ${error   && html`<div class="overlay-msg error">${error}</div>`}

          ${!focalNode && rawGraph.nodes.length === 0 && !loading && html`
            <div class="empty-state">
              <div class="empty-icon">⬡</div>
              <h2>Search for a node to start exploring</h2>
              <p>Type a table name, view, procedure, or entity in the search bar above.</p>
              <button class="btn-secondary" onClick=${showFullGraph}>Or load the full graph</button>
            </div>
          `}

          <${Graph}
            nodes=${graphData.nodes}
            edges=${graphData.edges}
            onNodeSelect=${handleNodeSelect}
            focalNodeId=${focalNode?.id}
          />
        </div>

        ${selectedNode && !showInsights && html`
          <${DetailPanel}
            node=${selectedNode}
            nodeDetail=${nodeDetail}
            nodeDetailLoading=${nodeDetailLoading}
            onClose=${() => { setSelectedNode(null); setNodeDetail(null); }}
            onExploreFrom=${handleExploreFrom}
            onShowFullGraph=${showFullGraph}
          />
        `}

        ${showInsights && html`
          <${InsightsPanel}
            data=${insightsData}
            loading=${insightsLoading}
            onClose=${() => setShowInsights(false)}
            onNodeSelect=${(node) => { handleSearchSelect(node); setShowInsights(false); }}
            onRefresh=${loadInsights}
          />
        `}
      </div>
    </div>
  `;
}
