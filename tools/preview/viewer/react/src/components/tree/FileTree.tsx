import { useState, useEffect, useImperativeHandle, forwardRef } from 'react';
import TreeNode from './TreeNode';
import './Tree.css';

interface FileNode {
    name: string;
    path: string;
    isDir: boolean;
    children?: FileNode[];
}

interface FileTreeProps {
    selectedFile: string | null;
    onFileSelect: (filePath: string | null) => void;
}

export interface FileTreeHandle {
    refresh: () => Promise<void>;
}

const FileTree = forwardRef<FileTreeHandle, FileTreeProps>(({ selectedFile, onFileSelect }, ref) => {
    const [fileTree, setFileTree] = useState<FileNode | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const loadFileTree = async () => {
        try {
            setLoading(true);
            setError(null);

            const response = await fetch('/api/tree');
            if (!response.ok) {
                throw new Error(`Failed to load tree: ${response.statusText}`);
            }

            const tree = await response.json();
            setFileTree(tree);
        } catch (err) {
            console.error('Failed to load directory tree:', err);
            setError(err instanceof Error ? err.message : 'Failed to load directory tree');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadFileTree();
    }, []);

    // Expose refresh function to parent
    useImperativeHandle(ref, () => ({
        refresh: loadFileTree
    }));

    if (loading) {
        return <div className="loading">Loading directory tree...</div>;
    }

    if (error) {
        return <div className="error">Failed to load directory tree: {error}</div>;
    }

    if (!fileTree) {
        return <div className="error">No directory tree data received</div>;
    }

    return (
        <div className="file-tree">
            <TreeNode
                node={fileTree}
                selectedFile={selectedFile}
                onFileSelect={onFileSelect}
            />
        </div>
    );
});

FileTree.displayName = 'FileTree';

export default FileTree; 