import { marked } from 'marked';
import mermaid from 'mermaid';

// Pure function for rendering markdown to HTML with SVGs directly rendered (no side effects, no closure dependencies)
export async function renderMarkdownToHtml(content: string, handleMermaidContextMenu: string, id: number): Promise<string> {
    if (!content) {
        return '';
    }

    // Initialize mermaid with optimal config
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

    // First pass: collect all mermaid code blocks
    const mermaidBlocks: { id: string; code: string; placeholder: string }[] = [];

    // Configure marked options for better rendering
    marked.setOptions({
        gfm: true, // GitHub Flavored Markdown
        breaks: true, // Convert \n to <br>
    });

    // Configure custom renderer to make links open in new tab and handle mermaid
    const renderer = new marked.Renderer();
    renderer.link = function (token: { href: string, title?: string | null, tokens: any[] }) {
        const titleAttr = token.title ? ` title="${token.title}"` : '';
        const text = this.parser.parseInline(token.tokens);
        return `<a href="${token.href}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
    };

    renderer.code = function (token: { text: string, lang?: string, escaped?: boolean }) {
        if (token.lang === 'mermaid') {
            const elementID = `mermaid-${id}`;
            const placeholder = `__MERMAID_PLACEHOLDER_${elementID}__`;
            mermaidBlocks.push({ id: elementID, code: token.text, placeholder });
            return placeholder;
        }
        // Default code block rendering
        const lang = token.lang || '';
        const langClass = lang ? ` class="language-${lang}"` : '';
        return `<pre><code${langClass}>${token.text}</code></pre>`;
    };

    // First pass: get HTML with placeholders
    let html = await marked(content, { renderer });

    // Second pass: render all mermaid diagrams and replace placeholders
    for (const block of mermaidBlocks) {
        try {
            // Render SVG directly using mermaid
            const { svg } = await mermaid.render(block.id, block.code);

            // Add context menu functionality to the SVG
            const svgWithEvents = svg.replace(
                '<svg',
                '<svg style="cursor: context-menu; max-width: 100%; height: auto; user-select: none;" oncontextmenu="window.' + handleMermaidContextMenu + '(event, this)"'
            );

            const finalSvg = `<div class="mermaid-container">${svgWithEvents}</div>`;
            html = html.replace(block.placeholder, finalSvg);

        } catch (err) {
            // Fall back to showing the code
            const fallback = `<pre style="text-align: left; padding: 16px; background: #f8f8f8; border: 1px solid #ddd; border-radius: 4px;"><code class="language-mermaid">${block.code}</code></pre>`;
            html = html.replace(block.placeholder, fallback);
        }
    }

    return html;
};
