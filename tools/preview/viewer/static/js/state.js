// Global state management
export const state = {
    fileTree: null,
    selectedFile: null,
    terminal: null,
    fitAddon: null,
    terminalVisible: false,
    terminalWebSocket: null,
    currentFileContent: null,
    isEditorMode: false,
    previewUpdateTimeout: null,
    autoSaveTimeout: null
};

// Check if a file is editable
export function isEditableFile(filePath) {
    const editableExtensions = ['.md', '.uml', '.puml', '.mmd', '.txt', '.json', '.yaml', '.yml'];
    const ext = filePath.toLowerCase().substring(filePath.lastIndexOf('.'));
    return editableExtensions.includes(ext);
}

// Update URL query parameter
export function updateUrlQuery(filePath) {
    const url = new URL(window.location);
    if (filePath) {
        url.searchParams.set('file', filePath);
    } else {
        url.searchParams.delete('file');
    }
    window.history.pushState({}, '', url);
}

// Get file path from URL query parameter
export function getFileFromUrl() {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get('file');
} 