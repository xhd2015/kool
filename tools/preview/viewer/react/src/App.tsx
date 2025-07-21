import { useState, useEffect, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import Layout from './components/layout/Layout';
import Preview from './components/preview/Preview';
import MultiTabTerminal from './components/terminal/MultiTabTerminal';
import { useResize } from './hooks/useResize';
import './styles/globals.css';

function App() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [terminalVisible, setTerminalVisible] = useState<boolean>(false);

  // Refs for vertical resizing
  const appContainerRef = useRef<HTMLDivElement>(null);
  const verticalResizerRef = useRef<HTMLDivElement>(null);

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
    // This effect will run when terminalVisible changes
    // The resize hook will maintain its state, but we adjust the layout accordingly
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

    // Update URL parameter
    const newParams = new URLSearchParams(searchParams);
    if (filePath) {
      newParams.set('file', filePath);
    } else {
      newParams.delete('file');
    }
    setSearchParams(newParams, { replace: true });
  };

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
          selectedFile={selectedFile}
          onFileSelect={handleFileSelect}
        >
          <div className="preview-section">
            <div className="preview-container-wrapper">
              <div className="preview-container">
                <Preview selectedFile={selectedFile} />
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
          isVisible={terminalVisible}
          onToggle={() => setTerminalVisible(!terminalVisible)}
        />
      </div>
    </div>
  );
}

export default App;
