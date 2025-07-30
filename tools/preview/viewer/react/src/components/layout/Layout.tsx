import { forwardRef } from 'react';
import Header from './Header';
import Sidebar from './Sidebar';
import { type FileTreeHandle } from '../tree/FileTree';
import './Layout.css';

interface LayoutProps {
    children?: React.ReactNode;
    selectedFile: string | null;
    onFileSelect: (filePath: string | null) => void;
    onExecuteTerminalCommand?: (command: string) => void;
}

const Layout = forwardRef<FileTreeHandle, LayoutProps>(({ children, selectedFile, onFileSelect, onExecuteTerminalCommand }, ref) => {
    return (
        <div className="app">
            <Header onExecuteTerminalCommand={onExecuteTerminalCommand} />
            <div className="container">
                <Sidebar
                    ref={ref}
                    selectedFile={selectedFile}
                    onFileSelect={onFileSelect}
                />
                <div className="content">
                    <div className="content-header">
                        <span id="content-title">
                            {selectedFile ? selectedFile : 'Select a file to preview'}
                        </span>
                    </div>
                    <div className="content-body">
                        {children}
                    </div>
                </div>
            </div>
        </div>
    );
});

Layout.displayName = 'Layout';

export default Layout; 