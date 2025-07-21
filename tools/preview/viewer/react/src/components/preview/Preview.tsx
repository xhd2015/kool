import { useState, useEffect } from 'react';
import { isEditableFile } from '../../utils/fileUtils';
import EditablePreview from './EditablePreview';
import UMLPreview from './UMLPreview';
import MermaidPreview from './MermaidPreview';
import MarkdownPreview from './MarkdownPreview';
import './Preview.css';

interface PreviewProps {
    selectedFile: string | null;
}

interface PreviewData {
    type: string;
    content: string;
}

const Preview = ({ selectedFile }: PreviewProps) => {
    const [previewData, setPreviewData] = useState<PreviewData | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!selectedFile) {
            setPreviewData(null);
            setError(null);
            return;
        }

        // If file is editable, don't load preview data here - EditablePreview will handle it
        if (isEditableFile(selectedFile)) {
            return;
        }

        const loadPreview = async () => {
            try {
                setLoading(true);
                setError(null);

                const response = await fetch(`/api/preview?path=${encodeURIComponent(selectedFile)}`);
                if (!response.ok) {
                    throw new Error(`Failed to load preview: ${response.statusText}`);
                }

                const data = await response.json();
                setPreviewData(data);
            } catch (err) {
                console.error('Failed to load preview:', err);
                setError(err instanceof Error ? err.message : 'Failed to load preview');
            } finally {
                setLoading(false);
            }
        };

        loadPreview();
    }, [selectedFile]);

    if (!selectedFile) {
        return <div className="empty-state">Select a file from the explorer to preview its contents</div>;
    }

    // Check if file is editable - show EditablePreview for editable files
    if (isEditableFile(selectedFile)) {
        // Determine file type for EditablePreview
        const ext = selectedFile.toLowerCase().substring(selectedFile.lastIndexOf('.'));
        let fileType = 'text';

        if (ext === '.md') {
            fileType = 'markdown';
        } else if (ext === '.uml' || ext === '.puml') {
            fileType = 'uml';
        } else if (ext === '.mmd') {
            fileType = 'mermaid';
        }

        return <EditablePreview selectedFile={selectedFile} fileType={fileType} />;
    }

    // For non-editable files, show regular preview
    if (loading) {
        return (
            <div className="loading">
                <div className="loading-spinner"></div>
                <div className="loading-text">Loading preview...</div>
            </div>
        );
    }

    if (error) {
        return <div className="error">Failed to load preview: {error}</div>;
    }

    if (!previewData) {
        return <div className="error">No preview data received</div>;
    }

    // Render based on content type with proper diagram components (for non-editable files)
    switch (previewData.type) {
        case 'uml':
            return <UMLPreview content={previewData.content} />;

        case 'mermaid':
            return <MermaidPreview content={previewData.content} />;

        case 'markdown':
            return <MarkdownPreview content={previewData.content} />;

        case 'text':
        default:
            return (
                <div className="preview-text">
                    <textarea
                        className="preview-text"
                        value={previewData.content}
                        readOnly
                    />
                </div>
            );
    }
};

export default Preview; 