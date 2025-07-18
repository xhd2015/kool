import { state } from './state.js';
import { loadPreview } from './preview.js';

// Setup editor mode for editable files
export async function setupEditorMode(filePath) {
    state.isEditorMode = true;

    // Show editor components
    document.getElementById('editor-section').style.display = 'flex';
    document.getElementById('horizontal-resizer').style.display = 'block';

    // Load file content into editor
    await loadFileContent(filePath);

    // Load initial preview
    loadPreview(filePath);
}

// Setup preview-only mode for non-editable files
export function setupPreviewOnlyMode() {
    state.isEditorMode = false;

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
        state.currentFileContent = data.content;

        const editor = document.getElementById('editor-textarea');
        editor.value = state.currentFileContent;

        // Setup real-time preview updates
        setupEditorEventListeners();

        // Initialize save button state
        updateSaveButtonState();

    } catch (error) {
        console.error('Failed to load file content:', error);
        document.getElementById('editor-textarea').value = 'Error loading file content';
    }
}

// Setup editor event listeners
function setupEditorEventListeners() {
    const editor = document.getElementById('editor-textarea');
    const saveButton = document.getElementById('save-button');

    // Real-time preview update and auto-save with debouncing
    editor.addEventListener('input', () => {
        // Clear existing timeouts
        clearTimeout(state.previewUpdateTimeout);
        clearTimeout(state.autoSaveTimeout);

        // Update preview with 500ms debounce
        state.previewUpdateTimeout = setTimeout(() => {
            if (state.isEditorMode && state.selectedFile) {
                loadPreview(state.selectedFile, editor.value);
            }
        }, 500);

        // Auto-save with 500ms throttle
        state.autoSaveTimeout = setTimeout(() => {
            if (state.isEditorMode && state.selectedFile && editor.value !== state.currentFileContent) {
                autoSaveFile(editor.value);
            }
        }, 500);

        // Update save button state
        updateSaveButtonState();
    });

    // Manual save functionality (kept for explicit saves)
    saveButton.addEventListener('click', async () => {
        if (!state.selectedFile) return;
        await saveFile(editor.value, true); // true = manual save
    });

    // Initially disable save button
    saveButton.disabled = true;
}

// Auto-save function
async function autoSaveFile(content) {
    if (!state.selectedFile) return;

    try {
        const response = await fetch('/api/save', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                path: state.selectedFile,
                content: content
            })
        });

        if (!response.ok) {
            throw new Error(`Failed to auto-save: ${response.statusText}`);
        }

        state.currentFileContent = content;
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
    if (!state.selectedFile) return;

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
                path: state.selectedFile,
                content: content
            })
        });

        if (!response.ok) {
            throw new Error(`Failed to save: ${response.statusText}`);
        }

        state.currentFileContent = content;

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

    if (editor.value === state.currentFileContent) {
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