import { state, getFileFromUrl } from './state.js';
import { loadTree, findAndSelectFile } from './tree.js';
import { initializeResizer, initializeHorizontalResizer, initializeVerticalResizer } from './resizer.js';
import { initializeTheme } from './theme.js';
import { initializeTerminal, initializeTerminalToggle, cleanupTerminal } from './terminal.js';

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
async function initializeApp() {
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
}

// Handle browser back/forward navigation
function handlePopState() {
    const fileFromUrl = getFileFromUrl();
    if (fileFromUrl !== state.selectedFile) {
        findAndSelectFile(fileFromUrl);
    }
}

// Initialize the application when DOM is loaded
window.addEventListener('load', initializeApp);

// Handle browser navigation
window.addEventListener('popstate', handlePopState);

// Clean up WebSocket on page unload
window.addEventListener('beforeunload', cleanupTerminal); 