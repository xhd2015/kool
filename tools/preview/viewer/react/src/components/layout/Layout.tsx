import Header from './Header';
import Sidebar from './Sidebar';
import './Layout.css';

interface LayoutProps {
    children?: React.ReactNode;
    selectedFile: string | null;
    onFileSelect: (filePath: string | null) => void;
}

const Layout = ({ children, selectedFile, onFileSelect }: LayoutProps) => {
    return (
        <div className="app">
            <Header />
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