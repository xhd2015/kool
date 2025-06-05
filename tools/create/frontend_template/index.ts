
import React from 'react';
import ReactDOM from 'react-dom/client';
import { HashRouter } from 'react-router';
export { AppRoutes } from './AppRoutes';

// el: e.g. document.getElementById('root')
export function renderComponent(el: HTMLElement, component: React.FC, props: React.Attributes) {
    const root = ReactDOM.createRoot(el);
    root.render(React.createElement(component, props));
}

export function renderRoute(el: HTMLElement, route: React.FC, props: React.Attributes) {
    const root = ReactDOM.createRoot(el);

    //  <BrowserRouter>
    // <App />
    // </BrowserRouter>
    return root.render(React.createElement(HashRouter, {
        children: React.createElement(route, props)
    }))
}

export { React as React };
