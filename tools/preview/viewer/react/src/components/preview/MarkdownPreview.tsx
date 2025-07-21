import { useState, useEffect } from 'react';
import { marked } from 'marked';
import './MarkdownPreview.css';

interface MarkdownPreviewProps {
    content: string;
}

const MarkdownPreview = ({ content }: MarkdownPreviewProps) => {
    const [htmlContent, setHtmlContent] = useState<string>('');
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const renderMarkdown = async () => {
            try {
                setError(null);

                // Configure marked options for better rendering
                marked.setOptions({
                    gfm: true, // GitHub Flavored Markdown
                    breaks: true, // Convert \n to <br>
                });

                // Configure custom renderer to make links open in new tab
                const renderer = new marked.Renderer();
                renderer.link = function (token: { href: string, title?: string | null, tokens: any[] }) {
                    const titleAttr = token.title ? ` title="${token.title}"` : '';
                    const text = this.parser.parseInline(token.tokens);
                    return `<a href="${token.href}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
                };

                const html = await marked(content, { renderer });
                setHtmlContent(html);
            } catch (err) {
                console.error('Failed to render markdown:', err);
                setError(err instanceof Error ? err.message : 'Failed to render markdown');
            }
        };

        if (content) {
            renderMarkdown();
        } else {
            setHtmlContent('');
        }
    }, [content]);

    if (error) {
        return (
            <div className="preview-markdown">
                <div className="error" style={{ margin: '16px' }}>
                    ⚠️ Failed to render markdown: {error}
                    <details style={{ marginTop: '8px' }}>
                        <summary>Markdown content:</summary>
                        <pre style={{ marginTop: '8px', fontSize: '12px' }}>{content}</pre>
                    </details>
                </div>
            </div>
        );
    }

    return (
        <div className="preview-markdown">
            <div
                className="markdown-content"
                dangerouslySetInnerHTML={{ __html: htmlContent }}
            />
        </div>
    );
};

export default MarkdownPreview; 