import { useState, useEffect } from 'react';
import mermaid from 'mermaid';
import { useZoom } from '../../hooks/useZoom';

interface MermaidPreviewProps {
    content: string;
}

const MermaidPreview = ({ content }: MermaidPreviewProps) => {
    const [displaySvg, setDisplaySvg] = useState<string>(''); // The SVG actually shown to user
    const [diagramReady, setDiagramReady] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [isFirstLoad, setIsFirstLoad] = useState(true);
    const { containerRef, zoomState, zoomIn, zoomOut, fitToScreen } = useZoom();

    useEffect(() => {
        const renderMermaid = async () => {
            try {
                setError(null);
                setDiagramReady(false); // Hide diagram while processing

                // Initialize Mermaid
                mermaid.initialize({
                    startOnLoad: false,
                    theme: document.body.classList.contains('dark-theme') ? 'dark' : 'default',
                    securityLevel: 'loose'
                });

                // Create unique ID for the diagram
                const diagramId = 'mermaid-diagram-' + Date.now();

                // Render the diagram
                const { svg: newSvg } = await mermaid.render(diagramId + '-svg', content);

                // For first load, apply fit-to-screen styling directly to the SVG
                let processedSvg = newSvg;
                if (isFirstLoad && containerRef.current) {
                    // Parse SVG dimensions
                    const widthMatch = newSvg.match(/width="([^"]+)"/);
                    const heightMatch = newSvg.match(/height="([^"]+)"/);

                    if (widthMatch && heightMatch) {
                        const svgWidth = parseFloat(widthMatch[1]);
                        const svgHeight = parseFloat(heightMatch[1]);
                        const containerRect = containerRef.current.getBoundingClientRect();

                        // Calculate fit-to-screen scale (same logic as useZoom)
                        const scaleX = (containerRect.width * 0.9) / svgWidth;
                        const scaleY = (containerRect.height * 0.9) / svgHeight;
                        const scale = Math.min(scaleX, scaleY, 1);

                        // Apply transform directly to SVG
                        if (scale < 1) {
                            processedSvg = newSvg.replace(
                                /<svg([^>]*)>/,
                                `<svg$1 style="transform: scale(${scale}); transform-origin: center center;">`
                            );
                        }
                    }
                    setIsFirstLoad(false);
                } else if (!isFirstLoad && zoomState.scale !== 1) {
                    // For updates, preserve zoom state by applying it directly to SVG
                    processedSvg = newSvg.replace(
                        /<svg([^>]*)>/,
                        `<svg$1 style="transform: scale(${zoomState.scale}) translate(${zoomState.translateX}px, ${zoomState.translateY}px); transform-origin: center center;">`
                    );
                }

                // Set the processed SVG content and show immediately
                setDisplaySvg(processedSvg);
                setDiagramReady(true);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to render Mermaid diagram');
                setDiagramReady(true); // Show error state
            }
        };

        if (content) {
            renderMermaid();
        }
    }, [content, fitToScreen, isFirstLoad, containerRef, zoomState]);

    if (error) {
        return (
            <div className="preview-mermaid">
                <div className="error" style={{ margin: '16px' }}>
                    ⚠️ Failed to render Mermaid diagram: {error}
                    <details style={{ marginTop: '8px' }}>
                        <summary>Mermaid content:</summary>
                        <pre style={{ marginTop: '8px', fontSize: '12px' }}>{content}</pre>
                    </details>
                </div>
            </div>
        );
    }

    return (
        <div className="preview-mermaid">
            <div className="zoom-controls">
                <button className="zoom-button" onClick={zoomIn} title="Zoom In">🔍+</button>
                <button className="zoom-button" onClick={zoomOut} title="Zoom Out">🔍-</button>
                <button className="zoom-button" onClick={fitToScreen} title="Fit to Screen">⚪</button>
            </div>

            <div className="zoomable-container" ref={containerRef}>
                {/* Current diagram */}
                <div
                    className="mermaid-container"
                    dangerouslySetInnerHTML={{ __html: displaySvg }}
                    style={{
                        opacity: diagramReady ? 1 : 0,
                        transition: diagramReady ? 'opacity 0.1s ease-in' : 'none',
                        position: 'relative',
                        zIndex: 1
                    }}
                />
            </div>
        </div>
    );
};

export default MermaidPreview; 