import { forwardRef } from 'react';
import FileTree, { type FileTreeHandle } from '../tree/FileTree';
import './Sidebar.css';

interface SidebarProps {
    selectedFile: string | null;
    onFileSelect: (filePath: string | null) => void;
}

const Sidebar = forwardRef<FileTreeHandle, SidebarProps>(({ selectedFile, onFileSelect }, ref) => {
    return (
        <div className="sidebar">
            <div className="sidebar-header">Explorer</div>
            <FileTree
                ref={ref}
                selectedFile={selectedFile}
                onFileSelect={onFileSelect}
            />
        </div>
    );
});

Sidebar.displayName = 'Sidebar';

export default Sidebar; 