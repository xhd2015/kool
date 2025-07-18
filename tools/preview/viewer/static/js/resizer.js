import { state } from './state.js';

// Initialize resizer
export function initializeResizer() {
    const resizer = document.querySelector('.resizer');
    if (!resizer) return;

    let isResizing = false;

    resizer.addEventListener('mousedown', (e) => {
        isResizing = true;
        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
    });

    function handleMouseMove(e) {
        if (!isResizing) return;

        const container = document.querySelector('.container');
        const sidebar = document.querySelector('.sidebar');
        const content = document.querySelector('.content');

        const containerRect = container.getBoundingClientRect();
        const relativeX = e.clientX - containerRect.left;

        const minWidth = 200;
        const maxSidebarWidth = containerRect.width - minWidth;
        const sidebarWidth = Math.max(minWidth, Math.min(maxSidebarWidth, relativeX));

        sidebar.style.width = sidebarWidth + 'px';
        content.style.width = (containerRect.width - sidebarWidth) + 'px';
    }

    function handleMouseUp() {
        isResizing = false;
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
    }
}

// Initialize horizontal resizer for editor/preview split
export function initializeHorizontalResizer() {
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

// Send terminal size to backend (imported from terminal.js logic)
function sendTerminalSize() {
    if (state.terminalWebSocket && state.terminalWebSocket.readyState === WebSocket.OPEN && state.terminal) {
        const cols = state.terminal.cols;
        const rows = state.terminal.rows;
        console.log('Sending terminal size from resizer:', { cols, rows });

        const message = JSON.stringify({
            resize: {
                cols: cols,
                rows: rows
            }
        });
        state.terminalWebSocket.send(message);
    }
}

// Initialize vertical resizer
export function initializeVerticalResizer() {
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

            if (state.fitAddon && state.terminalVisible) {
                setTimeout(() => {
                    state.fitAddon.fit();
                    // Send terminal size after fitting
                    sendTerminalSize();
                }, 50);
            }
        }
    }

    function handleMouseUp() {
        isResizing = false;
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
    }
} 