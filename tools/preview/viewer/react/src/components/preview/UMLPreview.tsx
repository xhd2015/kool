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
    const [dataUrl, setDataUrl] = useState<string>(''); // Base64 data URL for copying
    const [error, setError] = useState<string | null>(null);
    const [isFirstLoad, setIsFirstLoad] = useState(true);
    const [showContextMenu, setShowContextMenu] = useState(false);
    const [contextMenuPos, setContextMenuPos] = useState({ x: 0, y: 0 });
    const { containerRef, zoomState, zoomIn, zoomOut, fitToScreen } = useZoom();
    const imageRef = useRef<HTMLImageElement>(null);
    const hiddenImageRef = useRef<HTMLImageElement>(null); // For preloading new images

    // Convert SVG URL to data URL for copying
    const convertToDataUrl = async (url: string): Promise<string> => {
        try {
            const response = await fetch(url);
            if (!response.ok) {
                throw new Error(`Failed to fetch SVG: ${response.status}`);
            }
            const svgText = await response.text();

            // Create a data URL from the SVG content
            const dataUrl = `data:image/svg+xml;base64,${btoa(unescape(encodeURIComponent(svgText)))}`;
            return dataUrl;
        } catch (error) {
            console.error('Failed to convert to data URL:', error);
            return url; // Fallback to original URL
        }
    };

    useEffect(() => {
        try {
            const encoded = plantumlEncoder.encode(content);
            const url = `/planuml/svg/${encoded}`;
            setImageUrl(url);
            setError(null);

            // Convert to data URL for copying
            convertToDataUrl(url).then(setDataUrl);

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

        // Convert new image to data URL
        convertToDataUrl(imageUrl).then(setDataUrl);

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

    const handleContextMenu = (e: React.MouseEvent) => {
        e.preventDefault();
        setContextMenuPos({ x: e.clientX, y: e.clientY });
        setShowContextMenu(true);
    };

    const convertSvgToPng = async (svgDataUrl: string, scale: number = 2): Promise<string> => {
        return new Promise((resolve, reject) => {
            const img = new Image();
            img.onload = () => {
                const canvas = document.createElement('canvas');
                const ctx = canvas.getContext('2d');
                if (!ctx) {
                    reject(new Error('Failed to get canvas context'));
                    return;
                }

                // Set canvas size with scaling for better quality
                canvas.width = img.width * scale;
                canvas.height = img.height * scale;

                // Scale the context to match
                ctx.scale(scale, scale);

                // Set white background (important for Confluence)
                ctx.fillStyle = 'white';
                ctx.fillRect(0, 0, img.width, img.height);

                // Draw the SVG
                ctx.drawImage(img, 0, 0);

                // Convert to PNG data URL
                const pngDataUrl = canvas.toDataURL('image/png', 1.0);
                resolve(pngDataUrl);
            };
            img.onerror = () => reject(new Error('Failed to load SVG image'));
            img.src = svgDataUrl;
        });
    };

    const showSuccessMessage = (message: string) => {
        const notification = document.createElement('div');
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: #4CAF50;
            color: white;
            padding: 8px 16px;
            border-radius: 4px;
            z-index: 10000;
            font-family: Arial, sans-serif;
            font-size: 14px;
        `;
        document.body.appendChild(notification);
        setTimeout(() => document.body.removeChild(notification), 2000);
    };

    const handleCopyAsImage = async () => {
        setShowContextMenu(false);

        if (!dataUrl) {
            alert('Image data is not ready yet. Please wait a moment and try again.');
            return;
        }

        try {
            // Convert SVG to PNG for better Confluence compatibility
            const pngDataUrl = await convertSvgToPng(dataUrl);

            // Convert PNG data URL to blob
            const response = await fetch(pngDataUrl);
            const blob = await response.blob();

            // Copy to clipboard
            await navigator.clipboard.write([
                new ClipboardItem({
                    'image/png': blob
                })
            ]);

            showSuccessMessage('✓ Copied as PNG');

        } catch (error) {
            console.error('Failed to copy image:', error);
            alert('Failed to copy image to clipboard. Your browser may not support this feature.');
        }
    };

    const handleCopyAsSvg = async () => {
        setShowContextMenu(false);

        if (!dataUrl) {
            alert('Image data is not ready yet. Please wait a moment and try again.');
            return;
        }

        try {
            // Convert data URL to blob
            const response = await fetch(dataUrl);
            const blob = await response.blob();

            // Copy to clipboard
            await navigator.clipboard.write([
                new ClipboardItem({
                    [blob.type]: blob
                })
            ]);

            showSuccessMessage('✓ Copied as SVG');

        } catch (error) {
            console.error('Failed to copy SVG:', error);
            alert('Failed to copy SVG to clipboard. Your browser may not support this feature.');
        }
    };

    const handleDownloadPng = async () => {
        setShowContextMenu(false);

        if (!dataUrl) {
            alert('Image data is not ready yet. Please wait a moment and try again.');
            return;
        }

        try {
            // Convert SVG to PNG
            const pngDataUrl = await convertSvgToPng(dataUrl);

            // Create download link
            const link = document.createElement('a');
            link.href = pngDataUrl;
            link.download = 'uml-diagram.png';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        } catch (error) {
            console.error('Failed to convert to PNG:', error);
            alert('Failed to convert image to PNG.');
        }
    };

    const handleDownloadSvg = () => {
        setShowContextMenu(false);

        if (!dataUrl) {
            alert('Image data is not ready yet. Please wait a moment and try again.');
            return;
        }

        // Create download link
        const link = document.createElement('a');
        link.href = dataUrl;
        link.download = 'uml-diagram.svg';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    };

    // Close context menu when clicking elsewhere
    useEffect(() => {
        const handleClick = () => setShowContextMenu(false);
        if (showContextMenu) {
            document.addEventListener('click', handleClick);
            return () => document.removeEventListener('click', handleClick);
        }
    }, [showContextMenu]);

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
                    src={dataUrl || displayUrl}
                    alt="UML Diagram"
                    onLoad={handleDisplayImageLoad}
                    onError={handleImageError}
                    onContextMenu={handleContextMenu}
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

                {/* Custom context menu */}
                {showContextMenu && (
                    <div
                        style={{
                            position: 'fixed',
                            left: contextMenuPos.x,
                            top: contextMenuPos.y,
                            background: 'white',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                            boxShadow: '0 2px 10px rgba(0,0,0,0.1)',
                            zIndex: 10000,
                            minWidth: '150px'
                        }}
                    >
                        <button
                            onClick={handleCopyAsImage}
                            style={{
                                display: 'block',
                                width: '100%',
                                padding: '8px 12px',
                                border: 'none',
                                background: 'none',
                                textAlign: 'left',
                                cursor: 'pointer',
                                fontSize: '14px'
                            }}
                            onMouseEnter={(e) => e.currentTarget.style.background = '#f0f0f0'}
                            onMouseLeave={(e) => e.currentTarget.style.background = 'none'}
                        >
                            📋 Copy Image as PNG
                        </button>
                        <button
                            onClick={handleCopyAsSvg}
                            style={{
                                display: 'block',
                                width: '100%',
                                padding: '8px 12px',
                                border: 'none',
                                background: 'none',
                                textAlign: 'left',
                                cursor: 'pointer',
                                fontSize: '14px'
                            }}
                            onMouseEnter={(e) => e.currentTarget.style.background = '#f0f0f0'}
                            onMouseLeave={(e) => e.currentTarget.style.background = 'none'}
                        >
                            📋 Copy Image as SVG
                        </button>
                        <hr style={{ margin: '4px 0', border: 'none', borderTop: '1px solid #eee' }} />
                        <button
                            onClick={handleDownloadPng}
                            style={{
                                display: 'block',
                                width: '100%',
                                padding: '8px 12px',
                                border: 'none',
                                background: 'none',
                                textAlign: 'left',
                                cursor: 'pointer',
                                fontSize: '14px'
                            }}
                            onMouseEnter={(e) => e.currentTarget.style.background = '#f0f0f0'}
                            onMouseLeave={(e) => e.currentTarget.style.background = 'none'}
                        >
                            💾 Download PNG
                        </button>
                        <button
                            onClick={handleDownloadSvg}
                            style={{
                                display: 'block',
                                width: '100%',
                                padding: '8px 12px',
                                border: 'none',
                                background: 'none',
                                textAlign: 'left',
                                cursor: 'pointer',
                                fontSize: '14px'
                            }}
                            onMouseEnter={(e) => e.currentTarget.style.background = '#f0f0f0'}
                            onMouseLeave={(e) => e.currentTarget.style.background = 'none'}
                        >
                            💾 Download SVG
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
};

export default UMLPreview;
