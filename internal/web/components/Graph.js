import { html } from 'htm/react';
import { useEffect, useRef } from 'react';
import { nodeColor, edgeColor } from '../colors.js';

const CYTOSCAPE_STYLE = [
  {
    selector: 'node',
    style: {
      'label': 'data(name)',
      'background-color': (ele) => nodeColor(ele.data('label')),
      'color': '#fff',
      'text-valign': 'center',
      'text-halign': 'center',
      'font-size': '11px',
      'font-weight': '600',
      'text-wrap': 'wrap',
      'text-max-width': '100px',
      'width': 'label',
      'height': 'label',
      'padding': '10px',
      'shape': 'round-rectangle',
    },
  },
  {
    selector: 'node:selected',
    style: {
      'border-width': 3,
      'border-color': '#cba6f7',
    },
  },
  {
    selector: 'node.faded',
    style: { opacity: 0.15 },
  },
  {
    selector: 'node.highlighted',
    style: {
      'border-width': 3,
      'border-color': '#F1C40F',
    },
  },
  {
    selector: 'edge',
    style: {
      'label': 'data(rel)',
      'font-size': '9px',
      'color': '#64748b',
      'text-background-color': '#0f172a',
      'text-background-opacity': 0.7,
      'text-background-padding': '2px',
      'line-color': (ele) => edgeColor(ele.data('rel')),
      'target-arrow-color': (ele) => edgeColor(ele.data('rel')),
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier',
      'width': 1.5,
      'arrow-scale': 0.8,
    },
  },
  {
    selector: 'edge.faded',
    style: { opacity: 0.05 },
  },
];

export default function Graph({ nodes, edges, onNodeSelect, focalNodeId }) {
  const containerRef = useRef(null);
  const cyRef = useRef(null);

  useEffect(() => {
    if (!containerRef.current || !window.cytoscape) return;

    if (cyRef.current) {
      cyRef.current.destroy();
    }

    const elements = [
      ...nodes.map((n) => ({
        group: 'nodes',
        data: { id: n.id, label: n.label, name: n.name || n.id, system: n.system },
      })),
      ...edges.map((e, i) => ({
        group: 'edges',
        data: { id: 'e' + i, source: e.source, target: e.target, rel: e.rel },
      })),
    ];

    const cy = window.cytoscape({
      container: containerRef.current,
      elements,
      style: CYTOSCAPE_STYLE,
      layout: {
        name: 'dagre',
        rankDir: 'LR',
        nodeSep: 40,
        rankSep: 80,
        padding: 40,
        animate: false,
      },
      wheelSensitivity: 0.3,
    });

    cy.on('tap', 'node', (e) => {
      const node = e.target;
      cy.nodes().removeClass('faded highlighted');
      cy.edges().removeClass('faded');
      node.addClass('highlighted');
      node.connectedEdges().connectedNodes().not(node).each((n) => n.addClass('faded'));
      cy.edges().not(node.connectedEdges()).addClass('faded');
      onNodeSelect({
        id: node.data('id'),
        label: node.data('label'),
        name: node.data('name'),
        system: node.data('system'),
        indegree: node.indegree(),
        outdegree: node.outdegree(),
      });
    });

    cy.on('tap', (e) => {
      if (e.target === cy) {
        cy.nodes().removeClass('faded highlighted');
        cy.edges().removeClass('faded');
        onNodeSelect(null);
      }
    });

    if (focalNodeId) {
      const focal = cy.getElementById(focalNodeId);
      if (focal.length) focal.addClass('highlighted');
    }

    cyRef.current = cy;
  }, [nodes, edges]);

  useEffect(() => {
    if (!cyRef.current || !focalNodeId) return;
    cyRef.current.nodes().removeClass('highlighted');
    const focal = cyRef.current.getElementById(focalNodeId);
    if (focal.length) focal.addClass('highlighted');
  }, [focalNodeId]);

  const handleFit = () => cyRef.current?.fit(undefined, 40);

  const handleLayout = (name) => {
    if (!cyRef.current) return;
    const config = name === 'cose'
      ? { name: 'cose', animate: true, animationDuration: 500, nodeRepulsion: 4096, idealEdgeLength: 100, padding: 40 }
      : { name: 'dagre', rankDir: 'LR', nodeSep: 40, rankSep: 80, padding: 40, animate: false };
    cyRef.current.layout(config).run();
  };

  return html`
    <div style=${{ position: 'relative', flex: 1, display: 'flex', flexDirection: 'column' }}>
      <div style=${{ position: 'absolute', top: 10, right: 10, zIndex: 10, display: 'flex', gap: 6 }}>
        <button class="ctrl-btn" onClick=${() => handleLayout('dagre')}>Hierarchy</button>
        <button class="ctrl-btn" onClick=${() => handleLayout('cose')}>Force</button>
        <button class="ctrl-btn" onClick=${handleFit}>Fit</button>
      </div>
      <div ref=${containerRef} style=${{ flex: 1, background: '#0f172a' }} />
    </div>
  `;
}
