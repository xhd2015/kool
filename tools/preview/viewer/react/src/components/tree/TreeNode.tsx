import { useState } from 'react';
import './Tree.css';

interface FileNode {
    name: string;
    path: string;
    isDir: boolean;
    children?: FileNode[];
}

interface TreeNodeProps {
    node: FileNode;
    selectedFile: string | null;
    onFileSelect: (filePath: string | null) => void;
}

const TreeNode = ({ node, selectedFile, onFileSelect }: TreeNodeProps) => {
    const [isExpanded, setIsExpanded] = useState(true);

    // Use shared state for selection instead of local state
    const isSelected = selectedFile === node.path;

    const handleToggle = () => {
        if (node.isDir && node.children) {
            setIsExpanded(!isExpanded);
        }
    };

    const handleSelect = () => {
        // Only select files, not directories
        if (!node.isDir) {
            onFileSelect(node.path);
            console.log('Selected file:', node.path);
        }
    };

    const handleClick = () => {
        if (node.isDir) {
            handleToggle();
        } else {
            handleSelect();
        }
    };

    if (node.name == "overall_view.uml") {
        console.log("node: ", node);
    }

    return (
        <div>
            <div
                className={`tree-node ${node.isDir ? 'directory' : 'file'} ${isSelected ? 'selected' : ''}`}
                onClick={handleClick}
                data-path={node.path}
            >
                <span className="toggle">
                    {node.isDir && node.children && node.children.length > 0
                        ? (isExpanded ? '▼' : '▷')
                        : ''}
                </span>
                <span className="icon">
                    {node.isDir ? '📁' : '📄'}
                </span>
                <span className="name">{node.name}</span>
            </div>

            {node.isDir && node.children && isExpanded && (
                <div className="tree-children">
                    {node.children.map((child, index) => (
                        <TreeNode
                            key={`${child.path}-${index}`}
                            node={child}
                            selectedFile={selectedFile}
                            onFileSelect={onFileSelect}
                        />
                    ))}
                </div>
            )}
        </div>
    );
};

export default TreeNode; 