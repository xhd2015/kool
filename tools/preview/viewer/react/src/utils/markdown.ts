import { marked, type Token } from 'marked';
import mermaid from 'mermaid';
import './mermaid'; // initialize mermaid
import { encode } from 'plantuml-encoder';
import { highlightCode } from './syntaxHighlighting';


// Pure function for rendering markdown to HTML with SVGs directly rendered (no side effects, no closure dependencies)
export async function renderMarkdownToHtml(content: string, handleMermaidContextMenu: string, copySectionContent?: string, copyAllContent?: string): Promise<string> {
    if (!content) {
        return '';
    }

    // First pass: collect all mermaid and plantuml code blocks
    const mermaidBlocks: { id: string; code: string; placeholder: string }[] = [];
    const plantumlBlocks: { id: string; code: string; placeholder: string }[] = [];

    // Track first heading for copy all functionality
    let isFirstHeading = true;

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

    // Custom renderer for heading elements to add copy functionality
    renderer.heading = function (token: { tokens: Token[], depth: number }) {
        const text = this.parser.parseInline(token.tokens);
        const level = token.depth;

        if (copySectionContent) {
            // For all heading levels, add copy functionality
            const sectionTitle = text.replace(/<[^>]*>/g, ''); // Strip HTML tags for section title
            const isFirst = isFirstHeading;
            isFirstHeading = false; // Mark that we've processed the first heading

            let copyAllButton = '';
            if (isFirst && copyAllContent) {
                // Add copy all button for the first heading
                copyAllButton = `<button class="copy-all-btn" onclick="window.${copyAllContent}()" title="Copy entire document">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
                        <rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
                        <path d="M9 14l2 2 4-4"></path>
                    </svg>
                </button>`;
            }

            return `<h${level} class="copyable-section${isFirst ? ' first-heading' : ''}" data-section-title="${sectionTitle}" data-section-level="${level}">
                ${text}
                <button class="copy-section-btn" onclick="window.${copySectionContent}('${sectionTitle}', ${level})" title="Copy section content">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                    </svg>
                </button>
                ${copyAllButton}
            </h${level}>`;
        } else {
            // For other heading levels, use default rendering
            return `<h${level}>${text}</h${level}>`;
        }
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
