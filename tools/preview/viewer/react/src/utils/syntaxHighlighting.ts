// Lazy-loaded syntax highlighting utility

import Prism from 'prismjs';

import 'prismjs/components/prism-javascript';
import 'prismjs/components/prism-typescript';
import 'prismjs/components/prism-jsx';
import 'prismjs/components/prism-tsx';
import 'prismjs/components/prism-go';
import 'prismjs/components/prism-python';
import 'prismjs/components/prism-bash';
import 'prismjs/components/prism-sql';
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-yaml';
import 'prismjs/components/prism-markdown';
import 'prismjs/components/prism-css';
import 'prismjs/components/prism-scss';


import 'prismjs/themes/prism.css'; // Include a theme

// Map common aliases
const languageMap: Record<string, string> = {
    'js': 'javascript',
    'ts': 'typescript',
    'py': 'python',
    'sh': 'bash',
    'shell': 'bash',
    'yml': 'yaml',
    'golang': 'go'
} as const;

export function highlightCode(code: string, lang: string): string {
    lang = languageMap[lang] || lang;

    // Apply highlighting if language is supported
    let error = ''
    let highlightedCode = code
    try {
        if (Prism.languages[lang]) {
            highlightedCode = Prism.highlight(code, Prism.languages[lang], lang);
        }
    } catch (e) {
        error = `<p style="color: red;">(rendering code: ${e})</p>`
    }
    return error + `<pre><code class="language-${lang}">${highlightedCode}</code></pre>`
}