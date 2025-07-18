import { state, updateUrlQuery, isEditableFile } from './state.js';
import { setupEditorMode, setupPreviewOnlyMode } from './editor.js';
import { loadPreview } from './preview.js';

// Load and render the directory tree
export async function loadTree() {
    try {
        const response = await fetch('/api/tree');
        if (!response.ok) {
            throw new Error(`Failed to load tree: ${response.statusText}`);
        }
        state.fileTree = await response.json();
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
    if (state.fileTree) {
        container.appendChild(createTreeNode(state.fileTree));
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
    state.selectedFile = filePath;

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

// Find and select file in tree
export function findAndSelectFile(filePath) {
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