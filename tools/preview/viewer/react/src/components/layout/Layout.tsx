import Header from './Header';
import Sidebar from './Sidebar';
import './Layout.css';

interface LayoutProps {
    children?: React.ReactNode;
    selectedFile: string | null;
    onFileSelect: (filePath: string | null) => void;
    onExecuteTerminalCommand?: (command: string) => void;
}

const Layout = ({ children, selectedFile, onFileSelect, onExecuteTerminalCommand }: LayoutProps) => {
    return (
        <div className="app">
            <Header onExecuteTerminalCommand={onExecuteTerminalCommand} />
            <div className="container">
                <Sidebar
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
};

export default Layout; 