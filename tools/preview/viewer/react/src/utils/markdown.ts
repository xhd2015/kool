import { marked, type Token } from 'marked';
import mermaid from 'mermaid';
import { encode } from 'plantuml-encoder';
import { highlightCode } from './syntaxHighlighting';

// Pure function for rendering markdown to HTML with SVGs directly rendered (no side effects, no closure dependencies)
export async function renderMarkdownToHtml(content: string, handleMermaidContextMenu: string): Promise<string> {
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

    // First pass: collect all mermaid and plantuml code blocks
    const mermaidBlocks: { id: string; code: string; placeholder: string }[] = [];
    const plantumlBlocks: { id: string; code: string; placeholder: string }[] = [];

    // Configure marked options for better rendering
    marked.setOptions({
        gfm: true, // GitHub Flavored Markdown
        breaks: true, // Convert \n to <br>
    });

    // Configure custom renderer to make links open in new tab and handle mermaid
    const renderer = new marked.Renderer();
    renderer.link = function (token: { href: string, title?: string | null, tokens: Token[] }) {
        const titleAttr = token.title ? ` title="${token.title}"` : '';
        const text = this.parser.parseInline(token.tokens);
        return `<a href="${token.href}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
    };

    renderer.code = function (token: { text: string, lang?: string, escaped?: boolean }) {
        switch (token.lang) {
            case 'mermaid': {
                const elementID = `mermaid-${crypto.randomUUID()}`;
                const placeholder = `__MERMAID_PLACEHOLDER_${elementID}__`;
                mermaidBlocks.push({ id: elementID, code: token.text, placeholder });
                return placeholder;
            }
            case 'plantuml':
            case 'puml':
            case 'uml': {
                const elementID = `plantuml-${crypto.randomUUID()}`;
                const placeholder = `__PLANTUML_PLACEHOLDER_${elementID}__`;
                plantumlBlocks.push({ id: elementID, code: token.text, placeholder });
                return placeholder;
            }
            default:
                return highlightCode(token.text, token.lang || '');
        }
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
                '<svg style="max-width: 100%; height: auto; user-select: none;" oncontextmenu="window.' + handleMermaidContextMenu + '(event, this)"'
            );

            const finalSvg = `<div class="mermaid-container">${svgWithEvents}</div>`;
            html = html.replace(block.placeholder, finalSvg);

        } catch {
            // Fall back to showing the code
            const fallback = `<pre style="text-align: left; padding: 16px; background: #f8f8f8; border: 1px solid #ddd; border-radius: 4px;"><code class="language-mermaid">${block.code}</code></pre>`;
            html = html.replace(block.placeholder, fallback);
        }
    }

    // Third pass: render all PlantUML diagrams and replace placeholders
    for (const block of plantumlBlocks) {
        try {
            // Encode PlantUML code and create image element
            const encoded = encode(block.code);
            const plantUmlUrl = `/planuml/svg/${encoded}`;

            // Create img element with context menu support
            const imgElement = `<img src="${plantUmlUrl}" alt="PlantUML diagram" style="max-width: 100%; height: auto; user-select: none; display: block; margin: 0 auto;" oncontextmenu="window.${handleMermaidContextMenu}(event, this)" onload="this.style.border='none'" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';" />
            <div style="display: none; text-align: center; padding: 16px; background: #f8f8f8; border: 1px solid #ddd; border-radius: 4px; color: #666;">
                Failed to load PlantUML diagram
            </div>`;

            const finalImg = `<div class="plantuml-container">${imgElement}</div>`;
            html = html.replace(block.placeholder, finalImg);

        } catch {
            // Fall back to showing the code
            const fallback = `<pre style="text-align: left; padding: 16px; background: #f8f8f8; border: 1px solid #ddd; border-radius: 4px;"><code class="language-plantuml">${block.code}</code></pre>`;
            html = html.replace(block.placeholder, fallback);
        }
    }

    return html;
};
