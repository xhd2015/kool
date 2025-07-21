import FileTree from '../tree/FileTree';
import './Sidebar.css';

interface SidebarProps {
    selectedFile: string | null;
    onFileSelect: (filePath: string | null) => void;
}

const Sidebar = ({ selectedFile, onFileSelect }: SidebarProps) => {
    return (
        <div className="sidebar">
            <div className="sidebar-header">Explorer</div>
            <FileTree
                selectedFile={selectedFile}
                onFileSelect={onFileSelect}
            />
        </div>
    );
};

export default Sidebar; 