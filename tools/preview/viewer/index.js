let fileTree = null;
let selectedFile = null;
let terminal = null;
let fitAddon = null;
let terminalVisible = false;
let terminalWebSocket = null;
let currentFileContent = null;
let isEditorMode = false;
let previewUpdateTimeout = null;
let autoSaveTimeout = null;

// Check if a file is editable
function isEditableFile(filePath) {
    const editableExtensions = ['.md', '.uml', '.puml', '.mmd', '.txt', '.json', '.yaml', '.yml'];
    const ext = filePath.toLowerCase().substring(filePath.lastIndexOf('.'));
    return editableExtensions.includes(ext);
}

// Load and render the directory tree
async function loadTree() {
    try {
        const response = await fetch('/api/tree');
        if (!response.ok) {
            throw new Error(`Failed to load tree: ${response.statusText}`);
        }
        fileTree = await response.json();
        renderTree();
    } catch (error) {
        document.getElementById('tree-container').innerHTML =
            `<div class="error">Failed to load directory tree: ${error.message}</div>`;
    }
}



// Render the directory tree
function renderTree() {
    const container = document.getElementById('tree-container');
    container.innerHTML = '';
    if (fileTree) {
        container.appendChild(createTreeNode(fileTree));
    }
}

// Create a tree node element
function createTreeNode(node) {
    const nodeElement = document.createElement('div');

    const labelElement = document.createElement('div');
    labelElement.className = `tree-node ${node.isDir ? 'directory' : 'file'}`;
    labelElement.setAttribute('data-path', node.path);

    let content = '';

    if (node.isDir && node.children && node.children.length > 0) {
        content += '<span class="toggle">▼</span>';
    } else if (node.isDir) {
        content += '<span class="toggle">▷</span>';
    } else {
        content += '<span class="toggle"></span>';
    }

    if (node.isDir) {
        content += '<span class="icon">📁</span>';
    } else {
        content += '<span class="icon">📄</span>';
    }

    content += `<span class="name">${node.name}</span>`;
    labelElement.innerHTML = content;

    // Add click handlers
    if (node.isDir) {
        const toggle = labelElement.querySelector('.toggle');
        if (toggle) {
            toggle.addEventListener('click', (e) => {
                e.stopPropagation();
                toggleDirectory(nodeElement);
            });
        }
        labelElement.addEventListener('click', () => {
            toggleDirectory(nodeElement);
        });
    } else {
        labelElement.addEventListener('click', () => {
            selectFile(labelElement, node.path);
        });
    }

    nodeElement.appendChild(labelElement);

    // Add children
    if (node.isDir && node.children && node.children.length > 0) {
        const childrenContainer = document.createElement('div');
        childrenContainer.className = 'tree-children';

        node.children.forEach(child => {
            childrenContainer.appendChild(createTreeNode(child));
        });

        nodeElement.appendChild(childrenContainer);
    }

    return nodeElement;
}

// Toggle directory expansion
function toggleDirectory(nodeElement) {
    const childrenContainer = nodeElement.querySelector('.tree-children');
    const toggle = nodeElement.querySelector('.toggle');

    if (childrenContainer) {
        const isCollapsed = childrenContainer.classList.contains('collapsed');
        if (isCollapsed) {
            childrenContainer.classList.remove('collapsed');
            toggle.textContent = '▼';
        } else {
            childrenContainer.classList.add('collapsed');
            toggle.textContent = '▷';
        }
    }
}

// Select a file and load its preview
function selectFile(element, filePath) {
    // Remove previous selection
    document.querySelectorAll('.tree-node.selected').forEach(node => {
        node.classList.remove('selected');
    });

    // Add selection to current element
    element.classList.add('selected');
    selectedFile = filePath;

    // Update content title
    document.getElementById('content-title').textContent = filePath.split('/').pop();

    // Update URL query parameter
    updateUrlQuery(filePath);

    // Check if file is editable and setup editor mode
    if (isEditableFile(filePath)) {
        setupEditorMode(filePath);
    } else {
        setupPreviewOnlyMode();
        loadPreview(filePath);
    }
}

// Update URL query parameter
function updateUrlQuery(filePath) {
    const url = new URL(window.location);
    if (filePath) {
        url.searchParams.set('file', filePath);
    } else {
        url.searchParams.delete('file');
    }
    window.history.pushState({}, '', url);
}

// Setup editor mode for editable files
async function setupEditorMode(filePath) {
    isEditorMode = true;

    // Show editor components
    document.getElementById('editor-section').style.display = 'flex';
    document.getElementById('horizontal-resizer').style.display = 'block';

    // Load file content into editor
    await loadFileContent(filePath);

    // Load initial preview
    loadPreview(filePath);
}

// Setup preview-only mode for non-editable files
function setupPreviewOnlyMode() {
    isEditorMode = false;

    // Hide editor components
    document.getElementById('editor-section').style.display = 'none';
    document.getElementById('horizontal-resizer').style.display = 'none';
}

// Load file content into editor
async function loadFileContent(filePath) {
    try {
        const response = await fetch(`/api/content?path=${encodeURIComponent(filePath)}`);
        if (!response.ok) {
            throw new Error(`Failed to load file content: ${response.statusText}`);
        }

        const data = await response.json();
        currentFileContent = data.content;

        const editor = document.getElementById('editor-textarea');
        editor.value = currentFileContent;

        // Setup real-time preview updates
        setupEditorEventListeners();

        // Initialize save button state
        updateSaveButtonState();

    } catch (error) {
        console.error('Failed to load file content:', error);
        document.getElementById('editor-textarea').value = 'Error loading file content';
    }
}

// Load file preview
async function loadPreview(filePath, customContent = null) {
    const container = document.getElementById('preview-container');
    container.innerHTML = '<div class="loading">Loading preview...</div>';

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
            renderUMLPreview(data.content, container);
        } else if (data.type === 'mermaid') {
            renderMermaidPreview(data.content, container);
        } else if (data.type === 'markdown') {
            renderMarkdownPreview(data.content, container);
        } else {
            renderTextPreview(data.content, container);
        }
    } catch (error) {
        container.innerHTML = `<div class="error">Failed to load preview: ${error.message}</div>`;
    }
}

// Render UML preview
function renderUMLPreview(content, container) {
    try {
        const encoded = plantumlEncoder.encode(content);
        const url = `/planuml/svg/${encoded}`;

        // Show loading state
        container.innerHTML = `
            <div class="preview-uml">
                <div class="uml-loading">
                    <div class="loading-spinner"></div>
                    <div class="loading-text">Generating UML diagram...</div>
                </div>
            </div>
        `;

        // Create image element and handle loading
        const img = new Image();

        img.onload = function () {
            // Image loaded successfully, replace loading with image
            container.innerHTML = `
                <div class="preview-uml">
                    <img src="${url}" alt="UML Diagram" />
                </div>
            `;
        };

        img.onerror = function () {
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
function renderTextPreview(content, container) {
    const textarea = document.createElement('textarea');
    textarea.className = 'preview-text';
    textarea.value = content;
    textarea.readOnly = true;

    container.innerHTML = '';
    container.appendChild(textarea);
}

// Render Mermaid preview
function renderMermaidPreview(content, container) {
    try {
        // Show loading state
        container.innerHTML = `
            <div class="preview-mermaid">
                <div class="mermaid-loading">
                    <div class="loading-spinner"></div>
                    <div class="loading-text">Generating Mermaid diagram...</div>
                </div>
            </div>
        `;

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
                container.innerHTML = `
                    <div class="preview-mermaid">
                        <div class="mermaid-container">${svg}</div>
                    </div>
                `;
            }).catch(error => {
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
            container.innerHTML = `<div class="error">Mermaid library not loaded</div>`;
        }
    } catch (error) {
        container.innerHTML = `<div class="error">Failed to render Mermaid: ${error.message}</div>`;
    }
}

// Render Markdown preview
function renderMarkdownPreview(content, container) {
    try {
        // Show loading state
        container.innerHTML = `
            <div class="preview-markdown">
                <div class="markdown-loading">
                    <div class="loading-spinner"></div>
                    <div class="loading-text">Rendering Markdown...</div>
                </div>
            </div>
        `;

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

            container.innerHTML = `
                <div class="preview-markdown">
                    <div class="markdown-content">${htmlContent}</div>
                </div>
            `;
        } else {
            container.innerHTML = `<div class="error">Marked library not loaded</div>`;
        }
    } catch (error) {
        container.innerHTML = `<div class="error">Failed to render Markdown: ${error.message}</div>`;
    }
}

// Handle UML image load errors - updated to work with container
function handleUMLError(container, content) {
    container.innerHTML = `
        <div class="preview-uml">
            <div class="uml-error-compact">
                ⚠️ Failed to load UML diagram. Syntax might be invalid, or https://www.plantuml.com/plantuml/svg is overloaded.
                <button class="retry-button" onclick="retryUMLRender(this, '${content.replace(/'/g, "\\'")}')">Retry</button>
            </div>
        </div>
    `;
}

// Retry UML rendering
function retryUMLRender(button, content) {
    const container = button.closest('.preview-uml').parentElement;
    renderUMLPreview(content, container);
}

// Setup editor event listeners
function setupEditorEventListeners() {
    const editor = document.getElementById('editor-textarea');
    const saveButton = document.getElementById('save-button');

    // Real-time preview update and auto-save with debouncing
    editor.addEventListener('input', () => {
        // Clear existing timeouts
        clearTimeout(previewUpdateTimeout);
        clearTimeout(autoSaveTimeout);

        // Update preview with 500ms debounce
        previewUpdateTimeout = setTimeout(() => {
            if (isEditorMode && selectedFile) {
                loadPreview(selectedFile, editor.value);
            }
        }, 500);

        // Auto-save with 500ms throttle
        autoSaveTimeout = setTimeout(() => {
            if (isEditorMode && selectedFile && editor.value !== currentFileContent) {
                autoSaveFile(editor.value);
            }
        }, 500);

        // Update save button state
        updateSaveButtonState();
    });

    // Manual save functionality (kept for explicit saves)
    saveButton.addEventListener('click', async () => {
        if (!selectedFile) return;
        await saveFile(editor.value, true); // true = manual save
    });

    // Initially disable save button
    saveButton.disabled = true;
}

// Auto-save function
async function autoSaveFile(content) {
    if (!selectedFile) return;

    try {
        const response = await fetch('/api/save', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                path: selectedFile,
                content: content
            })
        });

        if (!response.ok) {
            throw new Error(`Failed to auto-save: ${response.statusText}`);
        }

        currentFileContent = content;
        showSaveStatus('Auto-saved', 'success', 1500);

    } catch (error) {
        console.error('Failed to auto-save file:', error);
        showSaveStatus('Auto-save failed', 'error', 2500);
    } finally {
        updateSaveButtonState();
    }
}

// Manual save function
async function saveFile(content, isManual = false) {
    if (!selectedFile) return;

    const saveButton = document.getElementById('save-button');

    try {
        if (isManual) {
            saveButton.disabled = true;
            showSaveStatus('Saving...', 'default', 10000); // Long duration, will be cleared on completion
        }

        const response = await fetch('/api/save', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                path: selectedFile,
                content: content
            })
        });

        if (!response.ok) {
            throw new Error(`Failed to save: ${response.statusText}`);
        }

        currentFileContent = content;

        if (isManual) {
            showSaveStatus('Saved', 'success', 2000);
        }

        updateSaveButtonState();

    } catch (error) {
        console.error('Failed to save file:', error);
        if (isManual) {
            showSaveStatus('Save failed', 'error', 3000);
        }
        updateSaveButtonState();
    }
}

// Update save button state based on content changes
function updateSaveButtonState() {
    const editor = document.getElementById('editor-textarea');
    const saveButton = document.getElementById('save-button');

    if (editor.value === currentFileContent) {
        saveButton.disabled = true;
    } else {
        saveButton.disabled = false;
    }

    // Keep button text simple
    saveButton.textContent = 'Save';
}

// Show subtle status indicator
function showSaveStatus(message, type = 'default', duration = 2000) {
    const statusElement = document.getElementById('save-status');

    // Remove existing classes
    statusElement.classList.remove('visible', 'success', 'error');

    // Set message and type
    statusElement.textContent = message;
    if (type === 'success') {
        statusElement.classList.add('success');
    } else if (type === 'error') {
        statusElement.classList.add('error');
    }

    // Show status
    statusElement.classList.add('visible');

    // Hide after duration
    setTimeout(() => {
        statusElement.classList.remove('visible');
    }, duration);
}

// Initialize sidebar resizing
function initializeResizer() {
    const resizer = document.querySelector('.resizer');
    const sidebar = document.querySelector('.sidebar');
    let isResizing = false;

    resizer.addEventListener('mousedown', (e) => {
        isResizing = true;
        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
    });

    function handleMouseMove(e) {
        if (!isResizing) return;

        const newWidth = e.clientX;
        if (newWidth > 200 && newWidth < 600) {
            sidebar.style.width = newWidth + 'px';
        }
    }

    function handleMouseUp() {
        isResizing = false;
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
    }
}

// Initialize horizontal resizer for editor/preview split
function initializeHorizontalResizer() {
    const resizer = document.querySelector('.horizontal-resizer');
    const editorSection = document.querySelector('.editor-section');
    const previewWrapper = document.querySelector('.preview-container-wrapper');
    let isResizing = false;

    resizer.addEventListener('mousedown', (e) => {
        isResizing = true;
        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
    });

    function handleMouseMove(e) {
        if (!isResizing) return;

        const container = document.querySelector('.preview-section');
        const containerRect = container.getBoundingClientRect();
        const relativeX = e.clientX - containerRect.left;
        const containerWidth = containerRect.width;

        const minWidth = 300;
        const maxEditorWidth = containerWidth - minWidth - 4; // 4px for resizer
        const editorWidth = Math.max(minWidth, Math.min(maxEditorWidth, relativeX - 4));
        const previewWidth = containerWidth - editorWidth - 4;

        if (previewWidth >= minWidth) {
            editorSection.style.flex = 'none';
            editorSection.style.width = editorWidth + 'px';
            previewWrapper.style.flex = 'none';
            previewWrapper.style.width = previewWidth + 'px';
        }
    }

    function handleMouseUp() {
        isResizing = false;
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
    }
}

// Theme management
function initializeTheme() {
    const themeToggle = document.getElementById('theme-toggle');
    const savedTheme = localStorage.getItem('theme');

    // Default to light theme
    if (savedTheme === 'dark') {
        document.body.classList.add('dark-theme');
        themeToggle.textContent = 'Light';
    } else {
        themeToggle.textContent = 'Dark';
    }

    themeToggle.addEventListener('click', () => {
        document.body.classList.toggle('dark-theme');
        const isDark = document.body.classList.contains('dark-theme');

        if (isDark) {
            themeToggle.textContent = 'Light';
            localStorage.setItem('theme', 'dark');
        } else {
            themeToggle.textContent = 'Dark';
            localStorage.setItem('theme', 'light');
        }
    });
}

// Initialize terminal with WebSocket streaming
function initializeTerminal() {
    if (typeof Terminal === 'undefined' || typeof FitAddon === 'undefined') {
        console.error('xterm libraries not loaded');
        return;
    }

    terminal = new Terminal({
        cursorBlink: true,
        theme: {
            background: '#1e1e1e',
            foreground: '#d4d4d4',
            cursor: '#ffffff',
            selection: '#264f78'
        },
        fontSize: 14,
        fontFamily: 'Consolas, Monaco, "Courier New", monospace'
    });

    fitAddon = new FitAddon.FitAddon();
    terminal.loadAddon(fitAddon);

    const terminalContainer = document.getElementById('terminal-container');
    terminal.open(terminalContainer);
    fitAddon.fit();

    // Handle terminal input - send each character directly to bash
    terminal.onData((data) => {
        console.log('Terminal input data:', JSON.stringify(data));

        // Don't echo locally - let bash handle all display
        // Just send the character directly to the bash session
        sendInput(data);
    });

    // Handle window resize
    window.addEventListener('resize', () => {
        if (fitAddon && terminalVisible) {
            fitAddon.fit();
        }
    });

    // Don't initialize WebSocket here - wait until terminal is shown
    // initializeTerminalWebSocket();
}

// Initialize WebSocket for terminal streaming
function initializeTerminalWebSocket() {
    if (terminalWebSocket) {
        console.log('Closing existing WebSocket...');
        terminalWebSocket.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/terminal/stream`;

    console.log('Connecting to WebSocket:', wsUrl);
    console.log('Current location:', window.location.href);

    terminalWebSocket = new WebSocket(wsUrl);

    terminalWebSocket.onopen = function (event) {
        console.log('Terminal WebSocket connected successfully');
    };

    terminalWebSocket.onmessage = function (event) {
        console.log('WebSocket message received:', event.data);
        try {
            const data = JSON.parse(event.data);
            console.log('Parsed WebSocket data:', data);

            if (data.output) {
                console.log('Writing output to terminal:', JSON.stringify(data.output));
                terminal.write(data.output);
            }
            if (data.error) {
                console.log('Writing error to terminal:', JSON.stringify(data.error));
                terminal.write(`\x1b[31m${data.error}\x1b[0m`);
            }
            if (data.keepalive) {
                console.log('Received keepalive');
                // Just a keepalive message, ignore
            }
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    };

    terminalWebSocket.onerror = function (event) {
        console.error('Terminal WebSocket error:', event);
    };

    terminalWebSocket.onclose = function (event) {
        console.log('Terminal WebSocket closed, code:', event.code, 'reason:', event.reason);
        console.log('Was clean close:', event.wasClean);

        // Try to reconnect after a delay only if terminal is still visible
        if (terminalVisible) {
            setTimeout(() => {
                console.log('Attempting to reconnect WebSocket...');
                initializeTerminalWebSocket();
            }, 5000);
        }
    };
}

// Send input to terminal via WebSocket
function sendInput(input) {
    console.log('Sending input to WebSocket:', JSON.stringify(input));
    if (terminalWebSocket && terminalWebSocket.readyState === WebSocket.OPEN) {
        const message = JSON.stringify({ input: input });
        console.log('WebSocket sending message:', message);
        terminalWebSocket.send(message);
    } else {
        console.log('WebSocket not ready, state:', terminalWebSocket ? terminalWebSocket.readyState : 'null');
    }
}

// Initialize terminal toggle
function initializeTerminalToggle() {
    const toggleButton = document.getElementById('terminal-toggle');
    const terminalContainer = document.getElementById('terminal-container');
    const verticalResizer = document.querySelector('.vertical-resizer');
    const previewSection = document.querySelector('.preview-section');

    toggleButton.addEventListener('click', () => {
        terminalVisible = !terminalVisible;
        console.log('Terminal visibility toggled to:', terminalVisible);

        if (terminalVisible) {
            terminalContainer.classList.remove('hidden');
            verticalResizer.style.display = 'block';
            toggleButton.textContent = 'Hide Terminal';

            // Reset to split layout (80% preview, 20% terminal)
            previewSection.style.flex = '4';
            previewSection.style.height = '';
            terminalContainer.style.flex = '1';

            if (fitAddon) {
                setTimeout(() => fitAddon.fit(), 100);
            }

            // Start terminal streaming when terminal becomes visible
            if (!terminalWebSocket || terminalWebSocket.readyState === WebSocket.CLOSED) {
                console.log('Starting WebSocket connection...');
                initializeTerminalWebSocket();
            } else {
                console.log('WebSocket already connected, state:', terminalWebSocket.readyState);
            }
        } else {
            terminalContainer.classList.add('hidden');
            verticalResizer.style.display = 'none';
            toggleButton.textContent = 'Show Terminal';

            // Make preview take more space
            previewSection.style.flex = '1';
            previewSection.style.height = '';

            // Close terminal streaming when terminal is hidden
            if (terminalWebSocket) {
                console.log('Closing WebSocket connection...');
                terminalWebSocket.close();
                terminalWebSocket = null;
            }
        }
    });
}

// Initialize vertical resizer
function initializeVerticalResizer() {
    const resizer = document.querySelector('.vertical-resizer');
    const previewSection = document.querySelector('.preview-section');
    const terminalContainer = document.getElementById('terminal-container');
    let isResizing = false;

    resizer.addEventListener('mousedown', (e) => {
        isResizing = true;
        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
    });

    function handleMouseMove(e) {
        if (!isResizing) return;

        const container = document.querySelector('.content-body');
        const containerRect = container.getBoundingClientRect();
        const relativeY = e.clientY - containerRect.top;
        const containerHeight = containerRect.height;
        const terminalHeaderHeight = 40; // Approximate height of terminal header

        const minTerminalHeight = Math.max(200, containerHeight * 0.2); // At least 20% or 200px
        const maxPreviewHeight = containerHeight - minTerminalHeight - 4 - terminalHeaderHeight;
        const previewHeight = Math.max(200, Math.min(maxPreviewHeight, relativeY - 4));
        const terminalHeight = containerHeight - previewHeight - 4 - terminalHeaderHeight; // 4px for resizer, subtract header height

        if (terminalHeight >= minTerminalHeight) {
            previewSection.style.flex = 'none';
            previewSection.style.height = previewHeight + 'px';
            terminalContainer.style.flex = 'none';
            terminalContainer.style.height = terminalHeight + 'px';

            if (fitAddon && terminalVisible) {
                setTimeout(() => fitAddon.fit(), 50);
            }
        }
    }

    function handleMouseUp() {
        isResizing = false;
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
    }
}

// Get file path from URL query parameter
function getFileFromUrl() {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get('file');
}



// Find and select file in tree
function findAndSelectFile(filePath) {
    if (!filePath) return;

    // Find the tree node with the matching path
    // First try exact match
    let treeNode = document.querySelector(`[data-path="${filePath}"]`);

    // If not found by exact match, try to find by relative path matching
    if (!treeNode) {
        const allNodes = document.querySelectorAll('[data-path]');
        for (const node of allNodes) {
            const nodePath = node.getAttribute('data-path');
            if (nodePath.endsWith(filePath) || nodePath.endsWith('/' + filePath)) {
                treeNode = node;
                break;
            }
        }
    }

    if (treeNode) {
        // Expand parent directories if needed
        expandParentDirectories(treeNode);

        // Select the file using the actual absolute path from the tree node
        const actualPath = treeNode.getAttribute('data-path');
        selectFile(treeNode, actualPath);

        // Scroll to the selected file
        treeNode.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
}

// Expand parent directories to make a file visible
function expandParentDirectories(fileNode) {
    let parent = fileNode.parentElement;
    while (parent) {
        if (parent.classList.contains('tree-children') && parent.classList.contains('collapsed')) {
            parent.classList.remove('collapsed');
            // Update the toggle icon
            const parentNode = parent.previousElementSibling;
            if (parentNode) {
                const toggle = parentNode.querySelector('.toggle');
                if (toggle) {
                    toggle.textContent = '▼';
                }
            }
        }
        parent = parent.parentElement;
    }
}

// Initialize file selection after tree is loaded
async function initializeFileSelection() {
    // Get file from URL query parameter
    const fileToSelect = getFileFromUrl();

    if (fileToSelect) {
        // Wait a bit for the tree to be fully rendered
        setTimeout(() => {
            findAndSelectFile(fileToSelect);
        }, 100);
    }
}

// Initialize the application
window.addEventListener('load', async () => {
    await loadTree();
    await initializeFileSelection();

    initializeResizer();
    initializeHorizontalResizer();
    initializeTheme();
    initializeTerminal();
    initializeTerminalToggle();
    initializeVerticalResizer();

    // Initialize layout - hide terminal container and resizer since terminal starts hidden
    const verticalResizer = document.querySelector('.vertical-resizer');
    const previewSection = document.querySelector('.preview-section');
    const terminalContainer = document.getElementById('terminal-container');

    if (verticalResizer && previewSection && terminalContainer) {
        verticalResizer.style.display = 'none';
        previewSection.style.flex = '1';
        terminalContainer.classList.add('hidden');
    }
});

// Handle browser back/forward navigation
window.addEventListener('popstate', () => {
    const fileFromUrl = getFileFromUrl();
    if (fileFromUrl !== selectedFile) {
        findAndSelectFile(fileFromUrl);
    }
});

// Clean up WebSocket on page unload
window.addEventListener('beforeunload', () => {
    if (terminalWebSocket) {
        terminalWebSocket.close();
    }
});