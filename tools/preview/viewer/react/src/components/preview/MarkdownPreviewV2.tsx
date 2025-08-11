import { useState, useEffect, useRef, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
// import remarkMermaid from 'remark-mermaid'; // Disabled due to type issues
// import rehypePrismPlus from 'rehype-prism-plus'; // Removed due to conflicts with custom code handling
import rehypeRaw from 'rehype-raw';
import mermaid from 'mermaid';
import '../../utils/mermaid'; // initialize mermaid
import { encode } from 'plantuml-encoder';
import { copyAsPng, svgToPng, pngBlobToDataUrl } from '../../utils/svg';
import { ScrollPositionManager } from '../../utils/scrollSync';
import { highlightCode } from '../../utils/syntaxHighlighting';

import './MarkdownPreviewV2.css';

interface MarkdownPreviewV2Props {
    content: string;
}

// Function to create SVG context menu callback
function createSvgCallback(setContextMenu: (menu: { x: number; y: number; svgData: string | null }) => void, setError: (error: string | null) => void): ((e: MouseEvent, element: SVGElement | HTMLImageElement) => void) {
    return function (e: MouseEvent, element: SVGElement | HTMLImageElement) {
        e.preventDefault();
        e.stopPropagation();

        try {
            let svgData: string;

            if (element.tagName.toLowerCase() === 'svg') {
                svgData = new XMLSerializer().serializeToString(element as SVGElement);
            } else if (element.tagName.toLowerCase() === 'img') {
                const img = element as HTMLImageElement;
                const canvas = document.createElement('canvas');
                const ctx = canvas.getContext('2d');
                if (!ctx) throw new Error('Could not get canvas context');

                canvas.width = img.naturalWidth || img.width;
                canvas.height = img.naturalHeight || img.height;
                ctx.drawImage(img, 0, 0);

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

// Helper function to convert SVG element to PNG data URL using svg.ts utilities
async function svgToPngDataUrl(svgElement: SVGElement): Promise<string> {
    try {
        const svgData = new XMLSerializer().serializeToString(svgElement);
        const pngBlob = await svgToPng(svgData);
        const pngDataUrl = await pngBlobToDataUrl(pngBlob);
        return pngDataUrl;
    } catch (err) {
        throw new Error(`Failed to convert SVG to PNG data URL: ${err}`);
    }
}

// Function to create copy section callback using unique section IDs
function createCopySectionCallback(
    getContainerRef: () => HTMLDivElement | null
): ((sectionId: string) => void) {
    return async function (sectionId: string) {
        console.log('DEBUG V2 copy section clicked:', { sectionId });

        const containerRef = getContainerRef();

        if (!containerRef) {
            console.log('DEBUG copy section failed: missing containerRef');
            return;
        }

        // Find the target heading element by ID
        const targetHeading = containerRef.querySelector(`#${sectionId}`) as HTMLElement;
        if (!targetHeading) {
            console.log('DEBUG copy section failed: target heading not found for ID:', sectionId);
            return;
        }

        const sectionLevel = parseInt(targetHeading.tagName.substring(1));
        console.log('DEBUG found target heading:', targetHeading.tagName, 'level:', sectionLevel);

        // Collect elements from this heading until the next heading of same or higher level
        const sectionElements: HTMLElement[] = [];
        let currentElement: HTMLElement | null = targetHeading;

        while (currentElement) {
            sectionElements.push(currentElement);

            // Get next sibling element
            currentElement = currentElement.nextElementSibling as HTMLElement;

            // Stop if we hit another heading of same or higher level
            if (currentElement &&
                currentElement.tagName.toLowerCase().match(/^h[1-6]$/) &&
                parseInt(currentElement.tagName.substring(1)) <= sectionLevel) {
                console.log('DEBUG stopping collection at heading:', currentElement.tagName, 'level:', parseInt(currentElement.tagName.substring(1)));
                break;
            }
        }

        console.log('DEBUG collected', sectionElements.length, 'elements for section');

        // Create temporary container and copy HTML
        const tempDiv = document.createElement('div');

        for (const element of sectionElements) {
            const clone = element.cloneNode(true) as HTMLElement;

            // Convert all SVG elements to PNG images
            const svgElements = clone.querySelectorAll('svg');
            for (const svgElement of svgElements) {
                try {
                    console.log('DEBUG converting SVG to PNG for copy');
                    const pngDataUrl = await svgToPngDataUrl(svgElement);

                    // Create an img element to replace the SVG
                    const imgElement = document.createElement('img');
                    imgElement.src = pngDataUrl;
                    imgElement.style.maxWidth = '100%';
                    imgElement.style.height = 'auto';

                    // Copy attributes from SVG to img if needed
                    const svgWidth = svgElement.getAttribute('width');
                    const svgHeight = svgElement.getAttribute('height');
                    if (svgWidth) imgElement.style.width = svgWidth.includes('px') ? svgWidth : `${svgWidth}px`;
                    if (svgHeight) imgElement.style.height = svgHeight.includes('px') ? svgHeight : `${svgHeight}px`;

                    // Replace SVG with PNG img element
                    svgElement.parentNode?.replaceChild(imgElement, svgElement);
                    console.log('DEBUG successfully converted SVG to PNG');
                } catch (err) {
                    console.error('DEBUG failed to convert SVG to PNG:', err);
                    // Keep the original SVG if conversion fails
                }
            }

            // Special handling for diagram containers
            const mermaidContainers = clone.querySelectorAll('.mermaid-container-v2');
            mermaidContainers.forEach(container => {
                // Ensure we're copying the rendered content, not any leftover code
                const svgElement = container.querySelector('svg');
                const imgElement = container.querySelector('img');
                if (svgElement || imgElement) {
                    // Clear container and add only the rendered element
                    const elementToKeep = imgElement || svgElement;
                    if (elementToKeep) {
                        container.innerHTML = '';
                        container.appendChild(elementToKeep.cloneNode(true));
                        console.log('DEBUG processed Mermaid diagram for copy');
                    }
                }
            });

            const plantumlContainers = clone.querySelectorAll('.plantuml-container-v2');
            plantumlContainers.forEach(container => {
                // Ensure we're copying the rendered image, not any leftover code
                const imgElement = container.querySelector('img');
                if (imgElement && imgElement.style.display !== 'none') {
                    // Clear container and add only the image
                    container.innerHTML = '';
                    container.appendChild(imgElement.cloneNode(true));
                    console.log('DEBUG processed PlantUML diagram for copy');
                }
            });

            // Remove copy buttons
            const copyButtons = clone.querySelectorAll('.copy-section-btn-v2, .copy-all-btn-v2');
            copyButtons.forEach(btn => btn.remove());

            tempDiv.appendChild(clone);
        }

        const htmlContent = tempDiv.innerHTML;
        console.log('DEBUG HTML content length:', htmlContent.length);
        console.log('DEBUG text content length:', tempDiv.textContent?.length || 0);

        // Copy to clipboard
        if (navigator.clipboard && (navigator.clipboard as any).write) {
            const clipboardItem = new ClipboardItem({
                'text/html': new Blob([htmlContent], { type: 'text/html' }),
                // 'text/plain': new Blob([htmlContent], { type: 'text/plain' }),
                // 'text/plain': new Blob([tempDiv.textContent || ''], { type: 'text/plain' })
            });

            (navigator.clipboard as any).write([clipboardItem]).then(() => {
                console.log('DEBUG copy section successful');
            }).catch((err: Error) => {
                console.error('DEBUG copy section failed (HTML):', err);
                navigator.clipboard.writeText(tempDiv.textContent || '').then(() => {
                    console.log('DEBUG copy section successful (text fallback)');
                }).catch(console.error);
            });
        } else {
            console.log('DEBUG using text-only clipboard fallback');
            navigator.clipboard.writeText(tempDiv.textContent || '').then(() => {
                console.log('DEBUG copy section successful (text only)');
            }).catch(console.error);
        }
    };
}

// Function to create copy all callback that accesses refs dynamically
function createCopyAllCallback(getContainerRef: () => HTMLDivElement | null): (() => void) {
    return function () {
        const containerRef = getContainerRef();
        if (!containerRef) return;

        const markdownContent = containerRef.querySelector('.markdown-content-v2');
        if (!markdownContent) return;

        const clone = markdownContent.cloneNode(true) as Element;

        // Special handling for diagram containers
        const mermaidContainers = clone.querySelectorAll('.mermaid-container-v2');
        mermaidContainers.forEach(container => {
            // Ensure we're copying the rendered SVG, not any leftover code
            const svgElement = container.querySelector('svg');
            if (svgElement) {
                // Clear container and add only the SVG
                container.innerHTML = '';
                container.appendChild(svgElement.cloneNode(true));
                console.log('DEBUG processed Mermaid diagram for copy all');
            }
        });

        const plantumlContainers = clone.querySelectorAll('.plantuml-container-v2');
        plantumlContainers.forEach(container => {
            // Ensure we're copying the rendered image, not any leftover code
            const imgElement = container.querySelector('img');
            if (imgElement && imgElement.style.display !== 'none') {
                // Clear container and add only the image
                container.innerHTML = '';
                container.appendChild(imgElement.cloneNode(true));
                console.log('DEBUG processed PlantUML diagram for copy all');
            }
        });

        const copyButtons = clone.querySelectorAll('.copy-section-btn-v2, .copy-all-btn-v2');
        copyButtons.forEach(btn => btn.remove());

        const htmlContent = clone.innerHTML;

        if (navigator.clipboard && (navigator.clipboard as any).write) {
            const clipboardItem = new ClipboardItem({
                'text/html': new Blob([htmlContent], { type: 'text/html' }),
                'text/plain': new Blob([clone.textContent || ''], { type: 'text/plain' })
            });

            (navigator.clipboard as any).write([clipboardItem]).catch((err: Error) => {
                console.error('Failed to copy all HTML:', err);
                navigator.clipboard.writeText(clone.textContent || '').catch(console.error);
            });
        } else {
            navigator.clipboard.writeText(clone.textContent || '').catch(console.error);
        }
    };
}

const MarkdownPreviewV2 = ({ content }: MarkdownPreviewV2Props) => {
    console.log('DEBUG MarkdownPreviewV2 component render, content length:', content.length);
    console.log('DEBUG MarkdownPreviewV2 COMPONENT ACTIVE - V2 PREVIEW');

    const [error, setError] = useState<string | null>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const [contextMenu, setContextMenu] = useState<{ x: number; y: number; svgData: string | null }>({ x: 0, y: 0, svgData: null });

    // Generate unique function names
    const handleMermaidContextMenuUniqFn = useMemo(() => `handleMermaidContextMenu_${crypto.randomUUID().replaceAll("-", "_")}`, [])
    const copySectionContentUniqFn = useMemo(() => `copySectionContent_${crypto.randomUUID().replaceAll("-", "_")}`, [])
    const copyAllContentUniqFn = useMemo(() => `copyAllContent_${crypto.randomUUID().replaceAll("-", "_")}`, [])

    // Enhanced scroll position management
    const scrollManagerRef = useRef(new ScrollPositionManager());

    // Track first heading
    const isFirstHeadingRef = useRef(true);

    // Setup global callbacks - recreate when content changes to ensure fresh refs
    useEffect(() => {
        console.log('DEBUG setting up global callbacks');

        // Create callbacks that dynamically access current refs
        console.log('DEBUG creating callbacks (ID-based approach)');
        const svgCallback = createSvgCallback(setContextMenu, setError);
        const copyCallback = createCopySectionCallback(
            () => containerRef.current
        );
        const copyAllCallback = createCopyAllCallback(
            () => containerRef.current
        );

        (window as unknown as Record<string, unknown>)[handleMermaidContextMenuUniqFn] = svgCallback;
        (window as unknown as Record<string, unknown>)[copySectionContentUniqFn] = copyCallback;
        (window as unknown as Record<string, unknown>)[copyAllContentUniqFn] = copyAllCallback;

        return () => {
            delete (window as unknown as Record<string, unknown>)[handleMermaidContextMenuUniqFn];
            delete (window as unknown as Record<string, unknown>)[copySectionContentUniqFn];
            delete (window as unknown as Record<string, unknown>)[copyAllContentUniqFn];
        };
    }, [handleMermaidContextMenuUniqFn, copySectionContentUniqFn, copyAllContentUniqFn, content]);

    // Context menu click handler
    const handleClick = () => {
        return
        if (!contextMenu || contextMenu.x === 0) return;
        setContextMenu(e => {
            if (!e || e.x === 0) {
                return e
            }
            return { x: 0, y: 0, svgData: null }
        })
    };



    // Custom heading component
    const HeadingComponent = ({ level, children, node, ...props }: { level: number; children: React.ReactNode; node?: any }) => {
        const HeadingTag = `h${level}` as 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';
        const sectionTitle = typeof children === 'string' ? children :
            Array.isArray(children) ? children.join('') : String(children);
        const isFirst = isFirstHeadingRef.current;

        // Generate unique ID for this section
        const sectionId = useMemo(() => `section-${crypto.randomUUID()}`, []);

        console.log('DEBUG HeadingComponent render:', {
            level,
            sectionTitle,
            sectionId,
            isFirst
        });

        if (isFirstHeadingRef.current) {
            isFirstHeadingRef.current = false;
        }

        const copyAllButton = isFirst ? (
            <button
                className="copy-all-btn-v2"
                onClick={() => (window as any)[copyAllContentUniqFn]()}
                title="Copy entire document"
            >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
                    <rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
                    <path d="M9 14l2 2 4-4"></path>
                </svg>
            </button>
        ) : null;

        return (
            <HeadingTag
                {...props}
                id={sectionId}
                className={`copyable-section-v2${isFirst ? ' first-heading-v2' : ''}`}
                data-section-title={sectionTitle}
                data-section-level={level}
            >
                {children}
                <button
                    className="copy-section-btn-v2"
                    onClick={() => (window as any)[copySectionContentUniqFn](sectionId)}
                    title="Copy section content"
                >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                    </svg>
                </button>
                {copyAllButton}
            </HeadingTag>
        );
    };

    // Custom code component for Mermaid and PlantUML
    const CodeComponent = ({ node, inline, className, children, ...props }: any) => {
        console.log("DEBUG CodeComponent render,props", className, props)
        const match = /language-(\w+)/.exec(className || '');
        const language = match ? match[1] : '';

        if (!inline) {
            if (language === 'mermaid') {
                const elementID = useMemo(() => `mermaid-${crypto.randomUUID()}`, []);
                console.log("DEBUG elementID", elementID)
                const code = String(children).replace(/\n$/, '');

                useEffect(() => {
                    const renderMermaid = async () => {
                        try {
                            const element = document.getElementById(elementID);
                            if (element) {
                                const { svg } = await mermaid.render(`${elementID}-svg`, code);
                                console.log("DEBUG recreate mermaid")
                                const svgWithEvents = svg.replace(
                                    '<svg',
                                    `<svg style="max-width: 100%; height: auto; user-select: none;" oncontextmenu="window.${handleMermaidContextMenuUniqFn}(event, this)"`
                                );
                                element.innerHTML = svgWithEvents;
                            }
                        } catch (err) {
                            const element = document.getElementById(elementID);
                            if (element) {
                                const errorMessage = err instanceof Error ? err.message : 'Unknown error';
                                element.innerHTML = `
                                    <div class="mermaid-error-v2">
                                        <strong>⚠️ Mermaid Rendering Error:</strong> ${errorMessage}
                                    </div>
                                    <pre class="code-block-v2" style="text-align: left;"><code class="language-mermaid">${code}</code></pre>
                                `;
                            }
                        }
                    };
                    renderMermaid();
                }, [elementID, code]);

                useEffect(() => {
                    console.log("DEBUG useEffect mount")
                    return () => {
                        console.log("DEBUG useEffect unmount")
                    }
                }, [])

                return (
                    <div className="mermaid-container-v2">
                        <div id={elementID}></div>
                    </div>
                );
            }

            if (language === 'plantuml' || language === 'puml' || language === 'uml') {
                const code = String(children).replace(/\n$/, '');
                const encoded = encode(code);
                const plantUmlUrl = `/planuml/svg/${encoded}`;

                return (
                    <div className="plantuml-container-v2">
                        <img
                            src={plantUmlUrl}
                            alt="PlantUML diagram"
                            style={{ maxWidth: '100%', height: 'auto', userSelect: 'none', display: 'block', margin: '0 auto' }}
                            onContextMenu={(e) => (window as any)[handleMermaidContextMenuUniqFn](e.nativeEvent, e.currentTarget)}
                            onLoad={(e) => { (e.currentTarget as HTMLImageElement).style.border = 'none' }}
                            onError={(e) => {
                                (e.currentTarget as HTMLImageElement).style.display = 'none';
                                const nextElement = (e.currentTarget as HTMLImageElement).nextElementSibling as HTMLElement;
                                if (nextElement) nextElement.style.display = 'block';
                            }}
                        />
                        <div style={{ display: 'none', textAlign: 'center', padding: '16px', background: '#f8f8f8', border: '1px solid #ddd', borderRadius: '4px', color: '#666' }}>
                            Failed to load PlantUML diagram
                        </div>
                    </div>
                );
            }
        }

        // Regular code block - handle syntax highlighting manually
        if (!inline && language) {
            // This is a code block with a language specified
            const code = String(children).replace(/\n$/, '');
            const highlightedHtml = highlightCode(code, language);

            // Return just the code element - the pre is handled by PreComponent
            return (
                <code
                    className={className}
                    dangerouslySetInnerHTML={{ __html: highlightedHtml }}
                />
            );
        }

        // Inline code or code without language
        return (
            <code
                className={className}
                {...props}
                style={inline ? {
                    backgroundColor: '#f5f5f5',
                    padding: '2px 4px',
                    borderRadius: '3px',
                    fontFamily: 'Consolas, Monaco, "Courier New", monospace',
                    fontSize: '0.9em'
                } : undefined}
            >
                {children}
            </code>
        );
    };

    // Custom paragraph component
    const ParagraphComponent = ({ node, ...props }: any) => (
        <p {...props} />
    );

    // Custom pre component
    const PreComponent = ({ node, ...props }: any) => {
        // Check if the child code element has diagram languages
        const child = props.children;
        const isDiagramCode = child && child.props && child.props.className && (
            child.props.className.includes('language-mermaid') ||
            child.props.className.includes('language-plantuml') ||
            child.props.className.includes('language-puml') ||
            child.props.className.includes('language-uml')
        );

        if (isDiagramCode) {
            // Return the child code component directly (which will render as diagram)
            // without pre wrapper
            return <div>{child}</div>;
        }

        // Regular code block - wrap in pre
        return (
            <pre
                {...props}
                className="code-block-v2"
            />
        );
    };

    // Reset first heading flag when content changes
    useEffect(() => {
        console.log('DEBUG content changed, preparing for new render');
        isFirstHeadingRef.current = true;

        // Save scroll position before updating content
        if (containerRef.current) {
            scrollManagerRef.current.savePosition(containerRef.current, content.length);
        }
    }, [content]);



    if (error) {
        return (
            <div className="preview-markdown-v2">
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
            className="preview-markdown-v2"
            ref={containerRef}
            onClick={handleClick}
            style={{
                height: '100%',
                overflow: 'auto'
            }}
        >
            <div className="markdown-content-v2">
                <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    rehypePlugins={[rehypeRaw]}
                    components={{
                        h1: ({ children, node, ...props }) => <HeadingComponent level={1} {...props}>{children}</HeadingComponent>,
                        h2: ({ children, node, ...props }) => <HeadingComponent level={2} {...props}>{children}</HeadingComponent>,
                        h3: ({ children, node, ...props }) => <HeadingComponent level={3} {...props}>{children}</HeadingComponent>,
                        h4: ({ children, node, ...props }) => <HeadingComponent level={4} {...props}>{children}</HeadingComponent>,
                        h5: ({ children, node, ...props }) => <HeadingComponent level={5} {...props}>{children}</HeadingComponent>,
                        h6: ({ children, node, ...props }) => <HeadingComponent level={6} {...props}>{children}</HeadingComponent>,
                        code: CodeComponent,
                        pre: PreComponent,
                        p: ParagraphComponent,
                        a: ({ href, title, children, node, ...props }) => (
                            <a href={href} title={title} target="_blank" rel="noopener noreferrer" {...props}>
                                {children}
                            </a>
                        ),
                    }}
                >
                    {content}
                </ReactMarkdown>
            </div>
            {contextMenu.svgData ? (
                <div
                    className="context-menu-v2"
                    style={{
                        position: 'fixed',
                        left: contextMenu.x,
                        top: contextMenu.y,
                        zIndex: 1000
                    }}
                    onClick={(e) => e.stopPropagation()}
                >
                    <button
                        className="context-menu-item-v2"
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

export default MarkdownPreviewV2;
