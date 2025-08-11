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

function createCopySectionCallback(containerRef: React.RefObject<HTMLDivElement | null>): ((sectionTitle: string, sectionLevel: number) => void) {
    return function (sectionTitle: string, sectionLevel: number) {
        if (!containerRef.current) return;

        // Find the heading element with the matching title and level
        const headingElements = containerRef.current.querySelectorAll('h1.copyable-section, h2.copyable-section, h3.copyable-section, h4.copyable-section, h5.copyable-section, h6.copyable-section');
        let targetHeading: Element | null = null;

        for (const heading of headingElements) {
            if (heading.getAttribute('data-section-title') === sectionTitle &&
                parseInt(heading.getAttribute('data-section-level') || '0') === sectionLevel) {
                targetHeading = heading;
                break;
            }
        }

        if (!targetHeading) return;

        // Collect all elements from this heading until the next heading of same or higher level
        const sectionElements: Element[] = [targetHeading];
        let currentElement = targetHeading.nextElementSibling;

        while (currentElement) {
            const tagName = currentElement.tagName.toLowerCase();

            // Check if it's a heading element
            if (tagName.match(/^h[1-6]$/)) {
                const currentLevel = parseInt(tagName.substring(1));
                // Stop if we encounter a heading of the same level or higher (lower number)
                if (currentLevel <= sectionLevel) {
                    break;
                }
            }

            sectionElements.push(currentElement);
            currentElement = currentElement.nextElementSibling;
        }

        // Create a temporary container to get the HTML
        const tempDiv = document.createElement('div');
        sectionElements.forEach(element => {
            // Clone the element to avoid modifying the original
            const clone = element.cloneNode(true) as Element;

            // Remove the copy button from heading clones
            if (clone.tagName.toLowerCase().match(/^h[1-6]$/)) {
                const copyBtn = clone.querySelector('.copy-section-btn');
                if (copyBtn) {
                    copyBtn.remove();
                }
            }

            tempDiv.appendChild(clone);
        });

        // Copy the HTML content to clipboard
        const htmlContent = tempDiv.innerHTML;

        // Use the modern clipboard API with HTML
        if (navigator.clipboard && (navigator.clipboard as any).write) {
            const clipboardItem = new ClipboardItem({
                'text/html': new Blob([htmlContent], { type: 'text/html' }),
                'text/plain': new Blob([tempDiv.textContent || ''], { type: 'text/plain' })
            });

            (navigator.clipboard as any).write([clipboardItem]).catch((err: Error) => {
                console.error('Failed to copy section HTML:', err);
                // Fallback to plain text
                navigator.clipboard.writeText(tempDiv.textContent || '').catch(console.error);
            });
        } else {
            // Fallback for browsers that don't support ClipboardItem
            navigator.clipboard.writeText(tempDiv.textContent || '').catch(console.error);
        }
    };
}

function createCopyAllCallback(containerRef: React.RefObject<HTMLDivElement | null>): (() => void) {
    return function () {
        if (!containerRef.current) return;

        // Get the entire markdown content container
        const markdownContent = containerRef.current.querySelector('.markdown-content');
        if (!markdownContent) return;

        // Clone the entire content to avoid modifying the original
        const clone = markdownContent.cloneNode(true) as Element;

        // Remove all copy buttons from the clone
        const copyButtons = clone.querySelectorAll('.copy-section-btn, .copy-all-btn');
        copyButtons.forEach(btn => btn.remove());

        // Copy the HTML content to clipboard
        const htmlContent = clone.innerHTML;

        // Use the modern clipboard API with HTML
        if (navigator.clipboard && (navigator.clipboard as any).write) {
            const clipboardItem = new ClipboardItem({
                'text/html': new Blob([htmlContent], { type: 'text/html' }),
                'text/plain': new Blob([clone.textContent || ''], { type: 'text/plain' })
            });

            (navigator.clipboard as any).write([clipboardItem]).catch((err: Error) => {
                console.error('Failed to copy all HTML:', err);
                // Fallback to plain text
                navigator.clipboard.writeText(clone.textContent || '').catch(console.error);
            });
        } else {
            // Fallback for browsers that don't support ClipboardItem
            navigator.clipboard.writeText(clone.textContent || '').catch(console.error);
        }
    };
}

const MarkdownPreview = ({ content }: MarkdownPreviewProps) => {
    const [htmlContent, setHtmlContent] = useState<string>('');
    const [error, setError] = useState<string | null>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; svgData: string | null }>({ x: 0, y: 0, svgData: null });
    const handleMermaidContextMenuUniqFn = useMemo(() => `handleMermaidContextMenu_${crypto.randomUUID().replaceAll("-", "_")}`, [])
    const copySectionContentUniqFn = useMemo(() => `copySectionContent_${crypto.randomUUID().replaceAll("-", "_")}`, [])
    const copyAllContentUniqFn = useMemo(() => `copyAllContent_${crypto.randomUUID().replaceAll("-", "_")}`, [])

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
        const svgCallback = createSvgCallback(setContextMenu, setError);
        const copyCallback = createCopySectionCallback(containerRef);
        const copyAllCallback = createCopyAllCallback(containerRef);

        (window as unknown as Record<string, unknown>)[handleMermaidContextMenuUniqFn] = svgCallback;
        (window as unknown as Record<string, unknown>)[copySectionContentUniqFn] = copyCallback;
        (window as unknown as Record<string, unknown>)[copyAllContentUniqFn] = copyAllCallback;

        return () => {
            delete (window as unknown as Record<string, unknown>)[handleMermaidContextMenuUniqFn];
            delete (window as unknown as Record<string, unknown>)[copySectionContentUniqFn];
            delete (window as unknown as Record<string, unknown>)[copyAllContentUniqFn];
        };
    }, [handleMermaidContextMenuUniqFn, copySectionContentUniqFn, copyAllContentUniqFn])


    useEffect(() => {
        debugger
        const processContent = async () => {
            try {
                setError(null);

                // Save scroll position before updating content
                if (containerRef.current) {
                    scrollManagerRef.current.savePosition(containerRef.current, content.length);
                }

                const html = await renderMarkdownToHtml(content, handleMermaidContextMenuUniqFn, copySectionContentUniqFn, copyAllContentUniqFn);
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
    }, [content, handleMermaidContextMenuUniqFn, copySectionContentUniqFn, copyAllContentUniqFn]);

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