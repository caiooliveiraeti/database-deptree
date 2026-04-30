import React from 'react';
import htm from 'htm';

// Bind htm to the single shared React instance so all components
// use the same React that react-dom/client uses. Importing htm/react
// directly bundles a second React copy, causing "invalid element" errors.
export const html = htm.bind(React.createElement);
