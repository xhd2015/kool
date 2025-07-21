import { useState, useEffect, useRef } from 'react';
// @ts-expect-error plantuml-encoder lacks TypeScript declarations
import plantumlEncoder from 'plantuml-encoder';
import { useZoom } from '../../hooks/useZoom';

interface UMLPreviewProps {
    content: string;
}

const UMLPreview = ({ content }: UMLPreviewProps) => {
    const [imageUrl, setImageUrl] = useState<string>('');
    const [displayUrl, setDisplayUrl] = useState<string>(''); // The URL actually shown to user
    const [error, setError] = useState<string | null>(null);
    const [isFirstLoad, setIsFirstLoad] = useState(true);
    const { containerRef, zoomState, zoomIn, zoomOut, fitToScreen } = useZoom();
    const imageRef = useRef<HTMLImageElement>(null);
    const hiddenImageRef = useRef<HTMLImageElement>(null); // For preloading new images

    useEffect(() => {
        try {
            const encoded = plantumlEncoder.encode(content);
            const url = `/planuml/svg/${encoded}`;
            setImageUrl(url);
            setError(null);

            // If this is the first load, set display URL immediately
            if (isFirstLoad) {
                setDisplayUrl(url);
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to encode UML content');
        }
    }, [content, isFirstLoad]);

    const handleDisplayImageLoad = () => {
        // This handles the load of the currently displayed image (first load only)
        if (isFirstLoad) {
            setError(null);

            // Use the useZoom hook's fitToScreen function for consistent behavior
            fitToScreen();

            // State change will trigger re-render and update opacity via JSX
            setIsFirstLoad(false);
        }
    };

    const handleNewImageLoad = () => {
        // This handles the load of the new image (for updates)
        setError(null);

        // Update display URL to show the new image
        setDisplayUrl(imageUrl);

        // Preserve current zoom state using requestAnimationFrame for proper DOM synchronization
        requestAnimationFrame(() => {
            if (containerRef.current && imageRef.current) {
                // Re-apply the current zoom state using the useZoom transform logic
                const image = imageRef.current;
                image.style.transform = `scale(${zoomState.scale}) translate(${zoomState.translateX}px, ${zoomState.translateY}px)`;
                image.style.transformOrigin = 'center center';
            }
        });
    };

    const handleImageError = () => {
        setError('Failed to load UML diagram. Syntax might be invalid, or PlantUML server is overloaded.');
    };

    const handleRetry = () => {
        setError(null);
        // Force reload by adding timestamp
        const encoded = plantumlEncoder.encode(content);
        const url = `/planuml/svg/${encoded}?t=${Date.now()}`;
        setImageUrl(url);
    };

    if (error) {
        return (
            <div className="preview-uml">
                <div className="error" style={{ margin: '16px', textAlign: 'center' }}>
                    ⚠️ {error}
                    <br />
                    <button
                        onClick={handleRetry}
                        className="save-button"
                        style={{ marginTop: '8px' }}
                    >
                        Retry
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="preview-uml">
            <div className="zoom-controls">
                <button className="zoom-button" onClick={zoomIn} title="Zoom In">🔍+</button>
                <button className="zoom-button" onClick={zoomOut} title="Zoom Out">🔍-</button>
                <button className="zoom-button" onClick={fitToScreen} title="Fit to Screen">⚪</button>
            </div>

            <div className="zoomable-container" ref={containerRef}>
                {/* Current image */}
                <img
                    ref={imageRef}
                    src={displayUrl}
                    alt="UML Diagram"
                    onLoad={handleDisplayImageLoad}
                    onError={handleImageError}
                    style={{
                        display: displayUrl ? 'block' : 'none',
                        // Hide image during first load to prevent showing clipped state
                        opacity: isFirstLoad ? 0 : 1,
                        transition: 'opacity 0.2s ease-in-out',
                        // For first load, use CSS to constrain size until fitToScreen() takes over
                        maxWidth: isFirstLoad ? '90%' : 'none',
                        maxHeight: isFirstLoad ? '90%' : 'none',
                        width: 'auto',
                        height: 'auto',
                        objectFit: isFirstLoad ? 'contain' : 'initial'
                    }}
                />
                {/* Hidden image for preloading new content */}
                {imageUrl !== displayUrl && (
                    <img
                        ref={hiddenImageRef}
                        src={imageUrl}
                        alt=""
                        onLoad={handleNewImageLoad}
                        onError={handleImageError}
                        style={{
                            position: 'absolute',
                            visibility: 'hidden',
                            pointerEvents: 'none'
                        }}
                    />
                )}
            </div>
        </div>
    );
};

export default UMLPreview; 