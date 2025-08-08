import { useState, useEffect, useRef, useMemo } from 'react';
import { renderMarkdownToHtml } from '../../utils/markdown';
import { copyAsPng } from '../../utils/svg';
import { ScrollPositionManager } from '../../utils/scrollSync';

import './MarkdownPreview.css';


interface MarkdownPreviewProps {
    content: string;
}

function createSvgCallback(setContextMenu: (menu: { x: number; y: number; svgData: string | null }) => void, setError: (error: string | null) => void): ((e: MouseEvent, element: SVGElement | HTMLImageElement) => void) {
    return function (e: MouseEvent, element: SVGElement | HTMLImageElement) {
        e.preventDefault();
        e.stopPropagation();

        try {
            let svgData: string;

            if (element.tagName.toLowerCase() === 'svg') {
                // Handle SVG elements (Mermaid)
                svgData = new XMLSerializer().serializeToString(element as SVGElement);
            } else if (element.tagName.toLowerCase() === 'img') {
                // Handle IMG elements (PlantUML) - convert to SVG
                const img = element as HTMLImageElement;
                const canvas = document.createElement('canvas');
                const ctx = canvas.getContext('2d');
                if (!ctx) throw new Error('Could not get canvas context');

                canvas.width = img.naturalWidth || img.width;
                canvas.height = img.naturalHeight || img.height;
                ctx.drawImage(img, 0, 0);

                // Create SVG from canvas
                svgData = `<svg xmlns="http://www.w3.org/2000/svg" width="${canvas.width}" height="${canvas.height}">
                    <foreignObject width="100%" height="100%">
                        <img src="${img.src}" width="${canvas.width}" height="${canvas.height}" />
                    </foreignObject>
                </svg>`;
            } else {
                throw new Error('Unsupported element type');
            }

            setContextMenu({
                x: e.clientX,
                y: e.clientY,
                svgData: svgData
            });
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error')
        }
    }
}

const MarkdownPreview = ({ content }: MarkdownPreviewProps) => {
    const [htmlContent, setHtmlContent] = useState<string>('');
    const [error, setError] = useState<string | null>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; svgData: string | null }>({ x: 0, y: 0, svgData: null });
    const handleMermaidContextMenuUniqFn = useMemo(() => `handleMermaidContextMenu_${crypto.randomUUID().replaceAll("-", "_")}`, [])

    // Enhanced scroll position management
    const scrollManagerRef = useRef(new ScrollPositionManager());

    // Context menu now handled globally - no longer needed!
    const handleClick = () => {
        setContextMenu(e => {
            if (!e || e.x === 0) {
                return e
            }
            return { x: 0, y: 0, svgData: null }
        })
    };

    useEffect(() => {
        const callback = createSvgCallback(setContextMenu, setError);

        (window as unknown as Record<string, unknown>)[handleMermaidContextMenuUniqFn] = callback;
        return () => {
            delete (window as unknown as Record<string, unknown>)[handleMermaidContextMenuUniqFn];
        };
    }, [handleMermaidContextMenuUniqFn])


    useEffect(() => {
        debugger
        const processContent = async () => {
            try {
                setError(null);

                // Save scroll position before updating content
                if (containerRef.current) {
                    scrollManagerRef.current.savePosition(containerRef.current, content.length);
                }

                const html = await renderMarkdownToHtml(content, handleMermaidContextMenuUniqFn);
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
    }, [content, handleMermaidContextMenuUniqFn]);

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
        <div
            className="preview-markdown"
            ref={containerRef}
            onClick={handleClick}
            style={{
                height: '100%',
                overflow: 'auto'
            }}
        >
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