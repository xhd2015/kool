import { useState, useEffect, useCallback } from 'react';
import DiffModal from './DiffModal';
import './Editor.css';

interface EditorProps {
    filePath: string;
    content: string;
    onChange: (content: string) => void;
    onSave?: (content: string) => void;
}

const Editor = ({ filePath, content, onChange, onSave }: EditorProps) => {
    const [currentContent, setCurrentContent] = useState(content);
    const [saveStatus, setSaveStatus] = useState<string>('');
    const [isModified, setIsModified] = useState(false);
    const [originalContent, setOriginalContent] = useState(content);
    const [showDiffModal, setShowDiffModal] = useState(false);
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
        setShowDiffModal(false);
        setConflictData(null);
    }, [content]);

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
    }, [filePath, originalContent]);

    // Auto-save with debouncing
    useEffect(() => {
        if (!isModified) return;

        const autoSaveTimer = setTimeout(async () => {
            try {
                setSaveStatus('Saving...');
                await saveContent(currentContent);
                setSaveStatus('Auto-saved');
                setTimeout(() => setSaveStatus(''), 1500);
            } catch (error) {
                console.error('Auto-save failed:', error);
                setSaveStatus('Auto-save failed');
                setTimeout(() => setSaveStatus(''), 2500);
            }
        }, 500);

        return () => clearTimeout(autoSaveTimer);
    }, [currentContent, isModified, saveContent]);

    const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const newContent = e.target.value;
        setCurrentContent(newContent);
        setIsModified(newContent !== content);
        onChange(newContent);
    };

    const handleManualSave = async () => {
        try {
            setSaveStatus('Saving...');
            await saveContent(currentContent);
            if (onSave) {
                onSave(currentContent);
            }
            setIsModified(false);
            setSaveStatus('Saved');
            setTimeout(() => setSaveStatus(''), 2000);
        } catch (error) {
            console.error('Save failed:', error);
            setSaveStatus('Save failed');
            setTimeout(() => setSaveStatus(''), 3000);
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
                            className="save-button"
                            onClick={handleManualSave}
                            disabled={!isModified}
                        >
                            Save
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