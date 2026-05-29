import { useState, useEffect } from 'react';
import { renderDotToSvg } from '../../utils/dot';
import { useZoom } from '../../hooks/useZoom';

interface DOTPreviewProps {
    content: string;
}

const DOTPreview = ({ content }: DOTPreviewProps) => {
    const [displaySvg, setDisplaySvg] = useState<string>('');
    const [diagramReady, setDiagramReady] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [isFirstLoad, setIsFirstLoad] = useState(true);
    const { containerRef, zoomState, zoomIn, zoomOut, fitToScreen } = useZoom();

    useEffect(() => {
        const renderDot = async () => {
            try {
                setError(null);
                setDiagramReady(false);

                let newSvg = await renderDotToSvg(content);

                if (isFirstLoad && containerRef.current) {
                    const widthMatch = newSvg.match(/width="([^"]+)"/);
                    const heightMatch = newSvg.match(/height="([^"]+)"/);

                    if (widthMatch && heightMatch) {
                        const svgWidth = parseFloat(widthMatch[1]);
                        const svgHeight = parseFloat(heightMatch[1]);
                        const containerRect = containerRef.current.getBoundingClientRect();

                        const scaleX = (containerRect.width * 0.9) / svgWidth;
                        const scaleY = (containerRect.height * 0.9) / svgHeight;
                        const scale = Math.min(scaleX, scaleY, 1);

                        if (scale < 1) {
                            newSvg = newSvg.replace(
                                /<svg([^>]*)>/,
                                `<svg$1 style="transform: scale(${scale}); transform-origin: center center;">`
                            );
                        }
                    }
                    setIsFirstLoad(false);
                } else if (!isFirstLoad && zoomState.scale !== 1) {
                    newSvg = newSvg.replace(
                        /<svg([^>]*)>/,
                        `<svg$1 style="transform: scale(${zoomState.scale}) translate(${zoomState.translateX}px, ${zoomState.translateY}px); transform-origin: center center;">`
                    );
                }

                setDisplaySvg(newSvg);
                setDiagramReady(true);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to render DOT diagram');
                setDiagramReady(true);
            }
        };

        if (content) {
            renderDot();
        }
    }, [content, fitToScreen, isFirstLoad, containerRef, zoomState]);

    if (error) {
        return (
            <div className="preview-dot">
                <div className="error" style={{ margin: '16px' }}>
                    ⚠️ Failed to render DOT diagram: {error}
                    <details style={{ marginTop: '8px' }}>
                        <summary>DOT content:</summary>
                        <pre style={{ marginTop: '8px', fontSize: '12px' }}>{content}</pre>
                    </details>
                </div>
            </div>
        );
    }

    return (
        <div className="preview-dot">
            <div className="zoom-controls">
                <button className="zoom-button" onClick={zoomIn} title="Zoom In">🔍+</button>
                <button className="zoom-button" onClick={zoomOut} title="Zoom Out">🔍-</button>
                <button className="zoom-button" onClick={fitToScreen} title="Fit to Screen">⚪</button>
            </div>

            <div className="zoomable-container" ref={containerRef}>
                <div
                    className="dot-container"
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

export default DOTPreview;
