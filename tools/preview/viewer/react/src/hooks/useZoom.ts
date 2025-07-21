import { useState, useEffect, useRef, useCallback } from 'react';

interface ZoomState {
    scale: number;
    translateX: number;
    translateY: number;
}

export const useZoom = () => {
    const [zoomState, setZoomState] = useState<ZoomState>({ scale: 1, translateX: 0, translateY: 0 });
    const containerRef = useRef<HTMLDivElement>(null);
    const isDraggingRef = useRef(false);
    const dragStartRef = useRef({ x: 0, y: 0, translateX: 0, translateY: 0 });

    const getContentElement = useCallback(() => {
        if (!containerRef.current) return null;
        return containerRef.current.querySelector('img, .mermaid-container') as HTMLElement;
    }, []);

    const updateTransform = useCallback((scale: number, translateX: number, translateY: number) => {
        const contentElement = getContentElement();
        if (contentElement && containerRef.current) {
            contentElement.style.transform = `scale(${scale}) translate(${translateX}px, ${translateY}px)`;
            contentElement.style.transformOrigin = 'center center';

            // Check if content is larger than container to determine cursor
            const container = containerRef.current;
            const containerRect = container.getBoundingClientRect();

            let contentWidth: number;
            let contentHeight: number;

            if (contentElement instanceof HTMLImageElement) {
                contentWidth = (contentElement.naturalWidth || contentElement.width) * scale;
                contentHeight = (contentElement.naturalHeight || contentElement.height) * scale;
            } else {
                const contentRect = contentElement.getBoundingClientRect();
                contentWidth = contentRect.width;
                contentHeight = contentRect.height;
            }

            const isContentLargerThanContainer = contentWidth > containerRect.width || contentHeight > containerRect.height;
            const cursor = isContentLargerThanContainer ? 'grab' : 'default';

            contentElement.style.cursor = cursor;
            containerRef.current.style.cursor = cursor;
        }
        setZoomState({ scale, translateX, translateY });
    }, [getContentElement]);

    const zoomIn = useCallback(() => {
        const newScale = Math.min(zoomState.scale * 1.2, 3); // Max zoom 3x
        updateTransform(newScale, zoomState.translateX, zoomState.translateY);
    }, [zoomState, updateTransform]);

    const zoomOut = useCallback(() => {
        const newScale = Math.max(zoomState.scale / 1.2, 0.1); // Min zoom 0.1x
        updateTransform(newScale, zoomState.translateX, zoomState.translateY);
    }, [zoomState, updateTransform]);

    const zoomReset = useCallback(() => {
        updateTransform(1, 0, 0);
    }, [updateTransform]);

    const fitToScreen = useCallback(() => {
        if (!containerRef.current) return;

        const contentElement = getContentElement();
        if (!contentElement) return;

        const container = containerRef.current;
        const containerRect = container.getBoundingClientRect();

        // Get the natural dimensions of the content
        let contentWidth: number;
        let contentHeight: number;

        if (contentElement instanceof HTMLImageElement) {
            contentWidth = contentElement.naturalWidth || contentElement.width;
            contentHeight = contentElement.naturalHeight || contentElement.height;
        } else {
            // For other elements, use their current dimensions
            const contentRect = contentElement.getBoundingClientRect();
            contentWidth = contentRect.width;
            contentHeight = contentRect.height;
        }

        if (contentWidth === 0 || contentHeight === 0) return;

        // Calculate the scale to fit the content within the container
        const scaleX = (containerRect.width * 0.9) / contentWidth; // 90% of container width for padding
        const scaleY = (containerRect.height * 0.9) / contentHeight; // 90% of container height for padding
        const scale = Math.min(scaleX, scaleY, 1); // Don't scale up, only down

        updateTransform(scale, 0, 0);
    }, [getContentElement, updateTransform]);

    // Mouse wheel zoom support
    useEffect(() => {
        const container = containerRef.current;
        if (!container) return;

        const handleWheel = (e: WheelEvent) => {
            if (e.ctrlKey || e.metaKey) {
                e.preventDefault();

                const currentScale = zoomState.scale;
                let newScale: number;

                if (e.deltaY < 0) {
                    // Zoom in
                    newScale = Math.min(currentScale * 1.1, 3);
                } else {
                    // Zoom out
                    newScale = Math.max(currentScale / 1.1, 0.1);
                }

                updateTransform(newScale, zoomState.translateX, zoomState.translateY);
            }
        };

        container.addEventListener('wheel', handleWheel);
        return () => container.removeEventListener('wheel', handleWheel);
    }, [zoomState, updateTransform]);

    // Drag and pan support
    useEffect(() => {
        const container = containerRef.current;
        if (!container) return;

        const handleMouseDown = (e: MouseEvent) => {
            const contentElement = getContentElement();
            if (!contentElement || !containerRef.current) return;

            // Check if the scaled content is larger than the container
            const container = containerRef.current;
            const containerRect = container.getBoundingClientRect();

            let contentWidth: number;
            let contentHeight: number;

            if (contentElement instanceof HTMLImageElement) {
                contentWidth = (contentElement.naturalWidth || contentElement.width) * zoomState.scale;
                contentHeight = (contentElement.naturalHeight || contentElement.height) * zoomState.scale;
            } else {
                const contentRect = contentElement.getBoundingClientRect();
                contentWidth = contentRect.width;
                contentHeight = contentRect.height;
            }

            // Only enable dragging if the scaled content is larger than the container
            const isContentLargerThanContainer = contentWidth > containerRect.width || contentHeight > containerRect.height;
            if (!isContentLargerThanContainer) return;

            isDraggingRef.current = true;
            dragStartRef.current = {
                x: e.clientX,
                y: e.clientY,
                translateX: zoomState.translateX,
                translateY: zoomState.translateY
            };

            // Disable CSS transitions during dragging for smooth movement
            contentElement.style.transition = 'none';

            // Change cursor to grabbing
            if (containerRef.current) {
                containerRef.current.style.cursor = 'grabbing';
            }

            // Prevent default to avoid text selection and image dragging
            e.preventDefault();
        };

        const handleMouseMove = (e: MouseEvent) => {
            if (!isDraggingRef.current) return;

            e.preventDefault();

            const deltaX = e.clientX - dragStartRef.current.x;
            const deltaY = e.clientY - dragStartRef.current.y;

            // Adjust translation by the inverse of the scale factor
            // This ensures the visual movement matches the cursor movement
            const scaleFactor = zoomState.scale;
            const newTranslateX = dragStartRef.current.translateX + deltaX / scaleFactor;
            const newTranslateY = dragStartRef.current.translateY + deltaY / scaleFactor;

            updateTransform(zoomState.scale, newTranslateX, newTranslateY);
        };

        const handleMouseUp = () => {
            if (!isDraggingRef.current) return;

            isDraggingRef.current = false;

            // Restore CSS transitions after dragging
            const contentElement = getContentElement();
            if (contentElement) {
                contentElement.style.transition = 'transform 0.1s ease-out';
            }

            // Restore cursor based on content size vs container size
            if (contentElement && containerRef.current) {
                const container = containerRef.current;
                const containerRect = container.getBoundingClientRect();

                let contentWidth: number;
                let contentHeight: number;

                if (contentElement instanceof HTMLImageElement) {
                    contentWidth = (contentElement.naturalWidth || contentElement.width) * zoomState.scale;
                    contentHeight = (contentElement.naturalHeight || contentElement.height) * zoomState.scale;
                } else {
                    const contentRect = contentElement.getBoundingClientRect();
                    contentWidth = contentRect.width;
                    contentHeight = contentRect.height;
                }

                const isContentLargerThanContainer = contentWidth > containerRect.width || contentHeight > containerRect.height;
                const cursor = isContentLargerThanContainer ? 'grab' : 'default';

                containerRef.current.style.cursor = cursor;
            }
        };

        container.addEventListener('mousedown', handleMouseDown);
        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);

        return () => {
            container.removeEventListener('mousedown', handleMouseDown);
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
        };
    }, [zoomState, updateTransform, getContentElement]);

    return {
        containerRef,
        zoomState,
        zoomIn,
        zoomOut,
        zoomReset,
        fitToScreen
    };
}; 