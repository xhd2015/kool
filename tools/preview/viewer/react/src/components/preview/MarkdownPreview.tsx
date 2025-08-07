import { useState, useEffect, useRef, useMemo } from 'react';
import './MarkdownPreview.css';
import { renderMarkdownToHtml } from '../../utils/markdown';
import { copyAsPng } from '../../utils/svg';

interface MarkdownPreviewProps {
    content: string;
}

// Counter for generating unique mermaid IDs
let mermaidIdCounter = 0;

function createSvgCallback(setContextMenu: (menu: { x: number; y: number; svgData: string | null }) => void, setError: (error: any) => void): ((e: MouseEvent, svgElement: SVGElement) => void) {
    return function (e: MouseEvent, svgElement: SVGElement) {
        e.preventDefault();
        e.stopPropagation();

        try {
            const svgData = new XMLSerializer().serializeToString(svgElement);
            setContextMenu({
                x: e.clientX,
                y: e.clientY,
                svgData: svgData
            });
        } catch (err) {
            setError(err)
        }
    }
}


const MarkdownPreview = ({ content }: MarkdownPreviewProps) => {
    const [htmlContent, setHtmlContent] = useState<string>('');
    const [error, setError] = useState<string | null>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; svgData: string | null }>({ x: 0, y: 0, svgData: null });
    const handleMermaidContextMenu = useMemo(() => "handleMermaidContextMenu_" + new Date().getTime(), [])

    // Context menu now handled globally - no longer needed!
    const handleClick = () => {
        setContextMenu({ x: 0, y: 0, svgData: null });
    };

    useEffect(() => {
        const callback = createSvgCallback(setContextMenu, setError);

        (window as any)[handleMermaidContextMenu] = callback;
        return () => {
            delete (window as any)[handleMermaidContextMenu];
        };
    }, [])


    useEffect(() => {
        const processContent = async () => {
            try {
                setError(null);
                mermaidIdCounter = 0; // Reset the global counter

                const html = await renderMarkdownToHtml(content, handleMermaidContextMenu, mermaidIdCounter++);
                setHtmlContent(html);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to render markdown');
            }
        };

        if (content) {
            processContent();
        } else {
            setHtmlContent('');
        }
    }, [content]);

    // No longer needed - SVGs are rendered directly in renderMarkdownToHtml!

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
        <div className="preview-markdown" ref={containerRef} onClick={handleClick}>
            <div
                className="markdown-content"
                dangerouslySetInnerHTML={{ __html: htmlContent }}
            />
            {contextMenu.svgData ? (
                <div
                    className="context-menu"
                    style={{
                        position: 'fixed',
                        left: contextMenu.x,
                        top: contextMenu.y,
                        zIndex: 1000
                    }}
                    onClick={(e) => e.stopPropagation()}
                >
                    <button
                        className="context-menu-item"
                        onClick={(e) => {
                            e.stopPropagation();
                            copyAsPng(contextMenu.svgData!);
                            setContextMenu({ x: 0, y: 0, svgData: null });
                        }}
                    >
                        Copy as PNG
                    </button>
                </div>
            ) : null}
        </div>
    );
};

export default MarkdownPreview; 