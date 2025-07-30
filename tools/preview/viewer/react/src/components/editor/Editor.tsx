import { useState, useEffect, useCallback } from 'react';
import DiffModal from './DiffModal';
import './Editor.css';

interface EditorProps {
    filePath: string;
    content: string;
    onChange: (content: string) => void;
    onSave?: (content: string) => void;
    fileModifiedExternally?: boolean;
    onReload?: () => void;
}

const Editor = ({ filePath, content, onChange, onSave, fileModifiedExternally, onReload }: EditorProps) => {
    const [currentContent, setCurrentContent] = useState(content);
    const [saveStatus, setSaveStatus] = useState<string>('');
    const [isModified, setIsModified] = useState(false);
    const [isModificationSaved, setIsModificationSaved] = useState(true);
    const [originalContent, setOriginalContent] = useState(content);
    const [showDiffModal, setShowDiffModal] = useState(false);
    const [isReloading, setIsReloading] = useState(false);
    const [isManualSaving, setIsManualSaving] = useState(false);
    const [conflictData, setConflictData] = useState<{
        currentContent: string;
        userDiff: string;
        currentDiff: string;
        message: string;
    } | null>(null);

    // Update content when file changes
    useEffect(() => {
        setCurrentContent(content);
        setOriginalContent(content);
        setIsModified(false);
        setIsModificationSaved(true);
        setShowDiffModal(false);
        setConflictData(null);
    }, [content]);

    const handleReload = useCallback(async (isManual = false) => {
        if (!onReload) return;

        // Only show visual feedback for manual reloads
        if (isManual) {
            setIsReloading(true);
        }

        const startTime = Date.now();

        try {
            await onReload();
        } catch (error) {
            console.error('Reload failed:', error);
        }

        // For manual reloads, ensure spinner shows for at least 200ms
        if (isManual) {
            const elapsed = Date.now() - startTime;
            const remainingTime = Math.max(0, 200 - elapsed);

            setTimeout(() => {
                setIsReloading(false);
            }, remainingTime);
        }
    }, [onReload]);

    // Auto-reload when file is modified externally and no unsaved changes
    useEffect(() => {
        if (fileModifiedExternally && isModificationSaved && onReload) {
            handleReload(false); // Auto-reload - no visual feedback
        }
    }, [fileModifiedExternally, isModificationSaved, onReload, handleReload]);

    const saveContent = useCallback(async (contentToSave: string) => {
        const response = await fetch('/api/save', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                path: filePath,
                content: contentToSave,
                oldContent: originalContent
            })
        });

        if (response.status === 409) {
            // Conflict detected
            const conflictResponse = await response.json();
            setConflictData({
                currentContent: conflictResponse.currentContent,
                userDiff: conflictResponse.userDiff,
                currentDiff: conflictResponse.currentDiff,
                message: conflictResponse.message
            });
            setShowDiffModal(true);
            throw new Error(conflictResponse.message);
        }

        if (!response.ok) {
            throw new Error(`Failed to save: ${response.statusText}`);
        }

        // Update original content after successful save
        setOriginalContent(contentToSave);
        setIsModificationSaved(true);
    }, [filePath, originalContent]);

    // Auto-save with debouncing
    useEffect(() => {
        if (!isModified || isModificationSaved) return;

        const autoSaveTimer = setTimeout(async () => {
            try {
                // Don't show "Saving..." status to avoid distraction
                await saveContent(currentContent);
                // Don't show "Auto-saved" status to avoid distraction
            } catch (error) {
                console.error('Auto-save failed:', error);
                setSaveStatus('Auto-save failed');
                setTimeout(() => setSaveStatus(''), 3000);
            }
        }, 500);

        return () => clearTimeout(autoSaveTimer);
    }, [currentContent, isModified, isModificationSaved, saveContent]);

    const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const newContent = e.target.value;
        setCurrentContent(newContent);
        const modified = newContent !== content;
        setIsModified(modified);

        // If content was modified, mark it as not saved yet
        // If content is back to original, mark it as saved
        if (modified) {
            setIsModificationSaved(false);
        } else {
            setIsModificationSaved(true);
        }

        onChange(newContent);
    };

    const handleManualSave = async () => {
        setIsManualSaving(true);

        try {
            setSaveStatus('Saving...');
            await saveContent(currentContent);
            if (onSave) {
                onSave(currentContent);
            }
            // Note: setIsModificationSaved(true) is already called in saveContent
            setSaveStatus('Saved');
            setTimeout(() => setSaveStatus(''), 1500);
        } catch (error) {
            console.error('Save failed:', error);
            setSaveStatus('Save failed');
            setTimeout(() => setSaveStatus(''), 3000);
        } finally {
            setIsManualSaving(false);
        }
    };

    const handleReloadFile = async () => {
        try {
            const response = await fetch(`/api/content?path=${encodeURIComponent(filePath)}`);
            if (!response.ok) {
                throw new Error(`Failed to reload file: ${response.statusText}`);
            }
            const data = await response.json();
            setCurrentContent(data.content);
            setOriginalContent(data.content);
            setIsModified(false);
            setShowDiffModal(false);
            setConflictData(null);
            onChange(data.content);
        } catch (error) {
            console.error('Failed to reload file:', error);
        }
    };

    const handleCloseDiffModal = () => {
        setShowDiffModal(false);
        setConflictData(null);
    };

    return (
        <>
            <div className="editor-section">
                <div className="editor-header">
                    <span>Editor</span>
                    <div className="save-controls">
                        <span className={`save-status ${saveStatus ? 'visible' : ''} ${saveStatus.includes('failed') ? 'error' : ''} ${saveStatus.includes('saved') || saveStatus.includes('Auto-saved') ? 'success' : ''}`}>
                            {saveStatus}
                        </span>
                        <button
                            className={`reload-button ${isReloading ? 'reloading' : ''}`}
                            onClick={() => {
                                if (isReloading) return;

                                if (!isModificationSaved) {
                                    const confirmReload = window.confirm(
                                        'The file has been modified externally. You have unsaved changes. Do you want to reload and lose your changes?'
                                    );
                                    if (confirmReload) {
                                        handleReload(true); // Manual reload - show visual feedback
                                    }
                                } else {
                                    handleReload(true); // Manual reload - show visual feedback
                                }
                            }}
                            title={isReloading ? "Reloading..." : fileModifiedExternally ? "File has been modified externally" : "Reload file"}
                        >
                            {isReloading ? '⟳' : '↻'}
                        </button>
                        <button
                            className={`save-button ${isManualSaving ? 'save-button-loading' : ''}`}
                            onClick={handleManualSave}
                            disabled={isManualSaving}
                        >
                            {isManualSaving ? 'Saving...' : 'Save'}
                        </button>
                    </div>
                </div>
                <textarea
                    className="editor-textarea"
                    value={currentContent}
                    onChange={handleContentChange}
                    placeholder="Start editing..."
                    style={{
                        flex: 1,
                        border: 'none',
                        outline: 'none',
                        resize: 'none',
                        fontFamily: 'Consolas, Monaco, "Courier New", monospace',
                        fontSize: '14px',
                        padding: '16px',
                        backgroundColor: 'var(--editor-bg, #ffffff)',
                        color: 'var(--editor-text, #333333)'
                    }}
                />
            </div>

            <DiffModal
                isOpen={showDiffModal}
                onClose={handleCloseDiffModal}
                onReload={handleReloadFile}
                userDiff={conflictData?.userDiff || ''}
                currentDiff={conflictData?.currentDiff || ''}
                filePath={filePath}
            />
        </>
    );
};

export default Editor; 