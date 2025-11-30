import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import BashHistoryManager from './component/BashHistoryManager.tsx'

const components: Record<string, any> = {
  App,
  BashHistoryManager,
};

(window as any).__KOOL_COMPONENTS = {
  ...components,
  renderComponent: (container: HTMLElement, componentName: string, props: any) => {
    const Component = components[componentName];
    if (!Component) {
      console.error(`Component ${componentName} not found`);
      return;
    }
    createRoot(container).render(
      <StrictMode>
        <Component {...props} />
      </StrictMode>
    );
  },
};
