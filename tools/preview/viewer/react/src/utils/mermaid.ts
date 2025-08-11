import mermaid from 'mermaid';

// initialize mermaid
mermaid.initialize({
    startOnLoad: false,
    theme: 'default',
    securityLevel: 'loose',
    flowchart: {
        useMaxWidth: true,
        htmlLabels: true
    },
    themeCSS: '',
    maxTextSize: 50000,
    darkMode: false
});