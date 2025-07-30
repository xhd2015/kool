import { useState, useEffect, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import Layout from './components/layout/Layout';
import Preview from './components/preview/Preview';
import MultiTabTerminal, { type MultiTabTerminalHandle } from './components/terminal/MultiTabTerminal';
import { type FileTreeHandle } from './components/tree/FileTree';
import { useResize } from './hooks/useResize';
import { useFileWatcher } from './hooks/useFileWatcher';
import './styles/globals.css';

function App() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [terminalVisible, setTerminalVisible] = useState<boolean>(false);
  const [fileNeedsReload, setFileNeedsReload] = useState<string | null>(null);

  // Refs for vertical resizing
  const appContainerRef = useRef<HTMLDivElement>(null);
  const verticalResizerRef = useRef<HTMLDivElement>(null);
  const terminalRef = useRef<MultiTabTerminalHandle>(null);
  const fileTreeRef = useRef<FileTreeHandle>(null);

  // Vertical resize between content and terminal
  const { size: contentSize, handleMouseDown: handleVerticalMouseDown } = useResize({
    containerRef: appContainerRef,
    direction: 'vertical',
    minSize: 30,
    maxSize: 80,
    defaultSize: 70,
    enabled: terminalVisible
  });

  // Reset content size when terminal visibility changes
  useEffect(() => {
    // When terminal visibility changes, we need to force a layout recalculation
    // This ensures the content area properly expands/contracts
    if (appContainerRef.current) {
      // Trigger a resize event to force layout recalculation
      const resizeEvent = new Event('resize');
      window.dispatchEvent(resizeEvent);
    }
  }, [terminalVisible]);

  // Initialize selectedFile from URL parameter on mount
  useEffect(() => {
    const fileParam = searchParams.get('file');
    if (fileParam) {
      setSelectedFile(fileParam);
    }
  }, [searchParams]);

  // Handle file selection and update URL parameter
  const handleFileSelect = (filePath: string | null) => {
    setSelectedFile(filePath);
    // Clear reload flag when switching files
    setFileNeedsReload(null);

    // Update URL parameter
    const newParams = new URLSearchParams(searchParams);
    if (filePath) {
      newParams.set('file', filePath);
    } else {
      newParams.delete('file');
    }
    setSearchParams(newParams, { replace: true });
  };

  // Handle terminal command execution
  const handleExecuteTerminalCommand = (command: string) => {
    if (terminalRef.current) {
      terminalRef.current.executeCommand(command);
    }
  };

  // File watcher integration
  useFileWatcher({
    onTreeRefresh: () => {
      // Refresh the file tree when files are added/deleted
      if (fileTreeRef.current) {
        fileTreeRef.current.refresh();
      }
    },
    onFileModified: (filePath) => {
      console.log('File modified:', filePath);
      // Show reload button for the modified file
      setFileNeedsReload(filePath);
    }
  });

  return (
    <div className="app" ref={appContainerRef}>
      <div style={{
        height: terminalVisible ? `${contentSize}%` : `calc(100% - 50px)`,
        minHeight: '300px',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden'
      }}>
        <Layout
          ref={fileTreeRef}
          selectedFile={selectedFile}
          onFileSelect={handleFileSelect}
          onExecuteTerminalCommand={handleExecuteTerminalCommand}
        >
          <div className="preview-section">
            <div className="preview-container-wrapper">
              <div className="preview-container">
                <Preview 
                  selectedFile={selectedFile} 
                  fileNeedsReload={fileNeedsReload}
                  onReloadComplete={() => setFileNeedsReload(null)}
                />
              </div>
            </div>
          </div>
        </Layout>
      </div>

      <div
        className="vertical-resizer"
        ref={verticalResizerRef}
        onMouseDown={handleVerticalMouseDown}
        style={{
          display: terminalVisible ? 'block' : 'none'
        }}
      ></div>

      <div style={{
        height: terminalVisible ? `${100 - contentSize}%` : '50px',
        minHeight: terminalVisible ? '200px' : '50px',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden'
      }}>
        <MultiTabTerminal
          ref={terminalRef}
          isVisible={terminalVisible}
          onToggle={() => setTerminalVisible(!terminalVisible)}
        />
      </div>
    </div>
  );
}

export default App;
