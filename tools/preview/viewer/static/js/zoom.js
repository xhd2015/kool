// Zoom functionality
export function zoomIn(button) {
    const container = button.closest('.preview-uml, .preview-mermaid');
    const zoomableContainer = container.querySelector('.zoomable-container');
    const contentElement = zoomableContainer?.querySelector('img') || zoomableContainer?.querySelector('.mermaid-container');

    if (contentElement) {
        const currentScale = getZoomScale(contentElement);
        const newScale = Math.min(currentScale * 1.2, 3); // Max zoom 3x
        setZoomScale(contentElement, newScale);
    }
}

export function zoomOut(button) {
    const container = button.closest('.preview-uml, .preview-mermaid');
    const zoomableContainer = container.querySelector('.zoomable-container');
    const contentElement = zoomableContainer?.querySelector('img') || zoomableContainer?.querySelector('.mermaid-container');

    if (contentElement) {
        const currentScale = getZoomScale(contentElement);
        const newScale = Math.max(currentScale / 1.2, 0.1); // Min zoom 0.1x
        setZoomScale(contentElement, newScale);
    }
}

export function zoomReset(button) {
    const container = button.closest('.preview-uml, .preview-mermaid');
    const zoomableContainer = container.querySelector('.zoomable-container');
    const contentElement = zoomableContainer?.querySelector('img') || zoomableContainer?.querySelector('.mermaid-container');

    if (contentElement) {
        // Reset both scale and translate
        contentElement.style.transform = 'scale(1) translate(0px, 0px)';
        contentElement.style.transformOrigin = 'center center';
        contentElement.style.cursor = 'default';
    }
}

function getZoomScale(element) {
    const transform = element.style.transform;
    if (transform && transform.includes('scale')) {
        const match = transform.match(/scale\(([^)]+)\)/);
        if (match) {
            return parseFloat(match[1]);
        }
    }
    return 1;
}

function getTranslateValues(element) {
    const transform = element.style.transform;
    if (transform && transform.includes('translate')) {
        const match = transform.match(/translate\(([^,]+),\s*([^)]+)\)/);
        if (match) {
            return {
                x: parseFloat(match[1]),
                y: parseFloat(match[2])
            };
        }
    }
    return { x: 0, y: 0 };
}

function setTranslateValues(element, x, y) {
    const currentScale = getZoomScale(element);
    element.style.transform = `scale(${currentScale}) translate(${x}px, ${y}px)`;
}

function setZoomScale(element, scale) {
    // Get current translate values if they exist
    const currentTransform = element.style.transform;
    let translateX = 0;
    let translateY = 0;

    if (currentTransform) {
        const translateMatch = currentTransform.match(/translate\(([^,]+),\s*([^)]+)\)/);
        if (translateMatch) {
            translateX = parseFloat(translateMatch[1]);
            translateY = parseFloat(translateMatch[2]);
        }
    }

    element.style.transform = `scale(${scale}) translate(${translateX}px, ${translateY}px)`;
    element.style.transformOrigin = 'center center';

    // Update cursor based on zoom level
    element.style.cursor = scale > 1 ? 'grab' : 'default';
}

// Add mouse wheel zoom support
export function addWheelZoomSupport(zoomableContainer) {
    zoomableContainer.addEventListener('wheel', (e) => {
        if (e.ctrlKey || e.metaKey) {
            e.preventDefault();

            const contentElement = zoomableContainer.querySelector('img') || zoomableContainer.querySelector('.mermaid-container');
            if (!contentElement) return;

            const currentScale = getZoomScale(contentElement);
            let newScale;

            if (e.deltaY < 0) {
                // Zoom in
                newScale = Math.min(currentScale * 1.1, 3);
            } else {
                // Zoom out
                newScale = Math.max(currentScale / 1.1, 0.1);
            }

            setZoomScale(contentElement, newScale);
        }
    });

    // Add drag/pan functionality
    addDragSupport(zoomableContainer);
}

// Add drag/pan support for zoomed images
function addDragSupport(zoomableContainer) {
    let isDragging = false;
    let startX = 0;
    let startY = 0;
    let startTranslateX = 0;
    let startTranslateY = 0;

    // Get the actual content element (img or mermaid-container)
    const getContentElement = () => {
        return zoomableContainer.querySelector('img') || zoomableContainer.querySelector('.mermaid-container');
    };

    zoomableContainer.addEventListener('mousedown', (e) => {
        // Only enable dragging if the image is zoomed (scale > 1)
        const contentElement = getContentElement();
        if (!contentElement) return;

        const currentScale = getZoomScale(contentElement);
        if (currentScale <= 1) return;

        isDragging = true;
        startX = e.clientX;
        startY = e.clientY;

        const currentTranslate = getTranslateValues(contentElement);
        startTranslateX = currentTranslate.x;
        startTranslateY = currentTranslate.y;

        // Change cursor to grabbing
        zoomableContainer.style.cursor = 'grabbing';

        // Prevent default to avoid text selection and image dragging
        e.preventDefault();
    });

    // Use document for mousemove and mouseup to handle mouse leaving container
    document.addEventListener('mousemove', (e) => {
        if (!isDragging) return;

        e.preventDefault();

        const contentElement = getContentElement();
        if (!contentElement) return;

        const deltaX = e.clientX - startX;
        const deltaY = e.clientY - startY;

        // Apply the movement as translate transform
        const newTranslateX = startTranslateX + deltaX;
        const newTranslateY = startTranslateY + deltaY;

        setTranslateValues(contentElement, newTranslateX, newTranslateY);
    });

    document.addEventListener('mouseup', () => {
        if (!isDragging) return;

        isDragging = false;
        const contentElement = getContentElement();
        if (contentElement) {
            const currentScale = getZoomScale(contentElement);
            zoomableContainer.style.cursor = currentScale > 1 ? 'grab' : 'default';
        }
    });

    // Update cursor based on zoom level
    const observer = new MutationObserver(() => {
        const contentElement = getContentElement();
        if (contentElement) {
            const currentScale = getZoomScale(contentElement);
            if (!isDragging) {
                zoomableContainer.style.cursor = currentScale > 1 ? 'grab' : 'default';
            }
        }
    });

    observer.observe(zoomableContainer, {
        attributes: true,
        attributeFilter: ['style'],
        subtree: true
    });
}

// Save zoom state from current container
export function saveZoomState(container) {
    const zoomableContainer = container.querySelector('.zoomable-container');
    if (!zoomableContainer) return null;

    const contentElement = zoomableContainer.querySelector('img') || zoomableContainer.querySelector('.mermaid-container');
    if (!contentElement) return null;

    const scale = getZoomScale(contentElement);
    const translate = getTranslateValues(contentElement);

    // Determine content type
    let type = null;
    if (container.querySelector('.preview-uml')) {
        type = 'uml';
    } else if (container.querySelector('.preview-mermaid')) {
        type = 'mermaid';
    }

    // Only save if there's meaningful zoom state (not default)
    if (scale !== 1 || translate.x !== 0 || translate.y !== 0) {
        return {
            type: type,
            scale: scale,
            translateX: translate.x,
            translateY: translate.y
        };
    }

    return null;
}

// Restore zoom state to a zoomable container
export function restoreZoomState(zoomableContainer, zoomState) {
    if (!zoomState || !zoomableContainer) return;

    const contentElement = zoomableContainer.querySelector('img') || zoomableContainer.querySelector('.mermaid-container');
    if (!contentElement) return;

    // Apply the saved zoom state
    contentElement.style.transform = `scale(${zoomState.scale}) translate(${zoomState.translateX}px, ${zoomState.translateY}px)`;
    contentElement.style.transformOrigin = 'center center';

    // Update cursor based on zoom level
    zoomableContainer.style.cursor = zoomState.scale > 1 ? 'grab' : 'default';
}

// Make zoom functions available globally for onclick handlers
window.zoomIn = zoomIn;
window.zoomOut = zoomOut;
window.zoomReset = zoomReset; 