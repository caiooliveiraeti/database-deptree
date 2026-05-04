export const LABEL_COLORS = {
  APPLICATION: '#F1C40F',
  ENTITY:      '#4A90D9',
  TABLE:       '#27AE60',
  REPOSITORY:  '#9B59B6',
  QUERY:       '#E67E22',
  PROCEDURE:   '#E74C3C',
  VIEW:        '#16A085',
  PACKAGE:     '#D35400',
  SYNONYM:     '#7F8C8D',
  TYPE:        '#8E44AD',
};

export const REL_COLORS = {
  STORED_IN:    '#27AE60',
  MANAGES:      '#9B59B6',
  QUERIES:      '#E67E22',
  USES_TABLE:   '#4A90D9',
  DEPENDS_ON:   '#E74C3C',
  READS:        '#00BCD4',
  MAPS_TO:      '#78909C',
  CALLS:        '#E91E63',
  CONTAINS:     '#F1C40F',
  ONE_TO_MANY:  '#1ABC9C',
  MANY_TO_ONE:  '#2ECC71',
  MANY_TO_MANY: '#3498DB',
  ONE_TO_ONE:   '#9B59B6',
};

export const DEFAULT_NODE_COLOR = '#95A5A6';
export const DEFAULT_EDGE_COLOR = '#94a3b8';

export function nodeColor(label) {
  return LABEL_COLORS[label] || DEFAULT_NODE_COLOR;
}

export function edgeColor(rel) {
  return REL_COLORS[rel] || DEFAULT_EDGE_COLOR;
}

export const SYSTEM_COLORS = ['#cba6f7','#89b4fa','#a6e3a1','#fab387','#f38ba8','#f9e2af'];
export function systemColor(i) {
  return SYSTEM_COLORS[i % SYSTEM_COLORS.length];
}
