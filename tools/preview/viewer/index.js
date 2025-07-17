let fileTree = null;
let selectedFile = null;

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

// Initialize the application
window.addEventListener('load', () => {
    loadTree();
    initializeResizer();
    initializeTheme();
});