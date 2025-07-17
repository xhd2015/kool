let fileTree = null;
let selectedFile = null;
let terminal = null;
let fitAddon = null;
let terminalVisible = false;
let terminalWebSocket = null;

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

    // Load file preview
    loadPreview(filePath);
}

// Load file preview
async function loadPreview(filePath) {
    const container = document.getElementById('preview-container');
    container.innerHTML = '<div class="loading">Loading preview...</div>';

    try {
        const response = await fetch(`/api/preview?path=${encodeURIComponent(filePath)}`);
        if (!response.ok) {
            throw new Error(`Failed to load preview: ${response.statusText}`);
        }

        const data = await response.json();

        if (data.type === 'uml') {
            renderUMLPreview(data.content, container);
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
        const url = `https://www.plantuml.com/plantuml/svg/${encoded}`;

        container.innerHTML = `
            <div class="preview-uml">
                <img src="${url}" alt="UML Diagram" onerror="handleUMLError(this)" />
            </div>
        `;
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

// Handle UML image load errors
function handleUMLError(img) {
    img.parentElement.innerHTML = '<div class="error">Failed to load UML diagram. The syntax might be invalid.</div>';
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

// Initialize the application
window.addEventListener('load', () => {
    loadTree();
    initializeResizer();
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

// Clean up WebSocket on page unload
window.addEventListener('beforeunload', () => {
    if (terminalWebSocket) {
        terminalWebSocket.close();
    }
});