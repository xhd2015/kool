import { useState, useEffect, useRef } from 'react';
import Editor from '../editor/Editor';
import UMLPreview from './UMLPreview';
import MermaidPreview from './MermaidPreview';
import MarkdownPreview from './MarkdownPreview';
import { useResize } from '../../hooks/useResize';

interface EditablePreviewProps {
    selectedFile: string;
    fileType: string;
}

const EditablePreview = ({ selectedFile, fileType }: EditablePreviewProps) => {
    const [originalContent, setOriginalContent] = useState<string>('');
    const [currentContent, setCurrentContent] = useState<string>('');
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // Refs for resizing
    const containerRef = useRef<HTMLDivElement>(null);

    // Horizontal resize between editor and preview
    const { size: editorWidth, handleMouseDown } = useResize({
        containerRef,
        direction: 'horizontal',
        minSize: 25,
        maxSize: 75,
        defaultSize: 50
    });

    // Load file content
    useEffect(() => {
        const loadFileContent = async () => {
            try {
                setLoading(true);
                setError(null);

                const response = await fetch(`/api/content?path=${encodeURIComponent(selectedFile)}`);
                if (!response.ok) {
                    throw new Error(`Failed to load file content: ${response.statusText}`);
                }

                const data = await response.json();
                setOriginalContent(data.content);
                setCurrentContent(data.content);
            } catch (err) {
                console.error('Failed to load file content:', err);
                setError(err instanceof Error ? err.message : 'Failed to load file content');
            } finally {
                setLoading(false);
            }
        };

        loadFileContent();
    }, [selectedFile]);

    const handleContentChange = (newContent: string) => {
        setCurrentContent(newContent);
    };

    const renderPreview = () => {
        if (!currentContent) {
            return <div className="empty-state">No content to preview</div>;
        }

        switch (fileType) {
            case 'uml':
                return <UMLPreview content={currentContent} />;
            case 'mermaid':
                return <MermaidPreview content={currentContent} />;
            case 'markdown':
                return <MarkdownPreview content={currentContent} />;
            default:
                return (
                    <div className="preview-text">
                        <textarea
                            className="preview-text"
                            value={currentContent}
                            readOnly
                        />
                    </div>
                );
        }
    };

    if (loading) {
        return (
            <div className="loading">
                <div className="loading-spinner"></div>
                <div className="loading-text">Loading file content...</div>
            </div>
        );
    }

    if (error) {
        return <div className="error">Failed to load file content: {error}</div>;
    }

    return (
        <div className="preview-section" ref={containerRef}>
            {/* Editor on the left */}
            <div style={{
                width: `${editorWidth}%`,
                minWidth: '300px',
                display: 'flex',
                flexDirection: 'column'
            }}>
                <Editor
                    filePath={selectedFile}
                    content={originalContent}
                    onChange={handleContentChange}
                />
            </div>

            {/* Horizontal resizer */}
            <div
                className="horizontal-resizer"
                onMouseDown={handleMouseDown}
            ></div>

            {/* Preview on the right */}
            <div style={{
                width: `${100 - editorWidth}%`,
                minWidth: '300px',
                display: 'flex',
                flexDirection: 'column'
            }}>
                <div className="preview-container-wrapper">
                    <div className="preview-container">
                        {renderPreview()}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default EditablePreview; 