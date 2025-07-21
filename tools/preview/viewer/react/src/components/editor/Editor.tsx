import { useState, useEffect, useCallback } from 'react';
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

    // Update content when file changes
    useEffect(() => {
        setCurrentContent(content);
        setIsModified(false);
    }, [content]);

    const saveContent = useCallback(async (contentToSave: string) => {
        const response = await fetch('/api/save', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                path: filePath,
                content: contentToSave
            })
        });

        if (!response.ok) {
            throw new Error(`Failed to save: ${response.statusText}`);
        }
    }, [filePath]);

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

    return (
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
    );
};

export default Editor; 