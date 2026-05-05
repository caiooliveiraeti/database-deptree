async function json(url) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`${url}: ${res.status}`);
  return res.json();
}

export async function fetchMeta() {
  return json('/api/meta');
}

export async function fetchFullGraph() {
  return json('/api/graph');
}

export async function fetchTraversal(nodeId, depth, direction) {
  const params = new URLSearchParams({ from: nodeId, depth, direction });
  return json('/api/traverse?' + params);
}

export async function searchNodes(term) {
  if (!term || term.length < 2) return [];
  const params = new URLSearchParams({ q: term });
  return json('/api/nodes/search?' + params);
}

export async function fetchNodeDetail(nodeId) {
  return json('/api/nodes/' + encodeURIComponent(nodeId));
}

export async function fetchInsights() {
  return json('/api/insights');
}
