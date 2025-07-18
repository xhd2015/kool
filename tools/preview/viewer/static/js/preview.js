import { addWheelZoomSupport, saveZoomState, restoreZoomState } from './zoom.js';

// Show subtle loading indicator overlay
function showLoadingIndicator(container) {
    // Remove existing indicator if any
    hideLoadingIndicator(container);

    const indicator = document.createElement('div');
    indicator.className = 'loading-indicator-overlay';
    indicator.innerHTML = `
        <div class="loading-indicator">
            <div class="loading-spinner-small"></div>
        </div>
    `;

    container.appendChild(indicator);
}

// Hide loading indicator overlay
function hideLoadingIndicator(container) {
    const indicator = container.querySelector('.loading-indicator-overlay');
    if (indicator) {
        indicator.remove();
    }
}

// Load file preview
export async function loadPreview(filePath, customContent = null) {
    const container = document.getElementById('preview-container');

    // Save current zoom state if there's zoomable content
    const currentZoomState = saveZoomState(container);

    // Check if this is a refresh (has existing content) or initial load
    const isRefresh = container.children.length > 0 && !container.querySelector('.loading') && !container.querySelector('.empty-state');

    if (isRefresh) {
        // Show subtle loading indicator overlay instead of replacing content
        showLoadingIndicator(container);
    } else {
        // Initial load - show full loading state
        container.innerHTML = '<div class="loading">Loading preview...</div>';
    }

    try {
        let data;

        if (customContent !== null) {
            // Use custom content (from editor) instead of loading from server
            const ext = filePath.toLowerCase().substring(filePath.lastIndexOf('.'));
            let type = 'text';

            if (ext === '.md') {
                type = 'markdown';
            } else if (ext === '.uml' || ext === '.puml') {
                type = 'uml';
            } else if (ext === '.mmd') {
                type = 'mermaid';
            }

            data = { type, content: customContent };
        } else {
            // Load from server
            const response = await fetch(`/api/preview?path=${encodeURIComponent(filePath)}`);
            if (!response.ok) {
                throw new Error(`Failed to load preview: ${response.statusText}`);
            }
            data = await response.json();
        }

        if (data.type === 'uml') {
            renderUMLPreview(data.content, container, currentZoomState, isRefresh);
        } else if (data.type === 'mermaid') {
            renderMermaidPreview(data.content, container, currentZoomState, isRefresh);
        } else if (data.type === 'markdown') {
            renderMarkdownPreview(data.content, container, currentZoomState, isRefresh);
        } else {
            renderTextPreview(data.content, container, currentZoomState, isRefresh);
        }
    } catch (error) {
        hideLoadingIndicator(container);
        container.innerHTML = `<div class="error">Failed to load preview: ${error.message}</div>`;
    }
}

// Render UML preview
function renderUMLPreview(content, container, zoomState = null, isRefresh = false) {
    try {
        const encoded = plantumlEncoder.encode(content);
        const url = `/planuml/svg/${encoded}`;

        // Show loading state only if not a refresh
        if (!isRefresh) {
            container.innerHTML = `
                <div class="preview-uml">
                    <div class="uml-loading">
                        <div class="loading-spinner"></div>
                        <div class="loading-text">Generating UML diagram...</div>
                    </div>
                </div>
            `;
        }

        // Create image element and handle loading
        const img = new Image();

        img.onload = function () {
            // Hide loading indicator if it was a refresh
            if (isRefresh) {
                hideLoadingIndicator(container);
            }

            // Image loaded successfully, replace loading with image
            container.innerHTML = `
                <div class="preview-uml">
                    <div class="zoom-controls">
                        <button class="zoom-button zoom-in" onclick="window.zoomIn(this)" title="Zoom In">🔍+</button>
                        <button class="zoom-button zoom-out" onclick="window.zoomOut(this)" title="Zoom Out">🔍-</button>
                        <button class="zoom-button zoom-reset" onclick="window.zoomReset(this)" title="Reset Zoom">⚪</button>
                    </div>
                    <div class="zoomable-container">
                        <img src="${url}" alt="UML Diagram" />
                    </div>
                </div>
            `;

            // Add wheel zoom support
            const zoomableContainer = container.querySelector('.zoomable-container');
            if (zoomableContainer) {
                addWheelZoomSupport(zoomableContainer);

                // Restore zoom state if available and same content type
                if (zoomState && zoomState.type === 'uml') {
                    restoreZoomState(zoomableContainer, zoomState);
                }
            }
        };

        img.onerror = function () {
            // Hide loading indicator if it was a refresh
            if (isRefresh) {
                hideLoadingIndicator(container);
            }

            // Image failed to load, show error
            handleUMLError(container, content);
        };

        // Start loading the image
        img.src = url;

    } catch (error) {
        container.innerHTML = `<div class="error">Failed to render UML: ${error.message}</div>`;
    }
}

// Render text preview
function renderTextPreview(content, container, zoomState = null, isRefresh = false) {
    if (isRefresh) {
        hideLoadingIndicator(container);
    }

    const textarea = document.createElement('textarea');
    textarea.className = 'preview-text';
    textarea.value = content;
    textarea.readOnly = true;

    container.innerHTML = '';
    container.appendChild(textarea);
}

// Render Mermaid preview
function renderMermaidPreview(content, container, zoomState = null, isRefresh = false) {
    try {
        // Show loading state only if not a refresh
        if (!isRefresh) {
            container.innerHTML = `
                <div class="preview-mermaid">
                    <div class="mermaid-loading">
                        <div class="loading-spinner"></div>
                        <div class="loading-text">Generating Mermaid diagram...</div>
                    </div>
                </div>
            `;
        }

        // Initialize Mermaid if not already done
        if (typeof mermaid !== 'undefined') {
            mermaid.initialize({
                startOnLoad: false,
                theme: document.body.classList.contains('dark-theme') ? 'dark' : 'default',
                securityLevel: 'loose'
            });

            // Create a unique ID for the diagram
            const diagramId = 'mermaid-diagram-' + Date.now();

            // Create container for the diagram
            const diagramContainer = document.createElement('div');
            diagramContainer.innerHTML = `<div id="${diagramId}" class="mermaid-diagram">${content}</div>`;

            // Render the diagram
            mermaid.render(diagramId + '-svg', content).then(({ svg }) => {
                // Hide loading indicator if it was a refresh
                if (isRefresh) {
                    hideLoadingIndicator(container);
                }

                container.innerHTML = `
                    <div class="preview-mermaid">
                        <div class="zoom-controls">
                            <button class="zoom-button zoom-in" onclick="window.zoomIn(this)" title="Zoom In">🔍+</button>
                            <button class="zoom-button zoom-out" onclick="window.zoomOut(this)" title="Zoom Out">🔍-</button>
                            <button class="zoom-button zoom-reset" onclick="window.zoomReset(this)" title="Reset Zoom">⚪</button>
                        </div>
                        <div class="zoomable-container">
                            <div class="mermaid-container">${svg}</div>
                        </div>
                    </div>
                `;

                // Add wheel zoom support
                const zoomableContainer = container.querySelector('.zoomable-container');
                if (zoomableContainer) {
                    addWheelZoomSupport(zoomableContainer);

                    // Restore zoom state if available and same content type
                    if (zoomState && zoomState.type === 'mermaid') {
                        restoreZoomState(zoomableContainer, zoomState);
                    }
                }
            }).catch(error => {
                // Hide loading indicator if it was a refresh
                if (isRefresh) {
                    hideLoadingIndicator(container);
                }

                container.innerHTML = `
                    <div class="preview-mermaid">
                        <div class="mermaid-error">
                            ⚠️ Failed to render Mermaid diagram: ${error.message}
                            <details>
                                <summary>Mermaid content:</summary>
                                <pre>${content}</pre>
                            </details>
                        </div>
                    </div>
                `;
            });
        } else {
            if (isRefresh) {
                hideLoadingIndicator(container);
            }
            container.innerHTML = `<div class="error">Mermaid library not loaded</div>`;
        }
    } catch (error) {
        if (isRefresh) {
            hideLoadingIndicator(container);
        }
        container.innerHTML = `<div class="error">Failed to render Mermaid: ${error.message}</div>`;
    }
}

// Render Markdown preview
function renderMarkdownPreview(content, container, zoomState = null, isRefresh = false) {
    try {
        // Show loading state only if not a refresh
        if (!isRefresh) {
            container.innerHTML = `
                <div class="preview-markdown">
                    <div class="markdown-loading">
                        <div class="loading-spinner"></div>
                        <div class="loading-text">Rendering Markdown...</div>
                    </div>
                </div>
            `;
        }

        // Initialize marked if not already done
        if (typeof marked !== 'undefined') {
            // Configure marked options
            marked.setOptions({
                breaks: true,
                gfm: true,
                sanitize: false,
                smartLists: true,
                smartypants: false
            });

            // Render the markdown
            const htmlContent = marked.parse(content);

            // Hide loading indicator if it was a refresh
            if (isRefresh) {
                hideLoadingIndicator(container);
            }

            container.innerHTML = `
                <div class="preview-markdown">
                    <div class="markdown-content">${htmlContent}</div>
                </div>
            `;
        } else {
            if (isRefresh) {
                hideLoadingIndicator(container);
            }
            container.innerHTML = `<div class="error">Marked library not loaded</div>`;
        }
    } catch (error) {
        if (isRefresh) {
            hideLoadingIndicator(container);
        }
        container.innerHTML = `<div class="error">Failed to render Markdown: ${error.message}</div>`;
    }
}

// Handle UML image load errors - updated to work with container
function handleUMLError(container, content) {
    container.innerHTML = `
        <div class="preview-uml">
            <div class="uml-error-compact">
                ⚠️ Failed to load UML diagram. Syntax might be invalid, or https://www.plantuml.com/plantuml/svg is overloaded.
                <button class="retry-button" onclick="window.retryUMLRender(this, '${content.replace(/'/g, "\\'")}')">Retry</button>
            </div>
        </div>
    `;
}

// Retry UML rendering
window.retryUMLRender = function (button, content) {
    const container = button.closest('.preview-uml').parentElement;
    renderUMLPreview(content, container, null);
}; 